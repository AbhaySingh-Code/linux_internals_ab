#!/bin/bash

# Exit immediately if any command fails
set -e

# --- CONFIGURATION ---
APP_NAME="good_program"
GO_ENTRY_FILE="main.go"
BINARY_DEST="/usr/local/bin/$APP_NAME"
SERVICE_FILE="/etc/systemd/system/$APP_NAME.service"
TIMER_FILE="/etc/systemd/system/$APP_NAME.timer"
RUN_USER="nobody"
RUN_GROUP="nogroup"

# --- CHECK PRIVILEGES ---
if [ "$EUID" -ne 0 ]; then
  echo "[-] Please run this script with sudo or as root."
  exit 1
fi

# --- 1. COMPILE GO APPLICATION ---
# echo "[+] Compiling Go application..."
# if [ ! -f "$GO_ENTRY_FILE" ]; then
#   echo "[-] Error: $GO_ENTRY_FILE not found in the current directory."
#   exit 1
# fi
# go build -ldflags="-s -w" -o "$APP_NAME" "$GO_ENTRY_FILE"

# --- 2. MOVE BINARY TO SYSTEM PATH ---
echo "[+] Moving binary to $BINARY_DEST..."
mv "$APP_NAME" "$BINARY_DEST"
chmod +x "$BINARY_DEST"

# --- 3. CREATE SYSTEMD SERVICE FILE ---
# Note: For timers, we remove the [Install] section from the service file.
# The timer handles the activation, not the service itself.
echo "[+] Creating systemd service file..."
cat << EOF > "$SERVICE_FILE"
[Unit]
Description=Go Application Service Triggered by Timer
After=network.target

[Service]
Type=oneshot
User=$RUN_USER
Group=$RUN_GROUP
ExecStart=$BINARY_DEST
WorkingDirectory=/usr/local/bin
EOF

# --- 4. CREATE SYSTEMD TIMER FILE ---
echo "[+] Creating systemd timer file..."
cat << EOF > "$TIMER_FILE"
[Unit]
Description=Runs My Go Application on a Schedule

[Timer]
OnBootSec=1min
OnUnitActiveSec=2min
AccuracySec=1s

[Install]
WantedBy=timers.target
EOF

# --- 5. START AND ENABLE THE TIMER ---
echo "[+] Reloading systemd daemon..."
systemctl daemon-reload

echo "[+] Disabling direct service boot (timer will manage it)..."
systemctl disable "$APP_NAME.service" 2>/dev/null || true

echo "[+] Enabling and starting the timer..."
systemctl enable "$APP_NAME.timer"
systemctl restart "$APP_NAME.timer"

# --- 6. VERIFY DEPLOYMENT ---
echo "[+] Deployment complete! Checking timer status..."
echo "------------------------------------------------"
systemctl list-timers --all | grep "$APP_NAME" || echo "Timer active but pending next execution."
echo "------------------------------------------------"
echo "[*] To view logs of execution, run: sudo journalctl -u $APP_NAME.service -f"
