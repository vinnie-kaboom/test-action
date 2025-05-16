package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	PlaybookPath  string
	InventoryPath string
	WatchInterval int
	WatchRepos    []string
	Branch        string
	GitHubToken   string
	GitHubUser    string
}

func (c *Config) Validate() error {
	if c.PlaybookPath == "" {
		return fmt.Errorf("playbook path is required")
	}
	if !fileExists(c.PlaybookPath) {
		return fmt.Errorf("playbook file does not exist: %s", c.PlaybookPath)
	}
	if c.InventoryPath != "" && !fileExists(c.InventoryPath) {
		return fmt.Errorf("inventory file does not exist: %s", c.InventoryPath)
	}
	if c.WatchInterval < 10 {
		return fmt.Errorf("watch interval must be at least 10 seconds")
	}
	if len(c.WatchRepos) == 0 {
		return fmt.Errorf("at least one repository must be specified")
	}
	if c.GitHubToken == "" {
		return fmt.Errorf("GitHub token is required")
	}
	if len(c.GitHubToken) < 40 {
		return fmt.Errorf("invalid GitHub token format")
	}
	return nil
}

func main() {
	fmt.Println("Starting Ansible GitOps Service...")
	fmt.Printf("Current working directory: %s\n", getCurrentDir())

	// Define command line flags
	playbookPath := flag.String("playbook", "", "Path to the Ansible playbook")
	inventoryPath := flag.String("inventory", "", "Path to the Ansible inventory")
	watchInterval := flag.Int("interval", 300, "Interval in seconds to check for changes")
	watchRepos := flag.String("repos", "", "Comma-separated list of repositories to watch")
	branch := flag.String("branch", "main", "Branch to watch for changes")
	githubToken := flag.String("github-token", "", "GitHub Personal Access Token")
	githubUser := flag.String("github-user", "", "GitHub username")

	flag.Parse()

	// Create config from flags
	config := Config{
		PlaybookPath:  *playbookPath,
		InventoryPath: *inventoryPath,
		WatchInterval: *watchInterval,
		Branch:        *branch,
		GitHubToken:   *githubToken,
		GitHubUser:    *githubUser,
	}

	// Parse watch repos
	if *watchRepos != "" {
		config.WatchRepos = strings.Split(*watchRepos, ",")
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Playbook Path: %s\n", config.PlaybookPath)
	fmt.Printf("  Inventory Path: %s\n", config.InventoryPath)
	fmt.Printf("  Watch Interval: %d seconds\n", config.WatchInterval)
	fmt.Printf("  Watch Repositories: %v\n", config.WatchRepos)
	fmt.Printf("  Watch Branch: %s\n", config.Branch)
	fmt.Printf("  GitHub User: %s\n", config.GitHubUser)

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Printf("Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	// Check if Ansible is available
	if err := checkAnsibleAvailability(); err != nil {
		fmt.Printf("Error: Ansible is not available: %v\n", err)
		os.Exit(1)
	}

	// Configure Git with GitHub credentials
	if err := configureGit(config); err != nil {
		fmt.Printf("Error configuring Git: %v\n", err)
		os.Exit(1)
	}

	// Start watching for changes
	watchForChanges(config)
}

func checkAnsibleAvailability() error {
	// Check if ansible-playbook is available
	cmd := exec.Command("ansible-playbook", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ansible-playbook not found: %v", err)
	}

	// Parse version output
	version := strings.TrimSpace(string(output))
	fmt.Printf("Ansible version: %s\n", version)

	// Check if ansible-galaxy is available
	cmd = exec.Command("ansible-galaxy", "--version")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ansible-galaxy not found: %v", err)
	}

	fmt.Printf("Ansible Galaxy version: %s\n", strings.TrimSpace(string(output)))
	return nil
}

func configureGit(config Config) error {
	// Configure Git to use the token for authentication
	commands := []struct {
		cmd  string
		args []string
	}{
		{"git", []string{"config", "--global", "credential.helper", "store"}},
		{"git", []string{"config", "--global", "user.name", config.GitHubUser}},
		{"git", []string{"config", "--global", "url.https://" + config.GitHubToken + "@github.com/.insteadOf", "https://github.com/"}},
	}

	for _, cmd := range commands {
		execCmd := exec.Command(cmd.cmd, cmd.args...)
		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("error running git config: %v", err)
		}
	}

	return nil
}

func watchForChanges(config Config) {
	// Initialize repository states
	repoStates := make(map[string]string)

	// Get initial states for all repositories
	for _, repo := range config.WatchRepos {
		hash, err := getGitHash(repo, config.Branch)
		if err != nil {
			fmt.Printf("Error getting initial hash for %s: %v\n", repo, err)
			continue
		}
		repoStates[repo] = hash
		fmt.Printf("Initial git hash for %s (branch: %s): %s\n", repo, config.Branch, hash)
	}

	for {
		// Periodically check Ansible availability
		if err := checkAnsibleAvailability(); err != nil {
			fmt.Printf("Warning: Ansible check failed: %v\n", err)
			time.Sleep(time.Duration(config.WatchInterval) * time.Second)
			continue
		}

		changesDetected := false

		// Check each repository for changes
		for repo, lastHash := range repoStates {
			currentHash, err := getGitHash(repo, config.Branch)
			if err != nil {
				fmt.Printf("Error checking %s: %v\n", repo, err)
				continue
			}

			if currentHash != lastHash {
				fmt.Printf("Detected changes in repository %s (branch: %s). Old hash: %s, New hash: %s\n",
					repo, config.Branch, lastHash, currentHash)

				// Pull latest changes
				if err := pullLatestChanges(repo, config.Branch); err != nil {
					fmt.Printf("Error pulling changes for %s: %v\n", repo, err)
					continue
				}

				changesDetected = true
				repoStates[repo] = currentHash
			}
		}

		// If any repository had changes, run the playbook
		if changesDetected {
			if err := runPlaybook(config); err != nil {
				fmt.Printf("Error running playbook: %v\n", err)
			} else {
				fmt.Println("Playbook executed successfully")
			}
		}

		// Wait for the specified interval
		time.Sleep(time.Duration(config.WatchInterval) * time.Second)
	}
}

func getGitHash(repoPath, branch string) (string, error) {
	// First ensure we're on the correct branch
	checkoutCmd := exec.Command("git", "-C", repoPath, "checkout", branch)
	if err := checkoutCmd.Run(); err != nil {
		return "", fmt.Errorf("error checking out branch %s: %v", branch, err)
	}

	// Then get the hash
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting git hash: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func pullLatestChanges(repoPath, branch string) error {
	// Ensure we're on the correct branch
	checkoutCmd := exec.Command("git", "-C", repoPath, "checkout", branch)
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("error checking out branch %s: %v", branch, err)
	}

	// Pull changes
	cmd := exec.Command("git", "-C", repoPath, "pull", "origin", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runPlaybook(config Config) error {
	// Construct ansible-playbook command
	cmd := exec.Command("ansible-playbook", config.PlaybookPath)

	// Add inventory if provided
	if config.InventoryPath != "" {
		cmd.Args = append(cmd.Args, "-i", config.InventoryPath)
	}

	fmt.Printf("Executing command: %v\n", cmd.Args)

	// Set up output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	return cmd.Run()
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "Error getting current directory"
	}
	return dir
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func runCommandWithTimeout(cmd *exec.Cmd, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd.WaitDelay = timeout
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %v", err)
		}
		return fmt.Errorf("command timed out")
	}
}
