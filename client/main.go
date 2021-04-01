package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var (
	addr = "localhost:8989"
)

func main() {
	// interrupt signal.
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt)

	url := url.URL{Scheme: "ws", Host: addr, Path: "/hello"}
	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
	}
	fmt.Printf("[DEBUG]url: %s\n", url.String())
	client, _, err := dialer.Dial(url.String(), nil)
	if err != nil {
		fmt.Printf("[ERROR]new client: %s\n", err.Error())
		return
	}
	defer client.Close()
	readDone := make(chan struct{})
	go func() {
		defer close((readDone))
		for {
			tp, data, err := client.ReadMessage()
			if err != nil {
				if err == io.EOF {
					fmt.Println("read message from server: EOF\n")
				} else {
					fmt.Printf("read message: %s\n", err.Error())
				}
				break
			}

			if tp == websocket.TextMessage {
				fmt.Printf("message from server: %s\n", string(data))
			}
		}
	}()

	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	for {
		select {
		case <-readDone:
			return
		case <-ticker.C:
			// send message to server
			msg := []byte("I'm client.")
			if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
				fmt.Printf("send message to server: %s\n", err.Error())
				return
			}
		case sig := <-interrupt:
			fmt.Printf("receive a signal from os: %s\n", sig.String())
			// shut down connection graciously.
			if err := client.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				fmt.Printf("send closure message to server: %s\n", err.Error())
			}
			// wait till read process is done.
			<-readDone
			return
		}
	}
}
