package main

import (
	"fmt"
	"os"
	"os/exec"
)

type Inputs struct {
	PlaybookPath  string `json:"playbook_path"`
	InventoryPath string `json:"inventory_path"`
	ExtraVars     string `json:"extra_vars"`
}

func main() {
	fmt.Println("Starting Ansible Action...")
	fmt.Printf("Current working directory: %s\n", getCurrentDir())

	// Read inputs from environment variables
	inputs := Inputs{
		PlaybookPath:  os.Getenv("INPUT_PLAYBOOK_PATH"),
		InventoryPath: os.Getenv("INPUT_INVENTORY_PATH"),
		ExtraVars:     os.Getenv("INPUT_EXTRA_VARS"),
	}

	fmt.Printf("Received inputs:\n")
	fmt.Printf("  Playbook Path: %s\n", inputs.PlaybookPath)
	fmt.Printf("  Inventory Path: %s\n", inputs.InventoryPath)
	fmt.Printf("  Extra Vars: %s\n", inputs.ExtraVars)

	// Validate inputs
	if inputs.PlaybookPath == "" {
		fmt.Println("Error: playbook_path is required")
		os.Exit(1)
	}

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
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running ansible-playbook: %v\n", err)
		os.Exit(1)
	}

	// Set output
	fmt.Printf("::set-output name=result::success\n")
	fmt.Println("Ansible Action completed successfully")
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "Error getting current directory"
	}
	return dir
}
