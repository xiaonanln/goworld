package netutil

import (
	"fmt"
	"net"

	"encoding/binary"

	"sync"

	"time"

	"sync/atomic"

	"compress/flate"

	"os"

	"io"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goNewlessPool"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/opmon"
)

const ( // Three different level of packet size
	PACKET_SIZE_SMALL = 1024
	PACKET_SIZE_BIG   = 1024 * 64
	PACKET_SIZE_HUGE  = 1024 * 1024 * 4
)

const (
	MAX_PACKET_SIZE    = 25 * 1024 * 1024
	SIZE_FIELD_SIZE    = 4
	PREPAYLOAD_SIZE    = SIZE_FIELD_SIZE
	MAX_PAYLOAD_LENGTH = MAX_PACKET_SIZE - PREPAYLOAD_SIZE
)

var (
	NETWORK_ENDIAN      = binary.LittleEndian
	errRecvAgain        = _ErrRecvAgain{}
	compressWritersPool = newless_pool.NewNewlessPool()
)

func init() {
	for i := 0; i < consts.COMPRESS_WRITER_POOL_SIZE; i++ {
		cw, err := flate.NewWriter(os.Stderr, flate.BestSpeed)
		if err != nil {
			gwlog.Fatal("create flate compressor failed: %v", err)
		}

		compressWritersPool.Put(cw)
	}

	gwlog.Info("%d compress writer created.", consts.COMPRESS_WRITER_POOL_SIZE)
}

type _ErrRecvAgain struct{}

func (err _ErrRecvAgain) Error() string {
	return "recv again"
}

func (err _ErrRecvAgain) Temporary() bool {
	return true
}

func (err _ErrRecvAgain) Timeout() bool {
	return false
}

type PacketConnection struct {
	conn               Connection
	compressed         bool
	pendingPackets     []*Packet
	pendingPacketsLock sync.Mutex
	sendBuffer         *SendBuffer // each PacketConnection uses 1 SendBuffer for sending packets

	// buffers and infos for receiving a packet
	payloadLenBuf         [SIZE_FIELD_SIZE]byte
	payloadLenBytesRecved int
	recvCompressed        bool
	recvTotalPayloadLen   uint32
	recvedPayloadLen      uint32
	recvingPacket         *Packet

	compressReader io.ReadCloser
}

func NewPacketConnection(conn Connection, compressed bool) *PacketConnection {
	pc := &PacketConnection{
		conn:       (conn),
		sendBuffer: NewSendBuffer(),
		compressed: compressed,
	}

	pc.compressReader = flate.NewReader(os.Stdin) // reader is always needed
	return pc
}

func (pc *PacketConnection) NewPacket() *Packet {
	return allocPacket()
}

func (pc *PacketConnection) SendPacket(packet *Packet) error {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s SEND PACKET %p: msgtype=%v, payload(%d)=%v", pc, packet,
			PACKET_ENDIAN.Uint16(packet.bytes[PREPAYLOAD_SIZE:PREPAYLOAD_SIZE+2]),
			packet.GetPayloadLen(),
			packet.bytes[PREPAYLOAD_SIZE+2:PREPAYLOAD_SIZE+packet.GetPayloadLen()])
	}
	if atomic.LoadInt64(&packet.refcount) <= 0 {
		gwlog.Panicf("sending packet with refcount=%d", packet.refcount)
	}

	packet.AddRefCount(1)
	pc.pendingPacketsLock.Lock()
	pc.pendingPackets = append(pc.pendingPackets, packet)
	pc.pendingPacketsLock.Unlock()
	return nil
}

