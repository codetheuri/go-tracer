package sender

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"network-go/internal/models"
	"network-go/pkg/config"
)

func StartSender(cfg *config.Config, eventChan <-chan models.FlowEvent) {
	log.Println("Sender started. Sending events in real-time...")

	client := &http.Client{Timeout: 10 * time.Second}

	for event := range eventChan {

		sendEvent(cfg.EndpointURL, event, client)
	}
}

func sendEvent(endpointURL string, event models.FlowEvent, client *http.Client) {

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	req, err := http.NewRequest("POST", endpointURL, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending event to %s: %v", endpointURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		log.Printf("Received non-OK response: %s", resp.Status)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading error response body: %v", err)
		} else {

			log.Printf("Validation Error Body: %s", string(body))
		}
	}
}
