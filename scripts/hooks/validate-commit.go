//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

const (
	Red   = "\033[31m"
	Green = "\033[32m"
	Cyan  = "\033[36m"
	Reset = "\033[0m"
)

var commitRegex = regexp.MustCompile(
	`^(feat|fix|docs|style|refactor|perf|test|chore|build|ci)(\([a-z0-9._-]+\))?!?: .+`,
)

func main() {
	fmt.Println(Cyan + "🔍 Validando mensagem de commit..." + Reset)

	if len(os.Args) < 2 {
		fmt.Println(Red + "❌ Erro: arquivo de commit não informado" + Reset)
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(Red+"❌ Erro ao abrir mensagem de commit: ", err, Reset)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		msg := scanner.Text()

		if !commitRegex.MatchString(msg) {
			fmt.Println(Red + "❌ Mensagem de commit inválida!" + Reset)
			fmt.Println("👉 Mensagem:", msg)
			fmt.Println(`
Formato correto:

  <tipo>[escopo]: descrição

Exemplos válidos:
  feat(login): adiciona captcha
  fix: corrige crash no parser
  chore!: remove suporte legado
`)
			os.Exit(1)
		}
	}

	fmt.Println(Green + "✅ Commit válido!" + Reset)
}
