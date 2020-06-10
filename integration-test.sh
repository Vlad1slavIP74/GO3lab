#!/usr/bin/env sh

cd ./integration || return
CGO_ENABLED=0 newbood
cat ./out/reports/test.txt