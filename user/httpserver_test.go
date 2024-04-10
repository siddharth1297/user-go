package user

import (
	"context"
	"log"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestHTTPServer1(t *testing.T) {
	port := 8080
	tcpConfig := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: false, NonBlock: true}
	server := NewHttpServer(&HttpConfig{TcpConfig: &tcpConfig})
	msg_send := "Hello"
	server.HandleFunc("/", func(header *HttpRequest, response *HttpResponseWriter) {
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
	tcpConfig := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: false, NonBlock: true}
	server := NewHttpServer(&HttpConfig{TcpConfig: &tcpConfig})
	msg_send := "Hello"
	server.HandleFunc("/", func(header *HttpRequest, response *HttpResponseWriter) {
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
	tcpConfig := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: false, NonBlock: true}
	server := NewHttpServer(&HttpConfig{TcpConfig: &tcpConfig})
	//msg_send := "Hello"
	server.HandleFunc("/", func(request *HttpRequest, response *HttpResponseWriter) {
		//data := msg_send
		//response.Write(&data)
		log.Println(request)
		log.Printf("%v \n", response)
		payload := "Working"
		resp := "HTTP/1.1 200 OK" + string(CRLF) +
			"Content-Length: " + strconv.Itoa(len(payload)) + string(CRLF) +
			"Connection: Closed" + string(CRLF) + string(CRLF) +
			payload
		response.Write(&resp)
		//response.conn.Conn.Write([]byte(resp), uint64(len(resp)))
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
