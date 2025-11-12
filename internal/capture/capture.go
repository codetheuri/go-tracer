package capture
import (
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"network-go/pkg/config"
)


// It sends captured packets to the provided channel.
func StartCapture(cfg *config.Config, packetChan chan<- gopacket.Packet) {
	log.Printf("Starting capture on interface: %s", cfg.Interface)

	// Open device
	handle, err := pcap.OpenLive(cfg.Interface, cfg.SnapLen, cfg.Promiscuous, 30*time.Second)
	if err != nil {
		log.Fatalf("Error opening device %s: %v", cfg.Interface, err)
	}
	defer handle.Close()

	// Set BPF filter
	if err := handle.SetBPFFilter(cfg.BPFFilter); err != nil {
		log.Fatalf("Error setting BPF filter: %v", err)
	}

	log.Println("Capture started. Waiting for packets...")
	

	// Use gopacket.PacketSource to correctly decode layers.
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// Send the captured packet to the processing channel
		packetChan <- packet
	}
	
}