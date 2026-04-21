# Argus

Lightweight server monitoring tool with a web dashboard and Telegram alerts.

Monitors CPU, memory, disk, Nginx access logs, and RabbitMQ queues. Sends alerts when thresholds are exceeded. Ships as a single binary with an embedded web UI.

## Features

- **System metrics** — CPU, memory, disk usage with configurable thresholds
- **Nginx monitoring** — error rate, slow requests, priority routes
- **RabbitMQ** — queue depth and unacked message tracking
- **Telegram alerts** — instant notifications with mentions
- **Web dashboard** — real-time metrics, alarm history, notification logs
- **Single binary** — web UI embedded, no external dependencies

## Installation

```sh
curl -fsSL https://raw.githubusercontent.com/yusupkhemraev/argus/main/install.sh | sh
```

This installs the binary to `/usr/local/bin/argus`, creates a default config at `/etc/argus/config.yaml`, and registers a system service (systemd on Linux, launchd on macOS).

**Supported platforms:** Linux amd64/arm64, macOS amd64/arm64

## Configuration

Edit `/etc/argus/config.yaml` after installation:

```yaml
name: my-server

collectors:
  memory:
    enabled: true
    threshold: 85
    severity: warning
  cpu:
    enabled: true
    threshold: 90
    severity: critical
  disk:
    enabled: true
    path: /
    threshold: 80

notifiers:
  telegram:
    enabled: true
    token: "your-bot-token"
    chat_id: -100123456789

server:
  enabled: true
  listen: 127.0.0.1:8080
```

See [config.yaml.example](config.yaml.example) for all available options.

## Service management

```sh
argus start      # start the service
argus stop       # stop the service
argus restart    # restart the service
argus service    # show service status
```

## CLI

```sh
argus status           # current system metrics
argus history          # recent alarm history
argus history -n 50 -severity critical
argus logs             # notification log
argus version
```

## Building from source

```sh
git clone https://github.com/yusupkhemraev/argus
cd argus
make build             # current platform
make build-all         # linux + darwin, amd64 + arm64
```

Requires Go 1.21+ and Node.js 20+.

## License

[MIT](LICENSE)