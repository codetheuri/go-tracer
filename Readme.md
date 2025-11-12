# Smart-Trace Sensor (network-go)

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org) [![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

# go‑Tracer Sensor 

go‑Tracer Sensor is a high‑performance, stateful network agent written in Go. It monitors traffic, summarizes flows into compact JSON `FlowEvent` objects, and posts them to the Python Smart‑Trace Brain for AI analysis.

## Key Features

- Stateful flow analysis using an in‑memory `FlowCache` (not a packet logger).
- High performance via Go goroutines and channels; handles thousands of packets/sec.
- Produces lightweight JSON summaries (`FlowDuration`, `PacketCount`, `ByteCount`) for ML/AI.
- Enriches flows with `hostname`, `interface`, and `direction` (inbound/outbound/local).
- Containerized with Docker for easy deployment.

## Architecture — the 3‑station "Factory"

1. Station 1 — capture (Collector)
    - Uses `gopacket/pcap` to sniff packets.
    - Emits raw packets to `packetChan`.

2. Station 2 — processor (Assembler / Brain)
    - Maintains `FlowCache` to track conversations.
    - Consumes `packetChan`, updates flow state, finalizes flows on FIN/RST or timeout.
    - Emits finalized `FlowEvent` to `eventChan`.
    - Runs a `janitor` goroutine to evict timed‑out flows.

3. Station 3 — sender (Shipper)
    - Consumes `eventChan`.
    - Sends each `FlowEvent` as a JSON `POST` to the backend API.

## Project Structure

```text
network-go/
├── cmd/agent/main.go       # Entry point — builds the factory
├── internal/
│   ├── capture/            # Station 1: packet capture
│   ├── processor/          # Station 2: stateful flow analysis
│   ├── sender/             # Station 3: JSON-over-HTTP sender
│   └── models/             # FlowEvent struct
├── pkg/
│   └── config/             # Env var configuration
├── go.mod
├── go.sum
├── Dockerfile
└── docker-compose.yml
```

## How to Run

This agent is designed to run in Docker.

1. Prerequisites
    - Docker
    - Docker Compose
    - A reachable Smart‑Trace backend (set via `AGENT_ENDPOINT_URL`)

2. Configuration (via `docker-compose.yml` or env vars)
    - `AGENT_INTERFACE` — network interface to monitor (e.g., `eth0`)
    - `AGENT_ENDPOINT_URL` — backend ingest URL (e.g., `http://127.0.0.1:8000/ingest`)

3. Build & Run
    - Find your interface:
      ```bash
      ip addr
      ```
    - Edit `docker-compose.yml`:
      - Set `AGENT_INTERFACE` to your interface.
      - Ensure `AGENT_ENDPOINT_URL` points to the running backend.
    - Build and start (sudo may be required for privileged network access):
      ```bash
      sudo docker-compose up --build
      ```

That's it — the agent will capture traffic, produce `FlowEvent` JSON summaries, and POST them to the configured backend.
