#!bin/bash

go env -w GOOS=linux GOARCH=amd64 && go build -o cranetest /Users/mpatel/Documents/GitHub/crane-testhook
docker build -t cranetest .
docker tag cranetest immnan/cranetest:0.1.0
docker push immnan/cranetest:0.1.0
docker images
