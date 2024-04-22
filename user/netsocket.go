// Go Socket library
package user

import (
	"log"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type Address struct {
	Ip   string
	Port uint16
}

type ConnectionConfig struct {
}

type Connection struct {
	Closed bool
	Sock   int
}

// Reads from socket.
func (conn *Connection) Read(buf []byte, offset uint64) (int, error) {
	log.Printf("Waiting for read on sock %d %v\n", conn.Sock, offset)
	n, err := unix.Read(conn.Sock, buf)
	if err != nil {
		log.Printf("Read error. %v", err)
	}
	log.Printf("Read %v %v\n", n, conn.Sock)
	return n, err
}

// Writes to socket.
func (conn *Connection) Write(buf []byte, size uint64) (int, error) {
	n, err := unix.Write(conn.Sock, buf)
	if err != nil {
		errno := err.(syscall.Errno)
		log.Printf("Write error. Code: %d %v %d", errno, err, n)
	}
	return n, err
}

// Writes to socket.
func (conn *Connection) Writev(bufs [][]byte) (int, error) {
	n, err := unix.Writev(conn.Sock, bufs)
	if err != nil {
		errno := err.(syscall.Errno)
		log.Printf("Writev error. Code: %d %v %d", errno, err, n)
	}
	return n, err
}

// Writes to socket. Raw syscall
func (conn *Connection) WritevRaw(iovs []syscall.Iovec) (int, error) {
	n, _, err := syscall.Syscall(syscall.SYS_WRITEV, uintptr(conn.Sock), uintptr(unsafe.Pointer(&iovs[0])), uintptr(len(iovs)))
	if err != 0 {
		return -1, err
	}
	return int(n), nil
}

// Close the connection
func (conn *Connection) Close() {
	if conn.Closed {
		return
	}
	if err := unix.Close(conn.Sock); err != nil {
		errno := err.(syscall.Errno)
		log.Printf("Close errror. No: %v %v", errno, err.Error())
	}
	conn.Closed = true
}

type TCPConnection struct {
	Conn          *Connection
	ClientAddress *Address
	Server        *TCPServer
}

func (tcpConn *TCPConnection) Read(buf []byte, offset uint64) (int, error) {
	n, err := tcpConn.Conn.Read(buf, offset)
	if err != nil {
		tcpConn.Server.Stats.TotalBytesRead += uint64(n)
	}
	return n, err
}

func (tcpConn *TCPConnection) Write(buf []byte, size uint64) (int, error) {
	n, err := tcpConn.Conn.Write(buf, size)
	if err != nil {
		tcpConn.Server.Stats.TotalBytesWritten += uint64(n)
	}
	return n, err
}

func (tcpConn *TCPConnection) Writev(bufs [][]byte) (int, error) {
	n, err := tcpConn.Conn.Writev(bufs)
	if err != nil {
		tcpConn.Server.Stats.TotalBytesWritten += uint64(n)
	}
	return n, err
}

func (tcpConn *TCPConnection) WritevRaw(iovs []syscall.Iovec) (int, error) {
	n, err := tcpConn.Conn.WritevRaw(iovs)
	if err != nil {
		tcpConn.Server.Stats.TotalBytesWritten += uint64(n)
	}
	return n, err
}

func (tcpConn *TCPConnection) Close() {
	tcpConn.Conn.Close()
	tcpConn.Server.Stats.OpennedConn--
}

type TCPClientConfig struct {
	ServerAddress Address
	ClientAddress Address
}
