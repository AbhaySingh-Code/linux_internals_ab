#!/bin/bash

# Exit immediately if any command fails
set -e

# --- CONFIGURATION ---
APP_NAME="good_program"
SERVICE_FILE="/etc/systemd/system/$APP_NAME.service"
TIMER_FILE="/etc/systemd/system/$APP_NAME.timer"

# --- CHECK PRIVILEGES ---
if [ "$EUID" -ne 0 ]; then
  echo "[-] Please run this script with sudo or as root."
  exit 1
fi

# --- 1. STOP AND DISABLE UNITS ---
echo "[+] Stopping and disabling systemd timer..."
systemctl stop "$APP_NAME.timer" 2>/dev/null || true
systemctl disable "$APP_NAME.timer" 2>/dev/null || true

echo "[+] Stopping and disabling systemd service..."
systemctl stop "$APP_NAME.service" 2>/dev/null || true
systemctl disable "$APP_NAME.service" 2>/dev/null || true

# --- 2. REMOVE CONFIGURATION FILES ---
echo "[+] Removing systemd configuration files..."
if [ -f "$TIMER_FILE" ]; then
  rm "$TIMER_FILE"
  echo "[*] Removed: $TIMER_FILE"
fi

if [ -f "$SERVICE_FILE" ]; then
  rm "$SERVICE_FILE"
  echo "[*] Removed: $SERVICE_FILE"
fi

#----- Remove binary from /usr/loca/bin
rm "/usr/local/bin/$APP_NAME" 2 > /dev/null || true

# --- 3. RELOAD DAEMON ---
echo "[+] Reloading systemd manager configuration..."
systemctl daemon-reload
systemctl reset-failed

echo "[+] Clean-up complete! The timer and service have been completely removed."
