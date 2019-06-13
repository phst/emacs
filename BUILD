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

go_library(
    name = "emacs",
    srcs = glob(
        ["*.go"],
        exclude = [
            "*_test.go",
            "example_main.go",
        ],
    ) + ["trampoline.h"],
    cdeps = ["@emacs_module_header//:header"],
    cgo = True,
    copts = COPTS,
    importpath = "github.com/phst/emacs",
)

go_test(
    name = "go_test",
    srcs = glob(
        ["*_test.go"],
        exclude = ["example_test.go"],
    ),
    embed = [":emacs"],
)

sh_test(
    name = "emacs_test",
    srcs = ["test.sh"],
    args = [
        "$(location :example)",
        "$(location test.el)",
    ],
    data = [
        "test.el",
        ":example",
    ],
)

go_library(
    name = "example_lib",
    srcs = glob(["*_test.go"]),
    embed = [":emacs"],
    importpath = "github.com/phst/emacs",
)

go_binary(
    name = "example",
    srcs = ["example/main.go"],
    linkmode = "c-shared",
    deps = [":example_lib"],
)

nogo(
    name = "nogo",
    vet = True,
    visibility = ["//visibility:public"],
)

# Local Variables:
# mode: bazel
# End:
