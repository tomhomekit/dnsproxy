@echo off
netsh interface ipv4 add dnsserver "以太网" 114.114.114.114 index=1
netsh interface ipv4 add dnsserver "以太网" 8.8.8.8 index=2