# Ansible GitHub Action

This GitHub Action allows you to run Ansible playbooks in your GitHub workflows.

## Setup Instructions

### 1. Setting up a Self-hosted Runner

1. Go to your GitHub repository
2. Click on "Settings" > "Actions" > "Runners"
3. Click "New self-hosted runner"
4. Choose your operating system and architecture
5. Follow the instructions provided to set up the runner on your machine:

```bash
# Download the runner package
curl -o actions-runner-linux-x64-2.311.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-linux-x64-2.311.0.tar.gz

# Extract the installer
tar xzf ./actions-runner-linux-x64-2.311.0.tar.gz

# Configure the runner
./config.sh --url https://github.com/YOUR-USERNAME/YOUR-REPO --token YOUR-TOKEN

# Install the runner service
./svc.sh install

# Start the runner service
./svc.sh start
```

### 2. Using the Action

Add the following to your workflow file:

```yaml
- name: Run Ansible Playbook
  uses: ./
  with:
    playbook_path: path/to/playbook.yml
    inventory_path: path/to/inventory.yml  # optional
    extra_vars: '{"key": "value"}'        # optional
```

## Inputs

- `playbook_path`: Path to the Ansible playbook (required)
- `inventory_path`: Path to the Ansible inventory file (optional)
- `extra_vars`: Extra variables to pass to Ansible (optional)

## Outputs

- `result`: Result of the playbook execution

## Development

To build and test the action locally:

```bash
# Build the Docker image
docker build -t ansible-action .

# Run the action
docker run -e INPUT_PLAYBOOK_PATH=playbook.yml ansible-action
```