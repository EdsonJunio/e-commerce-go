#!/bin/bash

# Define colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

current_branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)

if [ -z "$current_branch" ]; then
    echo -e "${RED}Error: Could not determine current branch.${NC}"
    exit 1
fi

# 1. PROTECTED BRANCHES CHECK
if [[ "$current_branch" == "main" || "$current_branch" == "develop" ]]; then
    echo "---------------------------------------------------"
    echo -e "${RED}Error: Direct push to protected branch '${current_branch}' is prohibited.${NC}"
    echo ""
    echo "You cannot push code directly to production or development branches."
    echo "Please create a new branch and open a Pull Request."
    echo ""
    echo -e "Command to fix: ${YELLOW}git checkout -b feature/PROJ-123-my-feature${NC}"
    echo "---------------------------------------------------"
    exit 1
fi

# 2. NAMING CONVENTION CHECK
pattern="^(feature|bugfix|fix|hotfix|chore|docs)\/[A-Z]+-[0-9]+.*$"

if [[ ! $current_branch =~ $pattern ]]; then
    echo "---------------------------------------------------"
    echo -e "${RED}Error: Invalid branch name detected.${NC}"
    echo "Current branch: '$current_branch'"
    echo ""
    echo "The branch name must start with a prefix and include the Jira Ticket ID."
    echo ""
    echo "Required Format:"
    echo -e "  ${CYAN}<prefix>/<PROJECT>-<NUMBER>-<description>${NC}"
    echo ""
    echo "Allowed prefixes:"
    echo -e "  - ${GREEN}feature, bugfix, fix, hotfix, chore, docs${NC}"
    echo ""
    echo "Examples:"
    echo -e "  - ${YELLOW}feature/PROJ-1234-login-page${NC}"
    echo -e "  - ${YELLOW}bugfix/BACK-404-fix-payment-error${NC}"
    echo "---------------------------------------------------"
    exit 1
fi

exit 0