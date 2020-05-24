#!/usr/bin/env sh

cd ./integration || return
CGO_ENABLED=0 bood
cat ./build/out/reports/test.txt