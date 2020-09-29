GO111MODULE=on

.PHONY: build
build: gocron

.PHONY: gocron
gocron:
	go build -o bin/server ./cmd/server.go
	go build -o bin/client ./cmd/client.go

.PHONY: clean
clean:
	rm bin/server
	rm bin/client