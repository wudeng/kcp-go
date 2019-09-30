package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"net/http"
	_ "net/http/pprof"

	"github.com/ejoy/kcp-go"
	reuse "github.com/libp2p/go-reuseport"
	"github.com/pkg/errors"
)

func listenWithOptions(laddr string) (*kcp.Listener, error) {
	conn, err := reuse.ListenPacket("udp", laddr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return kcp.ServeConn(nil, 0, 0, conn)
}

func main() {
	var listen string
	var count int
	var wg sync.WaitGroup

	flag.StringVar(&listen, "listen", "0.0.0.0:8551", "local listen port(0.0.0.0:1248)")
	flag.IntVar(&count, "count", 1, "count of listeners")
	flag.Parse()

	// runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)
	go func() {
		http.ListenAndServe(":8888", nil)
	}()
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(index int) {
			fmt.Fprintln(os.Stderr, "listener: ", index)
			if listener, err := listenWithOptions(listen); err == nil {
				// if listener, err := kcp.ListenWithOptions(listen, nil, 0, 0); err == nil {
				listener.SetReadBuffer(16 * 1024 * 1024)
				listener.SetWriteBuffer(16 * 1024 * 1024)
				// spin-up the client
				for {
					s, err := listener.AcceptKCP()
					if err != nil {
						log.Fatal(err)
					}
					go handleEcho(s)
				}
			} else {
				log.Fatal(err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

// handleEcho send back everything it received
func handleEcho(conn *kcp.UDPSession) {
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}

		n, err = conn.Write(buf[:n])
		if err != nil {
			log.Println(err)
			return
		}
	}
}
