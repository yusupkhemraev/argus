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

**Supported platforms:** Linux amd64/arm64

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
    # mentions:
    #   - username

server:
  enabled: true
  listen: 127.0.0.1:8765
  username: admin
  password: changeme
```

Any config field can be overridden with an environment variable using the `ARGUS_` prefix and `_` as separator — e.g. `ARGUS_NOTIFIERS_TELEGRAM_TOKEN=...`.

## Configuration reference

### `name`

Display name of the server, included in every notification. Defaults to the system hostname if empty.

---

### `collectors.memory`

| Field | Default | Description |
|-------|---------|-------------|
| `enabled` | `true` | Enable memory monitoring |
| `threshold` | `85` | Alert when usage exceeds this percentage |
| `severity` | `warning` | Alarm severity: `info`, `warning`, `critical` |
| `interval` | `10m` | How often to check the threshold and send an alarm |
| `refresh` | `5s` | How often to update the metric on the dashboard |

### `collectors.cpu`

Same fields as `memory`. Default threshold `90`, severity `critical`, interval `5m`.

### `collectors.disk`

Same fields as `memory`, plus:

| Field | Default | Description |
|-------|---------|-------------|
| `path` | `/` | Filesystem path to monitor |

Default threshold `80`, severity `warning`, interval `30m`.

---

### `collectors.nginx`

Tails the Nginx access log in real time and fires alarms on HTTP errors and slow requests.

| Field | Default | Description |
|-------|---------|-------------|
| `enabled` | `false` | Enable Nginx monitoring |
| `access_log` | `/var/log/nginx/access.log` | Path to the access log |
| `watch_statuses` | `["500-599"]` | Status codes or ranges to count as errors |
| `threshold` | `10` | Number of matching errors in `window` to trigger an alarm |
| `window` | `60s` | Sliding time window for counting errors |
| `severity` | `critical` | Alarm severity for HTTP errors |
| `slow_threshold` | `7.0` | Request time in seconds to consider a request slow (`rt` or `upstream_response_time`) |
| `slow_count` | `5` | Number of slow requests in `slow_window` to trigger an alarm |
| `slow_window` | `60s` | Sliding window for counting slow requests |
| `min_group_count` | `5` | Minimum hits per endpoint to include it in the alarm message |
| `interval` | `1m` | How often to evaluate thresholds |
| `refresh` | `5s` | How often to update the metric on the dashboard |

**`priority_routes`** — routes that trigger an alarm at a lower hit count, regardless of `threshold`:

```yaml
priority_routes:
  - method: POST          # HTTP method, or "*" for any
    pattern: /api/pay/*   # path pattern, "*" matches any segment
    min_count: 1          # alarm after this many errors on this route
    exclude_statuses:     # statuses to ignore for this route
      - "499"
```

**`ignore_routes`** — routes excluded from all error and slow-request counting:

```yaml
ignore_routes:
  - method: GET           # HTTP method, or "*" for any
    pattern: /healthz
```

---

### `collectors.rabbitmq`

Polls the RabbitMQ management API and fires alarms when queue depth exceeds limits.

| Field | Default | Description |
|-------|---------|-------------|
| `enabled` | `false` | Enable RabbitMQ monitoring |
| `management_url` | `http://localhost:15672` | RabbitMQ management plugin URL |
| `username` | `guest` | Management API username |
| `password` | | Management API password |
| `interval` | `30s` | Poll interval |

**`queues`** — list of queues to watch:

```yaml
queues:
  - name: orders
    threshold: 1000          # alarm when messages > threshold
    unacked_threshold: 50    # alarm when unacked messages > threshold
    severity: warning
```

---

### `notifiers.telegram`

| Field | Description |
|-------|-------------|
| `enabled` | Enable Telegram notifications |
| `token` | Bot token from [@BotFather](https://t.me/BotFather) |
| `chat_id` | Target chat or channel ID (negative for channels) |
| `mentions` | List of usernames to mention in every alert (without `@`) |

### `notifiers.webhook`

Sends a JSON POST (or configurable method) to any URL on each alarm.

| Field | Default | Description |
|-------|---------|-------------|
| `enabled` | `false` | Enable webhook notifications |
| `url` | | Endpoint URL |
| `method` | `POST` | HTTP method |
| `api_key` | | Value for the auth header |
| `api_key_header` | `Authorization` | Header name for `api_key` |
| `headers` | | Map of additional request headers |

Payload shape:

```json
{
  "server": "my-server",
  "collector": "cpu",
  "severity": "critical",
  "message": "CPU 92.3% / 90.0%",
  "value": 92.3,
  "threshold": 90.0,
  "timestamp": "2026-04-26T10:00:00Z"
}
```

### `notifiers.log_path`

Path to a file where every notification attempt is logged (JSON lines). Used by `argus logs` and the web dashboard. Leave empty to disable.

---

### `history`

| Field | Default | Description |
|-------|---------|-------------|
| `enabled` | `true` | Persist alarm history |
| `file_path` | `/var/log/argus/alarms.log` | Path to the history file (JSON lines) |
| `max_size_mb` | `100` | Rotate when the file exceeds this size |
| `max_backups` | `3` | Number of rotated files to keep |

---

### `server`

| Field | Default | Description |
|-------|---------|-------------|
| `enabled` | `true` | Enable the web dashboard |
| `listen` | `127.0.0.1:8765` | Address and port to listen on |
| `username` | | Basic auth username. Auth is disabled when either field is empty |
| `password` | | Basic auth password |

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
argus update           # download and install the latest release
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

## Roadmap

**Notifiers**
- [ ] Email (SMTP)
- [ ] Slack
- [ ] Discord

**Collectors**
- [ ] Kafka — consumer group lag, topic offsets
- [ ] PostgreSQL / MySQL — connections, replication lag, slow queries
- [ ] HTTP / URL — response code, latency, SSL expiry
- [ ] Process — check that a process is running
- [ ] Custom command — run a shell script and check its exit code or output

**Features**
- [ ] Alarm cooldown — suppress repeated notifications for the same alarm
- [ ] Tests

## License

[MIT](LICENSE)