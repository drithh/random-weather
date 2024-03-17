package broadcast

import "fmt"

var BroadcastMessage = make(chan string, 10)

func SendMessage(message string, done <-chan struct{}) {
	select {
	case BroadcastMessage <- message:
	case <-done:
		fmt.Println("Client disconnected, not adding message to broadcast")
	}
}
