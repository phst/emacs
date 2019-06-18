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

COPTS = [
    "-Werror",
    "-Wall",
    "-Wextra",
    "-Wno-unused-parameter",
    "-DEMACS_MODULE_GMP",
] + select({
    ":linux": [],
    ":macos": ["-I/usr/local/include"],
})

LINKOPTS = ["-lgmp"] + select({
    ":linux": [],
    ":macos": ["-L/usr/local/lib"],
})

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
    clinkopts = LINKOPTS,
    copts = COPTS,
    importpath = "github.com/phst/emacs",
) for suffix in SUFFIXES]

[go_test(
    name = "go" + suffix + "_test",
    size = "small",
    timeout = "short",
    srcs = glob(["*_test.go"]),
    embed = [":emacs" + suffix],
) for suffix in SUFFIXES]

[sh_test(
    name = "emacs" + suffix + "_test",
    size = "medium",
    timeout = "short",
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
    out = select(
        {
            ":linux": None,
            # Work around https://debbugs.gnu.org/cgi/bugreport.cgi?bug=36226.
            ":macos": "example" + suffix + ".so",
        },
        no_match_error = "unsupported platform",
    ),
    linkmode = "c-shared",
    deps = [":example" + suffix + "_lib"],
) for suffix in SUFFIXES]

nogo(
    name = "nogo",
    vet = True,
    visibility = ["//visibility:public"],
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
