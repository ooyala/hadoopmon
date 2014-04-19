# Copyright 2014 Ooyala, Inc. All rights reserved.
#
# This file is licensed under the MIT license.
# See the LICENSE file for details.

.PHONY: all build install test

CURRENT_DIR = $(shell  cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
SOURCE_DIR = htools

VERSION = $(shell cat $(CURRENT_DIR)/VERSION)
ifeq ($(strip $(shell git status --porcelain)),)
	GITCOMMIT = $(shell git rev-parse --short HEAD)
else
	GITCOMMIT = $(shell git rev-parse --short HEAD)-dirty
endif

LDFLAGS="-X main.GITCOMMIT '$(GITCOMMIT)' -X main.VERSION '$(VERSION)' -w"

default: all

all: build

build:
	go build -ldflags $(LDFLAGS)

install:
	go install -ldflags $(LDFLAGS)

test:
	cd $(CURRENT_DIR)/$(SOURCE_DIR); go test -v
