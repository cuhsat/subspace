#!/bin/sh
GO_RUN="go run -race"
SERVER="../../cmd/subspace/main.go"
CLIENT="../../cmd/ss/main.go"

HOST="localhost"
PORT="8080"

export SUBSPACE_RETENTION=10

$GO_RUN $SERVER &

sleep 1s

$GO_RUN proxy.go &

sleep 1s

echo "foo" | $GO_RUN $CLIENT $HOST $NAME
echo "bar" | $GO_RUN $CLIENT $HOST $NAME

sleep 1s

curl $HOST:$PORT/test

killall -INT main proxy
