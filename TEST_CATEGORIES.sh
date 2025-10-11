#!/bin/bash

echo "🧪 Testing Component Category Search"
echo "===================================="
echo ""

TEST_DIR="/tmp/test-categories-$$"
SUCCESS=0
FAILED=0

# Test function
test_component() {
    local type=$1
    local name=$2
    local category=$3
    
    echo "Testing $type: $name (expected in $category)"
    if ./cct --$type $name --directory $TEST_DIR 2>&1 | grep -q "installed successfully"; then
        echo "✅ PASS"
        ((SUCCESS++))
    else
        echo "❌ FAIL"
        ((FAILED++))
    fi
    echo ""
}

# Build first
echo "Building..."
go build -o cct ./cmd/cct
echo ""

# Test agents from different categories
test_component "agent" "api-documenter" "documentation"
test_component "agent" "prompt-engineer" "ai-specialists"
test_component "agent" "database-architect" "database"
test_component "agent" "git-flow-manager" "git"

# Test commands from different categories
test_component "command" "security-audit" "security"
test_component "command" "setup-linting" "setup"
test_component "command" "dependency-audit" "security"

# Test MCPs from different categories
test_component "mcp" "postgresql-integration" "database"
test_component "mcp" "supabase" "database"

# Summary
echo "===================================="
echo "📊 Test Summary"
echo "   ✅ Passed: $SUCCESS"
echo "   ❌ Failed: $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "🎉 All tests passed!"
    exit 0
else
    echo "❌ Some tests failed"
    exit 1
fi
