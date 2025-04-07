#!/bin/sh
set -e

host="$1"
port="$2"
shift 2
cmd="$@"

echo "wait for PostgreSQL at $host:$port..."
while ! nc -z "$host" "$port"; do
  echo "waiting..."
  sleep 1
done

echo "PostgreSQL is available, lets do: $cmd"
exec $cmd