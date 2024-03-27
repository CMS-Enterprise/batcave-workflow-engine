package pipelines

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"workflow-engine/pkg/shell"
)

type CodeScan struct {
	Stdout                        io.Writer
	Stderr                        io.Writer
	DryRunEnabled                 bool
	SemgrepExperimental           bool
	SemgrepErrorOnFindingsEnabled bool
	SemgrepRules                  string
	config                        *Config
	runtime                       struct {
		gitleaksFile      *os.File
		semgrepFile       *os.File
		bundleFilename    string
		gitleaksFilename  string
		semgrepFilename   string
		postSummaryBuffer *bytes.Buffer
	}
}

func (p *CodeScan) WithConfig(config *Config) *CodeScan {
	p.config = config
	return p
}

func NewCodeScan(stdout io.Writer, stderr io.Writer) *CodeScan {
	return &CodeScan{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
		config:        new(Config),
	}
}

// preRun is responsible for opening file needed during the run
func (p *CodeScan) preRun() error {
	var err error

	if err := MakeDirectoryP(p.config.ArtifactDir); err != nil {
		slog.Error("failed to create artifact directory", "name", p.config.ArtifactDir)
		return errors.New("Code Scan Pipeline failed to run.")
	}

	p.runtime.gitleaksFilename = path.Join(p.config.ArtifactDir, p.config.CodeScan.GitleaksFilename)
	p.runtime.gitleaksFile, err = OpenOrCreateFile(p.runtime.gitleaksFilename)
	if err != nil {
		slog.Error("cannot open gitleaks report file", "filename", p.runtime.gitleaksFile, "error", err)
		return err
	}

	p.runtime.semgrepFilename = path.Join(p.config.ArtifactDir, p.config.CodeScan.SemgrepFilename)
	p.runtime.semgrepFile, err = OpenOrCreateFile(p.runtime.semgrepFilename)
	if err != nil {
		slog.Error("cannot open semgrep report file", "filename", p.runtime.semgrepFilename, "error", err)
		return err
	}

	if err := InitGatecheckBundle(p.config, p.Stderr, p.DryRunEnabled); err != nil {
		slog.Error("cannot initialize gatecheck bundle", "error", err)
		return err
	}

	p.runtime.postSummaryBuffer = new(bytes.Buffer)
	p.runtime.bundleFilename = path.Join(p.config.ArtifactDir, p.config.GatecheckBundleFilename)

	return nil
}

func (p *CodeScan) Run() error {
	if !p.config.CodeScan.Enabled {
		slog.Warn("code scan pipeline is disabled, skip.")
		return nil
	}

	if err := p.preRun(); err != nil {
		return fmt.Errorf("Code Scan Pipeline Pre-Run Failed: %v", err)
	}

	fmt.Fprintln(p.Stdout, "******* Workflow Engine Code Scan Pipeline [Run] *******")

	// Create async tasks
	semgrepTask := NewAsyncTask("semgrep")
	gitleaksTask := NewAsyncTask("gitleaks")

	listTask := NewAsyncTask("gatecheck list")
	bundleTask := NewAsyncTask("gatecheck bundle")
	postRunTask := NewAsyncTask("post-run")

	// Run all jobs in the background

	go p.semgrepJob(semgrepTask)
	go p.gitleaksJob(gitleaksTask)

	go p.gatecheckListJob(listTask, semgrepTask, gitleaksTask)
	go p.gatecheckBundleJob(bundleTask, semgrepTask, gitleaksTask)

	go p.postRunJob(postRunTask, semgrepTask, gitleaksTask, listTask, bundleTask)

	// Stream Task results to stderr and block until the task completes
	// The order here is the stderr stream order, not the execution order.
	// execution order is sudo random because of the goroutines however,
	// dependencies are defined and handle by the jobs.

	allTasks := []*AsyncTask{
		semgrepTask,
		gitleaksTask,
		listTask,
		bundleTask,
		postRunTask,
	}

	allErrors := make([]error, 0)

	for _, task := range allTasks {
		err := task.StreamTo(p.Stderr)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("%s task failed. reason: %w", task.Name, err))
		}
	}

	// Dump the summary to stdout to prevent collisions during async runs
	_, _ = io.Copy(p.Stdout, p.runtime.postSummaryBuffer)

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}

	return nil
}

