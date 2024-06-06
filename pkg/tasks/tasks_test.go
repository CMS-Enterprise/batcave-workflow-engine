package tasks

import (
	"bufio"
	"bytes"
	"context"
	"math/rand"
	"testing"
	"time"
)

func TestSizeMonitorWriter(t *testing.T) {
	gen := rand.New(rand.NewSource(1))
	buf := new(bytes.Buffer)
	smWriter := NewSizeMonitorWriter("test", "test-file.txt", buf)
	smWriter.Interval = time.Microsecond * 100 // 1/10 of a millisecond
	ctx, cancel := context.WithCancel(context.Background())

	go smWriter.Start(ctx)

	for i := 0; i < 1_000; i++ {
		content := make([]byte, gen.Intn(100))
		smWriter.Write(content)
		time.Sleep(time.Microsecond * 1)
	}

	cancel()

	scanner := bufio.NewScanner(buf)

	for scanner.Scan() {
		t.Log(scanner.Text())
	}

}
