package user

import (
	"net"
	"strconv"
	"testing"
)

func TestTCPServer1(t *testing.T) {
	port := 8080
	config := TCPServerConfig{Port: port, Backlog: 5, reuseAddr: true, reusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	buf := make([]byte, 1024)
	var offset uint64 = 0

	conn := server.Accept()

	client.Write([]byte("Hello"))
	read_len, err := conn.Read(buf, offset)
	if err != nil {
		t.Fatalf("Error in reading. %v", err)
	}
	offset += uint64(read_len)
	if offset != 5 {
		t.Fatalf("offset mismatch")
	}
	client.Write([]byte("World"))
	read_len, err = conn.Read(buf, offset)
	if err != nil {
		t.Fatalf("Error in reading. %v", err)
	}
	offset += uint64(read_len)
	if offset != 10 {
		t.Fatalf("offset mismatch")
	}
	t.Logf("\"%v\" %v", string(buf), offset)

	client.Close()

	data := "Hii."
	write_len, _ := conn.Write([]byte(data), uint64(len(data)))
	t.Logf("WriteLen: %d", write_len)

	conn.Close()
	server.Stop()
}

func TestTCPServer2(t *testing.T) {
	port := 8080
	config := TCPServerConfig{Port: port, Backlog: 5, reuseAddr: true, reusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	buf := make([]byte, 1024)

	conn := server.Accept()

	msg1 := []byte("Hello")
	msg2 := []byte("Hii")
	msg3 := []byte("World")
	bufs := [][]byte{msg2, msg1, msg3}
	conn.Writev(bufs)
	conn.Close()

	read_len, err := client.Read(buf)
	if err != nil {
		t.Fatalf("Error in reading. %v", err)
	}
	read_str := buf[:read_len]
	client.Close()
	t.Logf("Read: %s", read_str)

	server.Stop()
}
