package user

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"
	"syscall"
	"testing"
	"unsafe"
)

func TestTCPServer1(t *testing.T) {
	port := 8080
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
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
	port := 8087
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	buf := make([]byte, 1024)

	conn := server.Accept()

	/*
		msg1 := []byte("Hello")
		msg2 := []byte("Hii")
		msg3 := []byte("World")
		bufs := [][]byte{msg2, msg1, msg3}
		conn.Writev(bufs)
	*/

	//msg2 := "World   "
	msg1 := "Hello"

	msg3 := " -- "

	a := uint32(302120960)
	a_in_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(a_in_bytes, a)
	log.Println("BYTES:: ", a_in_bytes)
	//a_bytes := (*[]byte)(unsafe.Pointer(&a))
	a_bytes := (*byte)(unsafe.Pointer(&a))

	iovs := []syscall.Iovec{
		{Base: (*byte)(unsafe.Pointer(&[]byte(msg1)[0])), Len: uint64(len(msg1))},
		{Base: (*byte)(unsafe.Pointer(&[]byte(msg3)[0])), Len: uint64(len(msg3))},
		{Base: (*byte)(unsafe.Pointer(&a)), Len: uint64(4)},
		//{Base: (*byte)(unsafe.Pointer(&[]byte(msg2)[0])), Len: uint64(len(msg2))},
	}
	size, _ := conn.Conn.WritevRaw(iovs)
	log.Printf("Wrote %v bytes\n", size)
	fmt.Println("BYTESDIGIT:: ", a_bytes)
	conn.Close()

	read_len, err := client.Read(buf)
	if err != nil {
		t.Fatalf("Error in reading. %v", err)
	}
	read_str := buf[:read_len]
	client.Close()
	t.Logf("Read: %s", read_str)

	var result int
	for _, b := range read_str[9:13] {
		result = (result << 8) | int(b)
	}
	//return result
	//digit := int(read_str[9:13])
	t.Logf("DIGIT: %v", result)
	log.Println(read_str[9:13])

	log.Println(read_str[9:13])
	log.Println(bytesToUint32LittleEndian(read_str[9:13]))
	server.Stop()
}

func bytesToUint32LittleEndian(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}
