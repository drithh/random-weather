package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"random-weather/broadcast"
	rs "random-weather/random_status"
	"time"

	"golang.org/x/net/websocket"
)

func statusHandler(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statusData, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Error reading status file", http.StatusInternalServerError)
			return
		}

		var status rs.Status
		err = json.Unmarshal(statusData, &status)
		if err != nil {
			http.Error(w, "Error unmarshalling status info", http.StatusInternalServerError)
			return
		}

		statusJSON, err := json.Marshal(status)

		if err != nil {
			http.Error(w, "Error marshalling status info", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(statusJSON)
	}
}

var connections = make(map[*websocket.Conn]struct{})

func wsHandler(ws *websocket.Conn) {
	fmt.Println("WebSocket connection established")
	connections[ws] = struct{}{}
	defer func() {
		delete(connections, ws)
		ws.Close()
	}()

	for message := range broadcast.BroadcastMessage {
		for conn := range connections {
			err := websocket.Message.Send(conn, message)
			if err != nil {
				fmt.Println("Error writing to WebSocket:", err)
				return
			}
		}
	}
}

func main() {
	filePath := "data/status.json"
	objectToGenerate := []rs.Object{
		// jika water dibawah 5 maka status aman, jika water antara 6 - 8 maka status siaga, jika water diatas 8 maka status bahaya
		{Name: "water", Unit: "m", Rules: rs.Rules{Safe: 5, Warning: 6, Danger: 8}},
		// jika wind dibawah 6 maka status aman, jika wind antara 7 - 15 maka status siaga, jika wind diatas 15 maka status bahaya
		{Name: "wind", Unit: "m/s", Rules: rs.Rules{Safe: 6, Warning: 7, Danger: 15}},
	}

	var sleepTime time.Duration = 15 // in seconds

	// go routine to update status JSON file and send WebSocket message
	go rs.UpdateStatusJSON(filePath, objectToGenerate, sleepTime)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "template/index.html")
	})
	http.HandleFunc("/api/v1/status", statusHandler(filePath))

	fs := http.FileServer(http.Dir("css"))
	http.Handle("/css/", http.StripPrefix("/css/", fs))

	http.Handle("/ws", websocket.Handler(wsHandler)) // WebSocket endpoint

	port := "8080"
	fmt.Printf("Starting server on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
