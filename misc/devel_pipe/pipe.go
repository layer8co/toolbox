// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package pipe

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/layer8co/toolbox/io/moreio"
	"github.com/layer8co/toolbox/os/oslite"
)

var null = []byte{0}

type Process func(in io.ReadCloser, out io.WriteCloser) error

func Pipeline(procs ...Process) {
}

func ReadFiles() Process {

	return func(in io.ReadCloser, out io.WriteCloser) (retErr error) {

		defer func() {
			in.Close()
			retErr = errors.Join(retErr, out.Close())
		}()

		s := bufio.NewScanner(in)
		s.Split(moreio.ScanSep(null))

		for s.Scan() {

			path := s.Text()

			file, err := oslite.Open(path)
			if err != nil {
				return fmt.Errorf("could not open file %q: %w", path, err)
			}

			_, err = io.Copy(out, file)
			file.Close()
			if err != nil {
				return fmt.Errorf("could not read file %q: %w", path, err)
			}
		}

		err := s.Err()
		if err != nil {
			return fmt.Errorf("could not read input: %w", err)
		}
		return nil
	}
}
