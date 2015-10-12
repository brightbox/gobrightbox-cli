#!/bin/sh

set -ex

go get -v -t -d

mkdir -p bin/

for GOOS in windows darwin linux; do
	for GOARCH in 386 amd64; do
		export GOOS GOARCH
		go build -o bin/gobrightbox-$GOOS-$GOARCH
	done
done

cd bin/
sha256sum * > SHA256SUM
