package tlstest

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
)

var (
	wait         sync.WaitGroup
	waitServerOk sync.WaitGroup
)

func TestTLS(t *testing.T) {
	wait.Add(2)
	waitServerOk.Add(1)
	go serverRoutine(t)
	waitServerOk.Wait()
	go clientRoutine(t)
	wait.Wait()
}

func serverRoutine(t *testing.T) {
	defer wait.Done()

	cert, err := tls.LoadX509KeyPair("../../rsa.crt", "../../rsa.key")
	checkError(err)
	t.Logf("server: certificate loaded")

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	ln, err := net.Listen("tcp", "127.0.0.1:10443")
	t.Log("server: listenning on :10443")
	checkError(err)
	defer ln.Close()

	waitServerOk.Done()

	conn, err := ln.Accept()
	checkError(err)

	go func() {
		defer conn.Close()
		conn = net.Conn(tls.Server(conn, tlsConfig))
		r := bufio.NewReader(conn)
		for {
			msg, err := r.ReadString('\n')
			if err == io.EOF {
				break
			}
			checkError(err)
			t.Logf("server: recv '%s'", msg)
			_, err = conn.Write([]byte("world\n"))
			checkError(err)
		}
	}()

}

func clientRoutine(t *testing.T) {
	defer wait.Done()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := net.Dial("tcp", "127.0.0.1:10443")
	checkError(err)

	defer conn.Close()
	conn = net.Conn(tls.Client(conn, tlsConfig))

	for i := 0; i < 10; i++ {
		n, err := conn.Write([]byte(fmt.Sprintf("hello %d\n", i)))
		checkError(err)
		buf := make([]byte, 100)
		n, err = conn.Read(buf)
		checkError(err)
		t.Logf("client: recv '%s'", string(buf[:n]))
	}
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