func (p *CodeScan) semgrepJob(task *AsyncTask) {
	defer task.Close()
	defer p.runtime.semgrepFile.Close()

	task.Logger.Debug("run semgrep sast scan")

	reportWriter := io.MultiWriter(p.runtime.semgrepFile, task.StdoutBuf)
	// using the syftscan pipe reader will block this job until syft is done
	task.ExitError = shell.SemgrepScan(
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithLogger(task.Logger),
		shell.WithStdout(reportWriter), // where the report goes
		shell.WithStderr(task.StderrPipeWriter),

		shell.WithSemgrep(p.config.CodeScan.SemgrepRules, p.SemgrepExperimental),
	)

	var commandError *shell.ErrCommand

	switch {
	case task.ExitError == nil:
		return
	case errors.As(task.ExitError, &commandError):
		// check for exit code 1 which means a "blocking" finding
		slog.Error("semgrep has findings", "exit status", commandError.Error())
		task.ExitError = nil
		return
	default:
		task.ExitError = fmt.Errorf("cannot inspect shell command error: %w", task.ExitError)
	}
}

func (p *CodeScan) gitleaksJob(task *AsyncTask) {
	defer task.Close()
	defer p.runtime.gitleaksFile.Close()

	task.Logger.Debug("gitleaks secret detection")

	// Gitleaks doesn't put it's report to STDOUT, we have to open
	// the file afterwards and dump to the tasks stdout
	task.ExitError = shell.GitLeaksDetect(
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithLogger(task.Logger),
		shell.WithStderr(task.StderrPipeWriter),
		shell.WithGitleaks(p.config.CodeScan.GitleaksSrcDir, p.runtime.gitleaksFilename),
	)

	if task.ExitError != nil {
		os.Remove(p.runtime.gitleaksFilename)
		return
	}
}

func (p *CodeScan) gatecheckListJob(task *AsyncTask, semgrepTask *AsyncTask, gitleaksTask *AsyncTask) {
	defer task.Close()

	opts := []shell.OptionFunc{
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithLogger(task.Logger),
		shell.WithErrorOnly(task.StderrPipeWriter),
		shell.WithStdout(p.runtime.postSummaryBuffer), // where the summary goes
	}
	semgrepOpts := append(
		opts,
		shell.WithWaitFunc(semgrepTask.Wait),
		shell.WithStdin(semgrepTask.StdoutBuf),
		shell.WithReportType("semgrep"),
	)
	gitleaksOpts := append(
		opts,
		shell.WithWaitFunc(gitleaksTask.Wait),
		shell.WithListTarget(p.runtime.gitleaksFilename),
		shell.WithReportType("gitleaks"),
	)

	task.Logger.Debug("list semgrep report after semgrep completes")
	semgrepListError := shell.GatecheckList(semgrepOpts...)

	fmt.Fprintln(task.StderrPipeWriter)

	task.Logger.Debug("list gitleaks report after gitleaks completes")
	gitleaksError := shell.GatecheckList(gitleaksOpts...)

	task.ExitError = errors.Join(semgrepListError, gitleaksError)
}

func (p *CodeScan) gatecheckBundleJob(task *AsyncTask, semgrep *AsyncTask, gitleaksTask *AsyncTask) {
	defer task.Close()

	opts := []shell.OptionFunc{
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithLogger(task.Logger),
		shell.WithStdout(p.runtime.postSummaryBuffer),
		shell.WithErrorOnly(task.StderrPipeWriter),
	}

	semgrepOpts := append(opts, shell.WithBundleFile(p.runtime.bundleFilename, p.runtime.semgrepFilename), shell.WithWaitFunc(semgrep.Wait))
	err := shell.GatecheckBundleAdd(semgrepOpts...)
	task.ExitError = errors.Join(task.ExitError, err)

	gitleaksOpts := append(opts, shell.WithBundleFile(p.runtime.bundleFilename, p.runtime.gitleaksFilename), shell.WithWaitFunc(gitleaksTask.Wait))
	err = shell.GatecheckBundleAdd(gitleaksOpts...)
	task.ExitError = errors.Join(task.ExitError, err)
}

func (p *CodeScan) postRunJob(task *AsyncTask, allTasks ...*AsyncTask) {
	defer task.Close()

	for _, task := range allTasks {
		_ = task.Wait()
	}

	opts := []shell.OptionFunc{
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithLogger(task.Logger),
		shell.WithErrorOnly(task.StderrPipeWriter),
		shell.WithStdout(p.runtime.postSummaryBuffer),
		shell.WithListTarget(p.runtime.bundleFilename),
	}

	// Add the bundle summary output to the summary buffer before dumping
	err := shell.GatecheckList(opts...)
	task.ExitError = errors.Join(task.ExitError, err)

	// print clamAV Report and grype summary
}
