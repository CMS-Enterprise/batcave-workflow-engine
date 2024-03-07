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

	if err := MakeDirectoryP(p.config.ArtifactsDir); err != nil {
		slog.Error("failed to create artifact directory", "name", p.config.ArtifactsDir)
		return errors.New("Code Scan Pipeline failed to run. See log for details.")
	}

	p.runtime.sbomFilename = path.Join(p.config.ArtifactsDir, p.config.ImageScan.SyftFilename)
	p.runtime.sbomFile, err = OpenOrCreateFile(p.runtime.sbomFilename)
	if err != nil {
		slog.Error("cannot open syft sbom file", "filename", p.runtime.sbomFilename, "error", err)
		return err
	}

	p.runtime.grypeFilename = path.Join(p.config.ArtifactsDir, p.config.ImageScan.GrypeFullFilename)
	p.runtime.grypeFile, err = OpenOrCreateFile(p.runtime.grypeFilename)
	if err != nil {
		slog.Error("cannot open grype sbom file", "filename", p.runtime.grypeFilename, "error", err)
		return err
	}

	p.runtime.clamavFilename = path.Join(p.config.ArtifactsDir, p.config.ImageScan.ClamavFilename)
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

	p.runtime.bundleFilename = path.Join(p.config.ArtifactsDir, p.config.GatecheckBundleFilename)

	return nil
}

func (p *ImageScan) Run() error {
	if !p.config.ImageScan.Enabled {
		slog.Warn("image scan pipeline is disabled, skip.")
		return nil
	}

	if err := p.preRun(); err != nil {
		return errors.New("Code Scan Pipeline Pre-Run Failed. See log for details.")
	}

	fmt.Fprintln(p.Stdout, "******* Workflow Engine Image Scan Pipeline [Run] *******")

	// Run in the background since this task takes a log time, we can stream the log after other jobs run
	clamscanTask := NewAsyncTask("clamscan")
	mw := io.MultiWriter(p.runtime.clamavFile, p.runtime.postSummaryBuffer)
	go RunClamScanJob(clamscanTask, mw, p.DryRunEnabled, shell.DockerAliasDocker, p.config.ImageScan.TargetImage)

	// Scope this way in-order to return without returning the entire run function
	syftGrypeError := func() error {
		syftBuf := new(bytes.Buffer)
		exitCode := shell.SyftScanImage(
			shell.WithScanImage(p.config.ImageScan.TargetImage),
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithStdout(io.MultiWriter(syftBuf, p.runtime.sbomFile)),
			shell.WithStderr(p.Stderr),
		)
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

		// List Report
		errBuf := new(bytes.Buffer) // Save stderr to dump if error
		exitCode = shell.GatecheckListAll(
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithReportType("grype"),
			shell.WithIO(grypeBuf, p.runtime.postSummaryBuffer, errBuf),
		)
		fmt.Fprintln(p.runtime.postSummaryBuffer)

		// Reduce the logging noise to only essential tasks, only dump stderr for errors
		if exitCode != shell.ExitOK {
			slog.Debug("dump gatecheck list stderr")
			_, _ = io.Copy(p.Stderr, errBuf)
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
	_, _ = io.Copy(p.Stdout, p.runtime.postSummaryBuffer)

	return err
}

func RunClamScanJob(task *AsyncTask, reportDst io.Writer, dryRunEnabled bool, alias shell.DockerAlias, targetImage string) {
	defer task.stdErrPipeWriter.Close()
	commonError := errors.New("Clam Scan Job Failed. See log for details.")
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
		exitCode := shell.DockerSave(
			shell.WithImage(targetImage),
			shell.WithDockerAlias(alias),
			shell.WithCtx(ctx),
			shell.WithFailTrigger(cancel),
			shell.WithDryRun(dryRunEnabled),
			shell.WithStdout(imageTarFile),
			shell.WithStderr(dockerSaveTask.stdErrPipeWriter))
		dockerSaveTask.exitError = exitCode.GetError("docker save")
	}()

	go func() {
		defer freshclamTask.stdErrPipeReader.Close()
		exitCode := shell.Freshclam(
			shell.WithCtx(ctx),
			shell.WithFailTrigger(cancel),
			shell.WithDryRun(dryRunEnabled),
			shell.WithStdout(freshclamTask.stdErrPipeWriter),
		)
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

	exitCode := shell.Clamscan(
		shell.WithDryRun(dryRunEnabled),
		shell.WithTarFilename(imageTarFilename),
		shell.WithIO(nil, reportDst, task.stdErrPipeWriter),
	)

	task.exitError = exitCode.GetError("clamscan")
}
