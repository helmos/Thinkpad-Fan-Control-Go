#!/bin/sh
echo "Copying binary to /usr/sbin..."
sudo cp thinkfancontrol /usr/sbin
echo "Copying configfile to /etc/thinkfancontrol..."
sudo mkdir -p /etc/thinkfancontrol
sudo cp config.yml /etc/thinkfancontrol
echo "Copying service object to /etc/systemd/system..."
sudo cp thinkfancontrol.service  /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start thinkfancontrol.service