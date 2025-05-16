# Ansible GitOps Service

This service continuously monitors Git repositories and runs Ansible playbooks when changes are detected.

## Installation

1. Clone this repository:
```bash
git clone https://github.com/yourusername/test-action.git
cd test-action
```

2. Make the installation script executable:
```bash
chmod +x install.sh
```

3. Run the installation script:
```bash
./install.sh
```

## Service Management

- Check service status:
```bash
sudo systemctl status ansible-gitops
```

- View logs:
```bash
sudo journalctl -u ansible-gitops -f
```

- Stop the service:
```bash
sudo systemctl stop ansible-gitops
```

- Start the service:
```bash
sudo systemctl start ansible-gitops
```

- Restart the service:
```bash
sudo systemctl restart ansible-gitops
```

## Configuration

The service is configured through the systemd service file (`ansible-gitops.service`). You can modify the following environment variables:

- `INPUT_PLAYBOOK_PATH`: Path to your Ansible playbook
- `INPUT_INVENTORY_PATH`: Path to your Ansible inventory
- `INPUT_WATCH_INTERVAL`: Interval in seconds to check for changes
- `INPUT_WATCH_REPOS`: Comma-separated list of repositories to watch

After modifying the service file, reload the configuration:
```bash
sudo systemctl daemon-reload
sudo systemctl restart ansible-gitops
```

## How it Works

1. The service runs as a systemd service, ensuring it:
   - Starts automatically on boot
   - Restarts if it crashes
   - Runs continuously in the background

2. It monitors the specified Git repositories for changes

3. When changes are detected:
   - Pulls the latest changes
   - Runs the specified Ansible playbook
   - Logs all activities

## Logs

View the service logs:
```bash
sudo journalctl -u ansible-gitops -f
```

## Uninstallation

To remove the service:
```bash
sudo systemctl stop ansible-gitops
sudo systemctl disable ansible-gitops
sudo rm /etc/systemd/system/ansible-gitops.service
sudo rm /usr/local/bin/ansible-gitops
sudo systemctl daemon-reload
```