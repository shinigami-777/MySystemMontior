package main

import (
	"MySystemMontior/cmd/hardware"
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type server struct {
	subscriberMessageBuffer int
	mux                     http.ServeMux
	subscribersMutex        sync.Mutex
	subscribers             map[*subscriber]struct{}
}

type subscriber struct {
	msgs chan []byte
}

func NewServer() *server {
	s := &server{
		subscriberMessageBuffer: 10,
		subscribers:             make(map[*subscriber]struct{}),
	}

	s.mux.Handle("/", http.FileServer(http.Dir("./cmd/htmx")))
	s.mux.HandleFunc("/ws", s.subscribeHandler)
	return s
}

func (s *server) subscribeHandler(writer http.ResponseWriter, req *http.Request) {
	err := s.subscribe(req.Context(), writer, req)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *server) subscribe(ctx context.Context, writer http.ResponseWriter, req *http.Request) error {
	var c *websocket.Conn
	subscriber := &subscriber{
		msgs: make(chan []byte, s.subscriberMessageBuffer),
	}
	s.addSubscriber(subscriber)

	c, err := websocket.Accept(writer, req, nil)
	if err != nil {
		return err
	}
	defer c.CloseNow()

	ctx = c.CloseRead(ctx)
	for {
		select {
		case msg := <-subscriber.msgs:
			ctx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			err := c.Write(ctx, websocket.MessageText, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *server) addSubscriber(subscriber *subscriber) {
	s.subscribersMutex.Lock()
	s.subscribers[subscriber] = struct{}{}
	s.subscribersMutex.Unlock()
	fmt.Println("Added subscriber ", subscriber)
}

func (s *server) broadcast(msg []byte) {
	s.subscribersMutex.Lock()
	defer s.subscribersMutex.Unlock()

	for s := range s.subscribers {
		s.msgs <- msg
	}
}

func main() {
	fmt.Println("Starting the Golang System Monitor....")
	srv := NewServer()

	go func(s *server) {
		for {
			systemSection, err := hardware.GetSystemSection()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(systemSection)
				//s.broadcast([]byte(systemSection))
			}

			diskSection, err := hardware.GetDiskSection()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(diskSection)
				//s.broadcast([]byte(diskSection))
			}

			cpuSection, err := hardware.GetCPUSection()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(cpuSection)
				//s.broadcast([]byte(cpuSection))
			}

			fmt.Println("---------------------------------------------------")

			timestamp := time.Now().Format("2006-01-02 15:01:05")
			msg := []byte(`
      <div hx-swap-oob="innerHTML:#update-timestamp">
        <p><i style="color: green" class="fa fa-circle"></i> ` + timestamp + `</p>
      </div>
      <div hx-swap-oob="innerHTML:#system-data">` + systemSection + `</div>
      <div hx-swap-oob="innerHTML:#cpu-data">` + cpuSection + `</div>
      <div hx-swap-oob="innerHTML:#disk-data">` + diskSection + `</div>`)
			s.broadcast(msg)

			//keep updating every 2 seconds
			time.Sleep(3 * time.Second)

		}
	}(srv)

	err := http.ListenAndServe(":6969", &srv.mux)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
