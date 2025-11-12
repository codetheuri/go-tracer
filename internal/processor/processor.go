package processor

import (
	"fmt"
	"log"
	"sync"
	"time"

	"network-go/internal/models"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)


type flowState struct {
	FlowID      string
	SrcIP       string
	DstIP       string
	SrcPort     string
	DstPort     string
	Protocol    string
	Direction   string
	StartTime   time.Time
	LastTime    time.Time
	PacketCount int64
	ByteCount   int64
}

// FlowCache holds the map of active flows and a mutex for thread-safety
type FlowCache struct {
	sync.Mutex
	flows map[string]*flowState

	
	hostname string
	iface    string
	localIPs map[string]bool

	// Channel to send finalized events
	eventChan chan<- models.FlowEvent
}

// NewFlowCache creates a new flow cache
func NewFlowCache(eventChan chan<- models.FlowEvent, hostname, iface string, localIPs map[string]bool) *FlowCache {
	return &FlowCache{
		flows:     make(map[string]*flowState),
		eventChan: eventChan,
		hostname:  hostname,
		iface:     iface,
		localIPs:  localIPs,
	}
}

// StartProcessor now creates a FlowCache and starts the janitor
func StartProcessor(
	packetChan <-chan gopacket.Packet,
	eventChan chan<- models.FlowEvent,
	hostname string,
	iface string,
	localIPs map[string]bool,
) {
	log.Println("Stateful Flow Processor started...")
	fc := NewFlowCache(eventChan, hostname, iface, localIPs)


	go fc.startJanitor()

	for packet := range packetChan {
		fc.processPacket(packet)
	}
}

// processPacket is now a method of FlowCache
func (fc *FlowCache) processPacket(packet gopacket.Packet) {
	// --- 1. Get Base Info ---
	netLayer := packet.NetworkLayer()
	transportLayer := packet.TransportLayer()
	if netLayer == nil || transportLayer == nil {
		return // Not a flow we can track
	}

	netFlow := netLayer.NetworkFlow()
	transFlow := transportLayer.TransportFlow()

	srcIP := netFlow.Src().String()
	dstIP := netFlow.Dst().String()
	srcPort := transFlow.Src().String()
	dstPort := transFlow.Dst().String()

	var protocol string
	var isFinOrRst bool

	switch transportLayer.LayerType() {
	case layers.LayerTypeTCP:
		protocol = "TCP"
		tcp, _ := transportLayer.(*layers.TCP)
		if tcp.FIN || tcp.RST {
			isFinOrRst = true
		}
	case layers.LayerTypeUDP:
		protocol = "UDP"
	default:
		return // Not TCP or UDP
	}

	// Create a unique flow ID
	flowID := fmt.Sprintf("%s:%s-%s:%s-%s", srcIP, srcPort, dstIP, dstPort, protocol)
	packetLength := int64(packet.Metadata().Length)
	now := packet.Metadata().Timestamp

	// --- 2. Lock Cache & Find Flow ---
	fc.Lock()
	defer fc.Unlock()

	flow, exists := fc.flows[flowID]
	if !exists {
		// --- 3. New Flow ---

		// Calculate Direction
		_, srcIsLocal := fc.localIPs[srcIP]
		_, dstIsLocal := fc.localIPs[dstIP]
		direction := "unknown"
		if srcIsLocal && !dstIsLocal {
			direction = "outbound"
		} else if !srcIsLocal && dstIsLocal {
			direction = "inbound"
		} else if srcIsLocal && dstIsLocal {
			direction = "local"
		}

		// Create new flow
		flow = &flowState{
			FlowID:      flowID,
			SrcIP:       srcIP,
			DstIP:       dstIP,
			SrcPort:     srcPort,
			DstPort:     dstPort,
			Protocol:    protocol,
			Direction:   direction,
			StartTime:   now,
			LastTime:    now,
			PacketCount: 1,
			ByteCount:   packetLength,
		}
		fc.flows[flowID] = flow
	} else {
		// --- 4. Existing Flow ---
		flow.PacketCount++
		flow.ByteCount += packetLength
		flow.LastTime = now
	}

	// --- 5. Finalize Flow (if TCP FIN/RST) ---
	if isFinOrRst {
		fc.finalizeAndSend(flow)
		delete(fc.flows, flowID)
	}
}

// finalizeAndSend converts flowState to FlowEvent and sends it
func (fc *FlowCache) finalizeAndSend(flow *flowState) {
	duration := flow.LastTime.Sub(flow.StartTime).Seconds()

	var avgPacketSize float64
	if flow.PacketCount > 0 {
		avgPacketSize = float64(flow.ByteCount) / float64(flow.PacketCount)
	}

	event := models.FlowEvent{
		FlowID:        flow.FlowID,
		Timestamp:     flow.LastTime, // Time flow ended
		Interface:     fc.iface,
		Hostname:      fc.hostname,
		Direction:     flow.Direction,
		SrcIP:         flow.SrcIP,
		DstIP:         flow.DstIP,
		SrcPort:       flow.SrcPort,
		DstPort:       flow.DstPort,
		Protocol:      flow.Protocol,
		FlowDuration:  duration,
		PacketCount:   flow.PacketCount,
		ByteCount:     flow.ByteCount,
		AvgPacketSize: avgPacketSize,
	}
	fc.eventChan <- event
}

// startJanitor runs periodically to evict old, timed-out flows
func (fc *FlowCache) startJanitor() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		<-ticker.C

		now := time.Now()

		fc.Lock()

		for flowID, flow := range fc.flows {
			// Evict flows older than 60 seconds (TCP) or 30 seconds (UDP)
			timeout := 60 * time.Second
			if flow.Protocol == "UDP" {
				timeout = 30 * time.Second
			}

			if now.Sub(flow.LastTime) > timeout {
				log.Printf("[Janitor] Evicting flow: %s", flowID)
				fc.finalizeAndSend(flow)
				delete(fc.flows, flowID)
			}
		}

		fc.Unlock()
	}
}
