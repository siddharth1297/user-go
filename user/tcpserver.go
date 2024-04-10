package user

import (
	"log"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

const DEFAULT_BACKLOG = 1024

type TCPServerConfig struct {
	AcceptFrom string // Default is "0.0.0.0". For specific address, set it
	Port       int    // Server port
	Device     string
	Backlog    int // == -1, 0. default value will be set
	ReuseAddr  bool
	ReusePort  bool
	NonBlock   bool // For Epoll
}

type ServerStats struct {
	TotalConn         uint64
	OpennedConn       uint64
	TotalBytesRead    uint64
	TotalBytesWritten uint64
}

type TCPServer struct {
	Config *TCPServerConfig
	sock   int
	Active bool
	Stats  *ServerStats
}

func CreateServerTCP(config *TCPServerConfig) *TCPServer {
	socketflags := 0
	if config.NonBlock {
		socketflags |= unix.O_NONBLOCK
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|socketflags, unix.IPPROTO_TCP)
	if err != nil {
		log.Fatalf("Unable to create socket. error: %v", err.Error())
	}
	reuse := 0
	if config.ReuseAddr {
		reuse |= unix.SO_REUSEADDR
	}
	if config.ReusePort {
		reuse |= unix.SO_REUSEPORT
	}
	if reuse != 0 {
		if err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, reuse, 1); err != nil {
			unix.Close(fd)
			log.Fatalf("reuse fail. %v", err)
		}
	}

	if config.AcceptFrom == "" {
		config.AcceptFrom = "0.0.0.0"
	}
	addr := syscall.SockaddrInet4{
		Port: config.Port,
		//Addr: [4]byte{0, 0, 0, 0},
	}
	copy(addr.Addr[:], net.ParseIP(config.AcceptFrom).To4())

	if err = syscall.Bind(fd, &addr); err != nil {
		unix.Close(fd)
		log.Fatalf("Unable to bind to address. error: %v", err)
	}

	if config.Device != "" {
		if err := unix.BindToDevice(fd, config.Device); err != nil {
			log.Fatalf("Unable to bind to device. error: %v", err.Error())
		}
	}

	if config.Backlog <= 0 {
		// TODO: Read the maximum possible value
		config.Backlog = DEFAULT_BACKLOG
	}
	return &TCPServer{Config: config, sock: fd, Active: true, Stats: &ServerStats{}}
}

func (server *TCPServer) StartListen() {
	if err := unix.Listen(server.sock, server.Config.Backlog); err != nil {
		log.Fatalf("Unable to Listen. error: %v", err.Error())
	}
	server.Active = true
}

func (server *TCPServer) Accept() *TCPConnection {

	fd, _, err := unix.Accept(server.sock)
	if err != nil {
		log.Fatalf("accept error. %v", err)
	}
	server.Stats.OpennedConn++
	server.Stats.TotalConn++
	// TODO: Set client address
	return &TCPConnection{Conn: &Connection{Closed: false, Sock: fd}, Server: server}
}

func (server *TCPServer) Stop() {
	unix.Close(server.sock)
}
