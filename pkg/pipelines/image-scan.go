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
	"sync"
	"time"
	"workflow-engine/pkg/shell"
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
		syftJobSuccess   bool
		grypeJobSuccess  bool
		clamJobSuccess   bool
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
		DockerOrAlias: shell.DockerCommand(nil, stdout, stderr),
	}
}

func (p *ImageScan) WithPodman() *ImageScan {
	slog.Debug("use podman cli")
	p.DockerOrAlias = shell.PodmanCommand(nil, p.Stdout, p.Stderr)
	return p
}

func (p *ImageScan) preRun() error {
	var err error

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

	// The follow are async tasks meaning that they should happen concurrently in the background since they
	// take so long.
	//
	// putting 'go' in front will run it in the background and continue execution of this function
	// the routine is handed off to the go scheduler so it won't stop once this function's scope exits

	// Using context, the execution can be signaled to cancel if one of the other tasks fails first.
	// If docker fails, we can't do a clamAV scan anyway so fail the entire prerun
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Results from the function can just be "error" because it would block execution of this function
	// By wrapping the results in a struct, we can monitor later via a channel
	resultChan := make(chan asyncResult)

	slog.Debug("start async task update virus definitions with fresh clam")
	go clamavDatabaseUpdate(ctx, resultChan, p.Stderr, p.DryRunEnabled)

	slog.Debug("start async task image tarball download")
	go dockerSave(ctx, resultChan, p.DockerOrAlias, p.runtime.imageTarFile, p.config.ImageScan.TargetImage, true, p.DryRunEnabled)
	// TODO: add a remote bool to the config object

	// create gatecheck bundle file
	if err := InitGatecheckBundle(p.config, p.Stderr, p.DryRunEnabled); err != nil {
		slog.Error("cannot initialize gatecheck bundle", "error", err)
		return err
	}

	p.runtime.bundleFilename = path.Join(p.config.ArtifactsDir, p.config.GatecheckBundleFilename)
	p.runtime.gatecheckListBuf = new(bytes.Buffer)

	var taskError error
	tasks := 2
	slog.Debug("wait for async tasks to complete")
	for i := 0; i < tasks; i++ {
		res := <-resultChan
		if res.success {
			slog.Info(res.msg, "task_name", res.taskName, "elapsed", res.elapsed)
			// wait for the other task to finish
			continue
		}
		// cancel the context for the async tasks, all tasks should return at some point even after cancel
		cancel()
		slog.Error(res.msg, "task_name", res.taskName, "elapsed", res.elapsed)
		taskError = errors.New("async task failure")
	}

	return taskError
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

	clamStdErrBuf := new(bytes.Buffer)
	var clamScanError error

	var wg sync.WaitGroup // used to determine when the clam av job finishes
	wg.Add(1)             // Only the clamAV job is run async

	// Scope the clam scan so it can run in the background since it will take longer than the others
	// careful not to modify values like p.runtime here because it's not thread safe to access
	// fields in a struct at the same time. in the future we could make it safe by using a mutux.
	clamScanJob := func() {
		clamScanError = RunClamScan(p.runtime.clamavFile, clamStdErrBuf, p.runtime.imageTarFilename, p.DryRunEnabled)
		wg.Done()
	}

	clamScanErrStream := func(w io.Writer) {
		for {
			time.Sleep(time.Millisecond * 10)
			_, _ = io.Copy(w, clamStdErrBuf)
		}
	}

	// Scope the vulnerability scan to prevent early return
	// Grype is dependent on syft output, so if syft fails, grype shouldn't run at all
	// Gatecheck list is dependent on Grype, so if it fails, the list should run either
	syftGrypeJob := func() error {
		syftReportBuf := new(bytes.Buffer)
		syftMW := io.MultiWriter(p.runtime.sbomFile, syftReportBuf)

		err := RunSyftScan(syftMW, p.Stderr, p.config, p.DryRunEnabled)
		if err != nil {
			slog.Error("syft sbom generation failed")
			return err
		}
		// Syft report passed so it can be added to the bundle
		p.runtime.syftJobSuccess = true

		grypeReportBuf := new(bytes.Buffer)
		grypeMW := io.MultiWriter(p.runtime.grypeFile, grypeReportBuf)

		err = RunGrypeScanSBOM(grypeMW, syftReportBuf, p.Stderr, p.config, p.DryRunEnabled)
		if err != nil {
			slog.Error("grype vulnerability scan failed")
			return err
		}
		// Grype report passed so it can be added to the bundle
		p.runtime.grypeJobSuccess = true

		slog.Debug("summarize grype report")
		err = RunGatecheckListAll(p.runtime.gatecheckListBuf, grypeReportBuf, p.Stderr, "grype", p.DryRunEnabled)
		if err != nil {
			slog.Error("cannot run gatecheck list all on grype report")
			return err
		}
		return nil
	}

	slog.Info("run image scan pipeline", "dry_run_enabled", p.DryRunEnabled, "artifact_directory", p.config.ArtifactsDir)
	slog.Debug("start clam virus scan in the background")
	go clamScanJob()

	syftGrypeError := syftGrypeJob()

	slog.Debug("waiting for clam virus scan to complete, streaming stderr")
	go clamScanErrStream(p.Stderr)
	wg.Wait() // Block here until clamAV finishes

	if clamScanError != nil {
		slog.Error("clam scan failed", "error", clamScanError)
	} else {
		p.runtime.clamJobSuccess = true
	}

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

	err := RunGatecheckBundleAdd(p.runtime.bundleFilename, p.Stderr, p.DryRunEnabled, bundleFiles...)
	if err != nil {
		slog.Error("cannot run gatecheck bundle add", "error", err)
	}

	// print the Gatecheck List Content
	_, _ = p.runtime.gatecheckListBuf.WriteTo(p.Stdout)

	// print clamAV Report Content

	return err
}

