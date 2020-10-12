GO111MODULE=on

.PHONY: build
build: gocron

.PHONY: gocron
gocron:
	go build -o bin/executor ./cmd/executor.go
	go build -o bin/dispacher ./cmd/dispacher.go

.PHONY: clean
clean:
	rm bin/executor
	rm bin/dispacher