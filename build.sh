#!/bin/sh
echo "Building amd64 version..."
CGO_ENABLED=0 GOOS=linux go build -o thinkfancontrol -a -ldflags '-s -w -extldflags "-static"' thinkfancontrol.go
echo "Done."
echo "comperessing binary with UPX"
upx --ultra-brute --best thinkfancontrol
ls -lh thinkfancontrol

