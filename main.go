package main

import (
	"bufio"
	"flag"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/chat"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	w := bufio.NewWriter(os.Stdout)
	r := bufio.NewReader(os.Stdin)

	toWrite := make(chan []byte)
	done := make(chan struct{})

	go readConn(c, toWrite, done)
	go writeConn(c, r)
	writeConsole(w, toWrite, done)
}

func writeConsole(w *bufio.Writer, in <-chan []byte, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case s := <-in:
			w.WriteString(string(s)) // try to handle err later
			w.Flush()                // try to handle err later
		}
	}
}

func writeConn(c *websocket.Conn, r *bufio.Reader) {
	for {
		input, err := r.ReadBytes('\n')
		input = input[:len(input)-2] // remove \n
		if err != nil {
			continue
		}
		err = c.WriteMessage(websocket.TextMessage, input)
		if err != nil {
			return
		}
	}
}

func readConn(c *websocket.Conn, out chan<- []byte, done chan struct{}) {
	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		out <- append(message, '\n')
	}
}
