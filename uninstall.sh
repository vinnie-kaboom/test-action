#!/bin/bash

# Exit on error
set -e

echo "=== Ansible GitOps Service Uninstaller ==="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (use sudo)"
    exit 1
fi

echo "Stopping ansible-gitops service..."
systemctl stop ansible-gitops || true

echo "Disabling ansible-gitops service..."
systemctl disable ansible-gitops || true

echo "Removing service file..."
rm -f /etc/systemd/system/ansible-gitops.service

echo "Removing binary..."
rm -f /usr/local/bin/ansible-gitops

echo "Reloading systemd..."
systemctl daemon-reload

echo "Cleaning up Git configuration..."
# Remove the GitHub token from Git config
git config --global --unset url.https://github.com/.insteadOf || true

echo "Verifying service removal..."
if ! systemctl status ansible-gitops 2>/dev/null; then
    echo "✅ Service successfully removed"
else
    echo "❌ Service removal may have failed"
    exit 1
fi

echo "=== Uninstallation complete ===" 