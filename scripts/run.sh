#!/bin/bash
SNAPSHOT_PATH=${1:-accounts.json}

if [ ! -f "$SNAPSHOT_PATH" ]; then
  echo -e "\033[31mSnapshot not found at $SNAPSHOT_PATH\033[0m"
  echo -e "\033[33mUsage: ./scripts/run.sh [path/to/accounts.json]\033[0m"
  exit 1
fi

echo -e "\033[32mRunning validator with snapshot: $SNAPSHOT_PATH\033[0m"
./bin/validator "$SNAPSHOT_PATH"