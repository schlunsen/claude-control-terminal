#!/bin/bash
# Quick test script for go-claude-templates

echo "🧪 Quick Test Suite for go-claude-templates"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
PASS=0
FAIL=0

# Test function
test_command() {
    echo -n "Testing: $1... "
    if eval "$2" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ PASS${NC}"
        ((PASS++))
    else
        echo -e "${RED}❌ FAIL${NC}"
        ((FAIL++))
    fi
}

# Test 1: Binary exists
test_command "Binary exists" "test -f ./cct"

# Test 2: Version command
test_command "Version command" "./cct --version | grep -q '2.0.0-go'"

# Test 3: Help command
test_command "Help command" "./cct --help | grep -q 'Claude Code'"

# Test 4: Build with make
test_command "Make build" "make clean && make build"

# Test 5: Component directory creation
echo -n "Testing: Component installation... "
mkdir -p /tmp/cct-test
./cct --agent test --directory /tmp/cct-test > /dev/null 2>&1
if [ -d "/tmp/cct-test/.claude/agents" ]; then
    echo -e "${GREEN}✅ PASS${NC}"
    ((PASS++))
else
    echo -e "${RED}❌ FAIL${NC}"
    ((FAIL++))
fi
rm -rf /tmp/cct-test

# Test 6: Analytics server (start and check)
echo -n "Testing: Analytics server... "
./cct --analytics > /tmp/cct-server.log 2>&1 &
SERVER_PID=$!
sleep 3

if curl -s http://localhost:3333/api/health | grep -q "ok"; then
    echo -e "${GREEN}✅ PASS${NC}"
    ((PASS++))
else
    echo -e "${RED}❌ FAIL${NC}"
    ((FAIL++))
fi

# Cleanup server
kill $SERVER_PID 2>/dev/null
rm -f /tmp/cct-server.log

# Test 7: Cross-platform build
test_command "Cross-platform build" "make build-all"

echo ""
echo "=========================================="
echo "Test Results:"
echo -e "${GREEN}Passed: $PASS${NC}"
echo -e "${RED}Failed: $FAIL${NC}"
echo "=========================================="

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some tests failed${NC}"
    exit 1
fi
