#!/bin/bash

# Define colors
RED='\033[0;31m'
GREEN='\033[0;32m'
GRAY='\033[0;90m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

echo -e "${BOLD}Running Unit Tests & Coverage Check...${NC}"
echo "---------------------------------------------------"

TEMP_LOG=$(mktemp)
COVERAGE_FILE=$(mktemp)

go test ./... -covermode=atomic -coverprofile="$COVERAGE_FILE" > "$TEMP_LOG" 2>&1
TEST_EXIT_CODE=$?

PASS_COUNT=0
FAIL_COUNT=0
SKIP_COUNT=0

while IFS= read -r line; do
    if [[ $line == ok* ]]; then
        ((PASS_COUNT++))
        pkg=$(echo "$line" | awk '{print $2}')
        time=$(echo "$line" | awk '{print $3}')
        cov=$(echo "$line" | grep -o "coverage: [0-9.]*%" | awk '{print $2}')
        if [ -z "$cov" ]; then cov="?"; fi
        echo -e "${GREEN}[PASS]${NC} $pkg ($time) - Cov: $cov"
    elif [[ $line == "?"* ]]; then
        ((SKIP_COUNT++))
    elif [[ $line == FAIL* ]]; then
        ((FAIL_COUNT++))
        pkg=$(echo "$line" | awk '{print $2}')
        echo -e "${RED}[FAIL]${NC} $pkg"
    fi
done < "$TEMP_LOG"

if [ -s "$COVERAGE_FILE" ]; then
    TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep "total:" | awk '{print $3}')
else
    TOTAL_COVERAGE="0.0%"
fi

echo "---------------------------------------------------"
echo -e "${BOLD}TEST SUMMARY:${NC}"
echo -e "  ${GREEN}✔ Passed:${NC}   $PASS_COUNT"
echo -e "  ${RED}✖ Failed:${NC}   $FAIL_COUNT"
echo -e "  ${GRAY}⏭ Skipped:${NC}  $SKIP_COUNT"
echo -e "  ${CYAN}📊 Coverage:${NC} $TOTAL_COVERAGE"
echo "---------------------------------------------------"

if [ $TEST_EXIT_CODE -ne 0 ]; then
    echo -e "${RED}FAILURE DETECTED. ERROR LOG:${NC}"
    echo ""
    grep -v "^ok" "$TEMP_LOG" | grep -v "^?"
    echo ""
    echo "---------------------------------------------------"
    echo -e "${RED}Commit aborted due to failed tests.${NC}"
    rm "$TEMP_LOG" "$COVERAGE_FILE"
    exit 1
fi

rm "$TEMP_LOG" "$COVERAGE_FILE"
echo -e "${GREEN}All checks passed. Proceeding with commit.${NC}"
exit 0
