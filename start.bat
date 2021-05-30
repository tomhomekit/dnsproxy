@echo off
netsh interface ipv4 add dnsserver "以太网" 127.0.0.1 index=1
netsh interface ipv4 add dnsserver "以太网" 114.114.114.114 index=2
dnsproxy.exe