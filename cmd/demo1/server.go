package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

var (
	globalSocketPath = filepath.Join("./", "global.sock")
	testFilePath     = filepath.Join("./", "1.txt")
)

func transferFd(uc *net.UnixConn, fd uintptr) error {
	buf := make([]byte, 1)
	buf[0] = 0
	rights := syscall.UnixRights(int(fd))
	n, oobn, err := uc.WriteMsgUnix(buf, rights, nil)
	if err != nil {
		return fmt.Errorf("WriteMsgUnix: %v", err)
	}
	if n != len(buf) || oobn != len(rights) {
		return fmt.Errorf("WriteMsgUnix = %d, %d; want 1, %d", n, oobn, len(rights))
	}

	return nil
}

func listenAndServer() {
	// 监听unix socket
	l, err := net.Listen("unix", globalSocketPath)
	if err != nil {
		log.Printf("listen %s failed: %s\n", globalSocketPath, err)
		return
	}
	defer l.Close()
	log.Println("listening...")

	// 读文件
	f, err := os.Open(testFilePath)
	if err != nil {
		log.Printf("open file err %v\n", err)
		return
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

	// 迁移文件fd
	ul := l.(*net.UnixListener)
	uc, err := ul.AcceptUnix()
	if err != nil {
		log.Printf("AcceptUnix failed: %s\n", err)
		return
	}

	err = transferFd(uc, f.Fd())
	if err != nil {
		log.Printf("transferFd err %v\n", err)
		return
	}
	log.Println("transferFd ok fd ", f.Fd())

	// 写消息
	//buf := []byte("hello client")
	//n, err := uc.Write(buf)
	//if err != nil {
	//	log.Printf("send message error %v\n", err)
	//	return
	//}
	//log.Printf("send message ok, len:%v\n", n)
}

func main() {
	log.Println("server start...")
	listenAndServer()
	time.Sleep(time.Second)
	log.Println("server stop")
}
