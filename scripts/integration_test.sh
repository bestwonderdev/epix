#!/bin/bash

apt install ginkgo
go install github.com/onsi/ginkgo/v2/ginkgo@latest
go install github.com/onsi/gomega/...@latest
export PATH=$PATH:$(go env GOPATH)/bin
cd "$PWD"/x/inflation/keeper && ginkgo
