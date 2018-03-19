#!/bin/sh
find . -name '*.go' -exec sed -i "s#github.com/Sirupsen#github.com/sirupsen#g" {} \;
