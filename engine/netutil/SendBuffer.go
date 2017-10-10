package netutil

import (
	"io"

	"github.com/xiaonanln/goworld/engine/gwioutil"
)

const (
	_SEND_BUFFER_SIZE = 8192 * 2
)

type sendBuffer struct {
	buffer [_SEND_BUFFER_SIZE]byte
	//sent    int
	written int
}

// newSendBuffer allocates a new send buffer
func newSendBuffer() *sendBuffer {
	return &sendBuffer{}
}

func (sb *sendBuffer) Write(b []byte) (n int, err error) {
	n = copy(sb.buffer[sb.written:], b)
	sb.written += n
	return
}

func (sb *sendBuffer) WriteAllTo(writer io.Writer) error {
	if sb.written == 0 {
		return nil
	}

	err := gwioutil.WriteAll(writer, sb.buffer[:sb.written])
	sb.reset()
	return err
}

func (sb *sendBuffer) FreeSpace() int {
	return _SEND_BUFFER_SIZE - sb.written
}

func (sb *sendBuffer) reset() {
	//sb.sent = 0
	sb.written = 0
}
