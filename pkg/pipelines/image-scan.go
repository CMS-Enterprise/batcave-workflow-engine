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
		sbomFile          *os.File
		grypeFile         *os.File
		clamavFile        *os.File
		sbomFilename      string
		grypeFilename     string
		clamavFilename    string
		bundleFilename    string
		syftJobSuccess    bool
		grypeJobSuccess   bool
		clamJobSuccess    bool
		postSummaryBuffer *bytes.Buffer
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

	if err := MakeDirectoryP(p.config.ArtifactDir); err != nil {
		slog.Error("failed to create artifact directory", "name", p.config.ArtifactDir)
		return errors.New("Code Scan Pipeline failed to run.")
	}

	p.runtime.sbomFilename = path.Join(p.config.ArtifactDir, p.config.ImageScan.SyftFilename)
	p.runtime.sbomFile, err = OpenOrCreateFile(p.runtime.sbomFilename)
	if err != nil {
		slog.Error("cannot open syft sbom file", "filename", p.runtime.sbomFilename, "error", err)
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

	// create gatecheck bundle file
	if err := InitGatecheckBundle(p.config, p.Stderr, p.DryRunEnabled); err != nil {
		slog.Error("cannot initialize gatecheck bundle", "error", err)
		return err
	}

	p.runtime.bundleFilename = path.Join(p.config.ArtifactDir, p.config.GatecheckBundleFilename)

	return nil
}

func (p *ImageScan) Run() error {
	if !p.config.ImageScan.Enabled {
		slog.Warn("image scan pipeline is disabled, skip.")
		return nil
	}

	if err := p.preRun(); err != nil {
		return errors.New("Code Scan Pipeline Pre-Run Failed.")
	}

	fmt.Fprintln(p.Stdout, "******* Workflow Engine Image Scan Pipeline [Run] *******")

	alias := shell.DockerAliasDocker
	// print the connection information, exit pipeline if failed
	switch strings.ToLower(p.DockerAlias) {
	case "podman":
		alias = shell.DockerAliasPodman
	case "docker":
		alias = shell.DockerAliasDocker
	}

	// Run in the background since this task takes a log time, we can stream the log after other jobs run
	clamscanTask := NewAsyncTask("clamscan")
	mw := io.MultiWriter(p.runtime.clamavFile, p.runtime.postSummaryBuffer)
	opts := []shell.OptionFunc{
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithImageTag(p.config.ImageTag), // clamscan target
		shell.WithDockerAlias(alias),
	}
	go RunClamScanJob(clamscanTask, mw, opts)

	// Scope this way in-order to return without returning the entire run function
	syftGrypeError := func() error {
		syftBuf := new(bytes.Buffer)
		opts := []shell.OptionFunc{
			shell.WithImageTag(p.config.ImageTag),
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithStdout(io.MultiWriter(syftBuf, p.runtime.sbomFile)),
			shell.WithStderr(p.Stderr),
		}
		exitCode := shell.SyftScanImage(opts...)
		if exitCode != shell.ExitOK {
			return exitCode.GetError("syft") // just return to syftGrypeError
		}
		grypeBuf := new(bytes.Buffer)
		exitCode = shell.GrypeScanSBOM(
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithIO(syftBuf, io.MultiWriter(p.runtime.grypeFile, grypeBuf), p.Stderr),
		)
		if exitCode != shell.ExitOK {
			return exitCode.GetError("grype")
		}
		p.runtime.grypeJobSuccess = true
		p.runtime.syftJobSuccess = true

		opts = []shell.OptionFunc{
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithReportType("grype"),
			shell.WithIO(grypeBuf, p.runtime.postSummaryBuffer, nil),
			// Reduce the logging noise to only essential tasks, only dump stderr for errors
			shell.WithErrorOnly(clamscanTask.stdErrPipeWriter),
		}

		// List Report
		exitCode = shell.GatecheckListAll(opts...)
		fmt.Fprintln(p.runtime.postSummaryBuffer)

		if exitCode != shell.ExitOK {
			return exitCode.GetError("gatecheck list")
		}

		return nil
	}() // Call function immediately

	// The stream order of tasks is important to prevent the log from gettings mangled
	// By reading from each task reader, it will block until the task is complete
	clamScanError := clamscanTask.Wait(p.Stdout)

	if clamScanError == nil {
		p.runtime.clamJobSuccess = true
	}

	postRunError := p.postRun()

	return errors.Join(clamScanError, syftGrypeError, postRunError)
}

