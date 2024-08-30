package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ahornerr/artifacts/character"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"log"
	"time"
)

type Event struct {
	Character *character.Character
	Bank      map[string]int
}

func marshalEvent(event Event) (byteArray []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered panic %+v for event %+v", r, event)
		}
	}()

	return json.Marshal(event)
}

func httpServer(events <-chan Event, onNewClient func()) *fiber.App {
	app := fiber.New()

	newClient := make(chan chan []byte)
	deleteClient := make(chan chan []byte)

	clients := map[chan []byte]bool{}

	go func() {
		for {
			select {
			case c := <-newClient:
				clients[c] = true
			case c := <-deleteClient:
				delete(clients, c)
			case ev := <-events:
				if len(clients) == 0 {
					continue
				}

				message, err := marshalEvent(ev)
				if err != nil {
					log.Println("Error marshalling event:", err)
					continue
				}

				for client := range clients {
					// Non-blocking write in case something goes wrong
					select {
					case client <- message:
						//log.Println("Wrote to client channel successfully")
					case <-time.After(200 * time.Millisecond):
						log.Println("Timeout writing to client channel")
					}
				}
			}
		}
	}()

	app.Get("/events", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-transform")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")
		c.Set("Access-Control-Allow-Origin", "*")

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			clientChan := make(chan []byte)

			defer func() {
				deleteClient <- clientChan
			}()

			newClient <- clientChan

			go func() {
				onNewClient()
			}()

			for {
				select {
				case message := <-clientChan:
					_, err := fmt.Fprintf(w, "data: %v\n\n", string(message))
					if err != nil {
						log.Println("Error writing to client:", err)
						return
					}

					err = w.Flush()
					if err != nil {
						return
					}
				}
			}
		})

		return nil
	})

	app.Get("/*", static.New("./frontend/build"))

	return app
}
