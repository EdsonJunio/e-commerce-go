//go:build ignore

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Reset  = "\033[0m"
)

func main() {
	fmt.Println(Cyan + "🚀 Running tests before commit..." + Reset)

	cmd := exec.Command("go", "test", "-v", "-cover", "./...")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	scanner := bufio.NewScanner(&out)

	failRegex := regexp.MustCompile(`--- FAIL: (\S+)`)
	passRegex := regexp.MustCompile(`--- PASS: (\S+)`)
	fileRegex := regexp.MustCompile(`\t(.+\.go):(\d+):`)

	var failedTests []string
	var failedFiles []string
	var passedTests []string

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case failRegex.MatchString(line):
			fmt.Println(Red + line + Reset)
			match := failRegex.FindStringSubmatch(line)
			failedTests = append(failedTests, match[1])
		case passRegex.MatchString(line):
			fmt.Println(Green + line + Reset)
			match := passRegex.FindStringSubmatch(line)
			passedTests = append(passedTests, match[1])
		default:
			fmt.Println(line)
			if match := fileRegex.FindStringSubmatch(line); match != nil {
				failedFiles = append(failedFiles, fmt.Sprintf("%s:%s", match[1], match[2]))
			}
		}
	}

	fmt.Println()
	fmt.Println(Cyan + "📊 Test Summary:" + Reset)
	fmt.Printf(Green+" - Passed: %d"+Reset+"\n", len(passedTests))
	fmt.Printf(Red+" - Failed: %d"+Reset+"\n", len(failedTests))

	if len(failedTests) > 0 || err != nil {
		if len(failedTests) > 0 {
			fmt.Println(Red + "\n❌ Fix the following tests/files before committing:" + Reset)
			for _, test := range failedTests {
				fmt.Printf(Red+" - Test: %s"+Reset+"\n", test)
			}
			for _, file := range failedFiles {
				fmt.Printf(Yellow+" - File: %s"+Reset+"\n", file)
			}
		} else {
			fmt.Println(Red + "\n❌ Some tests failed. Check details above." + Reset)
		}
		os.Exit(1)
	}

	fmt.Println(Green + "\n✅ All tests passed! Proceeding with commit..." + Reset)
}
