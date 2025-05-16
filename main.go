package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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

func (c *Config) LoadFromEnv() {
	if playbook := os.Getenv("ANSIBLE_PLAYBOOK"); playbook != "" {
		c.PlaybookPath = playbook
	}
	if inventory := os.Getenv("ANSIBLE_INVENTORY"); inventory != "" {
		c.InventoryPath = inventory
	}
	if interval := os.Getenv("WATCH_INTERVAL"); interval != "" {
		if val, err := strconv.Atoi(interval); err == nil {
			c.WatchInterval = val
		}
	}
	if repos := os.Getenv("WATCH_REPOS"); repos != "" {
		c.WatchRepos = strings.Split(repos, ",")
	}
	if branch := os.Getenv("WATCH_BRANCH"); branch != "" {
		c.Branch = branch
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		c.GitHubToken = token
	}
	if user := os.Getenv("GITHUB_USER"); user != "" {
		c.GitHubUser = user
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Ansible GitOps Service...")
	log.Printf("Current working directory: %s\n", getCurrentDir())

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

	// Load configuration from environment variables
	config.LoadFromEnv()

	log.Printf("Configuration:\n")
	log.Printf("  Playbook Path: %s\n", config.PlaybookPath)
	log.Printf("  Inventory Path: %s\n", config.InventoryPath)
	log.Printf("  Watch Interval: %d seconds\n", config.WatchInterval)
	log.Printf("  Watch Repositories: %v\n", config.WatchRepos)
	log.Printf("  Watch Branch: %s\n", config.Branch)
	log.Printf("  GitHub User: %s\n", config.GitHubUser)

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Check if Ansible is available
	if err := checkAnsibleAvailability(); err != nil {
		log.Fatalf("Ansible not available: %v", err)
	}

	// Configure Git with GitHub credentials
	if err := configureGit(config); err != nil {
		log.Fatalf("Git configuration error: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start watching for changes in a goroutine
	go watchForChanges(ctx, config)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Received shutdown signal, cleaning up...")
	cancel()
	time.Sleep(time.Second) // Give time for cleanup
	log.Println("Service stopped")
}

func checkAnsibleAvailability() error {
	// Check if ansible-playbook is available
	cmd := exec.Command("ansible-playbook", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ansible-playbook not found: %v", err)
	}

	version := strings.TrimSpace(string(output))
	log.Printf("Ansible version: %s", version)

	// Check if ansible-galaxy is available
	cmd = exec.Command("ansible-galaxy", "--version")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ansible-galaxy not found: %v", err)
	}

	log.Printf("Ansible Galaxy version: %s", strings.TrimSpace(string(output)))
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

func watchForChanges(ctx context.Context, config Config) {
	repoStates := make(map[string]string)

	// Get initial states
	for _, repo := range config.WatchRepos {
		hash, err := getGitHash(repo, config.Branch)
		if err != nil {
			log.Printf("Error getting initial hash for %s: %v", repo, err)
			continue
		}
		repoStates[repo] = hash
		log.Printf("Initial git hash for %s (branch: %s): %s", repo, config.Branch, hash)
	}

	ticker := time.NewTicker(time.Duration(config.WatchInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping watch loop")
			return
		case <-ticker.C:
			log.Printf("=== Checking repositories for changes (interval: %d seconds) ===", config.WatchInterval)

			if err := checkAnsibleAvailability(); err != nil {
				log.Printf("Warning: Ansible check failed: %v", err)
				continue
			}

			changesDetected := false

			for repo, lastHash := range repoStates {
				select {
				case <-ctx.Done():
					return
				default:
					log.Printf("Checking repository: %s (branch: %s)", repo, config.Branch)

					// Pull latest changes before checking hash
					if err := pullLatestChanges(repo, config.Branch); err != nil {
						log.Printf("Error pulling changes for %s: %v", repo, err)
						continue
					}

					currentHash, err := getGitHash(repo, config.Branch)
					if err != nil {
						log.Printf("Error checking %s: %v", repo, err)
						continue
					}

					if currentHash != lastHash {
						log.Printf("🔔 Detected changes in repository %s (branch: %s)", repo, config.Branch)
						log.Printf("   Old hash: %s", lastHash)
						log.Printf("   New hash: %s", currentHash)

						changesDetected = true
						repoStates[repo] = currentHash
					} else {
						log.Printf("✓ No changes detected in %s", repo)
					}
				}
			}

			if changesDetected {
				log.Println("=== Changes detected, running playbook ===")
				if err := runPlaybook(config); err != nil {
					log.Printf("Error running playbook: %v", err)
				} else {
					log.Println("✅ Playbook executed successfully")
				}
			} else {
				log.Println("=== No changes detected in any repository ===")
			}
		}
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
	cmd := exec.Command("ansible-playbook", config.PlaybookPath)

	if config.InventoryPath != "" {
		cmd.Args = append(cmd.Args, "-i", config.InventoryPath)
	}

	log.Printf("Executing command: %v", cmd.Args)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return runCommandWithTimeout(cmd, 30*time.Minute)
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
