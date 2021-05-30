#!/bin/bash

thread=$(ps -axf | grep dnsproxy | grep -v "grep")
if [[ "x"$thread  == "x" ]] ; then
  exit 0
fi
pid=$(echo $thread | awk -F " " '{print $2}')

echo "killing $pid"
kill -9 $pid

