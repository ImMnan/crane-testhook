#!/bin/bash
set -e

echo "Setting Go environment variables and building binary..."
go env -w GOOS=linux GOARCH=amd64
go build -o cranehook .

echo "Building Docker image..."
docker build -t cranehook .

echo "Cleaning up local binary..."
rm cranehook

echo "Tagging and pushing Docker image..."

docker tag cranehook immnan/cranehook:1.1.0
docker tag cranehook immnan/cranehook:latest
docker push immnan/cranehook:1.1.0
docker push immnan/cranehook:latest

echo "Listing Docker images..."
docker images