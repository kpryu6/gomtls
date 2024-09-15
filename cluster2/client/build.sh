#!/bin/bash
go build main.go
if [ $? -ne 0 ]; then
    echo "Build Fail.."
    exit 1
fi

sudo docker build . -t localhost:10000/helloworld-client:mtls
if [ $? -ne 0 ]; then
	echo "Poop"
	exit 1
fi

sudo docker push localhost:10000/helloworld-client:mtls
if [ $? -ne 0 ]; then
	echo "Poop"
	exit 1
fi

if [ $? -ne 0 ]; then
    echo "Build Fail.."
    exit 1
fi

echo "build complete"