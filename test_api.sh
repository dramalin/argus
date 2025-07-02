#!/bin/bash

echo "Testing Argus API endpoints..."
echo "================================"

echo -n "Health check: "
curl -s http://localhost:8080/health || echo "FAILED"

echo -e "\n\nCPU endpoint:"
curl -s http://localhost:8080/api/cpu || echo "FAILED"

echo -e "\n\nMemory endpoint:"
curl -s http://localhost:8080/api/memory || echo "FAILED"

echo -e "\n\nNetwork endpoint:"
curl -s http://localhost:8080/api/network || echo "FAILED"

echo -e "\n\nProcess endpoint (first 3 processes):"
curl -s http://localhost:8080/api/process | head -c 500 || echo "FAILED"

echo -e "\n\nWeb interface available at: http://localhost:8080" 