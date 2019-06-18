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

# We can’t link against GMP statically because Emacs links against the system
# GMP dynamically.  Therefore we add -lgmp to the linker options.  On macOS, we
# furthermore have to work around
# https://github.com/bazelbuild/bazel/issues/5391 by adding the local include
# and library directory.  We assume that the user installed GMP using Homebrew
# or similar, using the prefix /usr/local.

load(":def.bzl", "emacs_module")

SRCS = glob(
    [
        "*.go",
        "*.h",
        "*.c",
    ],
    exclude = ["*_test.go"],
)

TEST_SRCS = glob(["*_test.go"])

emacs_module(
    name = "stable",
    srcs = SRCS,
    header = "@emacs_module_header//:header",
    test_srcs = TEST_SRCS,
)

emacs_module(
    name = "master",
    srcs = SRCS,
    header = "@emacs_module_header_master//:header",
    test_srcs = TEST_SRCS,
)

load("@io_bazel_rules_go//go:def.bzl", "go_binary", "nogo")

go_binary(
    name = "genheader",
    srcs = ["genheader/main.go"],
    visibility = ["@emacs_module_header_master//:__pkg__"],
)

nogo(
    name = "nogo",
    vet = True,
    visibility = ["//visibility:public"],
)

load("@com_github_bazelbuild_buildtools//buildifier:def.bzl", "buildifier")

buildifier(
    name = "buildifier",
    lint_mode = "warn",
    lint_warnings = ["all"],
    mode = "fix",
)

config_setting(
    name = "linux",
    constraint_values = ["@bazel_tools//platforms:linux"],
)

config_setting(
    name = "macos",
    constraint_values = ["@bazel_tools//platforms:osx"],
)

# Local Variables:
# mode: bazel
# End:
