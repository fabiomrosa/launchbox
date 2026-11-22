#!/usr/bin/env bash

start_port=3000
max_port=3999

is_port_free() {
  ! lsof -i TCP:"$1" -sTCP:LISTEN > /dev/null 2>&1
}

command -v lsof >/dev/null 2>&1 || {
  echo "Error: lsof is not installed."
  exit 1
}

port=$start_port
while ! is_port_free "$port"; do
  port=$((port + 1))
  [ "$port" -gt "$max_port" ] && {
    echo "No free ports available in range $start_port-$max_port."
    exit 1
  }
done

exec env NODE_ENV=development PORT="$port" node server.js
