#!/bin/bash -ex
go run cmd/font2go/*.go webfonts/*.svg
mv fonts.go gerber
