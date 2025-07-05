#!/bin/bash
mkdir -p bin
set -e
echo -e "\033[34mBuilding validator binary...\033[0m"
go build -o bin/validator main.go
echo -e "\033[32mDone. Binary is at bin/validator\033[0m"