//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger := newLogger()
	defer logger.Sync()

	if len(os.Args) < 2 {
		logger.Error("missing commit message file argument")
		fmt.Fprintln(os.Stderr, "commit-msg hook: missing file argument")
		os.Exit(1)
	}

	msgFile := os.Args[1]
	f, err := os.Open(msgFile)
	if err != nil {
		logger.Error("failed to open commit message file", zap.String("file", msgFile), zap.Error(err))
		fmt.Fprintln(os.Stderr, "commit-msg hook: failed to open commit message file:", err)
		os.Exit(1)
	}
	defer f.Close()

	firstLine, err := readFirstNonEmptyLine(f)
	if err != nil {
		logger.Error("failed to read commit message", zap.String("file", msgFile), zap.Error(err))
		fmt.Fprintln(os.Stderr, "commit-msg hook: failed to read commit message:", err)
		os.Exit(1)
	}

	// Conventional Commits: type(optional scope): description
	// types allowed: feat, fix, chore, docs, style, refactor, perf, test, build, ci, revert
	pattern := regexp.MustCompile(`^(feat|fix|chore|docs|style|refactor|perf|test|build|ci|revert)(\([\w\-\_]+\))?: .+`)

	if !pattern.MatchString(firstLine) {
		logger.Error("invalid commit message", zap.String("message", firstLine))
		printValidationHelp(firstLine)
		os.Exit(1)
	}

	logger.Info("commit message validated", zap.String("message", firstLine))
	os.Exit(0)
}

func readFirstNonEmptyLine(f *os.File) (string, error) {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments created by git when opening editor (lines starting with #)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line, nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("commit message is empty")
}

func printValidationHelp(msg string) {
	fmt.Fprintln(os.Stderr, "Invalid commit message.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commit message must follow Conventional Commits format:")
	fmt.Fprintln(os.Stderr, "  type(optional scope): description")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Allowed types:")
	fmt.Fprintln(os.Stderr, "  feat | fix | chore | docs | style | refactor | perf | test | build | ci | revert")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  feat: add product search endpoint")
	fmt.Fprintln(os.Stderr, "  fix(api): handle nil pointer in product handler")
	fmt.Fprintln(os.Stderr, "  chore(deps): update go.mod")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commit message provided:")
	fmt.Fprintln(os.Stderr, "  ", msg)
}

func newLogger() *zap.Logger {
	cfg := zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "msg",
			LevelKey:    "level",
			TimeKey:     "time",
			EncodeTime:  zapcore.ISO8601TimeEncoder,
			EncodeLevel: zapcore.CapitalLevelEncoder,
		},
	}
	logger, _ := cfg.Build()
	return logger
}
