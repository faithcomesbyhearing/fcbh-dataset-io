package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("=== Testing So-VITS-SVC Import ===\n")

	// Check environment variable
	soVitsRoot := os.Getenv("SO_VITS_SVC_ROOT")
	if soVitsRoot == "" {
		fmt.Println("ERROR: SO_VITS_SVC_ROOT environment variable not set")
		fmt.Println("Set it to the path where so-vits-svc is cloned, e.g.:")
		fmt.Println("  export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc")
		os.Exit(1)
	}

	fmt.Printf("SO_VITS_SVC_ROOT: %s\n", soVitsRoot)

	// Check if path exists
	if _, err := os.Stat(soVitsRoot); os.IsNotExist(err) {
		fmt.Printf("ERROR: SO_VITS_SVC_ROOT path does not exist: %s\n", soVitsRoot)
		os.Exit(1)
	}

	fmt.Printf("✓ Path exists\n")

	// Find Python script
	goproj := os.Getenv("GOPROJ")
	if goproj == "" {
		// Try to infer from current directory
		cwd, _ := os.Getwd()
		// Look for arti directory
		for {
			if filepath.Base(cwd) == "arti" {
				goproj = cwd
				break
			}
			parent := filepath.Dir(cwd)
			if parent == cwd {
				break
			}
			cwd = parent
		}
		if goproj == "" {
			fmt.Println("ERROR: GOPROJ environment variable not set and could not infer")
			os.Exit(1)
		}
	}

	testScript := filepath.Join(goproj, "revise_audio", "vits", "python", "test_import.py")
	if _, err := os.Stat(testScript); os.IsNotExist(err) {
		fmt.Printf("ERROR: Test script not found: %s\n", testScript)
		os.Exit(1)
	}

	fmt.Printf("Test script: %s\n", testScript)

	// Find Python
	pythonPath := os.Getenv("FCBH_VITS_PYTHON")
	if pythonPath == "" {
		// Try conda environment
		if condaPrefix := os.Getenv("CONDA_PREFIX"); condaPrefix != "" {
			pythonPath = filepath.Join(condaPrefix, "bin", "python")
			if _, err := os.Stat(pythonPath); err != nil {
				pythonPath = ""
			}
		}
		if pythonPath == "" {
			// Try system python
			pythonPath = "python3"
		}
	}

	fmt.Printf("Python: %s\n\n", pythonPath)

	// Run test script
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, pythonPath, testScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		fmt.Printf("\n✗ Test FAILED: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ All tests passed!")
}

