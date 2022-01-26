package main

import (
	"bufio"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

var (
	globalSocketPath = filepath.Join("./", "global.sock")
)

func recvFd(uc *net.UnixConn) error {
	buf := make([]byte, 1)
	oob := make([]byte, 1024)
	_, oobn, _, _, err := uc.ReadMsgUnix(buf, oob)
	if err != nil {
		log.Printf("ReadMsgUnix err: %v\n", err)
		return err
	}
	scms, err := unix.ParseSocketControlMessage(oob[0:oobn])
	if err != nil {
		log.Printf("ParseSocketControlMessage failed: %v\n", err)
		return err
	}
	if len(scms) != 1 {
		log.Printf("ParseSocketControlMessage failed: expected 1 bug get scms = %#v\n", scms)
		return err
	}
	gotFds, err := unix.ParseUnixRights(&scms[0])
	if err != nil {
		log.Printf("ParseUnixRights err: %s\n", err)
		return err
	}

	fd := uintptr(gotFds[0])
	log.Printf("recv fd %v\n", fd)
	f := os.NewFile(fd, "")
	if f == nil {
		log.Printf("create new file failed fd %d\n", fd)
		return err
	}
	defer f.Close()
	log.Printf("open file ok, fd is %v\n", f.Fd())

	br := bufio.NewReader(f)
	for {
		str, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		fmt.Print(string(str))
	}

	return nil
}

func openFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("open file err %v\n", err)
		return
	}
	//defer f.Close()
	log.Printf("open file %s ok, fd is %v\n", path, f.Fd())
}

func main() {
	log.Println("client start...")
	unixConn, err := net.DialTimeout("unix", globalSocketPath, 1*time.Second)
	if err != nil {
		log.Printf("connect failed: %s\n", err)
		return
	}
	defer unixConn.Close()

	log.Println("dail ok")
	//buf := make([]byte, 128)
	//n, err := unixConn.Read(buf)
	//if err != nil {
	//	log.Printf("read err %v\n", err)
	//	return
	//}
	//log.Printf("read ok len %v, msg:%s\n", n, string(buf))
	openFile("2.txt")
	openFile("3.txt")
	err = recvFd(unixConn.(*net.UnixConn))
	if err != nil {
		log.Printf("recvFd err %v\n", err)
	}
	log.Println("client stop")
}
