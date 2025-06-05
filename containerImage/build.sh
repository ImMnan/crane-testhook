#!/bin/bash
set -e

echo "Setting Go environment variables and building binary..."
go env -w GOOS=linux GOARCH=amd64
go build -o cranetest ../.

echo "Building Docker image..."
docker build -t cranetest .

echo "Cleaning up local binary..."
rm cranetest

echo "Tagging and pushing Docker image..."

docker tag cranetest immnan/cranetest:0.2.10
docker tag cranetest immnan/cranetest:latest
docker push immnan/cranetest:0.2.10
docker push immnan/cranetest:latest

echo "Listing Docker images..."
docker images