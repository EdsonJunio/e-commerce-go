#!/bin/sh

HOOKS=".git/hooks"

mkdir -p "$HOOKS"

echo "Instalando hook pre-commit..."
cat << 'EOF' > "$HOOKS/pre-commit"
#!/bin/sh
cd "$(git rev-parse --show-toplevel)"
go run scripts/hooks/validate-branch.go || exit 1
go run scripts/hooks/pre-commit.go
exit $?
EOF

echo "Instalando hook commit-msg..."
cat << 'EOF' > "$HOOKS/commit-msg"
#!/bin/sh
cd "$(git rev-parse --show-toplevel)"
go run scripts/hooks/validate-commit.go "$1"
exit $?
EOF

chmod +x "$HOOKS/pre-commit"
chmod +x "$HOOKS/commit-msg"

echo "✅ Hooks instalados com sucesso!"
