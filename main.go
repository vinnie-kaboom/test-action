package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Inputs struct {
	PlaybookPath  string   `json:"playbook_path"`
	InventoryPath string   `json:"inventory_path"`
	ExtraVars     string   `json:"extra_vars"`
	WatchInterval int      `json:"watch_interval"` // in seconds
	WatchRepos    []string `json:"watch_repos"`    // list of repositories to watch
}

type RepoState struct {
	Path string
	Hash string
}

func main() {
	fmt.Println("Starting Ansible GitOps Action...")
	fmt.Printf("Current working directory: %s\n", getCurrentDir())

	// Read inputs from environment variables
	inputs := Inputs{
		PlaybookPath:  os.Getenv("INPUT_PLAYBOOK_PATH"),
		InventoryPath: os.Getenv("INPUT_INVENTORY_PATH"),
		ExtraVars:     os.Getenv("INPUT_EXTRA_VARS"),
		WatchInterval: 300, // default 5 minutes
	}

	// Parse watch repos from environment
	if watchRepos := os.Getenv("INPUT_WATCH_REPOS"); watchRepos != "" {
		inputs.WatchRepos = strings.Split(watchRepos, ",")
	}

	fmt.Printf("Received inputs:\n")
	fmt.Printf("  Playbook Path: %s\n", inputs.PlaybookPath)
	fmt.Printf("  Inventory Path: %s\n", inputs.InventoryPath)
	fmt.Printf("  Extra Vars: %s\n", inputs.ExtraVars)
	fmt.Printf("  Watch Interval: %d seconds\n", inputs.WatchInterval)
	fmt.Printf("  Watch Repositories: %v\n", inputs.WatchRepos)

	// Validate inputs
	if inputs.PlaybookPath == "" {
		fmt.Println("Error: playbook_path is required")
		os.Exit(1)
	}

	// Start watching for changes
	watchForChanges(inputs)
}

func watchForChanges(inputs Inputs) {
	// Initialize repository states
	repoStates := make(map[string]RepoState)

	// Get initial states for all repositories
	for _, repo := range inputs.WatchRepos {
		hash, err := getGitHash(repo)
		if err != nil {
			fmt.Printf("Error getting initial hash for %s: %v\n", repo, err)
			continue
		}
		repoStates[repo] = RepoState{Path: repo, Hash: hash}
		fmt.Printf("Initial git hash for %s: %s\n", repo, hash)
	}

	for {
		changesDetected := false

		// Check each repository for changes
		for repo, state := range repoStates {
			currentHash, err := getGitHash(repo)
			if err != nil {
				fmt.Printf("Error checking %s: %v\n", repo, err)
				continue
			}

			if currentHash != state.Hash {
				fmt.Printf("Detected changes in repository %s. Old hash: %s, New hash: %s\n",
					repo, state.Hash, currentHash)

				// Pull latest changes
				if err := pullLatestChanges(repo); err != nil {
					fmt.Printf("Error pulling changes for %s: %v\n", repo, err)
					continue
				}

				changesDetected = true
				repoStates[repo] = RepoState{Path: repo, Hash: currentHash}
			}
		}

		// If any repository had changes, run the playbook
		if changesDetected {
			if err := runPlaybook(inputs); err != nil {
				fmt.Printf("Error running playbook: %v\n", err)
			} else {
				fmt.Println("Playbook executed successfully")
			}
		}

		// Wait for the specified interval
		time.Sleep(time.Duration(inputs.WatchInterval) * time.Second)
	}
}

func getGitHash(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting git hash: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func pullLatestChanges(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runPlaybook(inputs Inputs) error {
	// Construct ansible-playbook command
	cmd := exec.Command("ansible-playbook", inputs.PlaybookPath)

	// Add inventory if provided
	if inputs.InventoryPath != "" {
		cmd.Args = append(cmd.Args, "-i", inputs.InventoryPath)
	}

	// Add extra vars if provided
	if inputs.ExtraVars != "" {
		cmd.Args = append(cmd.Args, "--extra-vars", inputs.ExtraVars)
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
