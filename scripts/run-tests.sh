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

TEMP_LOG=$(mktemp) || { echo -e "${RED}Error: Failed to create temp file${NC}"; exit 1; }
COVERAGE_FILE=$(mktemp) || { echo -e "${RED}Error: Failed to create coverage file${NC}"; exit 1; }

# Check if go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH.${NC}"
    rm "$TEMP_LOG" "$COVERAGE_FILE"
    exit 1
fi

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

        # Try to extract coverage if available
        cov=$(echo "$line" | grep -o "coverage: [0-9.]*%" | awk '{print $2}')
        if [ -z "$cov" ]; then cov="N/A"; fi

        echo -e "${GREEN}[PASS]${NC} $pkg ($time) - Cov: $cov"
    elif [[ $line == "?"* ]]; then
        ((SKIP_COUNT++))
    elif [[ $line == FAIL* ]]; then
        ((FAIL_COUNT++))
        pkg=$(echo "$line" | awk '{print $2}')
        echo -e "${RED}[FAIL]${NC} $pkg"
    fi
done < "$TEMP_LOG"

TOTAL_COVERAGE="0.0%"
if [ -s "$COVERAGE_FILE" ]; then
    COVERAGE_OUTPUT=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null)
    if [ $? -eq 0 ]; then
        TOTAL_COVERAGE=$(echo "$COVERAGE_OUTPUT" | grep "total:" | awk '{print $3}')
    fi
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