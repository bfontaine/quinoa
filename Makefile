all: quinoa

.PHONY: all parser-tests

quinoa: main.go */*.go
	go build -o $@ ./main.go

parser/quinoa.peg.go: parser/quinoa.peg
	peg $<

parser-tests: parser/quinoa.peg.go
	go test -v ./parser/...
