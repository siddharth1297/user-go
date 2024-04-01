// inspired from https://gist.github.com/Lisprez/7b52f4a55cd0fcf96324b5f02b865e54
package user

/*
import (
	"log"
	//"os"
	"syscall"
)

const DEFAULT_TIMEOUT = 0

// Only accept
const DEFAULT_SERVER_EVENTS uint32 = syscall.EPOLLIN  | syscall.EPOLLOUT | syscall.EPOLLRDHUP |syscall.EPOLLERR | syscall.EPOLLET
const DEFAULT_CONN_READ_EVENTS uint32 = 0
const DEFAULT_CONN_WRITE_EVENTS uint32 = 0
const DEFAULT_CONN_READ_WRITE_EVENTS uint32 = 0

type EpollInstance struct {
	timeout     int
	maxevents   int
	epfd        int // epoll fd
	epoll_event syscall.EpollEvent
	event_list  []syscall.EpollEvent
}

func CreateEpollInstance(timeout int, maxevents int) *EpollInstance {
	epfd, err := syscall.EpollCreate1(DEFAULT_TIMEOUT)
	if err != nil {
		log.Fatalf("error in epoll_create1: %s", err.Error())
	}
	return &EpollInstance{timeout: timeout, maxevents: maxevents, epfd: epfd, event_list: make([]syscall.EpollEvent, maxevents)}
}

func (ep_instance *EpollInstance) AddConnection(fd int, events uint32) {
	ep_instance.epoll_event.Events = events
	ep_instance.epoll_event.Fd = int32(fd)
	if err := syscall.EpollCtl(ep_instance.epfd, syscall.EPOLL_CTL_ADD, fd, &ep_instance.epoll_event); err != nil {
		log.Fatalf("epoll_ctl_add: %s", err.Error())
	}
}

func (ep_instance *EpollInstance) RemoveConnection(fd int) {
	if err := syscall.EpollCtl(ep_instance.epfd, syscall.EPOLL_CTL_DEL, fd, nil); err != nil {
		log.Fatalf("epoll_ctl_del: %s", err.Error())
	}
}

func (epollServer *EpollInstance) Close() {
	syscall.Close(epollServer.epfd)
}

func (epollServer *EpollInstance) CollectEvents() {

	
}
*/