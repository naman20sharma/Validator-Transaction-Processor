#!/bin/bash
echo "Starting mock batch receiver server on port 2002..."
chmod +x scripts/server.sh
go run test_server/server.go