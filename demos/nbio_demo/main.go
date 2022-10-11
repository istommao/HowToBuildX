package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var (
	qps   uint64 = 0
	total uint64 = 0

	svr *nbhttp.Server
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		c.SetReadDeadline(time.Now().Add(time.Second * 60))

		// fmt.Println("Recv: ", string(data))

		c.WriteMessage(messageType, data)
		atomic.AddUint64(&qps, 1)
	})

	return u
}

func onWebsocket(w http.ResponseWriter, r *http.Request) {
	upgrader := newUpgrader()
	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	ServerPort := 3333
	fmt.Printf("Listening service on http://0.0.0.0:%d\n", ServerPort)

	mux := &http.ServeMux{}
	mux.HandleFunc("/", HomePage)
	mux.HandleFunc("/websocket/", onWebsocket)

	var addrs = []string{
		// fmt.Sprintf("localhost:%d", ServerPort),
		fmt.Sprintf("127.0.0.1:%d", ServerPort),
		// fmt.Sprintf("0.0.0.0:%d", ServerPort),
	}

	svr = nbhttp.NewServer(nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   addrs,
		MaxLoad:                 1000000,
		ReleaseWebsocketPayload: true,
		Handler:                 mux,
		ReadBufferSize:          1024 * 4,
	})

	err := svr.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer svr.Stop()

	// ticker := time.NewTicker(time.Second)
	// for i := 1; true; i++ {
	// 	<-ticker.C
	// 	n := atomic.SwapUint64(&qps, 0)
	// 	total += n
	// 	fmt.Printf("running for %v seconds, NumGoroutine: %v, qps: %v, total: %v\n", i, runtime.NumGoroutine(), n, total)
	// }

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

	signalCode := <-ch
	switch signalCode {
	case syscall.SIGHUP:
		fmt.Println(">>>")
	case syscall.SIGUSR1:
	case syscall.SIGUSR2:
	default:
		fmt.Println(">>>")
	}
}
