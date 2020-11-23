#!/bin/bash

CGO_ENABLED=0 go build -mod=vendor -o buildpacks/dependent/bin/detect ./vendor/github.com/vaikas/buildpackstuff/cmd/detect
pack package-buildpack "${@}"
