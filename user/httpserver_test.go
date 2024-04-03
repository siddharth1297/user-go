package user

import (
	"bytes"
	"context"
	"log"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestHTTPServer1(t *testing.T) {
	port := 8080
	tcpConfig := TCPServerConfig{Port: port, Backlog: 5, reuseAddr: true, reusePort: false, NonBlock: true}
	server := NewHttpServer(&HttpConfig{TcpConfig: &tcpConfig})
	msg_send := "Hello"
	server.HandleFunc("/", func(header *HttpHeader, response *HttpResponseWriter) {
		data := msg_send
		response.Write(&data)
	})
	ctx, cancelFunc := context.WithCancel(context.Background())

	go server.ListenAndServe(ctx)
	defer cancelFunc()
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
	tcpConfig := TCPServerConfig{Port: port, Backlog: 5, reuseAddr: true, reusePort: false, NonBlock: true}
	server := NewHttpServer(&HttpConfig{TcpConfig: &tcpConfig})
	msg_send := "Hello"
	server.HandleFunc("/", func(header *HttpHeader, response *HttpResponseWriter) {
		data := msg_send
		response.Write(&data)
	})
	ctx, cancelFunc := context.WithCancel(context.Background())

	go server.ListenAndServe(ctx)
	defer cancelFunc()
	time.Sleep(time.Second)

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))
	if err != nil {
		t.Fatalf("server not running: %v", err)
	}

	t.Logf("Client. Writing")
	//w_len, err := client.Write([]byte("Hii from dial client"))
	buf := make([]byte, 1e9)
	w_len, err := client.Write(buf)
	t.Logf("Client WriteDone. %d %v", w_len, err)
	/*
		resp_read := make([]byte, 1024)
		_, err = client.Read(resp_read)
		if err != nil {
			t.Logf("Error while reading. Err: %v", err.Error())
		}
		client.Close()
	*/
	/*
		msg := string(resp_read[:n])
		if msg != msg_send {
			t.Logf("msg mismatch. Read: \"%v\", n: %v \"%v\" %v", msg, n, msg_send, len(msg_send))
		}
	*/
	time.Sleep(time.Second * 2)
}

func TestHTTPServer3(t *testing.T) {
	port := 8080
	tcpConfig := TCPServerConfig{Port: port, Backlog: 5, reuseAddr: true, reusePort: false, NonBlock: true}
	server := NewHttpServer(&HttpConfig{TcpConfig: &tcpConfig})
	msg_send := "Hello"
	server.HandleFunc("/", func(header *HttpHeader, response *HttpResponseWriter) {
		data := msg_send
		response.Write(&data)
	})
	ctx, cancelFunc := context.WithCancel(context.Background())

	go server.ListenAndServe(ctx)
	defer cancelFunc()
	time.Sleep(time.Second)

	//client, err := net.Dial("http", "127.0.0.1"+":"+strconv.Itoa(port))
	//if err != nil {
	//	t.Fatalf("server not running: %v", err)
	//}

	//t.Logf("Client. Writing")
	//w_len, err := client.Write([]byte("Hii from dial client"))
	//buf := make([]byte, DEFA)
	//w_len, err := client.Write(buf)
	//t.Logf("Client WriteDone. %d %v", w_len, err)
	/*
		resp_read := make([]byte, 1024)
		_, err = client.Read(resp_read)
		if err != nil {
			t.Logf("Error while reading. Err: %v", err.Error())
		}
		client.Close()
	*/
	/*
		msg := string(resp_read[:n])
		if msg != msg_send {
			t.Logf("msg mismatch. Read: \"%v\", n: %v \"%v\" %v", msg, n, msg_send, len(msg_send))
		}
	*/

	time.Sleep(time.Second * 50)
}

func parseHeader(buf []byte, buf_len int) {
	header_end_index := bytes.Index(buf, []byte("\r\n"))
	if header_end_index != -1 {
		log.Fatalf("Header end not found. Len: %v buf %v", buf_len, string(buf))
	}

}
