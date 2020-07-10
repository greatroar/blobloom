#!/bin/sh

set -e -x

if [ "$(go env GOARCH)" = amd64 ]; then
	go test
	go test -tags purego
	GOARCH=386 go test
else
	go test
fi

for e in examples/*; do
	(cd $e && go build && rm $(basename $e))
done
