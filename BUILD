# Copyright 2019, 2021, 2022, 2023, 2024 Google LLC
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

load("@bazel_skylib//rules:copy_file.bzl", "copy_file")
load("@buildifier_prebuilt//:rules.bzl", "buildifier", "buildifier_test")
load("@io_bazel_rules_go//go:def.bzl", "TOOLS_NOGO", "go_binary", "go_library", "go_test", "nogo")
load("@phst_rules_elisp//elisp:defs.bzl", "elisp_library", "elisp_test")

TEST_SRCS = glob(
    ["*_test.go"],
    allow_empty = False,
)

go_library(
    name = "go_default_library",
    srcs = glob(
        [
            "*.go",
            "*.h",
            "*.c",
        ],
        allow_empty = False,
        exclude = ["*_test.go"],
    ),
    cdeps = ["@phst_rules_elisp//emacs:module_header"],
    cgo = True,
    copts = [
        "-Werror",
        "-Wall",
        "-Wconversion",
        "-Wextra",
        "-Wno-sign-conversion",
        "-Wno-unused-parameter",
        "-fvisibility=hidden",
    ],
    importpath = "github.com/phst/emacs",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_test",
    size = "medium",
    timeout = "short",
    srcs = TEST_SRCS,
    embed = [":go_default_library"],
)

elisp_test(
    name = "elisp_test",
    size = "medium",
    timeout = "short",
    srcs = ["test.el"],
    deps = [
        ":example_elisp_lib",
        "@aio//:library",
    ],
)

elisp_library(
    name = "example_elisp_lib",
    srcs = ["example-module.so"],
)

go_library(
    name = "example_lib",
    srcs = TEST_SRCS,
    embed = [":go_default_library"],
    importpath = "github.com/phst/emacs",
)

go_binary(
    name = "example",
    srcs = ["example/main.go"],
    linkmode = "c-shared",
    deps = [":example_lib"],
)

# We copy the module file so that it’s guaranteed to be in the “bin”
# directory of the “elisp_library” rule.  “go_binary” seems to add a
# configuration transition.  This should better be addressed in the
# implementation of “elisp_library” itself.
copy_file(
    name = "example_copy",
    src = "example",
    # Output the module with a fixed name so that (require 'example-module)
    # works.  Note that we use the .so suffix on macOS as well due to
    # https://debbugs.gnu.org/cgi/bugreport.cgi?bug=36226.  We can switch to
    # .dylib once we drop support for Emacs 27.
    out = "example-module.so",
)

nogo(
    name = "nogo",
    config = "nogo.json",
    visibility = ["//visibility:public"],
    deps = TOOLS_NOGO,
)

buildifier(
    name = "buildifier",
    lint_mode = "warn",
    lint_warnings = ["all"],
    mode = "fix",
)

buildifier_test(
    name = "buildifier_test",
    size = "small",
    srcs = [
        "BUILD",
        "WORKSPACE",
    ] + glob(
        ["*.bazel"],
        allow_empty = False,
    ),
    lint_mode = "warn",
    mode = "check",
)

# Local Variables:
# mode: bazel-build
# End:
