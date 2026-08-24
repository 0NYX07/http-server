#!/bin/bash
SERVER="http://localhost:8080"
echo "🔥 Running 2,000 concurrent requests test..."
start=$(date +%s.%N)
success=0
for batch in {1..20}; do
    for i in {1..100}; do
        curl -s $SERVER/ > /dev/null 2>&1 && success=$((success+1)) &
    done
    wait
done
end=$(date +%s.%N)
duration=$(echo "$end - $start" | bc)
rps=$(echo "scale=2; 2000 / $duration" | bc)
echo "✓ Completed 2,000 requests in ${duration}s"
echo "✓ Throughput: $rps req/s"
echo ""
echo "📊 Fetching detailed metrics..."
curl -s $SERVER/__metrics__ | python3 -c "
import sys, json
data = json.load(sys.stdin)
rt = data['performance']['response_times']
print()
print('╔═══════════════════════════════════════════════════════════════════════╗')
print('║                    IMPROVED PERFORMANCE METRICS                       ║')
print('╚═══════════════════════════════════════════════════════════════════════╝')
print()
print(f'Server Statistics:')
print(f'  Total Requests:       {data[\"requests\"][\"total\"]:,}')
print(f'  Total Responses:      {data[\"responses\"][\"total\"]:,}')
print(f'  Data Served:          {data[\"responses\"][\"total_bytes_served\"]/1024/1024:.2f} MB')
print(f'  Requests/Second:      {data[\"performance\"][\"requests_per_second\"]:.2f}')
print()
print(f'Response Time Statistics (6-decimal precision):')
print(f'  Sample Count:         {rt[\"count\"]:,} measurements')
print(f'  Mean:                 {rt[\"mean_ms\"]:.6f} ms')
print(f'  Median (P50):         {rt[\"median_ms\"]:.6f} ms')
print(f'  75th Percentile:      {rt[\"p75_ms\"]:.6f} ms')
print(f'  90th Percentile:      {rt[\"p90_ms\"]:.6f} ms')
print(f'  95th Percentile:      {rt[\"p95_ms\"]:.6f} ms')
print(f'  99th Percentile:      {rt[\"p99_ms\"]:.6f} ms')
print(f'  99.9th Percentile:    {rt[\"p999_ms\"]:.6f} ms')
print(f'  Minimum:              {rt[\"min_ms\"]:.6f} ms')
print(f'  Maximum:              {rt[\"max_ms\"]:.6f} ms')
print(f'  Standard Deviation:   {rt[\"stddev_ms\"]:.6f} ms')
print()
print(f'Connection Statistics:')
print(f'  Active:               {data[\"connections\"][\"active\"]}')
print(f'  Total Handled:        {data[\"connections\"][\"total\"]:,}')
print()
"
