package pipelines

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/fatih/color"
	"github.com/lmittmann/tint"
)

// AsyncTask is a simple object for running tasks in the background.
//
// The owner of the task MUST close the stdErrPipeWriter or the Wait function
// will never stop waiting. Best method for this is to defer stdErrPipeWriter.Close()
// at the top of the function scope.
// The owner should use the internal logger in a go routine to prevent stderr interweaving
//
// When ready, the caller can use the Wait function to block until this tasks reader has closed.
//
// Any errors during should be directly set to exitError.
type AsyncTask struct {
	Name string

	StderrPipeReader *io.PipeReader
	StderrPipeWriter *io.PipeWriter

	StdoutBuf *bytes.Buffer

	Logger    *slog.Logger
	ExitError error

	ctx context.Context

	cancelFunc func()
}

func NewAsyncTask(name string) *AsyncTask {
	task := new(AsyncTask)
	task.Name = name
	task.ctx, task.cancelFunc = context.WithCancel(context.Background())

	task.StderrPipeReader, task.StderrPipeWriter = io.Pipe()
	task.StdoutBuf = new(bytes.Buffer)

	task.Logger = slog.New(tint.NewHandler(task.StderrPipeWriter, &tint.Options{Level: slog.LevelDebug, TimeFormat: time.TimeOnly}))

	return task
}

// StreamTo will stream stderr from the task until the task is marked complete
//
// The return will be a combination of any IO error occured during the stream and the task.ExitError
func (t *AsyncTask) StreamTo(stderrWriter io.Writer) error {
	defer t.StderrPipeReader.Close()

	fmt.Fprintf(stderrWriter, "[%s:stderr]   streaming stderr log...\n", t.Name)
	start := time.Now()
	_, writeError := io.Copy(stderrWriter, t.StderrPipeReader)
	if writeError != nil {
		writeError = fmt.Errorf("%s Async Task failed to write line to stderr: %v", t.Name, writeError)
	}

	c := color.New(color.FgRed)
	if t.ExitError == nil {
		c = color.New(color.FgGreen)
	}
	c.Fprintf(stderrWriter, "[%s:stderr]   %s \n\n", t.Name, time.Since(start))
	return errors.Join(t.ExitError, writeError)
}

func (t *AsyncTask) Wait() error {
	<-t.ctx.Done()
	return t.ExitError
}

// Close closes the stdout and stderr writers, any reader would be unblocked after this function is called
//
// it should be defered at the top of a function scope to prevent dead locking
func (t *AsyncTask) Close() {
	t.StderrPipeWriter.Close()

	t.cancelFunc() // end the context so Done will "release" other jobs
}
