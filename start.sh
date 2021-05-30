#!/bin/bash

./stop.sh

networksetup -setdnsservers Wi-Fi 127.0.0.1 114.114.114.114
nohup dnsproxy > /tmp/dns.log 2>&1 &

echo $!