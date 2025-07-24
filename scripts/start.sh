#!/bin/bash

cd /home/ec2-user/app

# Stop old container (optional)
docker stop myapp || true
docker rm myapp || true

# Build and run app (adjust based on your setup)
docker build -t myapp .
docker run -d -p 80:80 --name myapp myapp
