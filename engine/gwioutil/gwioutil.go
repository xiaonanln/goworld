package gwioutil

import (
	"io"

	"github.com/pkg/errors"
)

type timeoutError interface {
	Timeout() bool // Is it a timeout error
}

// IsTimeoutError checks if the error is a timeout error
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	err = errors.Cause(err)
	ne, ok := err.(timeoutError)
	return ok && ne.Timeout()
}

// WriteAll write all bytes of data to the writer
func WriteAll(conn io.Writer, data []byte) error {
	left := len(data)
	for left > 0 {
		n, err := conn.Write(data)
		if n == left && err == nil { // handle most common case first
			return nil
		}

		if n > 0 {
			data = data[n:]
			left -= n
		}

		if err != nil && !IsTimeoutError(err) {
			return err
		}
	}
	return nil
}

// ReadAll reads from the reader until all bytes in data is filled
func ReadAll(conn io.Reader, data []byte) error {
	left := len(data)
	for left > 0 {
		n, err := conn.Read(data)
		if n == left && err == nil { // handle most common case first
			return nil
		}

		if n > 0 {
			data = data[n:]
			left -= n
		}

		if err != nil && !IsTimeoutError(err) {
			return err
		}
	}
	return nil
}
