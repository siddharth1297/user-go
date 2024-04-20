package user

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	// "path"
)

const (
	DEFAULT_HTTP_REQ_HEADER_SIZE = (1 << 10) // (1KB)
	DEFAULT_HTTP_REQ_BODY_SIZE   = (1 << 20) // (1MB)
)

var (
	CRLF     = []byte("\r\n")
	CRLFCRLF = []byte("\r\n\r\n")
)

type callbackfunc_t func(*HttpRequest, *HttpResponseWriter)

type HttpServer struct {
	tcpServer           *TCPServer
	Handles             map[string]callbackfunc_t
	ep_instance         *EpollInstance
	activeConnectionMap map[int]*TCPConnection
}

type HttpConfig struct {
	TcpConfig *TCPServerConfig
}

type HttpRequest struct {
	Headers      map[string]string
	Body         []byte
	Method       string
	AbsolutePath string
	Queries      url.Values
	File         string
	Version      string
	Payload      []byte
	PayloadStart int

	Path string
}

type HttpResponseWriter struct {
	conn   *TCPConnection
	server *HttpServer
}

func (resp *HttpResponseWriter) Write(msg *string) {
	resp.conn.Write([]byte(*msg), uint64(len(*msg)))
	resp.server.ep_instance.RemoveConnection(resp.conn.Conn.Sock)
	delete(resp.server.activeConnectionMap, resp.conn.Conn.Sock)
	resp.conn.Close()
}

// Similar to NewServeMux
func NewHttpServer(httpConfig *HttpConfig) *HttpServer {
	server := &HttpServer{Handles: make(map[string]callbackfunc_t), activeConnectionMap: make(map[int]*TCPConnection)}
	server.ep_instance = NewEpollInstance(DEFAULT_TIMEOUT, DEFAULT_MAX_EVENTS, server.onEpollReadEvent, server.onEpollWriteEvent)
	server.tcpServer = CreateServerTCP(httpConfig.TcpConfig)
	return server
}

func (server *HttpServer) HandleFunc(path string, callbackFunc callbackfunc_t) {
	server.Handles[path] = callbackFunc
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
			server.Handles[path](reqHeader, respWriter)
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
	size, _ := conn.Read(buf, uint64(data_len))

	/*
		fmt.Printf("Conn: %v\n", conn.Conn)
		size, err := unix.Read(conn.Conn.Sock, buf)
	*/
	//fmt.Printf("Read: %v %s %v %v\n", size, string(buf[:size]), err, len(buf))
	//fmt.Printf("%v\n", string(buf[:size]))
	req := server.parseHeader(buf, size)
	respWriter := &HttpResponseWriter{conn: conn, server: server}
	fmt.Printf("%v Body: \"%v\"\n", size, string(buf[:size]))
	fmt.Printf("%v\n", req)
	server.Handles[req.AbsolutePath](req, respWriter)
	log.Printf("----CallbackDone---\n")
	// TODO: Shrink buf
	// Start from small size then grow.
	// If default size is biggrt, then also shrink it
}

func (server *HttpServer) onEpollWriteEvent(sock int) {

}

func (server *HttpServer) onEpollCloseEvent(sock int) {

}

func (server *HttpServer) parseHeader(buf []byte, buf_len int) *HttpRequest {
	req := &HttpRequest{}

	header_end_index := bytes.Index(buf[:buf_len], CRLFCRLF)
	if header_end_index == -1 {
		log.Fatalf("Header end not found. Len: %v buf \n\"%v\"", buf_len, (buf[:buf_len]))
	}
	startindex := 0
	index := bytes.Index(buf[startindex:], CRLF)
	req.Method, req.Path, req.Version = parseRequestLine(buf, startindex, index)
	path_parser, err := url.Parse(req.Path)
	if err != nil {
		panic(err)
	}
	req.AbsolutePath = path_parser.Path
	// req.AbsolutePath, req.File = path.Split(req.AbsolutePath)
	req.Queries, err = url.ParseQuery(path_parser.RawQuery)
	if err != nil {
		log.Fatalf("Error in query string. path: \"%v\" err: %v", req.Path, err.Error())
	}
	// TODO: Parse headers
	if req.Method == "POST" {
		req.Payload = buf[header_end_index+4 : buf_len]
		req.PayloadStart = header_end_index + 4
	}
	return req
}

func parseRequestLine(buf []byte, start int, end int) (string, string, string) {
	fs := bytes.Index(buf[start:], []byte(" "))
	method := string(buf[start:fs])
	ss := bytes.Index(buf[fs+1:], []byte(" "))
	path := string(buf[fs+1 : fs+1+ss])
	version := "HTTP/1.1"
	return method, path, version
}
