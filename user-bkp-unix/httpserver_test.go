package user

import (
	"net"
	"strconv"
	"testing"
	"time"
)

func TestHTTPServer1(t *testing.T) {
	port := 8080
	tcpConfig := TCPServerConfig{Port: port, Backlog: 5, reuseAddr: true, reusePort: false}
	server := NewHttpServer(&HttpConfig{TcpConfig: &tcpConfig})
	msg_send := "Hello"
	server.HandleFunc("/", func(header *HttpHeader, response *HttpResponseWriter) {
		data := msg_send
		response.Write(&data)
	})
	go server.ListenAndServe()
	time.Sleep(time.Second)
	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))
	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	resp_read := make([]byte, 1024)
	n, err := client.Read(resp_read)
	if err != nil {
		t.Logf("Error while reading. Err: %v", err.Error())
	}
	client.Close()
	msg := string(resp_read[:n])
	if msg != msg_send {
		t.Logf("msg mismatch. Read: \"%v\", n: %v \"%v\" %v", msg, n, msg_send, len(msg_send))
	}
}

func TestHTTPServer2(t *testing.T) {
	port := 8080
	tcpConfig := TCPServerConfig{Port: port, Backlog: 5, reuseAddr: true, reusePort: false}
	server := NewHttpServer(&HttpConfig{TcpConfig: &tcpConfig})
	msg1 := "Hello"
	msg2 := "World"
	msg3 := "Hii"
	msg_send := [msg3, msg1, msg2]
	server.HandleFunc("/", func(header *HttpHeader, response *HttpResponseWriter) {
		data := msg_send
		response.Writev(&data)
	})
	go server.ListenAndServe()
	time.Sleep(time.Second)
	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))
	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	resp_read := make([]byte, 1024)
	n, err := client.Read(resp_read)
	if err != nil {
		t.Logf("Error while reading. Err: %v", err.Error())
	}
	client.Close()
	msg := string(resp_read[:n])
	if msg != msg_send {
		t.Logf("msg mismatch. Read: \"%v\", n: %v \"%v\" %v", msg, n, msg_send, len(msg_send))
	}
}