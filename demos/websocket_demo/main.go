package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	context := map[string]string{}

	fp := path.Join("templates", "index.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, context); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WebsocketHandler(conn *websocket.Conn) {
	fmt.Println("New Websocket Conn: ", conn.RemoteAddr().String())

	defer conn.Close()

	for {
		msgType, rawBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v", err)
			} else {
				fmt.Println("read error: ", err.Error())
			}
			break
		}

		fmt.Println(msgType, string(rawBytes))
		conn.WriteMessage(msgType, []byte("resp"))
	}
}

func RegisterRouters(mux *http.ServeMux) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	mux.HandleFunc("/", HomePage)

	// Websocket entry
	mux.HandleFunc("/websocket/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			println("Upgrade Error:", err.Error())
			return
		}

		WebsocketHandler(conn)

	})
}

func APIMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func NewHttpServer(ServerPort int) *http.Server {
	mux := &http.ServeMux{}

	HttpServer := &http.Server{
		Handler: mux,
	}

	corsHandler := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		Debug:            false,
	})

	RegisterRouters(mux)

	handler := corsHandler.Handler(APIMiddleware(mux))
	go func() {
		err := http.ListenAndServe(fmt.Sprint(":", ServerPort), handler)
		if err != nil {
			fmt.Printf("server.Start failed: %v\n", err)
		}
	}()

	return HttpServer
}

func main() {
	ctx := context.Background()

	ServerPort := 3333
	fmt.Printf("Listening service on http://0.0.0.0:%d\n", ServerPort)

	HttpServer := NewHttpServer(ServerPort)
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

	signalCode := <-ch
	switch signalCode {
	case syscall.SIGHUP:
		fmt.Println(">>>")
	case syscall.SIGUSR1:
	case syscall.SIGUSR2:
	default:
		HttpServer.Shutdown(ctx)
	}
}
