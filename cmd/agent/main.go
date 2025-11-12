package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/gopacket"
	"network-go/internal/capture"
	"network-go/internal/models" 
	"network-go/internal/processor"
	"network-go/internal/sender"
	"network-go/pkg/config"
)

func main() {
	// --- LOAD CONFIG & HOST INFO ---
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Could not get hostname, using 'unknown': %v", err)
		hostname = "unknown"
	}

	localIPs, err := getLocalIPs()
	if err != nil {
		log.Fatalf("Failed to get local IPs: %v", err)
	}
	log.Printf("Local IPs detected: %v", localIPs)

	// --- CHANNELS ---
	packetChan := make(chan gopacket.Packet, 1000)
	

	eventChan := make(chan models.FlowEvent, 1000) 

	// --- START PIPELINE ---
	go capture.StartCapture(cfg, packetChan)

	go processor.StartProcessor(packetChan, eventChan, hostname, cfg.Interface, localIPs)

	go sender.StartSender(cfg, eventChan)

	log.Println("Network agent is running. Press Ctrl+C to stop.")


	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutdown signal received. Closing channels...")
	close(packetChan)
	close(eventChan)
	log.Println("Agent shut down gracefully.")
}

// getLocalIPs
func getLocalIPs() (map[string]bool, error) {
	ips := make(map[string]bool)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil {
			continue
		}
		ips[ip.String()] = true
	}
	ips["127.0.0.1"] = true
	ips["::1"] = true
	return ips, nil
}