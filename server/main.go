package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	//"github.com/gorilla/websocket"
	"leslieran.com/websocketTest/server/websocket"
)

var (
	upgrade = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	addr = "localhost:8989"
)

func main() {
	server := http.NewServeMux()

	server.HandleFunc("/hello", Hello)

	http.ListenAndServe(addr, server)

}

func Hello(rw http.ResponseWriter, req *http.Request) {
	wsConn, err := upgrade.Upgrade(rw, req, nil)
	if err != nil {
		fmt.Printf("[ERROR]upgrade: %s\n", err.Error())
		return
	}
	defer wsConn.Close()
	readDone := make(chan struct{})
	go func() {
		defer close(readDone) // inform that write process should be terminated and send a signal to close connection.
		for {                 // receive message from client and response.
			tp, data, err := wsConn.ReadMessage()
			if err != nil {
				if err == io.EOF {
					fmt.Println("[DEBUG]there is nothing from client.")

				} else {
					fmt.Printf("[ERROR]read message: %s\n", err.Error())
				}
				break
			}
			if tp == websocket.TextMessage {
				fmt.Printf("[DEBUG]data from client: %s\n", string(data))
			} else if tp == websocket.CloseMessage {
				fmt.Println("[WARNING]client has closed the connection.")
				// terminate write process.
				return
			}

		}
	}()
	// send message to client forwardly.
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-readDone:
			fmt.Printf("read process has been terminated\n")
			// shut down connection graciously.
			if err := wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				fmt.Printf("send closure message to client: %s\n", err.Error())
			}
			return
		case <-ticker.C:
			msg := []byte("I'm server.")
			if err := wsConn.WriteMessage(websocket.TextMessage, msg); err != nil {
				fmt.Printf("send message forwardly: %s\n", err.Error())
				break
			}
		}

	}

}
