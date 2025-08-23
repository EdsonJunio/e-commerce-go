//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
)

// This script is meant to be used as a Git pre-commit hook.
// It runs all Go tests with coverage before allowing a commit.
// If any test fails, the commit will be blocked.
func main() {
	fmt.Println("🚀 Running tests before commit...")

	cmd := exec.Command("go", "test", "-v", "-cover", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("\n❌ Tests failed! Please fix the errors before committing.")
		os.Exit(1)
	}

	fmt.Println("\n✅ All tests passed successfully! Proceeding with commit...")
}
