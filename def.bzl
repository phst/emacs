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

"""Contains the macro emacs_module."""

load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

# We can’t link against GMP statically because Emacs links against the system
# GMP dynamically.  Therefore we add -lgmp to the linker options.  On macOS, we
# furthermore have to work around
# https://github.com/bazelbuild/bazel/issues/5391 by adding the local include
# and library directory.  We assume that the user installed GMP using Homebrew
# or similar, using the prefix /usr/local.

_COPTS = [
    "-Werror",
    "-Wall",
    "-Wextra",
    "-Wno-unused-parameter",
    "-DEMACS_MODULE_GMP",
    "-fvisibility=hidden",
] + select({
    ":linux": [],
    ":macos": ["-isystem", "/usr/local/include"],
})

_LINKOPTS = ["-lgmp"] + select({
    ":linux": [],
    ":macos": ["-L/usr/local/lib"],
})

def emacs_module(name, srcs, header, test_srcs):
    """Generates an Emacs module library and associated tests.

    Args:
      name: name of the library rule
      srcs: Go sources for the library
      header: a cc_library containing the emacs-module.h header
      test_srcs: Go sources for the tests

    Generates:
      NAME: a Go library that implements an Emacs module
      NAME_test: a test for NAME
    """
    go_library(
        name = name,
        srcs = srcs,
        cdeps = [header],
        cgo = True,
        clinkopts = _LINKOPTS,
        copts = _COPTS,
        importpath = "github.com/phst/emacs",
    )
    bin_name = "_" + name + "_example"
    lib_name = bin_name + "_lib"
    go_test(
        name = name + "_test",
        size = "medium",
        timeout = "short",
        srcs = test_srcs,
        embed = [name],
        data = [bin_name, "//:test.el"],
        args = [
            "--module=$(location " + bin_name + ")",
            "--ert_tests=$(location //:test.el)",
        ],
    )
    go_library(
        name = lib_name,
        srcs = test_srcs,
        embed = [name],
        importpath = "github.com/phst/emacs",
    )
    go_binary(
        name = bin_name,
        srcs = ["//:example/main.go"],
        out = select(
            {
                ":linux": None,
                # Work around https://debbugs.gnu.org/cgi/bugreport.cgi?bug=36226.
                ":macos": bin_name + ".so",
            },
            no_match_error = "unsupported platform",
        ),
        linkmode = "c-shared",
        deps = [lib_name],
    )
