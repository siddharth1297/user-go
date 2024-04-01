package user

import (
	"context"
	"fmt"
)

type callbackfunc_t func(*HttpHeader, *HttpResponseWriter)

type HttpServer struct {
	tcpServer   *TCPServer
	handles     map[string]callbackfunc_t
	ep_instance *EpollInstance
}

type HttpConfig struct {
	TcpConfig *TCPServerConfig
}

type HttpHeader struct {
}

type HttpResponseWriter struct {
	conn *TCPConnection
}

func (resp *HttpResponseWriter) Write(msg *string) {
	resp.conn.Write([]byte(*msg), uint64(len(*msg)))
}

// Similar to NewServeMux
func NewHttpServer(httpConfig *HttpConfig) *HttpServer {
	ep_instance := NewEpollInstance(DEFAULT_TIMEOUT, DEFAULT_MAX_EVENTS)
	tcpServer := CreateServerTCP(httpConfig.TcpConfig)
	return &HttpServer{tcpServer: tcpServer, handles: make(map[string]callbackfunc_t), ep_instance: ep_instance}
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
			fmt.Println("cccccccccccccc")
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
