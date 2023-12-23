# Copyright 2019, 2021, 2022, 2023 Google LLC
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
load(":def.bzl", "COPTS")

SRCS = glob(
    [
        "*.go",
        "*.h",
        "*.c",
    ],
    exclude = ["*_test.go"],
)

TEST_SRCS = glob(["*_test.go"])

go_library(
    name = "go_default_library",
    srcs = SRCS,
    cdeps = ["@phst_rules_elisp//emacs:module_header"],
    cgo = True,
    copts = COPTS,
    importpath = "github.com/phst/emacs",
    visibility = ["//visibility:public"],
)

go_library(
    name = "stable",
    srcs = SRCS,
    cdeps = ["@phst_rules_elisp//emacs:module_header"],
    cgo = True,
    copts = COPTS,
    importpath = "github.com/phst/emacs",
)

go_test(
    name = "stable_go_test",
    size = "medium",
    timeout = "short",
    srcs = TEST_SRCS,
    embed = ["stable"],
)

# The Emacs Lisp Bazel rules don’t allow multiple libraries with
# overlapping source files, so make a per-target copy of the test file.
copy_file(
    name = "_stable_copy",
    src = "//:test.el",
    out = "_stable_test.el",
)

elisp_test(
    name = "stable_elisp_test",
    size = "medium",
    timeout = "short",
    srcs = ["_stable_test.el"],
    deps = [
        "_stable_example_elisp_lib",
        "@aio",
    ],
)

elisp_library(
    name = "_stable_example_elisp_lib",
    srcs = ["stable/example-module.so"],
    load_path = ["stable"],
)

go_library(
    name = "_stable_example_lib",
    srcs = TEST_SRCS,
    embed = ["stable"],
    importpath = "github.com/phst/emacs",
)

go_binary(
    name = "_stable_example",
    srcs = ["//:example/main.go"],
    linkmode = "c-shared",
    deps = ["_stable_example_lib"],
)

# We copy the module file so that it’s guaranteed to be in the “bin”
# directory of the “elisp_library” rule.  “go_binary” seems to add a
# configuration transition.  This should better be addressed in the
# implementation of “elisp_library” itself.
copy_file(
    name = "_stable_example_copy",
    src = "_stable_example",
    # Output the module with a fixed name so that (require 'example-module)
    # works.  Note that we use the .so suffix on macOS as well due to
    # https://debbugs.gnu.org/cgi/bugreport.cgi?bug=36226.  We can switch to
    # .dylib once we drop support for Emacs 27.
    out = "stable/example-module.so",
)

go_binary(
    name = "genheader",
    srcs = ["genheader/main.go"],
    visibility = ["@emacs_module_header_master//:__pkg__"],
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
    ] + glob([
        "*.BUILD",
        "*.WORKSPACE",
        "*.bazel",
        "*.bzl",
    ]),
    lint_mode = "warn",
    mode = "check",
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
# mode: bazel-build
# End:
