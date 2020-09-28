// +build darwin

package main

import (
	"errors"
	"log"
	"net"
	"sync"
	"syscall"
)


type epoll struct {
	fd          int
	connections map[int]net.Conn
	lock        *sync.RWMutex
}

func MkEpoll() (*epoll, error) {
	fd, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}
	kevent := syscall.Kevent_t{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}
	if _, err = syscall.Kevent(fd, []syscall.Kevent_t{kevent}, nil, nil); err != nil {
		return nil, err
	}
	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: map[int]net.Conn{},
	}, nil
}

func (e *epoll) Add(conn net.Conn) error {
	fd := websocketFD(conn)
	if fd < 0 {
		return errors.New("invalid conn")
	}
	newE := syscall.Kevent_t{
		Ident:  uint64(fd),
		Flags:  syscall.EV_ADD,
		Filter: syscall.EVFILT_READ,
	}

	_, err := syscall.Kevent(e.fd, []syscall.Kevent_t{
		newE,
	}, nil, nil)
	if err != nil {
		return err
	}
	e.connections[fd] = conn
	return nil
}

func (e *epoll) Remove(fd int) error {
	newE := syscall.Kevent_t{
		Ident:  uint64(fd),
		Flags:  syscall.EV_DELETE,
		Filter: syscall.EVFILT_READ,
	}
	_, _ = syscall.Kevent(e.fd, []syscall.Kevent_t{
		newE,
	}, nil, nil)
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.connections, fd)
	if len(e.connections)%100 == 0 {
		log.Printf("Total number of connections: %v", len(e.connections))
	}
	_ = syscall.Close(fd)
	return nil
}

func (e *epoll) Wait() ([]int, error) {
	events := make([]syscall.Kevent_t, len(e.connections)+1)
	n, err := syscall.Kevent(e.fd, nil, events, &syscall.Timespec{
		Sec:  1000,
		Nsec: 0,
	})
	if err != nil {
		return nil, err
	}
	e.lock.RLock()
	defer e.lock.RUnlock()
	var connectionsFD []int
	for i := 0; i < n; i++ {
		connectionsFD = append(connectionsFD, int(events[i].Ident))
	}
	return connectionsFD, nil

}

func websocketFD(conn net.Conn) int {
	tln := conn.(*net.TCPConn)
	f, err := tln.File()
	if err != nil{
		return  -1
	}
	return int(f.Fd())
}



