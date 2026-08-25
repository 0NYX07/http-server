#!/bin/bash

# Comprehensive benchmark script for HTTP server
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         HTTP SERVER PERFORMANCE BENCHMARK SUITE               ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════════╝${NC}"
echo ""

SERVER="http://localhost:8080"

# Test 1: Single Request Latency
echo -e "${YELLOW}📊 Test 1: Single Request Latency${NC}"
echo "Testing response time for a single request..."
time_result=$(curl -o /dev/null -s -w "%{time_total}" $SERVER/)
echo -e "${GREEN}✓ Response Time: ${time_result}s ($(echo "$time_result * 1000" | bc)ms)${NC}"
echo ""

# Test 2: Sequential Requests
echo -e "${YELLOW}📊 Test 2: Sequential Throughput (100 requests)${NC}"
start_time=$(date +%s.%N)
for i in {1..100}; do
    curl -s $SERVER/ > /dev/null
done
end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc)
rps=$(echo "100 / $duration" | bc -l)
echo -e "${GREEN}✓ Duration: ${duration}s${NC}"
echo -e "${GREEN}✓ Throughput: $(printf "%.2f" $rps) requests/second${NC}"
echo ""

# Test 3: Concurrent Requests (Small Scale)
echo -e "${YELLOW}📊 Test 3: Concurrent Performance (20 parallel x 5 batches)${NC}"
start_time=$(date +%s.%N)
for batch in {1..5}; do
    for i in {1..20}; do
        curl -s $SERVER/ > /dev/null &
    done
    wait
done
end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc)
rps=$(echo "100 / $duration" | bc -l)
echo -e "${GREEN}✓ Duration: ${duration}s${NC}"
echo -e "${GREEN}✓ Throughput: $(printf "%.2f" $rps) requests/second${NC}"
echo ""

# Test 4: File Size Performance
echo -e "${YELLOW}📊 Test 4: Different File Sizes${NC}"

# Small file (HTML)
html_size=$(curl -s $SERVER/ | wc -c)
html_time=$(curl -o /dev/null -s -w "%{time_total}" $SERVER/)
html_speed=$(echo "$html_size / $html_time / 1024" | bc -l)
echo -e "  HTML Page ($(numfmt --to=iec $html_size)):"
echo -e "    Response time: ${html_time}s"
echo -e "    Speed: $(printf "%.2f" $html_speed) KB/s"

# Medium file (Image)
if [ -f "resources/logo.png" ]; then
    img_size=$(curl -s $SERVER/logo.png | wc -c)
    img_time=$(curl -o /dev/null -s -w "%{time_total}" $SERVER/logo.png)
    img_speed=$(echo "$img_size / $img_time / 1024" | bc -l)
    echo -e "  PNG Image ($(numfmt --to=iec $img_size)):"
    echo -e "    Response time: ${img_time}s"
    echo -e "    Speed: $(printf "%.2f" $img_speed) KB/s"
fi

# Large file
if [ -f "resources/large-image.png" ]; then
    large_size=$(curl -s $SERVER/large-image.png | wc -c)
    large_time=$(curl -o /dev/null -s -w "%{time_total}" $SERVER/large-image.png)
    large_speed=$(echo "$large_size / $large_time / 1024" | bc -l)
    echo -e "  Large PNG ($(numfmt --to=iec $large_size)):"
    echo -e "    Response time: ${large_time}s"
    echo -e "    Speed: $(printf "%.2f" $large_speed) KB/s"
fi
echo ""

# Test 5: Keep-Alive Performance
echo -e "${YELLOW}📊 Test 5: Keep-Alive Connection (10 requests on same connection)${NC}"
start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s -H "Connection: keep-alive" $SERVER/ > /dev/null
done
end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc)
avg_time=$(echo "$duration / 10" | bc -l)
echo -e "${GREEN}✓ Average time per request: $(printf "%.3f" $avg_time)s${NC}"
echo ""

# Test 6: POST Performance
echo -e "${YELLOW}📊 Test 6: POST Request Performance (JSON upload)${NC}"
start_time=$(date +%s.%N)
for i in {1..20}; do
    curl -s -X POST $SERVER/upload \
        -H "Content-Type: application/json" \
        -d '{"test":"data","index":'$i'}' > /dev/null
done
end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc)
rps=$(echo "20 / $duration" | bc -l)
echo -e "${GREEN}✓ Duration: ${duration}s${NC}"
echo -e "${GREEN}✓ Throughput: $(printf "%.2f" $rps) POST requests/second${NC}"
echo ""

# Final Metrics
echo -e "${BLUE}╔═══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                     FINAL METRICS                             ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════════╝${NC}"
curl -s $SERVER/__metrics__ | python3 -c "
import sys, json
data = json.load(sys.stdin)
print(f\"Total Requests:       {data['requests']['total']}\")
print(f\"Total Responses:      {data['responses']['total']}\")
print(f\"Total Data Served:    {data['responses']['total_bytes_served']/1024:.2f} KB\")
print(f\"Success Rate:         {data['responses']['by_status']['200']/data['responses']['total']*100:.1f}%\")
print(f\"Avg Response Time:    {data['performance']['avg_response_time_ms']:.2f} ms\")
print(f\"Server Uptime:        {data['server']['uptime_human']}\")
"

echo ""
echo -e "${GREEN}✓ Benchmark Complete!${NC}"

