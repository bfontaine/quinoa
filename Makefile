all: quinoa

quinoa: main.go */*.go
	go build -o $@ ./main.go

parser/quinoa.peg.go: parser/quinoa.peg
	peg $<
