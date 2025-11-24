#!/bin/bash

# Define colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🔧 Setting up Git Hooks..."

# 1. Find git root
GIT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null)

if [ -z "$GIT_ROOT" ]; then
    echo -e "${RED}Error: Not a git repository or git command failed.${NC}"
    exit 1
fi

HOOKS_DIR="$GIT_ROOT/.git/hooks"
SCRIPTS_DIR="$GIT_ROOT/scripts"

# Validate scripts directory
if [ ! -d "$SCRIPTS_DIR" ]; then
    echo -e "${RED}Error: Scripts directory not found at $SCRIPTS_DIR${NC}"
    exit 1
fi

# --- Make scripts executable ---
echo "⚙️  Making scripts executable..."
if ls "$SCRIPTS_DIR"/*.sh 1> /dev/null 2>&1; then
    chmod +x "$SCRIPTS_DIR"/*.sh
else
    echo -e "${RED}Warning: No .sh files found in scripts directory.${NC}"
fi
# ------------------------------------------------

# Function to safely create hook
create_hook() {
    local hook_name=$1
    local script_name=$2
    local hook_path="$HOOKS_DIR/$hook_name"

    echo "#!/bin/bash" > "$hook_path"
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to create $hook_name hook.${NC}"
        exit 1
    fi

    echo "$SCRIPTS_DIR/$script_name \"\$1\"" >> "$hook_path"
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to write to $hook_name hook.${NC}"
        exit 1
    fi

    chmod +x "$hook_path"
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to make $hook_name hook executable.${NC}"
        rm -f "$hook_path"
        exit 1
    fi

    echo -e "✅ ${GREEN}$hook_name hook installed.${NC}"
}

# 2. Setup Hooks

# ✅ Commit Messages
create_hook "commit-msg" "validate-commit.sh"

# ✅ Tests (ENABLED for this project)
create_hook "pre-commit" "run-tests.sh"

# ✅ Branch Validation
create_hook "pre-push" "validate-branch.sh"

# 3. Configure Git hooks path
git config core.hooksPath .git/hooks
if [ $? -ne 0 ]; then
    echo -e "${RED}Warning: Failed to configure git hooks path.${NC}"
fi

echo "------------------------------------------------"
echo -e "🎉 ${GREEN}Git hooks configured successfully!${NC}"