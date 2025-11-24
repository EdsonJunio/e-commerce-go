.PHONY: setup test

# Configuration command for new developers
setup:
	@echo "🔧 Configuring development environment..."
	@echo "⚙️  Making scripts executable..."
	@chmod +x scripts/*.sh
	@echo "🔗 Installing Git Hooks..."
	@./scripts/install-hooks.sh

# Short command to run tests manually
test:
	@./scripts/run-tests.sh