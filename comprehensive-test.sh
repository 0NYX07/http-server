#!/bin/bash

# Comprehensive endpoint and process test
# Tests all GET, POST endpoints and verifies test-data.json

SERVER="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║        COMPREHENSIVE HTTP SERVER TEST SUITE                  ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

PASSED=0
FAILED=0

test_endpoint() {
    local name="$1"
    local method="$2"
    local path="$3"
    local expected="$4"
    local extra_args="$5"
    
    echo -e "${BLUE}Testing: $name${NC}"
    
    if [ "$method" = "GET" ]; then
        RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$SERVER$path")
    else
        RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" $extra_args "$SERVER$path")
    fi
    
    if [ "$RESPONSE" = "$expected" ]; then
        echo -e "${GREEN}✓ PASS: Got $RESPONSE${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAIL: Expected $expected, got $RESPONSE${NC}"
        ((FAILED++))
    fi
    echo ""
}

echo "═══════════════════════════════════════════════════════════════"
echo "  PART 1: HTML FILES (GET)"
echo "═══════════════════════════════════════════════════════════════"
echo ""

test_endpoint "GET / (index.html)" "GET" "/" "200"
test_endpoint "GET /index.html" "GET" "/index.html" "200"
test_endpoint "GET /about.html" "GET" "/about.html" "200"
test_endpoint "GET /contact.html" "GET" "/contact.html" "200"
test_endpoint "GET /features.html" "GET" "/features.html" "200"

echo "═══════════════════════════════════════════════════════════════"
echo "  PART 2: PNG IMAGES (GET Binary)"
echo "═══════════════════════════════════════════════════════════════"
echo ""

test_endpoint "GET /logo.png (small PNG)" "GET" "/logo.png" "200"
test_endpoint "GET /large-image.png (118KB PNG)" "GET" "/large-image.png" "200"
test_endpoint "GET /large-test-image.png (23MB PNG)" "GET" "/large-test-image.png" "200"

echo "═══════════════════════════════════════════════════════════════"
echo "  PART 3: JPEG IMAGES (GET Binary)"
echo "═══════════════════════════════════════════════════════════════"
echo ""

test_endpoint "GET /photo.jpg" "GET" "/photo.jpg" "200"
test_endpoint "GET /image.jpg" "GET" "/image.jpg" "200"

echo "═══════════════════════════════════════════════════════════════"
echo "  PART 4: TEXT FILES (GET Binary)"
echo "═══════════════════════════════════════════════════════════════"
echo ""

test_endpoint "GET /sample.txt" "GET" "/sample.txt" "200"
test_endpoint "GET /readme.txt" "GET" "/readme.txt" "200"

echo "═══════════════════════════════════════════════════════════════"
echo "  PART 5: JSON FILE (GET)"
echo "═══════════════════════════════════════════════════════════════"
echo ""

echo -e "${BLUE}Testing: GET /test-data.json${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" "$SERVER/test-data.json")
STATUS=$(echo "$RESPONSE" | tail -1)
CONTENT=$(echo "$RESPONSE" | sed '$d')

if [ "$STATUS" = "200" ] && echo "$CONTENT" | grep -q "name"; then
    echo -e "${GREEN}✓ PASS: Got 200 and valid JSON content${NC}"
    echo "  Content preview: $(echo "$CONTENT" | head -c 80)..."
    ((PASSED++))
else
    echo -e "${RED}✗ FAIL: Status=$STATUS${NC}"
    ((FAILED++))
fi
echo ""

echo "═══════════════════════════════════════════════════════════════"
echo "  PART 6: POST REQUESTS (JSON)"
echo "═══════════════════════════════════════════════════════════════"
echo ""

echo -e "${BLUE}Testing: POST /upload (inline JSON)${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"test":"inline","value":123}' \
    "$SERVER/upload")
if [ "$RESPONSE" = "201" ]; then
    echo -e "${GREEN}✓ PASS: Got 201${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ FAIL: Expected 201, got $RESPONSE${NC}"
    ((FAILED++))
fi
echo ""

echo -e "${BLUE}Testing: POST /upload (from test-data.json)${NC}"
if [ -f resources/test-data.json ]; then
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d @resources/test-data.json \
        "$SERVER/upload")
    if [ "$RESPONSE" = "201" ]; then
        echo -e "${GREEN}✓ PASS: Got 201 (test-data.json uploaded successfully)${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAIL: resources/test-data.json not found${NC}"
    ((FAILED++))
fi
echo ""

echo -e "${BLUE}Testing: POST /upload (invalid Content-Type)${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: text/plain" \
    -d "not json" \
    "$SERVER/upload")
if [ "$RESPONSE" = "415" ]; then
    echo -e "${GREEN}✓ PASS: Got 415${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ FAIL: Expected 415, got $RESPONSE${NC}"
    ((FAILED++))
fi
echo ""

echo "═══════════════════════════════════════════════════════════════"
echo "  PART 7: METRICS ENDPOINT"
echo "═══════════════════════════════════════════════════════════════"
echo ""

echo -e "${BLUE}Testing: GET /__metrics__${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" "$SERVER/__metrics__")
STATUS=$(echo "$RESPONSE" | tail -1)
CONTENT=$(echo "$RESPONSE" | sed '$d')

if [ "$STATUS" = "200" ] && echo "$CONTENT" | grep -q "requests"; then
    echo -e "${GREEN}✓ PASS: Got 200 and valid metrics JSON${NC}"
    echo "  Metrics preview:"
    echo "$CONTENT" | grep -E '"total"|"active"|"uptime"' | head -3
    ((PASSED++))
else
    echo -e "${RED}✗ FAIL: Status=$STATUS${NC}"
    ((FAILED++))
fi
echo ""

echo "═══════════════════════════════════════════════════════════════"
echo "  PART 8: ERROR HANDLING"
echo "═══════════════════════════════════════════════════════════════"
echo ""

test_endpoint "GET /nonexistent.html (404)" "GET" "/nonexistent.html" "404"
test_endpoint "PUT request (405)" "PUT" "/" "405"
echo -e "${BLUE}Testing: Path traversal /../etc/passwd (403 or 404)${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$SERVER/../etc/passwd")
if [ "$RESPONSE" = "403" ] || [ "$RESPONSE" = "404" ]; then
    echo -e "${GREEN}✓ PASS: Got $RESPONSE (path traversal blocked)${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
    ((FAILED++))
fi
echo ""

echo "═══════════════════════════════════════════════════════════════"
echo "  PART 9: CONNECTION MANAGEMENT"
echo "═══════════════════════════════════════════════════════════════"
echo ""

test_endpoint "Keep-alive connection" "GET" "/" "200" '-H "Connection: keep-alive"'

echo "═══════════════════════════════════════════════════════════════"
echo "  FINAL RESULTS"
echo "═══════════════════════════════════════════════════════════════"
echo ""
TOTAL=$((PASSED + FAILED))
echo -e "Total Tests: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED${NC}"
else
    echo -e "${GREEN}Failed: $FAILED${NC}"
fi
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║          ALL TESTS PASSED! ✓                                 ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║          SOME TESTS FAILED!                                  ║${NC}"
    echo -e "${RED}╚══════════════════════════════════════════════════════════════╝${NC}"
    exit 1
fi
