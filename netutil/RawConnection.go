package bigworld_netutil

import "net"

type RawConnection struct {
	Conn net.Conn
}

func NewConnection(conn net.Conn) RawConnection {
	return RawConnection{conn}
}

func (c RawConnection) String() string {
	return c.Conn.RemoteAddr().String()
}

func (c RawConnection) RecvByte() (byte, error) {
	buf := []byte{0}
	for {
		n, err := c.Conn.Read(buf)
		if n >= 1 {
			return buf[0], nil
		} else if err != nil {
			if IsTemporaryNetError(err) {
				continue
			}
			return 0, err
		}
	}
}

func (c RawConnection) SendByte(b byte) error {
	buf := []byte{b}
	for {
		n, err := c.Conn.Write(buf)
		if n >= 1 {
			return nil
		} else if err != nil {
			if IsTemporaryNetError(err) {
				continue
			}
			return err
		}
	}
}

func (c RawConnection) RecvAll(buf []byte) error {
	for len(buf) > 0 {
		n, err := c.Conn.Read(buf)
		if err != nil {
			if IsTemporaryNetError(err) {
				continue
			}

			return err
		}
		buf = buf[n:]
	}
	return nil
}

func (c RawConnection) SendAll(data []byte) error {
	for len(data) > 0 {
		n, err := c.Conn.Write(data)
		if err != nil {
			if IsTemporaryNetError(err) {
				continue
			}

			return err
		}
		data = data[n:]
	}
	return nil
}

func (c RawConnection) Read(data []byte) (int, error) {
	return c.Conn.Read(data)
}

func (c RawConnection) Write(data []byte) (int, error) {
	return c.Conn.Write(data)
}

func (c RawConnection) Close() error {
	return c.Conn.Close()
}