func RunSyftScan(reportDst io.Writer, stdErr io.Writer, config *Config, dryRunEnabled bool) error {
	return shell.SyftCommand(nil, reportDst, stdErr).ScanImage(config.ImageScan.TargetImage).WithDryRun(dryRunEnabled).Run()
}

func RunGrypeScanSBOM(reportDst io.Writer, syftSrc io.Reader, stdErr io.Writer, config *Config, dryRunEnabled bool) error {
	return shell.GrypeCommand(syftSrc, reportDst, stdErr).ScanSBOM().WithDryRun(dryRunEnabled).Run()
}

func RunClamScan(reportDst io.Writer, stdErr io.Writer, targetDirectory string, dryRunEnabled bool) error {
	return shell.ClamScanCommand(nil, reportDst, stdErr).Scan(targetDirectory).WithDryRun(dryRunEnabled).Run()
}

type asyncResult struct {
	taskName string
	success  bool
	msg      string
	elapsed  time.Duration
}

func clamavDatabaseUpdate(ctx context.Context, resultChan chan<- asyncResult, stdErr io.Writer, dryRunEnabled bool) {
	start := time.Now()
	res := asyncResult{success: true, taskName: "clamav database update", msg: "clamav database successfully updated with freshclam"}

	err := shell.FreshClamCommand(nil, stdErr, stdErr).FreshClam().WithDryRun(dryRunEnabled).RunWithContext(ctx)
	if err != nil {
		res.success = false
		res.msg = err.Error()
	}

	res.elapsed = time.Since(start)
	resultChan <- res
}

func dockerSave(ctx context.Context, resultChan chan<- asyncResult, docker dockerOrAliasCommand, w io.Writer, image string, remote bool, dryRunEnabled bool) {
	start := time.Now()
	res := asyncResult{success: true, taskName: "docker save", msg: "docker save complete"}
	if remote {
		err := docker.Pull(image).WithDryRun(dryRunEnabled).RunWithContext(ctx)
		if err != nil {
			res.success = false
			res.msg = fmt.Sprintf("cannot pull remote image. error: %v", err)
			res.elapsed = time.Since(start)
			resultChan <- res
			return
		}
	}

	err := docker.Save(image, w).WithDryRun(dryRunEnabled).RunWithContext(ctx)
	if err != nil {
		res.success = false
		res.msg = fmt.Sprintf("cannot save image. error: %v", err)
	}

	res.elapsed = time.Since(start)
	resultChan <- res
}
