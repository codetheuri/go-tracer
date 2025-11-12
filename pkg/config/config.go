package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// Config holds the application configuration.
type Config struct {
	// Interface is the network interface to capture packets from.
	Interface string `envconfig:"INTERFACE" default:"eth0"`
	// BPF Filer to apply to the packet capture.
	BPFFilter string `envconfig:"BPF_FILTER" default:"tcp"`
	// SnapLen is the maximum size to read for each packet.
	SnapLen int32 `envconfig:"SNAP_LEN" default:"1024"`
	// Promiscuous mode for the network interface.
	Promiscuous bool `envconfig:"PROMISCUOUS" default:"true"`

	// Sender configuration
	EndpointURL  string `envconfig:"ENDPOINT_URL" default:"http://localhost:8080/post"`
	BatchSize    int    `envconfig:"BATCH_SIZE" default:"10"`
	SendInterval int    `envconfig:"SEND_INTERVAL" default:"5"` // In seconds
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	var cfg Config
	log.Println("Loading configuration...")
	err := envconfig.Process("AGENT", &cfg)
	if err != nil {
		return nil, err
	}
	log.Printf("Configuration loaded: %+v\n", cfg)
	return &cfg, nil
}
