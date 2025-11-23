#!/bin/bash

# Define colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

current_branch=$(git rev-parse --abbrev-ref HEAD)
pattern="^(main|develop|(feature|bugfix|hotfix)\/.+)$"

if [[ ! $current_branch =~ $pattern ]]; then
    echo -e "${RED}Error: Invalid branch name detected.${NC}"
    echo "Current branch: '$current_branch'"
    echo ""
    echo "The branch name does not follow the project's naming conventions."
    echo ""
    echo "Allowed patterns:"
    echo -e "  - ${GREEN}main${NC}, ${GREEN}develop${NC}"
    echo -e "  - ${GREEN}feature/${NC}${CYAN}<name>${NC}"
    echo -e "  - ${GREEN}bugfix/${NC}${CYAN}<name>${NC}"
    echo -e "  - ${GREEN}hotfix/${NC}${CYAN}<name>${NC}"
    echo ""
    echo "Examples:"
    echo -e "  - ${YELLOW}feature/user-login${NC}"
    echo -e "  - ${YELLOW}bugfix/payment-error${NC}"
    echo "---------------------------------------------------"
    exit 1
fi

exit 0
