#!/usr/bin/env bash
export version=$(grep '"version"' plugin.json | sed 's/"version":[[:space:]]*"//' | sed 's/",//' | tr -d [:space:])
go run build/make.go --all-platforms --distro
cd deploy
curl -sSfL https://github.com/getgauge/gauge/raw/master/build/github_release.sh | sh