#!/usr/bin/env sh
set -e

REPO="yusupkhemraev/argus"
BIN_DIR="/usr/local/bin"
CONFIG_DIR="/etc/argus"
SERVICE_NAME="argus"

# ── helpers ────────────────────────────────────────────────────────────────────

info()  { printf '\033[1;34m==>\033[0m %s\n' "$*"; }
ok()    { printf '\033[1;32m OK\033[0m %s\n' "$*"; }
die()   { printf '\033[1;31mERR\033[0m %s\n' "$*" >&2; exit 1; }

need() {
    for cmd in "$@"; do
        command -v "$cmd" >/dev/null 2>&1 || die "required command not found: $cmd"
    done
}

gen_password() {
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -hex 10
    else
        LC_ALL=C tr -dc 'a-zA-Z0-9' < /dev/urandom | dd bs=1 count=20 2>/dev/null
    fi
}

# ── detect platform ────────────────────────────────────────────────────────────

detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    [ "$OS" = "Linux" ] || die "unsupported OS: $OS (only Linux is supported)"

    case "$ARCH" in
        x86_64)          ARCH_KEY="amd64" ;;
        aarch64 | arm64) ARCH_KEY="arm64" ;;
        *) die "unsupported architecture: $ARCH" ;;
    esac

    ASSET="argus-linux-${ARCH_KEY}"
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

    info "Downloading argus ${VERSION} (linux/${ARCH_KEY})..."
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
        CONFIG_CREATED=0
        return
    fi

    WEB_PASSWORD="$(gen_password)"
    CONFIG_CREATED=1

    info "Creating default config at ${CONFIG_DIR}/config.yaml..."
    $SUDO tee "${CONFIG_DIR}/config.yaml" >/dev/null <<EOF
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
  listen: 127.0.0.1:8765
  username: admin
  password: ${WEB_PASSWORD}
EOF
    ok "Default config created"
}

# ── create runtime directories ─────────────────────────────────────────────────

create_dirs() {
    info "Creating runtime directories..."
    $SUDO mkdir -p /var/log/argus
    if [ -n "$SUDO" ]; then
        $SUDO chown "$(id -un)" /var/log/argus 2>/dev/null || true
    fi
    ok "Directories ready"
}

# ── systemd ────────────────────────────────────────────────────────────────────

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

# ── main ───────────────────────────────────────────────────────────────────────

main() {
    CONFIG_CREATED=0
    WEB_PASSWORD=""

    detect_platform

    if [ "$(id -u)" -eq 0 ]; then
        SUDO=""
    else
        need sudo
        SUDO="sudo"
    fi

    fetch_latest_version
    install_binary
    create_dirs
    create_config
    install_systemd

    echo ""
    info "Argus ${VERSION} installed successfully"
    printf '    Config:  %s\n' "${CONFIG_DIR}/config.yaml"
    printf '    Binary:  %s\n' "${BIN_DIR}/argus"
    printf '    Version: %s\n' "$(argus version 2>/dev/null || echo "$VERSION")"
    if [ "$CONFIG_CREATED" -eq 1 ]; then
        echo ""
        printf '\033[1;33m  Web UI credentials\033[0m\n'
        printf '    URL:      http://127.0.0.1:8765\n'
        printf '    Username: admin\n'
        printf '    Password: %s\n' "$WEB_PASSWORD"
    fi
    echo ""
    printf '  Edit the config, then: argus restart\n'
}

main "$@"