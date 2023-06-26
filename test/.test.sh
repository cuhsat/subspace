#!/bin/sh
GO_RUN="go run -race"
SERVER="../cmd/subspace/main.go"
CLIENT="../cmd/ss/main.go"

HOST="localhost"
NAME="test"

$GO_RUN $SERVER 10 &

sleep 1s

echo "foo" | $GO_RUN $CLIENT $HOST
echo "bar" | $GO_RUN $CLIENT $HOST

$GO_RUN $CLIENT $HOST $NAME
$GO_RUN $CLIENT $HOST $NAME

killall -INT main
