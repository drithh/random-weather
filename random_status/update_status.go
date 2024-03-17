package randomstatus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	broadcast "random-weather/broadcast"
	"text/template"
	"time"
)

type Object struct {
	Name  string `json:"name"`
	Unit  string `json:"unit"`
	Rules Rules  `json:"rules"`
}

type Rules struct {
	Safe    int `json:"safe"`
	Warning int `json:"warning"`
	Danger  int `json:"danger"`
}

type Status struct {
	Status map[string]int `json:"status"`
}

func generateRandomValue(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func UpdateStatusJSON(filePath string, objectToGenerate []Object, interval time.Duration) {
	for {
		statuses := make(map[string]int)
		for _, object := range objectToGenerate {
			statuses[object.Name] = generateRandomValue(1, 100)
		}

		// convert to meet requirements
		statusObject := Status{Status: statuses}

		statusJSON, err := json.Marshal(statusObject)
		if err != nil {
			fmt.Println("Error marshalling status info:", err)
			continue
		}

		// write to /data/status.json
		err = os.WriteFile(filePath, statusJSON, 0644)
		if err != nil {
			fmt.Println("Error writing status file:", err)
			continue
		}

		var result = []map[string]interface{}{}

		for _, object := range objectToGenerate {
			var condition = getCondition(statusObject.Status[object.Name], object.Rules)
			var status = map[string]interface{}{
				"time":      time.Now().Format("15:04:05"),
				"name":      object.Name,
				"unit":      object.Unit,
				"value":     statusObject.Status[object.Name],
				"condition": condition,
			}
			result = append(result, status)
		}

		// websocket dashboard
		tmpl, err := template.ParseFiles("template/dashboard.html")
		if err != nil {
			fmt.Println("Error parsing template:", err)
			return
		}

		var renderedDashboard bytes.Buffer
		if err := tmpl.Execute(&renderedDashboard, result); err != nil {
			fmt.Println("Error executing template:", err)
			return
		}

		// websocket table
		tmpl, err = template.ParseFiles("template/table.html")

		if err != nil {
			fmt.Println("Error parsing template:", err)
			return
		}

		var renderedTable bytes.Buffer
		if err := tmpl.Execute(&renderedTable, result); err != nil {
			fmt.Println("Error executing template:", err)
			return
		}

		broadcast.SendMessage(renderedDashboard.String(), nil)
		broadcast.SendMessage(renderedTable.String(), nil)

		time.Sleep(interval * time.Second)
	}
}

func getCondition(value int, rules Rules) int {
	if value <= rules.Safe {
		return 0
	} else if value <= rules.Warning {
		return 1
	} else {
		return 2
	}
}
