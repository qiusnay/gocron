GO111MODULE=on

.PHONY: build
build: gocron

.PHONY: gocron
gocron:
	go build $(RACE) -o bin/gocron ./bash/

.PHONY: clean
clean:
	rm bin/gocron

.PHONY: test
test:
	go test $(RACE) ./...

.PHONY: enable-race
enable-race:
	$(eval RACE = -race)