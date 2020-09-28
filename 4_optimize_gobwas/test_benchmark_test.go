package main

import (
	"net"
	"reflect"
	"testing"
)

var conn net.Conn
func init()  {
	var err error
	conn, err = net.Dial("tcp", "127.0.0.1:6060")
	if err != nil{
		panic(err)
	}
}

func getFD(conn net.Conn) int {
	tln := conn.(*net.TCPConn)
	f, err := tln.File()
	if err != nil{
		return  -1
	}else {
		_ = conn.Close()
	}
	return int(f.Fd())
}

func BenchmarkTest(b *testing.B)  {
	b.ResetTimer()
	for i:=0; i < b.N; i++{
		getFD(conn)
	}
}

func getFDReflect(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return int(pfdVal.FieldByName("Sysfd").Int())
}

func BenchmarkReflect(b *testing.B)  {
	b.ResetTimer()
	for i:=0; i < b.N; i++{
		getFDReflect(conn)
	}
}