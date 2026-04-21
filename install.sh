#!/usr/bin/env sh
set -e

REPO="yusupkhemraev/argus"
BIN_DIR="/usr/local/bin"
CONFIG_DIR="/etc/argus"
SERVICE_NAME="argus"
MAC_PLIST_LABEL="io.github.yusupkhemraev.argus"
MAC_PLIST_PATH="$HOME/Library/LaunchAgents/io.github.yusupkhemraev.argus.plist"

# ── helpers ────────────────────────────────────────────────────────────────────

info()  { printf '\033[1;34m==>\033[0m %s\n' "$*"; }
ok()    { printf '\033[1;32m OK\033[0m %s\n' "$*"; }
die()   { printf '\033[1;31mERR\033[0m %s\n' "$*" >&2; exit 1; }

need() {
    for cmd in "$@"; do
        command -v "$cmd" >/dev/null 2>&1 || die "required command not found: $cmd"
    done
}

# ── detect platform ────────────────────────────────────────────────────────────

detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux)  OS_KEY="linux"  ;;
        Darwin) OS_KEY="darwin" ;;
        *) die "unsupported OS: $OS" ;;
    esac

    case "$ARCH" in
        x86_64)          ARCH_KEY="amd64" ;;
        aarch64 | arm64) ARCH_KEY="arm64" ;;
        *) die "unsupported architecture: $ARCH" ;;
    esac

    ASSET="argus-${OS_KEY}-${ARCH_KEY}"
}

# ── fetch latest release ───────────────────────────────────────────────────────

fetch_latest_version() {
    need curl

    API_URL="https://api.github.com/repos/${REPO}/releases/latest"
    VERSION="$(curl -fsSL "$API_URL" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"

    [ -n "$VERSION" ] || die "failed to fetch latest release version"
}

# ── download & install binary ──────────────────────────────────────────────────

install_binary() {
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

    info "Downloading argus ${VERSION} (${OS_KEY}/${ARCH_KEY})..."
    TMP="$(mktemp)"
    curl -fsSL -o "$TMP" "$DOWNLOAD_URL" || die "download failed: $DOWNLOAD_URL"
    chmod +x "$TMP"

    info "Installing binary to ${BIN_DIR}/argus..."
    $SUDO mv "$TMP" "${BIN_DIR}/argus"
    ok "Binary installed"
}

# ── create default config ──────────────────────────────────────────────────────

create_config() {
    $SUDO mkdir -p "$CONFIG_DIR"

    if [ -f "${CONFIG_DIR}/config.yaml" ]; then
        ok "Config already exists at ${CONFIG_DIR}/config.yaml — skipping"
        return
    fi

    info "Creating default config at ${CONFIG_DIR}/config.yaml..."
    $SUDO tee "${CONFIG_DIR}/config.yaml" >/dev/null <<'EOF'
name: ""

collectors:
  memory:
    enabled: true
    threshold: 85.0
    severity: warning
  cpu:
    enabled: true
    threshold: 90.0
    severity: critical
  disk:
    enabled: true
    path: /
    threshold: 80.0
    severity: warning
  nginx:
    enabled: false
    access_log: /var/log/nginx/access.log

notifiers:
  telegram:
    enabled: false
    token: ""
    chat_id: 0

server:
  enabled: true
  listen: 127.0.0.1:8080
EOF
    ok "Default config created — edit ${CONFIG_DIR}/config.yaml before starting"
}

# ── systemd (Linux) ────────────────────────────────────────────────────────────

install_systemd() {
    info "Installing systemd service..."

    $SUDO tee /etc/systemd/system/argus.service >/dev/null <<EOF
[Unit]
Description=Argus Server Monitor
After=network.target

[Service]
Type=simple
ExecStart=${BIN_DIR}/argus
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    $SUDO systemctl daemon-reload
    $SUDO systemctl enable "$SERVICE_NAME"
    $SUDO systemctl start  "$SERVICE_NAME"
    ok "Service started (systemctl status argus)"
}

# ── launchd (macOS) ────────────────────────────────────────────────────────────

install_launchd() {
    info "Installing launchd service..."

    mkdir -p "$HOME/Library/LaunchAgents"

    cat >"$MAC_PLIST_PATH" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>              <string>${MAC_PLIST_LABEL}</string>
  <key>ProgramArguments</key>  <array><string>${BIN_DIR}/argus</string></array>
  <key>RunAtLoad</key>          <true/>
  <key>KeepAlive</key>          <true/>
  <key>StandardOutPath</key>    <string>/tmp/argus.log</string>
  <key>StandardErrorPath</key>  <string>/tmp/argus.log</string>
</dict>
</plist>
EOF

    launchctl unload "$MAC_PLIST_PATH" 2>/dev/null || true
    launchctl load -w "$MAC_PLIST_PATH"
    ok "Service started (launchctl list | grep argus)"
}

# ── main ───────────────────────────────────────────────────────────────────────

main() {
    detect_platform

    # use sudo only when needed
    if [ "$(id -u)" -eq 0 ]; then
        SUDO=""
    else
        need sudo
        SUDO="sudo"
    fi

    fetch_latest_version
    install_binary
    create_config

    case "$OS_KEY" in
        linux)  install_systemd ;;
        darwin) install_launchd ;;
    esac

    echo ""
    info "Argus ${VERSION} installed successfully"
    printf '    Config:  %s\n' "${CONFIG_DIR}/config.yaml"
    printf '    Binary:  %s\n' "${BIN_DIR}/argus"
    printf '    Version: %s\n' "$(argus version 2>/dev/null || echo "$VERSION")"
    echo ""
    printf '  Edit the config, then: argus restart\n'
}

main "$@"