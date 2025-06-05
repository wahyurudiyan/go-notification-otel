# Open Telemetry Context Propagation Example

📡 Distributed Context Propagation Example using OpenTelemetry in Go

## 📘 Overview

This project demonstrates how to propagate trace context across services using [OpenTelemetry](https://opentelemetry.io/) in Go. It shows how HTTP and gRPC-based services can carry trace information such as trace ID and span ID to enable end-to-end observability in a distributed system.

**Context propagation** is a core concept in OpenTelemetry that enables telemetry signals (traces, metrics, logs) to be correlated across process and network boundaries by transmitting trace context (e.g., `traceparent` headers).

## 🎯 Objectives

- Show how to propagate trace context between services (HTTP and gRPC).
- Use OpenTelemetry Go SDK to instrument services.
- Export trace data to Jaeger.
- Implement graceful shutdown for all services.
- Provide reusable modules for setting up telemetry and context handling.

## 🧱 Project Structure

```
├── cmd/ # Main applications
│ ├── client/ # HTTP or gRPC client
│ └── server/ # Server handling requests and propagating context
├── contract/ # Shared definitions (e.g., proto files)
├── pkg/
│ ├── graceful/ # Graceful shutdown helper
│ └── telemetry/ # OpenTelemetry setup
├── telemetry/ # Tracing configuration
├── go.mod
├── go.sum
├── README.md
└── playground.http # Sample HTTP requests for testing
```

## ✅ Graceful Shutdown

The project includes safe shutdown handling using the graceful package. This ensures services flush telemetry data and release resources before terminating.

## 📊 Observability Stack

- Tracing Backend: Jaeger (Not implemented yet)
- Instrumentation: OpenTelemetry SDK for Go
