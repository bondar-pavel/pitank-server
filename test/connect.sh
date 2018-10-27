#!/bin/bash

host="localhost"
port="8080"

while getopts :h:p: opt; do
  case $opt in
    h)
      host=$OPTARG
      ;;
    p)
      port=$OPTARG
      ;;
  esac
done

shift $((OPTIND-1))

if [ -z "$1" ]; then
    echo "usage: ./connect.sh [-h host] [-p port] pitank_name"
    exit -1
fi

curl -i -N --output - -H "Connection: Upgrade" -H "Upgrade: websocket" \
 -H "Sec-Websocket-Key: TEST" -H "Sec-Websocket-Version: 13" \
 http://$host:$port/api/connect/$1
