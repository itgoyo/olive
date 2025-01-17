package parser

import (
	"io"
	"os/exec"
	"path/filepath"
	"sync"

	l "github.com/go-olive/olive/src/log"
	"github.com/sirupsen/logrus"
)

func init() {
	SharedManager.Register(
		new(streamlink),
	)
}

type streamlink struct {
	cmd      *exec.Cmd
	cmdStdIn io.WriteCloser

	closeOnce sync.Once
	stop      chan struct{}
}

func (s *streamlink) New() Parser {
	return &streamlink{
		stop: make(chan struct{}),
	}
}

func (s *streamlink) Stop() {
	s.closeOnce.Do(func() {
		close(s.stop)
	})
}

func (s *streamlink) Type() string {
	return "streamlink"
}

// streamlink -o "a.mp4"  https://www.twitch.tv/nnabi best -f
func (s *streamlink) Parse(streamURL string, out string) (err error) {
	ext := filepath.Ext(out)
	out = out[0:len(out)-len(ext)] + ".flv"

	l.Logger.WithFields(logrus.Fields{
		// "streamURL": streamURL,
		"out": out,
	}).Debug("streamlink working")

	s.cmd = exec.Command(
		"streamlink",
		"-o", out,
		streamURL,
		"best",
		"-f",
	)
	// s.cmd.Stderr = os.Stderr
	if s.cmdStdIn, err = s.cmd.StdinPipe(); err != nil {
		return err
	}
	if err = s.cmd.Start(); err != nil {
		s.cmd.Process.Kill()
		return err
	}
	go func() {
		<-s.stop
		s.cmdStdIn.Write([]byte("\x03"))
	}()
	return s.cmd.Wait()
}