func (pc *PacketConnection) Flush() (err error) {
	pc.pendingPacketsLock.Lock()
	if len(pc.pendingPackets) == 0 { // no packets to send, common to happen, so handle efficiently
		pc.pendingPacketsLock.Unlock()
		return
	}
	packets := make([]*Packet, 0, len(pc.pendingPackets))
	packets, pc.pendingPackets = pc.pendingPackets, packets
	pc.pendingPacketsLock.Unlock()

	// flush should only be called in one goroutine
	op := opmon.StartOperation("FlushPackets")
	defer op.Finish(time.Millisecond * 100)

	var cw *flate.Writer
	if pc.compressed {
		_cw := compressWritersPool.TryGet() // try to get a usable compress writer, might fail
		if _cw != nil {
			cw = _cw.(*flate.Writer)
			defer compressWritersPool.Put(cw)
		} else {
			gwlog.Warn("Fail to get compressor, packet is not compressed")
		}
	}

	if len(packets) == 1 {
		// only 1 packet to send, just send it directly, no need to use send buffer
		packet := packets[0]
		if cw != nil {
			packet.compress(cw)
		}
		err = WriteAll(pc.conn, packet.data())
		packet.Release()
		if err == nil {
			err = pc.conn.Flush()
		}
		return
	}

	sendBuffer := pc.sendBuffer // the send buffer

send_packets_loop:
	for _, packet := range packets {
		if cw != nil {
			packet.compress(cw)
		}

		packetData := packet.data()
		if len(packetData) > sendBuffer.FreeSpace() {
			// can not append data to send buffer, so clear send buffer first
			if err = sendBuffer.WriteTo(pc.conn); err != nil {
				return err
			}

			if len(packetData) >= SEND_BUFFER_SIZE {
				// packet is too large, impossible to put to send buffer
				err = WriteAll(pc.conn, packetData)
				packet.Release()

				if err != nil {
					return
				}
				continue send_packets_loop
			}
		}

		// now we are sure that len(packetData) <= sendBuffer.FreeSize()
		n, _ := sendBuffer.Write(packetData)
		if n != len(packetData) {
			gwlog.Panicf("packet is not fully written")
		}
		packet.Release()
	}

	// now we send all data in the send buffer
	err = sendBuffer.WriteTo(pc.conn)
	if err == nil {
		err = pc.conn.Flush()
	}
	return
}

func (pc *PacketConnection) SetRecvDeadline(deadline time.Time) error {
	return pc.conn.SetReadDeadline(deadline)
}

func (pc *PacketConnection) RecvPacket() (*Packet, error) {
	if pc.payloadLenBytesRecved < SIZE_FIELD_SIZE {
		// receive more of payload len bytes
		n, err := pc.conn.Read(pc.payloadLenBuf[pc.payloadLenBytesRecved:])
		pc.payloadLenBytesRecved += n
		if pc.payloadLenBytesRecved < SIZE_FIELD_SIZE {
			if err == nil {
				err = errRecvAgain
			}
			return nil, err // packet not finished yet
		}

		pc.recvTotalPayloadLen = NETWORK_ENDIAN.Uint32(pc.payloadLenBuf[:])
		//pc.recvCompressed = false
		if pc.recvCompressed {
			gwlog.Panicf("should be false")
		}
		if pc.recvTotalPayloadLen&COMPRESSED_BIT_MASK != 0 {
			pc.recvTotalPayloadLen &= PAYLOAD_LEN_MASK
			pc.recvCompressed = true
		}

		if pc.recvTotalPayloadLen == 0 || pc.recvTotalPayloadLen > MAX_PAYLOAD_LENGTH {
			err := errors.Errorf("invalid payload length: %v", pc.recvTotalPayloadLen)
			pc.resetRecvStates()
			pc.Close()
			return nil, err
		}

		pc.recvedPayloadLen = 0
		pc.recvingPacket = NewPacket()
		pc.recvingPacket.assureCapacity(pc.recvTotalPayloadLen)
	}

	// now all bytes of payload len is received, start receiving payload
	n, err := pc.conn.Read(pc.recvingPacket.bytes[PREPAYLOAD_SIZE+pc.recvedPayloadLen : PREPAYLOAD_SIZE+pc.recvTotalPayloadLen])
	pc.recvedPayloadLen += uint32(n)

	if pc.recvedPayloadLen == pc.recvTotalPayloadLen {
		// full packet received, return the packet
		packet := pc.recvingPacket
		packet.setPayloadLenCompressed(pc.recvTotalPayloadLen, pc.recvCompressed)
		pc.resetRecvStates()
		packet.decompress(pc.compressReader)

		return packet, nil
	}

	if err == nil {
		err = errRecvAgain
	}
	return nil, err
}
func (pc *PacketConnection) resetRecvStates() {
	pc.payloadLenBytesRecved = 0
	pc.recvTotalPayloadLen = 0
	pc.recvedPayloadLen = 0
	pc.recvingPacket = nil
	pc.recvCompressed = false
}

func (pc *PacketConnection) Close() error {
	return pc.conn.Close()
}

func (pc *PacketConnection) RemoteAddr() net.Addr {
	return pc.conn.RemoteAddr()
}

func (pc *PacketConnection) LocalAddr() net.Addr {
	return pc.conn.LocalAddr()
}

func (pc *PacketConnection) String() string {
	return fmt.Sprintf("[%s >>> %s]", pc.LocalAddr(), pc.RemoteAddr())
}
