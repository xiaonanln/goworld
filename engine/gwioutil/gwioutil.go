package gwioutil

import "io"

type temporaryError interface {
	Temporary() bool // Is the error temporary?
}

func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	terr, ok := err.(temporaryError)
	return ok && terr.Temporary()
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

		if err != nil && !isTemporaryError(err) {
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

		if err != nil && !isTemporaryError(err) {
			return err
		}
	}
	return nil
}
