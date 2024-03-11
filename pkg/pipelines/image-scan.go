package pipelines

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"
	"workflow-engine/pkg/shell"
)

// ImageScan Pipeline
//
// Job status can be used in post to determine what clean up or bundling should be done
// TODO: docker save is really slow, only use it if the scan target is remote
type ImageScan struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
	config        *Config
	DockerAlias   string
	runtime       struct {
		syftFile          *os.File
		grypeFile         *os.File
		clamavFile        *os.File
		tarImageFile      *os.File
		syftFilename      string
		grypeFilename     string
		clamavFilename    string
		bundleFilename    string
		tarImageFilename  string
		postSummaryBuffer *bytes.Buffer
		dockerAlias       shell.DockerAlias
		ctx               context.Context
		cancelFunc        func()
	}
}

func (p *ImageScan) WithConfig(config *Config) *ImageScan {
	p.config = config
	return p
}

func NewImageScan(stdout io.Writer, stderr io.Writer) *ImageScan {
	return &ImageScan{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
		config:        new(Config),
	}
}

func (p *ImageScan) preRun() error {
	fmt.Fprintln(p.Stdout, "******* Workflow Engine Image Scan Pipeline [Pre-Run] *******")
	var err error

	// In Memory Buffer
	p.runtime.postSummaryBuffer = new(bytes.Buffer)

	// Early cancel context for certain tasks
	p.runtime.ctx, p.runtime.cancelFunc = context.WithCancel(context.Background())

	if err := MakeDirectoryP(p.config.ArtifactDir); err != nil {
		slog.Error("failed to create artifact directory", "name", p.config.ArtifactDir)
		return errors.New("Code Scan Pipeline failed to run.")
	}

	p.runtime.syftFilename = path.Join(p.config.ArtifactDir, p.config.ImageScan.SyftFilename)
	p.runtime.syftFile, err = OpenOrCreateFile(p.runtime.syftFilename)
	if err != nil {
		slog.Error("cannot open syft sbom file", "filename", p.runtime.syftFilename, "error", err)
		return err
	}

	p.runtime.grypeFilename = path.Join(p.config.ArtifactDir, p.config.ImageScan.GrypeFullFilename)
	p.runtime.grypeFile, err = OpenOrCreateFile(p.runtime.grypeFilename)
	if err != nil {
		slog.Error("cannot open grype sbom file", "filename", p.runtime.grypeFilename, "error", err)
		return err
	}

	p.runtime.clamavFilename = path.Join(p.config.ArtifactDir, p.config.ImageScan.ClamavFilename)
	p.runtime.clamavFile, err = OpenOrCreateFile(p.runtime.clamavFilename)
	if err != nil {
		slog.Error("cannot open clam virus report file", "filename", p.runtime.clamavFilename, "error", err)
		return err
	}

	// Create temporary image tar file for writing
	slog.Debug("create temporary file for image tar, used for clam virus scan")
	p.runtime.tarImageFile, err = os.CreateTemp(os.TempDir(), "*-image.tar")
	if err != nil {
		slog.Error("cannot create temp image tar file", "temp_dir", os.TempDir())
		return err
	}
	p.runtime.tarImageFilename = p.runtime.tarImageFile.Name()

	// create gatecheck bundle file
	if err := InitGatecheckBundle(p.config, p.Stderr, p.DryRunEnabled); err != nil {
		slog.Error("cannot initialize gatecheck bundle", "error", err)
		return err
	}

	p.runtime.bundleFilename = path.Join(p.config.ArtifactDir, p.config.GatecheckBundleFilename)

	p.runtime.dockerAlias = shell.DockerAliasDocker
	// print the connection information, exit pipeline if failed
	switch strings.ToLower(p.DockerAlias) {
	case "podman":
		p.runtime.dockerAlias = shell.DockerAliasPodman
	case "docker":
		p.runtime.dockerAlias = shell.DockerAliasDocker
	}

	return nil
}

