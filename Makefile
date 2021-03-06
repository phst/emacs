# Copyright 2019, 2021 Google LLC
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

SHELL := /bin/sh

.DEFAULT: all
.SUFFIXES:

BAZEL := bazel
BAZELFLAGS :=

# All potentially supported Emacs versions.
versions := 27.1 27.2

kernel := $(shell uname -s)
ifeq ($(kernel),Linux)
  # GNU/Linux supports all Emacs versions.
else ifeq ($(kernel),Darwin)
  ifneq ($(shell uname -m),x86_64)
    # Apple Silicon doesn’t support Emacs 27.1.
    unsupported := 27.1
  endif
else
  $(error Unsupported kernel $(kernel))
endif

versions := $(filter-out $(unsupported),$(versions))

all: check $(versions)

check:
	$(BAZEL) test --test_output=errors $(BAZELFLAGS) -- //...

$(versions):
	$(MAKE) check BAZELFLAGS='$(BAZELFLAGS) --extra_toolchains=@phst_rules_elisp//elisp:emacs_$@_toolchain'

.PHONY: all check $(versions)
