//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	Red   = "\033[31m"
	Green = "\033[32m"
	Cyan  = "\033[36m"
	Reset = "\033[0m"
)

func main() {
	fmt.Println(Cyan + "🔍 Validando nome da branch..." + Reset)

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(Red + "❌ Erro ao obter nome da branch" + Reset)
		os.Exit(1)
	}

	branch := strings.TrimSpace(string(output))

	regex := regexp.MustCompile(`^(feature|bugfix|hotfix|chore|docs|refactor|test)\/[a-z0-9._-]+$`)

	if !regex.MatchString(branch) {
		fmt.Println(Red + "❌ Nome da branch inválido!" + Reset)
		fmt.Println("👉 Branch atual:", branch)
		fmt.Println(`
Padrões aceitos:

  feature/<nome>
  bugfix/<nome>
  hotfix/<nome>
  chore/<nome>
  docs/<nome>
  refactor/<nome>
  test/<nome>
`)
		os.Exit(1)
	}

	fmt.Println(Green + "✅ Branch válida!" + Reset)
}
