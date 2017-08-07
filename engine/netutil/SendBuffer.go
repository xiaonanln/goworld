package netutil

import "io"

const (
	SEND_BUFFER_SIZE = 8192 * 2
)

type SendBuffer struct {
	buffer [SEND_BUFFER_SIZE]byte
	//sent    int
	written int
}

// Allocate a new send buffer
func NewSendBuffer() *SendBuffer {
	return &SendBuffer{}
}

func (sb *SendBuffer) Write(b []byte) (n int, err error) {
	n = copy(sb.buffer[sb.written:], b)
	sb.written += n
	return
}

//func (sb *SendBuffer) WriteAllTo(writer io.Writer) (n int, err error) {
//	n, err = writer.Write(sb.buffer[:sb.written])
//	sb.sent += n
//	return
//}

func (sb *SendBuffer) WriteAllTo(writer io.Writer) error {
	if sb.written == 0 {
		return nil
	}

	err := WriteAll(writer, sb.buffer[:sb.written])
	sb.reset()
	return err
}

func (sb *SendBuffer) FreeSpace() int {
	return SEND_BUFFER_SIZE - sb.written
}

func (sb *SendBuffer) reset() {
	//sb.sent = 0
	sb.written = 0
}
