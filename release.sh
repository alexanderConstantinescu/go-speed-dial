#!/bin/bash

set -e 

arr=("amd64" "386")

for CPU_V in ${arr[@]}; do
	env GOOS=linux GOARCH=${CPU_V} go build sd.go			
	env GOOS=windows GOARCH=${CPU_V} go build sd.go
	tar -czvf linux-${CPU_V}.tar.gz sd	
	zip windows-${CPU_V}.zip sd.exe
done

rm sd sd.exe
