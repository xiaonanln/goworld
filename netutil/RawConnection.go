package netutil

import "net"

type RawConnection struct {
	net.Conn
}

func NewRawConnection(conn net.Conn) RawConnection {
	return RawConnection{conn}
}

func (rc RawConnection) String() string {
	return rc.Conn.RemoteAddr().String()
}

//func (rc RawConnection) RecvByte() (byte, error) {
//	buf := []byte{0}
//	for {
//		n, err := rc.Conn.Read(buf)
//		if n >= 1 {
//			return buf[0], nil
//		} else if err != nil {
//			if IsTemporaryNetError(err) {
//				continue
//			}
//			return 0, err
//		}
//	}
//}
//
//func (rc RawConnection) SendByte(b byte) error {
//	buf := []byte{b}
//	for {
//		n, err := rc.Conn.Write(buf)
//		if n >= 1 {
//			return nil
//		} else if err != nil {
//			if IsTemporaryNetError(err) {
//				continue
//			}
//			return err
//		}
//	}
//}

func (rc RawConnection) Recv(buf []byte) error {
	for len(buf) > 0 {
		n, err := rc.Conn.Read(buf)
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

func (rc RawConnection) Send(data []byte) error {
	for len(data) > 0 {
		n, err := rc.Conn.Write(data)
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
