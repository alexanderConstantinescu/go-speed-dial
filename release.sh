#!/bin/bash

set -e 

go test .

arr=("amd64" "386")

for CPU_V in ${arr[@]}; do
	echo "Building for Linux and $CPU_V"
	env GOOS=linux GOARCH=${CPU_V} go build sd.go			
	tar -czvf linux-${CPU_V}.tar.gz sd &> /dev/null
	rm sd
	echo "Building for Darwin and $CPU_V"
	env GOOS=darwin GOARCH=${CPU_V} go build sd.go			
	tar -czvf darwin-${CPU_V}.tar.gz sd &> /dev/null
	rm sd
	echo "Building for Windows and $CPU_V"
	env GOOS=windows GOARCH=${CPU_V} go build sd.go
	zip -q windows-${CPU_V}.zip sd.exe
	rm sd.exe
done
