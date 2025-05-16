# Ansible GitOps Service

A service that watches Git repositories for changes and runs Ansible playbooks when changes are detected.

## Features

- Monitors specified Git repositories for changes
- Runs Ansible playbooks when changes are detected
- Supports GitHub authentication
- Runs as a systemd service
- Configurable watch interval
- Supports multiple repositories

## Prerequisites

- Go 1.16 or later
- Ansible installed
- Git installed
- GitHub Personal Access Token with `repo` scope

## Installation

1. Clone the repository:
```bash
git clone https://github.com/vinnie-kaboom/ansible-gitops-test.git
cd ansible-gitops-test
```

2. Build the service:
```bash
go build -o ansible-gitops
```

3. Install the service:
```bash
chmod +x install.sh
./install.sh
```

4. Configure the service:
Edit `/etc/systemd/system/ansible-gitops.service` and update:
- GitHub token
- Repository paths
- Playbook path
- Watch interval

5. Start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl start ansible-gitops
```

## Configuration

The service can be configured through command-line arguments:

- `--playbook`: Path to the Ansible playbook
- `--inventory`: (Optional) Path to the Ansible inventory
- `--interval`: Watch interval in seconds (default: 300)
- `--repos`: Comma-separated list of repositories to watch
- `--branch`: Branch to watch (default: main)
- `--github-token`: GitHub Personal Access Token
- `--github-user`: GitHub username

## Usage

1. Create your Ansible playbook in the watched repository
2. The service will automatically detect changes
3. When changes are detected, it will:
   - Pull the latest changes
   - Run the playbook
   - Log the results

## Service Management

Check service status:
```bash
sudo systemctl status ansible-gitops
```

View logs:
```bash
journalctl -u ansible-gitops -f
```

Restart service:
```bash
sudo systemctl restart ansible-gitops
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

## Security Notes

- Never commit GitHub tokens to the repository
- Use environment variables or a secrets manager in production
- Set appropriate token permissions (least privilege)
- Consider using a dedicated service account

## License

MIT License