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

COPTS = [
    "-Werror",
    "-Wall",
    "-Wextra",
    "-Wno-unused-parameter",
]

load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test", "nogo")

SUFFIXES = [
    "",
    "_master",
]

[go_library(
    name = "emacs" + suffix,
    srcs = glob(
        ["*.go"],
        exclude = [
            "*_test.go",
            "example_main.go",
        ],
    ) + ["trampoline.h"],
    cdeps = ["@emacs_module_header" + suffix + "//:header"],
    cgo = True,
    copts = COPTS,
    importpath = "github.com/phst/emacs",
) for suffix in SUFFIXES]

[go_test(
    name = "go" + suffix + "_test",
    srcs = glob(
        ["*_test.go"],
        exclude = ["example_test.go"],
    ),
    embed = [":emacs" + suffix],
) for suffix in SUFFIXES]

[sh_test(
    name = "emacs" + suffix + "_test",
    srcs = ["test.sh"],
    args = [
        "$(location :example" + suffix + ")",
        "$(location test.el)",
    ],
    data = [
        "test.el",
        ":example" + suffix,
    ],
) for suffix in SUFFIXES]

[go_library(
    name = "example" + suffix + "_lib",
    srcs = glob(["*_test.go"]),
    embed = [":emacs" + suffix],
    importpath = "github.com/phst/emacs",
) for suffix in SUFFIXES]

[go_binary(
    name = "example" + suffix,
    srcs = ["example/main.go"],
    linkmode = "c-shared",
    deps = [":example" + suffix + "_lib"],
) for suffix in SUFFIXES]

nogo(
    name = "nogo",
    vet = True,
    visibility = ["//visibility:public"],
)

# Local Variables:
# mode: bazel
# End:
