package user

import (
	"bytes"
	"context"
	"fmt"
	"log"
)

const (
	DEFAULT_HTTP_REQ_HEADER_SIZE = (1 << 10) // (1KB)
	DEFAULT_HTTP_REQ_BODY_SIZE   = (1 << 20) // (1MB)
)

type callbackfunc_t func(*HttpHeader, *HttpResponseWriter)

type HttpServer struct {
	tcpServer           *TCPServer
	handles             map[string]callbackfunc_t
	ep_instance         *EpollInstance
	activeConnectionMap map[int]*TCPConnection
}

type HttpConfig struct {
	TcpConfig *TCPServerConfig
}

type HttpHeader struct {
	Headers map[string]string
	Body    []byte
}

type HttpResponseWriter struct {
	conn *TCPConnection
}

func (resp *HttpResponseWriter) Write(msg *string) {
	resp.conn.Write([]byte(*msg), uint64(len(*msg)))
}

// Similar to NewServeMux
func NewHttpServer(httpConfig *HttpConfig) *HttpServer {
	server := &HttpServer{handles: make(map[string]callbackfunc_t), activeConnectionMap: make(map[int]*TCPConnection)}
	server.ep_instance = NewEpollInstance(DEFAULT_TIMEOUT, DEFAULT_MAX_EVENTS, server.onEpollReadEvent, server.onEpollWriteEvent)
	server.tcpServer = CreateServerTCP(httpConfig.TcpConfig)
	return server
}

func (server *HttpServer) HandleFunc(path string, callbackFunc callbackfunc_t) {
	server.handles[path] = callbackFunc
}

/*
func (server *HttpServer) ListenAndServe(ctx context.Context) {
	server.tcpServer.StartListen()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("cccccccccccccc")
			break
		default:
			conn := server.tcpServer.Accept()
			path := "/"
			respWriter := &HttpResponseWriter{conn: conn}
			reqHeader := &HttpHeader{}
			server.handles[path](reqHeader, respWriter)
		}
	}
	// Close all the active requests
	fmt.Println("Wxisted")
}
*/

func (server *HttpServer) ListenAndServe(ctx context.Context) {
	server.ep_instance.AddConnection(server.tcpServer.sock, DEFAULT_SERVER_EVENTS)
	server.tcpServer.StartListen()
	for {
		select {
		case <-ctx.Done():
			//fmt.Println("cccccccccccccc")
			break
		default:
			server.ep_instance.CollectEvents()
		}
	}
	// Close all the active requests
	fmt.Println("Wxisted")
}

func (server *HttpServer) waitForEpollEvents() {

}

func (server *HttpServer) onEpollReadEvent(sock int) {
	log.Printf("Read Event %v", sock)
	if sock == server.tcpServer.sock {
		tcp_conn := server.tcpServer.Accept()
		server.ep_instance.AddConnection(tcp_conn.Conn.Sock, DEFAULT_CONN_READ_EVENTS)
		server.activeConnectionMap[tcp_conn.Conn.Sock] = tcp_conn
		return
	}
	conn := server.activeConnectionMap[sock]
	//data_len := int(1e9) //2048
	data_len := DEFAULT_HTTP_REQ_BODY_SIZE
	buf := make([]byte, data_len)

	log.Printf("Waiting for read. bufLen: %v\n", data_len)
	/*for i := 0; i < 20; i++ {
		buf[i] = 'c'
	}*/
	//fmt.Println(string(buf))
	size, err := conn.Read(buf, uint64(data_len))

	/*
		fmt.Printf("Conn: %v\n", conn.Conn)
		size, err := unix.Read(conn.Conn.Sock, buf)
	*/
	fmt.Printf("Read: %v %s %v %v", size, string(buf[:size]), err, len(buf))
	server.parseHeader(buf, size)
}

func (server *HttpServer) onEpollWriteEvent(sock int) {

}

func (server *HttpServer) onEpollCloseEvent(sock int) {

}

func (server *HttpServer) parseHeader(buf []byte, buf_len int) {
	break_line := []byte("\r\n")
	header_end_index := bytes.Index(buf, break_line)
	if header_end_index != -1 {
		log.Fatalf("Header end not found. Len: %v buf %v", buf_len, string(buf))
	}
	startindex := 0
	index := bytes.Index(buf[startindex:], break_line)
	log.Println("RequestLine: ", string(buf[startindex:index]))
}