func (p *ImageScan) Run() error {
	if !p.config.ImageScan.Enabled {
		slog.Warn("image scan pipeline is disabled, skip.")
		return nil
	}

	if err := p.preRun(); err != nil {
		return fmt.Errorf("Image Scan Pipeline Pre-Run Failed: %v", err)
	}

	fmt.Fprintln(p.Stdout, "******* Workflow Engine Image Scan Pipeline [Run] *******")

	// Create async tasks
	freshclamTask := NewAsyncTask("freshclam")
	dockerSaveTask := NewAsyncTask("docker save")
	clamscanTask := NewAsyncTask("clamscan")
	syftTask := NewAsyncTask("syft")
	grypeTask := NewAsyncTask("grype")
	listTask := NewAsyncTask("gatecheck list")
	bundleTask := NewAsyncTask("gatecheck bundle")
	postRunTask := NewAsyncTask("cleanup")

	// Run all jobs in the background
	go p.freshclamJob(freshclamTask)
	go p.dockerSaveJob(dockerSaveTask)
	go p.clamscanJob(clamscanTask, dockerSaveTask, freshclamTask)

	go p.syftJob(syftTask, dockerSaveTask)
	go p.grypeJob(grypeTask, syftTask)

	go p.gatecheckListJob(listTask, grypeTask)
	go p.gatecheckBundleJob(bundleTask, syftTask, grypeTask, clamscanTask)

	go p.postRunJob(postRunTask, syftTask, grypeTask, clamscanTask)

	// Stream Task results to stderr and block until the task completes
	// The order here is the stderr stream order, not the execution order.
	// execution order is sudo random because of the goroutines however,
	// dependencies are defined and handle by the jobs.

	allTasks := []*AsyncTask{
		freshclamTask,
		dockerSaveTask,
		syftTask,
		grypeTask,
		clamscanTask,
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

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}
	return nil
}

func (p *ImageScan) dockerSaveJob(task *AsyncTask) {
	defer task.Close()
	defer p.runtime.tarImageFile.Close()
	ctx, cancel := context.WithCancel(p.runtime.ctx)
	// Report to the log every few seconds while the command is running
	// docker save doesn't have a progress meter or anything so
	// when it runs for a long time, it seems like the pipeline freezes
	go func() {
		for {
			after := time.After(time.Second * 5)
			select {
			case <-ctx.Done():
				return
			case <-after:
				task.Logger.Debug("docker save running...")
			}
		}
	}()

	go func() {
		defer cancel()
		task.ExitError = shell.DockerSave(
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithCtx(p.runtime.ctx), // runtime ctx, not internal function ctx
			shell.WithFailTrigger(p.runtime.cancelFunc),
			shell.WithErrorOnly(task.StderrPipeWriter),
			shell.WithLogger(task.Logger),
			shell.WithImageTag(p.config.ImageTag),
			shell.WithDockerAlias(p.runtime.dockerAlias),
			shell.WithStdout(p.runtime.tarImageFile),
			shell.WithStderr(task.StderrPipeWriter),
		)
	}()

	<-ctx.Done()

	if task.ExitError != nil {
		_ = os.Remove(p.runtime.tarImageFilename)
	}
}

func (p *ImageScan) freshclamJob(task *AsyncTask) {
	defer task.Close()

	task.Logger.Debug("update clamav virus definitions database")

	task.ExitError = shell.Freshclam(
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithCtx(p.runtime.ctx),
		shell.WithFailTrigger(p.runtime.cancelFunc),
		shell.WithErrorOnly(task.StderrPipeWriter),
		shell.WithLogger(task.Logger),
		shell.WithStdout(task.StderrPipeWriter), // this is on purpose, we don't want fresh clam in stdout
		shell.WithStderr(task.StderrPipeWriter),
	)
}

func (p *ImageScan) clamscanJob(task *AsyncTask, dockerSaveTask *AsyncTask, freshClamTask *AsyncTask) {
	defer task.Close()
	defer p.runtime.clamavFile.Close()

	task.Logger.Debug("run clamav virus scan after freshclam and docker save complete")

	err := dockerSaveTask.Wait()
	if err != nil {
		task.Logger.Debug("docker save task failed before clamscan could run")
		task.ExitError = errors.New("clamscan task canceled")
		return
	}

	err = freshClamTask.Wait()
	if err != nil {
		task.Logger.Debug("freshclam task failed before clamscan could run")
		task.ExitError = errors.New("clamscan task canceled")
		return
	}
	ctx, cancel := context.WithCancel(p.runtime.ctx)
	// Report to the log every few seconds while the command is running
	go func() {
		for {
			after := time.After(time.Second * 5)
			select {
			case <-ctx.Done():
				return
			case <-after:
				task.Logger.Debug("clamscan running...")
			}
		}
	}()

	go func() {
		defer cancel()
		reportWriter := io.MultiWriter(p.runtime.clamavFile, p.runtime.postSummaryBuffer)
		task.ExitError = shell.Clamscan(
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithLogger(task.Logger),
			shell.WithStdout(reportWriter), // where the report goes
			shell.WithStderr(task.StderrPipeWriter),
			shell.WithTarFilename(p.runtime.tarImageFilename),
		)
	}()
	<-ctx.Done()
	if task.ExitError != nil {
		_ = os.Remove(p.runtime.clamavFilename)
	}
}

