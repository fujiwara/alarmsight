.PHONY: clean test

alarmsight: go.* *.go
	go build -o $@ cmd/alarmsight/main.go

clean:
	rm -rf alarmsight dist/

test:
	go test -v ./...

install:
	go install github.com/fujiwara/alarmsight/cmd/alarmsight

dist:
	goreleaser build --snapshot --rm-dist
