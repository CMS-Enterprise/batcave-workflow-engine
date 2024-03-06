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
	shellLegacy "workflow-engine/pkg/shell/legacy"
)

// ImageScan Pipeline
//
// Job status can be used in post to determine what clean up or bundling should be done
type ImageScan struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
	config        *Config
	DockerOrAlias dockerOrAliasCommand
	runtime       struct {
		sbomFile         *os.File
		grypeFile        *os.File
		clamavFile       *os.File
		imageTarFile     *os.File
		sbomFilename     string
		grypeFilename    string
		clamavFilename   string
		bundleFilename   string
		imageTarFilename string
		gatecheckListBuf *bytes.Buffer
		clamavReportBuf  *bytes.Buffer
		taskChan         chan asyncTask
		syftJobSuccess   bool
		grypeJobSuccess  bool
		clamJobSuccess   bool
	}
}

type asyncTask struct {
	name    string
	success bool
	logBuf  *bytes.Buffer
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
		// will declare at runtime
		DockerOrAlias: shellLegacy.DockerCommand(nil, nil, nil),
	}
}

func (p *ImageScan) WithPodman() *ImageScan {
	slog.Debug("use podman cli")
	p.DockerOrAlias = shellLegacy.PodmanCommand(nil, p.Stdout, p.Stderr)
	return p
}