func (p *ImageScan) syftJob(task *AsyncTask, dockerSaveTask *AsyncTask) {
	defer task.Close()
	task.Logger.Debug("run syft sbom scan after image tar is created")

	reportWriter := io.MultiWriter(p.runtime.syftFile, task.StdoutBuf)
	task.ExitError = shell.SyftScanImage(
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithLogger(task.Logger),
		shell.WithWaitFunc(dockerSaveTask.Wait),           // defines a dependency
		shell.WithStdout(reportWriter),                    // where the report goes
		shell.WithStderr(p.Stderr),                        // bypass the async stderr, wait function will prevent stderr collisions
		shell.WithStdin(new(bytes.Buffer)),                // This prevents stdin blocking in certain environments
		shell.WithTarFilename(p.runtime.tarImageFilename), // target docker archive to scan
	)

	if task.ExitError != nil {
		_ = os.Remove(p.runtime.syftFilename)
	}
}

func (p *ImageScan) grypeJob(task *AsyncTask, syftTask *AsyncTask) {
	defer task.Close()
	defer p.runtime.grypeFile.Close()

	task.Logger.Debug("run grype vulnerability scan after syft sbom is created")

	reportWriter := io.MultiWriter(p.runtime.grypeFile, task.StdoutBuf)
	// using the syftscan pipe reader will block this job until syft is done
	task.ExitError = shell.GrypeScanSBOM(
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithWaitFunc(syftTask.Wait), // defines a dependency
		shell.WithLogger(task.Logger),
		shell.WithStdin(syftTask.StdoutBuf), // Where the syft SBOM will come from
		shell.WithStdout(reportWriter),      // where the report goes
		shell.WithStderr(p.Stderr),
	)
	if task.ExitError != nil {
		_ = os.Remove(p.runtime.grypeFilename)
	}
}

func (p *ImageScan) gatecheckListJob(task *AsyncTask, grypeTask *AsyncTask) {
	defer task.Close()

	task.Logger.Debug("run gatecheck list after grype report is created")
	task.ExitError = shell.GatecheckListAll(
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithLogger(task.Logger),
		shell.WithWaitFunc(grypeTask.Wait),
		shell.WithStdin(grypeTask.StdoutBuf),
		shell.WithStdout(p.runtime.postSummaryBuffer), // where the report goes
		shell.WithErrorOnly(task.StderrPipeWriter),
		shell.WithReportType("grype"),
	)
}

func (p *ImageScan) gatecheckBundleJob(task *AsyncTask, syftTask *AsyncTask, grypeTask *AsyncTask, clamscanTask *AsyncTask) {
	defer task.Close()

	opts := []shell.OptionFunc{
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithLogger(task.Logger),
		shell.WithWaitFunc(grypeTask.Wait),
		shell.WithStdout(p.Stdout),
		shell.WithErrorOnly(task.StderrPipeWriter),
	}

	syftOpts := append(opts, shell.WithBundleFile(p.runtime.bundleFilename, p.runtime.syftFilename), shell.WithWaitFunc(syftTask.Wait))
	task.ExitError = errors.Join(task.ExitError, shell.GatecheckBundleAdd(syftOpts...))

	grypeOpts := append(opts, shell.WithBundleFile(p.runtime.bundleFilename, p.runtime.grypeFilename), shell.WithWaitFunc(grypeTask.Wait))
	task.ExitError = errors.Join(task.ExitError, shell.GatecheckBundleAdd(grypeOpts...))

	clamavOpts := append(opts, shell.WithBundleFile(p.runtime.bundleFilename, p.runtime.clamavFilename), shell.WithWaitFunc(clamscanTask.Wait))
	task.ExitError = errors.Join(task.ExitError, shell.GatecheckBundleAdd(clamavOpts...))
}

func (p *ImageScan) postRunJob(task *AsyncTask, allTasks ...*AsyncTask) {
	defer task.Close()

	for _, task := range allTasks {
		task.Wait()
	}

	_ = os.Remove(p.runtime.tarImageFilename)

	opts := []shell.OptionFunc{
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithLogger(task.Logger),
		shell.WithStdout(p.Stdout),
		shell.WithErrorOnly(task.StderrPipeWriter),
		shell.WithStdout(p.runtime.postSummaryBuffer),
		shell.WithListTarget(p.runtime.bundleFilename),
	}

	// Add the bundle summary output to the summary buffer before dumping
	err := shell.GatecheckList(opts...)
	task.ExitError = errors.Join(task.ExitError, err)

	// print gatecheck list summary content
	_, _ = p.runtime.postSummaryBuffer.WriteTo(p.Stdout)
}
