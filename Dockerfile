FROM golang:alpine AS build

WORKDIR /app

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o /bin/subspace cmd/subspace/main.go

FROM scratch
COPY --from=build /bin/subspace /bin/subspace

EXPOSE 8211/udp
EXPOSE 8212/udp

ENTRYPOINT ["/bin/subspace"]
