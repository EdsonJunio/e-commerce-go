#!/bin/bash

# Define colors
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "🔧 Setting up Git Hooks..."

# 1. Setup Commit-Msg Hook
echo "#!/bin/bash" > .git/hooks/commit-msg
echo "./scripts/validate-commit.sh \"\$1\"" >> .git/hooks/commit-msg
chmod +x .git/hooks/commit-msg
echo -e "✅ ${GREEN}commit-msg hook installed.${NC}"

# 2. Setup Pre-Commit Hook
echo "#!/bin/bash" > .git/hooks/pre-commit
echo "./scripts/run-tests.sh" >> .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
echo -e "✅ ${GREEN}pre-commit hook installed.${NC}"

# 3. Setup Pre-Push Hook
echo "#!/bin/bash" > .git/hooks/pre-push
echo "./scripts/validate-branch.sh" >> .git/hooks/pre-push
chmod +x .git/hooks/pre-push
echo -e "✅ ${GREEN}pre-push hook installed.${NC}"

# 4. Configure Git to use local hooks (just in case)
git config core.hooksPath .git/hooks

echo "------------------------------------------------"
echo -e "🎉 ${GREEN}Git hooks configured successfully!${NC}"
echo "   You are now protected by the project's quality standards."