func (p *ImageScan) postRun() error {
	fmt.Fprintln(p.Stdout, "\n******* Workflow Engine Image Scan Pipeline [Post-Run] *******")
	cleanUpFiles := []string{}
	bundleFiles := []string{}

	_ = p.runtime.sbomFile.Close()
	_ = p.runtime.grypeFile.Close()
	_ = p.runtime.clamavFile.Close()

	if p.runtime.syftJobSuccess {
		bundleFiles = append(bundleFiles, p.runtime.sbomFilename)
	} else {
		cleanUpFiles = append(cleanUpFiles, p.runtime.sbomFilename)
	}

	if p.runtime.grypeJobSuccess {
		bundleFiles = append(bundleFiles, p.runtime.grypeFilename)
	} else {
		cleanUpFiles = append(cleanUpFiles, p.runtime.grypeFilename)
	}

	if p.runtime.clamJobSuccess {
		bundleFiles = append(bundleFiles, p.runtime.clamavFilename)
	} else {
		cleanUpFiles = append(cleanUpFiles, p.runtime.clamavFilename)
	}

	// delete temporary or incomplete file
	for _, filename := range cleanUpFiles {
		if err := os.RemoveAll(filename); err != nil {
			slog.Warn("during post run, file could not be removed", "filename", filename, "error", err)
		}
	}

	errBuf := new(bytes.Buffer)
	err := RunGatecheckBundleAdd(p.runtime.bundleFilename, errBuf, p.DryRunEnabled, bundleFiles...)
	if err != nil {
		slog.Error("cannot run gatecheck bundle add, dumping logs", "error", err)
		_, _ = io.Copy(p.Stderr, errBuf)
	}
	// print clamAV Report and grype summary
	_, _ = p.runtime.postSummaryBuffer.WriteTo(p.Stdout)
	return err
}

func RunClamScanJob(task *AsyncTask, reportDst io.Writer, options []shell.OptionFunc) {
	defer task.stdErrPipeWriter.Close()
	commonError := errors.New("Clam Scan Job Failed.")
	// Create temporary image tar file for writing
	slog.Debug("create temporary file for image tar, used for clam virus scan")
	imageTarFile, err := os.CreateTemp(os.TempDir(), "*-image.tar")
	if err != nil {
		task.logger.Error("cannot create temp image tar file", "temp_dir", os.TempDir())
		task.exitError = commonError
		return
	}
	imageTarFilename := imageTarFile.Name()

	// Clean up to run before function scope ends
	defer func() {
		_ = imageTarFile.Close()
		_ = os.Remove(imageTarFilename)
	}()

	dockerSaveTask := NewAsyncTask("docker save")
	freshclamTask := NewAsyncTask("freshclam")
	ctx, cancel := context.WithCancel(context.Background())

	// Run the docker save task in the background with the option to termination early
	// logging will be sent to the async task.
	// If the command fails, it will call cancel to trigger an interupt on the freshclam job
	go func() {
		defer dockerSaveTask.stdErrPipeReader.Close()
		opts := append(
			options,
			shell.WithCtx(ctx),
			shell.WithFailTrigger(cancel),
			shell.WithStdout(imageTarFile),
			shell.WithStderr(dockerSaveTask.stdErrPipeWriter),
		)
		exitCode := shell.DockerSave(opts...)
		dockerSaveTask.exitError = exitCode.GetError("docker save")
	}()

	go func() {
		defer freshclamTask.stdErrPipeReader.Close()
		opts := append(
			options,
			shell.WithCtx(ctx),
			shell.WithFailTrigger(cancel),
			shell.WithStdout(freshclamTask.stdErrPipeWriter),
			shell.WithStderr(freshclamTask.stdErrPipeWriter),
		)

		exitCode := shell.Freshclam(opts...)
		freshclamTask.exitError = exitCode.GetError("freshclam")
	}()

	task.logger.Debug("clamscan wait for freshclam update and docker save to complete")
	defer task.stdErrPipeWriter.Close()
	// Wait until these tasks finish first, streaming their outputs in order to this tasks logger
	_ = dockerSaveTask.Wait(task.stdErrPipeWriter)
	_ = freshclamTask.Wait(task.stdErrPipeWriter)

	// determine if both tasks passed before moving on
	prescanErrors := errors.Join(dockerSaveTask.exitError, freshclamTask.exitError)
	if err := errors.Join(prescanErrors); err != nil {
		task.logger.Error("cannot run clamscan without image tar and freshclam update")
		task.exitError = prescanErrors
		return
	}

	opts := append(
		options,
		shell.WithTarFilename(imageTarFilename),
		shell.WithIO(nil, reportDst, task.stdErrPipeWriter),
	)
	exitCode := shell.Clamscan(opts...)

	task.exitError = exitCode.GetError("clamscan")
}
