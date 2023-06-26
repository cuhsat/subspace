GO=go
CLIENT=ss
SERVER=subspace

.PHONY: all clean

all: build

build:
	mkdir -p bin
	${GO} build -o bin/${CLIENT} cmd/${CLIENT}/main.go
	${GO} build -o bin/${SERVER} cmd/${SERVER}/main.go

clean:
	rm -rf bin
	${GO} clean
