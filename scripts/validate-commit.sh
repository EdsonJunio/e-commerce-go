#!/bin/bash

# Define colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

commit_msg_file=$1
commit_msg=$(head -n 1 "$commit_msg_file")
types="(feat|fix|docs|style|refactor|test|chore|build|ci|perf)"
pattern="^$types(\([a-z0-9\-\_]+\))?!?: .+$"

if [[ ! $commit_msg =~ $pattern ]]; then
    echo -e "${RED}Error: Invalid commit message format.${NC}"
    echo "Your message: '$commit_msg'"
    echo ""
    echo "Please follow the Conventional Commits standard:"
    echo -e "  ${CYAN}<type>(<scope>): <subject>${NC}"
    echo ""
    echo "Common examples:"
    echo -e "  - ${GREEN}feat:${NC}     add a new feature"
    echo -e "              Ex: ${YELLOW}feat(auth): add login endpoint${NC}"
    echo ""
    echo -e "  - ${GREEN}fix:${NC}      fix a bug"
    echo -e "              Ex: ${YELLOW}fix(ui): correct button alignment${NC}"
    echo ""
    echo -e "  - ${GREEN}chore:${NC}    maintenance tasks (deps, config, etc)"
    echo -e "              Ex: ${YELLOW}chore: update go.mod dependencies${NC}"
    echo ""
    echo -e "Other allowed types: ${CYAN}docs, style, refactor, test, build, ci, perf${NC}"
    echo "---------------------------------------------------"
    exit 1
fi

exit 0
