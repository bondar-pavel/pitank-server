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
    echo "usage: ./client.sh [-h host] [-p port] pitank_name"
    exit -1
fi

# To install wscat run
# go get github.com/emulbreh/wscat
wscat ws://$host:$port/api/tanks/$1/connect
