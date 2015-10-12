#!/bin/sh

set -ex

go get -v -t -d

mkdir -p bin/
rm -f bin/* || true

export GOOS GOARCH

for GOOS in darwin linux; do
	for GOARCH in 386 amd64; do
		go build -o bin/gobrightbox-$GOOS-$GOARCH
	done
done

GOOS=windows
for GOARCH in 386 amd64; do
	go build -o bin/gobrightbox-$GOOS-$GOARCH.exe
done

cd bin/
sha256sum * > SHA256SUM
