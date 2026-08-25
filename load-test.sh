#!/bin/bash

# Advanced Load Testing Script
# Tests with 10,000+ concurrent requests

SERVER="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         ADVANCED LOAD TEST - 10,000+ CONCURRENT REQUESTS             ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if server is running
if ! curl -s --connect-timeout 2 $SERVER/ > /dev/null 2>&1; then
    echo -e "${RED}ERROR: Server is not running on $SERVER${NC}"
    echo "Please start the server first: ./server"
    exit 1
fi

echo -e "${GREEN}✓ Server is running${NC}"
echo ""

# Function to run concurrent requests
run_concurrent_test() {
    local num_requests=$1
    local batch_size=$2
    local test_name=$3
    
    echo -e "${YELLOW}Test: $test_name${NC}"
    echo "  Requests: $num_requests"
    echo "  Batch size: $batch_size"
    echo ""
    
    start_time=$(date +%s.%N)
    success_count=0
    error_count=0
    
    # Run in batches to avoid overwhelming the shell
    for ((batch=0; batch<num_requests/batch_size; batch++)); do
        for ((i=0; i<batch_size; i++)); do
            (
                response=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 $SERVER/ 2>/dev/null)
                if [ "$response" = "200" ]; then
                    echo "success" > /tmp/load_test_$$_$batch\_$i
                else
                    echo "error:$response" > /tmp/load_test_$$_$batch\_$i
                fi
            ) &
        done
        wait
        
        # Show progress
        completed=$((batch * batch_size))
        printf "\r  Progress: %d/%d requests" $completed $num_requests
    done
    wait
    
    echo ""
    
    # Count results
    for file in /tmp/load_test_$$_*; do
        if [ -f "$file" ]; then
            result=$(cat "$file")
            if [ "$result" = "success" ]; then
                success_count=$((success_count + 1))
            else
                error_count=$((error_count + 1))
            fi
            rm "$file"
        fi
    done
    
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    rps=$(echo "scale=2; $success_count / $duration" | bc)
    success_rate=$(echo "scale=2; ($success_count * 100) / $num_requests" | bc)
    
    echo -e "${GREEN}✓ Completed${NC}"
    echo "  Duration: ${duration}s"
    echo "  Success: $success_count/$num_requests (${success_rate}%)"
    echo "  Errors: $error_count"
    echo "  Throughput: $rps req/s"
    echo ""
}

# Test 1: 1,000 concurrent requests
run_concurrent_test 1000 100 "1,000 Concurrent Requests (Warm-up)"

# Test 2: 5,000 concurrent requests
run_concurrent_test 5000 200 "5,000 Concurrent Requests"

# Test 3: 10,000 concurrent requests (THE BIG ONE)
run_concurrent_test 10000 250 "10,000 Concurrent Requests"

# Test 4: Sustained load test
echo -e "${YELLOW}Test: Sustained Load (1,000 req/s for 10 seconds)${NC}"
start_time=$(date +%s)
success=0
errors=0

for i in {1..100}; do
    for j in {1..10}; do
        curl -s $SERVER/ > /dev/null 2>&1 && success=$((success+1)) || errors=$((errors+1)) &
    done
    sleep 0.1
    wait
done

end_time=$(date +%s)
duration=$((end_time - start_time))

echo -e "${GREEN}✓ Completed${NC}"
echo "  Duration: ${duration}s"
echo "  Success: $success"
echo "  Errors: $errors"
echo "  Average: $(echo "scale=2; $success / $duration" | bc) req/s"
echo ""

# Fetch final metrics
echo -e "${BLUE}╔═══════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                      FINAL SERVER METRICS                             ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

curl -s $SERVER/__metrics__ | python3 -c "
import sys, json
data = json.load(sys.stdin)

print(f\"Server Uptime:        {data['server']['uptime_human']}\")
print(f\"Total Requests:       {data['requests']['total']:,}\")
print(f\"Total Responses:      {data['responses']['total']:,}\")
print(f\"Total Data Served:    {data['responses']['total_bytes_served']/1024/1024:.2f} MB\")
print(f\"Requests/Second:      {data['performance']['requests_per_second']:.2f}\")
print(f\"Throughput:           {data['performance']['throughput_bytes_per_second']/1024:.2f} KB/s\")
print()
rt = data['performance']['response_times']
print(f\"Response Time Stats (ms):\")
print(f\"  Count:              {rt['count']:,}\")
print(f\"  Mean:               {rt['mean_ms']:.6f}\")
print(f\"  Median (P50):       {rt['median_ms']:.6f}\")
print(f\"  P75:                {rt['p75_ms']:.6f}\")
print(f\"  P90:                {rt['p90_ms']:.6f}\")
print(f\"  P95:                {rt['p95_ms']:.6f}\")
print(f\"  P99:                {rt['p99_ms']:.6f}\")
print(f\"  P99.9:              {rt['p999_ms']:.6f}\")
print(f\"  Min:                {rt['min_ms']:.6f}\")
print(f\"  Max:                {rt['max_ms']:.6f}\")
print(f\"  StdDev:             {rt['stddev_ms']:.6f}\")
print()
print(f\"Active Connections:   {data['connections']['active']}\")
print(f\"Total Connections:    {data['connections']['total']:,}\")
"

echo ""
echo -e "${GREEN}✓ Load Test Complete!${NC}"

