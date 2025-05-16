#!/bin/bash

# Build the Go binary
go build -o ansible-gitops

# Install the binary
sudo cp ansible-gitops /usr/local/bin/

# Install the service
sudo cp ansible-gitops.service /etc/systemd/system/

# Reload systemd
sudo systemctl daemon-reload

# Enable and start the service
sudo systemctl enable ansible-gitops
sudo systemctl start ansible-gitops

# Check status
sudo systemctl status ansible-gitops 