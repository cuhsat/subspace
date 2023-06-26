#!/bin/sh
GO_RUN="go run -race"
SERVER="../../cmd/subspace/main.go"
CLIENT="../../cmd/ss/main.go"

HOST="localhost"
PORT="8081"

$GO_RUN $SERVER 10 &

sleep 1s

$GO_RUN stats.go &

sleep 1s

echo "foo" | $GO_RUN $CLIENT $HOST $NAME
echo "bar" | $GO_RUN $CLIENT $HOST $NAME

sleep 1s

nc $HOST $PORT

killall -INT main stats