func (p *ImageScan) preRun() error {
	var err error
	fmt.Fprintln(p.Stderr, "******* Workflow Engine Image Scan Pipeline [Pre-Run] *******")

	// In Memory buffers for runtime reports, async logs, etc.
	p.runtime.clamavReportBuf = new(bytes.Buffer)
	p.runtime.gatecheckListBuf = new(bytes.Buffer)

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

	slog.Debug("create temporary file for image tar, used for clam virus scan")
	p.runtime.imageTarFile, err = os.CreateTemp(os.TempDir(), "*-image.tar")
	if err != nil {
		slog.Error("cannot create temp image tar file", "temp_dir", os.TempDir())
		return err
	}
	p.runtime.imageTarFilename = p.runtime.imageTarFile.Name()

	// create gatecheck bundle file
	if err := InitGatecheckBundle(p.config, p.Stderr, p.DryRunEnabled); err != nil {
		slog.Error("cannot initialize gatecheck bundle", "error", err)
		return err
	}

	p.runtime.bundleFilename = path.Join(p.config.ArtifactsDir, p.config.GatecheckBundleFilename)

	// The follow are async tasks meaning that they should happen concurrently in the background since they
	// take so long.
	//
	// putting 'go' in front will run it in the background and continue execution of this function
	// the routine is handed off to the go scheduler so it won't stop once this function's scope exits

	// Using context, the execution can be signaled to cancel if one of the other tasks fails first.
	// If docker fails, we can't do a clamAV scan  or vice versa, so fail the entire pre-run
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Results from the function can just be "error" because it would block execution of this function
	slog.Debug("start async task update virus definitions with fresh clam")
	go func() {
		task := asyncTask{name: "fresh clam", success: true, logBuf: new(bytes.Buffer)}
		exitCode := shell.Freshclam(
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithStdout(task.logBuf),
			shell.WithCtx(ctx),
		)
		if exitCode != shell.ExitOK {
			cancel() // cancel other async tasks
			task.success = false
			_, _ = fmt.Fprintf(task.logBuf, "\n Exit Code: %d", exitCode)
		}
		p.runtime.taskChan <- task
	}()

	slog.Debug("start async task image tarball download")
	go func() {
		task := asyncTask{name: "docker save", success: true, logBuf: new(bytes.Buffer)}
		exitCode := shell.DockerSave(
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithCtx(ctx),
			shell.WithImage(p.runtime.imageTarFilename),
			shell.WithIO(nil, p.runtime.imageTarFile, task.logBuf),
		)
		if exitCode != shell.ExitOK {
			cancel() // cancel other async tasks
			task.success = false
			_, _ = fmt.Fprintf(task.logBuf, "\n Exit Code: %d", exitCode)
		}
		p.runtime.taskChan <- task
	}()

	_ = p.runtime.imageTarFile.Close()
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

	defer func() {
		_ = p.runtime.sbomFile.Close()
		_ = p.runtime.grypeFile.Close()
		_ = p.runtime.clamavFile.Close()
	}()

	fmt.Fprintln(p.Stderr, "******* Workflow Engine Image Scan Pipeline [Run] *******")
	// make a new channel for only this function scope
	runTaskChan := make(chan asyncTask, 1)
	// Scope the clam scan so it can run in the background since it will take longer than the others
	clamScanJob := func() {
		taskCount := 2
		var clamScanError error
		// wait for freshclam and dockersave to complete
		for i := 0; i < taskCount; i++ {
			task := <-p.runtime.taskChan
			if !task.success {
				err := fmt.Errorf("cannot run clamscan, dependent task failed: %s", task.name)
				clamScanError = errors.Join(clamScanError, err)
			}
		}
		// Create a new task
		clamScanTask := asyncTask{name: "clamscan", success: true, logBuf: new(bytes.Buffer)}

		// fail early if the freshclam or docker save failed
		if clamScanError != nil {
			clamScanTask.success = false
			runTaskChan <- clamScanTask
			return
		}

		mw := io.MultiWriter(p.runtime.clamavFile, p.runtime.clamavReportBuf)
		exitCode := shell.Clamscan(
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithIO(nil, mw, clamScanTask.logBuf),
		)
		if exitCode != shell.ExitOK {
			clamScanTask.success = false
		}
		runTaskChan <- clamScanTask
	}

	// Scope the vulnerability scan to prevent early return
	// Grype is dependent on syft output, so if syft fails, grype shouldn't run at all
	// Gatecheck list is dependent on Grype, so if it fails, the list should run either
	syftGrypeJob := func() error {
		syftReportBuf := new(bytes.Buffer)
		syftMW := io.MultiWriter(p.runtime.sbomFile, syftReportBuf)

		syftExit := shell.SyftScanImage(
			shell.WithScanImage(p.config.ImageScan.TargetImage),
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithIO(nil, syftMW, p.Stderr),
		)

		if syftExit != shell.ExitOK {
			slog.Error("syft sbom generation failed")
			return fmt.Errorf("syft non-zero exit: %d", syftExit)
		}
		// Syft report passed so it can be added to the bundle
		p.runtime.syftJobSuccess = true

		grypeReportBuf := new(bytes.Buffer)
		grypeMW := io.MultiWriter(p.runtime.grypeFile, grypeReportBuf)

		grypeExit := shell.GrypeScanSBOM(
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithIO(syftReportBuf, grypeMW, p.Stderr),
		)
		if grypeExit != shell.ExitOK {
			slog.Error("grype vulnerability scan failed")
			return fmt.Errorf("grype non-zero exit: %d", grypeExit)
		}

		// Grype report passed so it can be added to the bundle
		p.runtime.grypeJobSuccess = true

		slog.Debug("summarize grype report")
		listErrBuf := new(bytes.Buffer)
		err := RunGatecheckListAll(p.runtime.gatecheckListBuf, grypeReportBuf, listErrBuf, "grype", p.DryRunEnabled)
		if err != nil {
			slog.Error("cannot run gatecheck list all on grype report, dumping stderr log")
			_, _ = io.Copy(p.Stderr, listErrBuf)
			return err
		}
		return nil
	}

	slog.Debug("start clam virus scan in the background")
	go clamScanJob()

	syftGrypeError := syftGrypeJob()

	slog.Debug("waiting for clam virus scan to complete, suppressing stderr unless command fails")
	var clamScanError error

	// blocking
	task := <-runTaskChan
	if !task.success {
		clamScanError = errors.New("Clamscan failed. See log for details")
		slog.Error("clam scan failed, dumping logs")
		_, _ = io.Copy(p.Stderr, task.logBuf)
	}

	p.runtime.clamJobSuccess = task.success

	postRunError := p.postRun()
	if postRunError != nil {
		slog.Error("post run failed", "error", postRunError)
	}

	// return all possible pipeline run errors
	if err := errors.Join(syftGrypeError, clamScanError, postRunError); err != nil {
		return errors.New("Image Scan Pipeline failed. See log for details")
	}
	return nil
}

func (p *ImageScan) postRun() error {
	fmt.Fprintln(p.Stderr, "******* Workflow Engine Image Scan Pipeline [Post-Run] *******")
	cleanUpFiles := []string{p.runtime.imageTarFilename}
	bundleFiles := []string{}

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

	// print clamAV Report Content
	_, _ = p.runtime.clamavFile.Seek(0, io.SeekStart)
	_, _ = io.Copy(p.Stderr, p.runtime.clamavReportBuf)

	// print the Gatecheck List Content
	_, _ = p.runtime.gatecheckListBuf.WriteTo(p.Stdout)

	return err
}
