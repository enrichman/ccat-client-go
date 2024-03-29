package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
)

type ChatService struct {
	client *Client
}

func (c *ChatService) Chat(ctx context.Context, in, out chan string) error {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:1865/ws/user", nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	errChan := make(chan error)
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				errChan <- err
				return
			}

			type response struct {
				Content string `json:"content"`
			}
			var res response
			err = json.Unmarshal(message, &res)
			if err != nil {
				errChan <- err
				return
			}

			out <- res.Content
		}
	}()

	for {
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			return conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		case line := <-in:
			line = strings.TrimSpace(line)
			msg := fmt.Sprintf(`{"text":"%s"}`, line)
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				return err
			}
		}
	}
}
