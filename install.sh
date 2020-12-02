#!/bin/sh
sudo systemctl stop thinkfancontrol.service
echo "Copying binary to /usr/sbin..."
sudo cp thinkfancontrol /usr/sbin
if [ ! -f /etc/thinkfancontrol/config.yml ]; then
    echo "Copying configfile to /etc/thinkfancontrol..."
    sudo mkdir -p /etc/thinkfancontrol
    sudo cp config.yml /etc/thinkfancontrol
fi
if [ ! -f /etc/systemd/system/thinkfancontrol.service ]; then

    echo "Copying service object to /etc/systemd/system..."
    sudo cp thinkfancontrol.service  /etc/systemd/system/
    sudo systemctl daemon-reload
fi
sudo systemctl restart thinkfancontrol.service
sudo systemctl status thinkfancontrol.service