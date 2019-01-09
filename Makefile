# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SHELL := /bin/bash
EMACS := emacs

export CGO_CFLAGS := -pedantic-errors -Werror -Wall -Wextra -Weverything \
  -Wno-language-extension-token \
  -Wno-missing-prototypes \
  -Wno-missing-variable-declarations \
  -Wno-packed \
  -Wno-reserved-id-macro \
  -Wno-sign-conversion \
  -Wno-strict-prototypes \
  -Wno-unused-macros \
  -Wno-unused-parameter \
  -Wno-used-but-marked-unused

check: go-test emacs-test

go-test: *.go *.h
	go build
	go vet
	golint -set_exit_status -min_confidence=0.3
	go test

emacs-test: example.so test.el
	$(EMACS) --quick --batch --module-assertions \
	  --load=ert --load=example.so --load=test.el \
	  --funcall=ert-run-tests-batch-and-exit

example.so: *.go *.h casts.go
	go test -c -o $@ -buildmode=c-shared -tags=example

.PHONY: check go-test emacs-test
