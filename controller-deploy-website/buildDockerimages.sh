#! /bin/bash

# Load to Minikube local repository
eval $(minikube docker-env)

# Controller image
docker build -f controller/Dockerfile -t website-controller controller

# Website images
docker build -f websites/homepage-1/Dockerfile -t homepage-1 websites/homepage-1
docker build -f websites/homepage-2/Dockerfile -t homepage-2 websites/homepage-2
docker build -f websites/homepage-3/Dockerfile -t homepage-3 websites/homepage-3
