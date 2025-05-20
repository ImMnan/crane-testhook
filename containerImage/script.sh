docker build -t cranetest .
docker tag cranetest immnan/cranetest:0.1.0
docker push immnan/cranetest:0.1.0
docker images
