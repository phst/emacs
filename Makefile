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

GO := go
GOFLAGS :=

CGO_CFLAGS := -pedantic-errors -Werror -Wall -Wextra \
  -Wno-sign-compare \
  -Wno-unused-parameter \
  -Wno-language-extension-token

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

bazel_major := $(shell $(BAZEL) --version | sed -E -n -e 's/^bazel ([[:digit:]]+)\..*$$/\1/p')

# The Buildifier target doesn’t work well on old Bazel versions.
buildifier_supported := $(shell test $(bazel_major) -ge 4 && echo yes)

all: buildifier vet check $(versions)

buildifier:
  ifeq ($(buildifier_supported),yes)
	$(BAZEL) run $(BAZELFLAGS) -- \
	  @com_github_bazelbuild_buildtools//buildifier \
	  --mode=check --lint=warn -r -- "$${PWD}"
  else
    $(warn Buildifier not supported on Bazel $(bazel_major))
  endif

vet:
        # Ensure that emacs-module.h exists, for the “go vet” command below.
	$(BAZEL) build $(BAZELFLAGS) -- '@gnu_emacs_27.2//:emacs-module.h'
	bin_dir="$$($(BAZEL) info bazel-bin)" \
	  && CGO_CFLAGS="$(CGO_CFLAGS) -isystem $${bin_dir}/external/gnu_emacs_27.2" \
	  $(GO) $(GOFLAGS) vet

check:
	$(BAZEL) test --test_output=errors $(BAZELFLAGS) -- //...

$(versions):
	$(MAKE) check BAZELFLAGS='$(BAZELFLAGS) --extra_toolchains=@phst_rules_elisp//elisp:emacs_$@_toolchain'

.PHONY: all buildifier vet check $(versions)
