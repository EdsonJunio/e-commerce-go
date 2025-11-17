# Caminho do diretório .git/hooks
$repoRoot = git rev-parse --show-toplevel
$hooksPath = "$repoRoot/.git/hooks"

# Cria o diretório de hooks se não existir
if (-not (Test-Path $hooksPath)) {
    New-Item -ItemType Directory -Path $hooksPath -Force | Out-Null
}

# Conteúdo do hook pre-commit real
$hookContent = @'
#!/bin/sh

# Garantir execução na raiz do repositório
cd "$(git rev-parse --show-toplevel)"

# Executar o script pre-commit em Go
go run scripts/hooks/pre-commit.go

# Retornar status da execução
exit $?
'@

# Caminho final do hook real
$hookFile = "$hooksPath/pre-commit"

# Criar o hook
[System.IO.File]::WriteAllText($hookFile, $hookContent)

# Tornar executável para todos os usuários
try {
    bash -c "chmod 755 $hookFile" 2>$null
} catch {
    try {
        icacls $hookFile /grant "*S-1-1-0:(RX)" 2>$null
    } catch {}
}

Write-Host "Hook pre-commit instalado com sucesso."
Write-Host "Validação de branch, commit message e testes ativados."
