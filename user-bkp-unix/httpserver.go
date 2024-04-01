package user

type callbackfunc_t func(*HttpHeader, *HttpResponseWriter)

type HttpServer struct {
	tcpServer *TCPServer
	handles   map[string]callbackfunc_t
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
	tcpServer := CreateServerTCP(httpConfig.TcpConfig)
	return &HttpServer{tcpServer: tcpServer, handles: make(map[string]callbackfunc_t)}
}

func (server *HttpServer) HandleFunc(path string, callbackFunc callbackfunc_t) {
	server.handles[path] = callbackFunc
}

func (server *HttpServer) ListenAndServe() {
	server.tcpServer.StartListen()
	for {
		conn := server.tcpServer.Accept()
		path := "/"
		respWriter := &HttpResponseWriter{conn: conn}
		reqHeader := &HttpHeader{}
		server.handles[path](reqHeader, respWriter)
	}

}
