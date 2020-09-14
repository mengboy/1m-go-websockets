// +build darwin

package main

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

import (
	"errors"
	"reflect"
	"syscall"
)

type epoll struct {
	fd          int
	connections map[int]*websocket.Conn
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
		connections: map[int]*websocket.Conn{},
	}, nil
}

func (e *epoll) Add(conn *websocket.Conn) error {
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

func (e *epoll) Remove(conn *websocket.Conn) error {
	fd := websocketFD(conn)
	if fd < 0 {
		return errors.New("")
	}
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

func (e *epoll) Wait() ([]*websocket.Conn, error) {
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
	var connections []*websocket.Conn
	for i := 0; i < n; i++ {
		conn := e.connections[int(events[i].Ident)]
		connections = append(connections, conn)
	}
	return connections, nil

}

func websocketFD(conn *websocket.Conn) int {
	connVal := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn").Elem()
	tcpConn := reflect.Indirect(connVal).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
