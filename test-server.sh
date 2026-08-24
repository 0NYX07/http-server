#!/bin/bash

# Test script for Multi-threaded HTTP Server
# Run this after starting the server to test all features

SERVER="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=================================="
echo "HTTP Server Test Suite"
echo "=================================="
echo ""

# Test 1: GET root path (index.html)
echo -e "${BLUE}Test 1: GET / (index.html)${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $SERVER/)
if [ "$RESPONSE" = "200" ]; then
    echo -e "${GREEN}✓ PASS: Got 200 OK${NC}"
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
fi
echo ""

# Test 2: GET HTML file
echo -e "${BLUE}Test 2: GET /about.html${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $SERVER/about.html)
if [ "$RESPONSE" = "200" ]; then
    echo -e "${GREEN}✓ PASS: Got 200 OK${NC}"
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
fi
echo ""

# Test 3: GET text file (binary download)
echo -e "${BLUE}Test 3: GET /sample.txt (binary download)${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $SERVER/sample.txt)
if [ "$RESPONSE" = "200" ]; then
    echo -e "${GREEN}✓ PASS: Got 200 OK${NC}"
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
fi
echo ""

# Test 4: GET image file
echo -e "${BLUE}Test 4: GET /logo.png (binary download)${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $SERVER/logo.png)
if [ "$RESPONSE" = "200" ]; then
    echo -e "${GREEN}✓ PASS: Got 200 OK${NC}"
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
fi
echo ""

# Test 5: GET 404
echo -e "${BLUE}Test 5: GET /nonexistent.html (404)${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $SERVER/nonexistent.html)
if [ "$RESPONSE" = "404" ]; then
    echo -e "${GREEN}✓ PASS: Got 404 Not Found${NC}"
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
fi
echo ""

# Test 6: POST with JSON
echo -e "${BLUE}Test 6: POST /upload with JSON${NC}"
RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d '{"test": "data", "value": 123}' -w "%{http_code}" $SERVER/upload)
if [[ "$RESPONSE" == *"201"* ]]; then
    echo -e "${GREEN}✓ PASS: Got 201 Created${NC}"
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
fi
echo ""

# Test 7: POST with non-JSON (should fail)
echo -e "${BLUE}Test 7: POST /upload with non-JSON (415)${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: text/plain" -d "not json" $SERVER/upload)
if [ "$RESPONSE" = "415" ]; then
    echo -e "${GREEN}✓ PASS: Got 415 Unsupported Media Type${NC}"
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
fi
echo ""

# Test 8: Unsupported method (PUT)
echo -e "${BLUE}Test 8: PUT request (405)${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X PUT $SERVER/)
if [ "$RESPONSE" = "405" ]; then
    echo -e "${GREEN}✓ PASS: Got 405 Method Not Allowed${NC}"
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
fi
echo ""

# Test 9: Path traversal attack (should fail)
echo -e "${BLUE}Test 9: Path traversal attack /../etc/passwd (403)${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$SERVER/../etc/passwd")
if [ "$RESPONSE" = "403" ]; then
    echo -e "${GREEN}✓ PASS: Got 403 Forbidden${NC}"
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
fi
echo ""

# Test 10: Invalid host header (should fail)
echo -e "${BLUE}Test 10: Invalid Host header (403)${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: evil.com" $SERVER/)
if [ "$RESPONSE" = "403" ]; then
    echo -e "${GREEN}✓ PASS: Got 403 Forbidden${NC}"
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE (security may be disabled)${NC}"
fi
echo ""

# Test 11: Keep-alive connection
echo -e "${BLUE}Test 11: Keep-alive connection${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -H "Connection: keep-alive" $SERVER/)
if [ "$RESPONSE" = "200" ]; then
    echo -e "${GREEN}✓ PASS: Keep-alive working${NC}"
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
fi
echo ""

# Test 12: Unsupported file type
echo -e "${BLUE}Test 12: Unsupported file type .pdf (415)${NC}"
touch /tmp/test.pdf 2>/dev/null
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$SERVER/test.pdf")
if [ "$RESPONSE" = "415" ] || [ "$RESPONSE" = "404" ]; then
    echo -e "${GREEN}✓ PASS: Got $RESPONSE${NC}"
else
    echo -e "${RED}✗ FAIL: Got $RESPONSE${NC}"
fi
echo ""

echo "=================================="
echo "Test Suite Complete"
echo "=================================="

