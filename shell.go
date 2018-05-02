// Copyright (c) 2017 Gorillalabs. All rights reserved.

package powershell

import (
	"bytes"
	"fmt"
	"io"

	"github.com/google/martian"
	"github.com/inosvaruag/go-powershell/backend"
	"github.com/inosvaruag/go-powershell/utils"
	"github.com/pkg/errors"
)

const newline = "\r\n"

type Shell interface {
	Execute(cmd string) (string, string, error)
	Exit()
}

type shell struct {
	handle backend.Waiter
	stdin  io.Writer
	stdout io.Reader
	stderr io.Reader
}

func New(backend backend.Starter) (Shell, error) {
	handle, stdin, stdout, stderr, err := backend.StartProcess("powershell.exe", "-NoExit", "-Command", "-")
	if err != nil {
		return nil, err
	}

	return &shell{handle, stdin, stdout, stderr}, nil
}

func (s *shell) Execute(cmd string) (string, string, error) {
	if s.handle == nil {
		return "", "", errors.Wrap(errors.New(cmd), "cannot execute commands on closed shells.")
	}

	outBoundary := createBoundary()
	errBoundary := createBoundary()

	// wrap the command in special markers so we know when to stop reading from the pipes
	full := fmt.Sprintf("%s; echo '%s'; [Console]::Error.WriteLine('%s')%s", cmd, outBoundary, errBoundary, newline)

	_, err := s.stdin.Write([]byte(full))
	if err != nil {
		return "", "", errors.Wrap(errors.Wrap(err, cmd), "could not send PowerShell command")
	}

	// read stdout and stderr
	sout := ""
	serr := ""

	chErr := make(chan error)
	go func() { chErr <- streamReader(s.stdout, outBoundary, &sout, "stdout") }()
	go func() { chErr <- streamReader(s.stderr, errBoundary, &serr, "stderr") }()

	mErr := martian.NewMultiError()
	for i := 0; i < 2; i++ {
		select {
		case stdErr := <-chErr:
			if stdErr != nil {
				mErr.Add(stdErr)
			}
		}
	}

	if !mErr.Empty() {
		return sout, serr, errors.Wrap(errors.Wrap(mErr, cmd), "could not read from std stream")
	}

	if len(serr) > 0 {
		return sout, serr, errors.Wrap(errors.New(cmd), serr)
	}

	return sout, serr, nil
}

func (s *shell) Exit() {
	s.stdin.Write([]byte("exit" + newline))

	// if it's possible to close stdin, do so (some backends, like the local one,
	// do support it)
	closer, ok := s.stdin.(io.Closer)
	if ok {
		closer.Close()
	}

	s.handle.Wait()

	s.handle = nil
	s.stdin = nil
	s.stdout = nil
	s.stderr = nil
}

// streamReader reads and returns the stream up to and excluding boundary
func streamReader(stream io.Reader, boundary string, buffer *string, streamName string) error {
	var buf bytes.Buffer
	marker := []byte(boundary + newline)

	for {
		if _, err := buf.ReadFrom(stream); err != nil {
			return errors.Wrapf(err, "cannot read from %v", streamName)
		}

		if bytes.HasSuffix(buf.Bytes(), marker) {
			break
		}
	}

	*buffer = string(bytes.TrimSuffix(buf.Bytes(), marker))
	return nil
}

func createBoundary() string {
	return "$gorilla" + utils.CreateRandomString(12) + "$"
}
