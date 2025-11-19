# Caminho para o diretório .git/hooks
$hooksPath = ".git/hooks"

# Cria o diretório de hooks se não existir
if (-not (Test-Path $hooksPath)) {
    New-Item -ItemType Directory -Path $hooksPath -Force | Out-Null
}

# Cria o script do hook de pre-commit
$hookContent = @'
#!/bin/sh

# Navega para a raiz do repositório
cd "$(git rev-parse --show-toplevel)"

# Compila e executa o script de pre-commit em Go
go run scripts/hooks/pre-commit.go

# Se o script falhar, aborta o commit
exit $?
'@

# Salva o script de hook
$hookFile = "$hooksPath/pre-commit"
[System.IO.File]::WriteAllText($hookFile, $hookContent)

# Torna o script executável (para sistemas Unix-like)
try {
    # Tenta usar chmod (WSL/Linux/Mac)
    bash -c "chmod +x $hookFile" 2>$null
} catch {
    # Se falhar, tenta usar icacls (Windows)
    try {
        $user = [System.Environment]::UserName
        icacls $hookFile /grant "${user}:(RX)" 2>$null
    } catch {}
}

Write-Host "✅ Hook de pre-commit configurado com sucesso!" -ForegroundColor Green
Write-Host "Agora, antes de cada commit, todos os testes serão executados." -ForegroundColor Cyan
Write-Host "Se algum teste falhar, o commit será abortado." -ForegroundColor Cyan
