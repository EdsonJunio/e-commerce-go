//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("🚀 Executando testes antes do commit...")

	// Executa os testes do Go
	cmd := exec.Command("go", "test", "-v", "-cover", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("\n❌ Testes falharam! Por favor, corrija os erros antes de fazer o commit.")
		os.Exit(1)
	}

	fmt.Println("\n✅ Todos os testes passaram com sucesso! Prosseguindo com o commit...")
}
