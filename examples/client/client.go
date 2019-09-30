package main

import (
	"flag"
	"sync"
	"time"

	kcp "github.com/ejoy/kcp-go"
)

func dialEchoServer(listen string) (*kcp.UDPSession, error) {
	sess, err := kcp.DialWithOptions(listen, nil, 0, 0)
	if err != nil {
		panic(err)
	}

	sess.SetStreamMode(true)
	sess.SetWindowSize(1024, 1024)
	sess.SetReadBuffer(16 * 1024 * 1024)
	sess.SetWriteBuffer(16 * 1024 * 1024)
	sess.SetNoDelay(1, 10, 2, 1)
	sess.SetMtu(1400)
	sess.SetACKNoDelay(true)
	return sess, err
}

func main() {
	var server string
	var count int
	var pkts int
	var size int

	flag.StringVar(&server, "server", "127.0.0.1:8551", "local listen port(0.0.0.0:1248)")
	flag.IntVar(&count, "count", 8, "count of clients")
	flag.IntVar(&pkts, "pkts", 1000, "packet count")
	flag.IntVar(&size, "size", 1000, "packet size")
	flag.Parse()

	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			cli, err := dialEchoServer(server)
			if err != nil {
				panic(err)
			}
			buf := make([]byte, size)
			for j := 0; j < pkts; j++ {
				if _, err := cli.Write(buf); err != nil {
					break
				}
				if _, err := cli.Read(buf); err != nil {
					break
				}
			}
			time.Sleep(5 * time.Second)
			cli.Close()
			wg.Done()
		}()
	}
	wg.Wait()
	time.Sleep(1 * time.Second)
}
