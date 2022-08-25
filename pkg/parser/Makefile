install:
	go get golang.org/x/tools/cmd/goyacc & go get github.com/blynn/nex
	go install golang.org/x/tools/cmd/goyacc & go install github.com/blynn/nex

parser:
	nex lexer.nex
	goyacc -o parser.go parser.y 2>&1  | egrep "(shift|reduce)/reduce" | awk '{print} END {if (NR > 0) {print "Find conflict in parser.y. Please check y.output for more information."; exit 1;}}'
	rm -f y.output

clean:
	go clean -i ./...
	rm -rf *.out
	rm -rf parser.go