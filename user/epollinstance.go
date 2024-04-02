// inspired from https://gist.github.com/Lisprez/7b52f4a55cd0fcf96324b5f02b865e54
package user

import (
	"log"

	"golang.org/x/sys/unix"
)

const DEFAULT_TIMEOUT = 0
const DEFAULT_MAX_EVENTS = 1024

// Only accept
const DEFAULT_SERVER_EVENTS uint32 = unix.EPOLLIN | unix.EPOLLRDHUP | unix.EPOLLERR | unix.EPOLLET
const DEFAULT_CONN_READ_EVENTS uint32 = unix.EPOLLIN | unix.EPOLLRDHUP | unix.EPOLLERR | unix.EPOLLET
const DEFAULT_CONN_WRITE_EVENTS uint32 = 0
const DEFAULT_CONN_READ_WRITE_EVENTS uint32 = 0

type epollcallbackfunc_t func(int)

type EpollInstance struct {
	timeout           int
	maxevents         int
	epfd              int // epoll fd
	epoll_event       unix.EpollEvent
	event_list        []unix.EpollEvent
	readCallBackFunc  epollcallbackfunc_t
	writeCallBackFunc epollcallbackfunc_t
}

func NewEpollInstance(timeout int, maxevents int, readcallbackfunc epollcallbackfunc_t, writecallbackfunc epollcallbackfunc_t) *EpollInstance {
	epfd, err := unix.EpollCreate1(0)
	if err != nil {
		log.Fatalf("error in epoll_create1: %s", err.Error())
	}
	return &EpollInstance{timeout: timeout, maxevents: maxevents, epfd: epfd, event_list: make([]unix.EpollEvent, maxevents), readCallBackFunc: readcallbackfunc, writeCallBackFunc: writecallbackfunc}
}

func (ep_instance *EpollInstance) AddConnection(fd int, events uint32) {
	ep_instance.epoll_event.Events = events
	ep_instance.epoll_event.Fd = int32(fd)
	if err := unix.EpollCtl(ep_instance.epfd, unix.EPOLL_CTL_ADD, fd, &ep_instance.epoll_event); err != nil {
		log.Fatalf("epoll_ctl_add: %s", err.Error())
	}
	log.Printf("Added socket %d to epoll List", fd)
}

func (ep_instance *EpollInstance) RemoveConnection(fd int) {
	if err := unix.EpollCtl(ep_instance.epfd, unix.EPOLL_CTL_DEL, fd, nil); err != nil {
		log.Fatalf("epoll_ctl_del: %s", err.Error())
	}
}

func (epollServer *EpollInstance) Close() {
	unix.Close(epollServer.epfd)
}

func (epollServer *EpollInstance) CollectEvents() {
	read_fds, err := unix.EpollWait(epollServer.epfd, epollServer.event_list, epollServer.timeout)

	if err != nil {
		log.Fatalf("epoll_wait error. %v %v", unix.ErrnoName(err.(unix.Errno)), err.Error())
	}
	if DEFAULT_TIMEOUT == 0 && read_fds == 0 {
		return
	}
	log.Printf("ReadySocks: %d", read_fds)
	for i := 0; i < read_fds; i++ {
		if (epollServer.event_list[i].Events & unix.EPOLLRDHUP) > 0 {
			// Closed
			log.Printf("EpollEvt %v EPOLLRDHUP", epollServer.event_list[i].Fd)
		}
		if (epollServer.event_list[i].Events & unix.EPOLLIN) > 0 {
			// Read
			epollServer.readCallBackFunc(int(epollServer.event_list[i].Fd))
		}
		if (epollServer.event_list[i].Events & unix.EPOLLOUT) > 0 {
			// Write
			log.Printf("EpollEvt %v EPOLLOUT", epollServer.event_list[i].Fd)
		}
	}
}
