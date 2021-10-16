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

"""Contains the macro emacs_module."""

load("@bazel_skylib//lib:paths.bzl", "paths")
load("@bazel_skylib//rules:copy_file.bzl", "copy_file")
load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")
load("@phst_rules_elisp//elisp:defs.bzl", "elisp_library", "elisp_test")

COPTS = [
    "-Werror",
    "-Wall",
    "-Wconversion",
    "-Wextra",
    "-Wno-sign-conversion",
    "-Wno-unused-parameter",
    "-fvisibility=hidden",
]

def emacs_module(name, srcs, header, test_srcs):
    """Generates an Emacs module library and associated tests.

    Args:
      name: name of the library rule
      srcs: Go sources for the library
      header: a cc_library containing the emacs-module.h header
      test_srcs: Go sources for the tests

    Generates:
      NAME: a Go library that implements an Emacs module
      NAME_go_test: a Go test for NAME
      NAME_elisp_test: an Emacs test for NAME
    """
    go_library(
        name = name,
        srcs = srcs,
        cdeps = [header],
        cgo = True,
        copts = COPTS,
        importpath = "github.com/phst/emacs",
    )
    bin_name = "_" + name + "_example"
    lib_name = bin_name + "_lib"
    elisp_lib_name = bin_name + "_elisp_lib"
    go_test(
        name = name + "_go_test",
        size = "medium",
        timeout = "short",
        srcs = test_srcs,
        embed = [name],
    )

    # Output the module with a fixed name so that (require 'example-module)
    # works.  Note that we use the .so suffix on macOS as well due to
    # https://debbugs.gnu.org/cgi/bugreport.cgi?bug=36226.  We can switch to
    # .dylib once we drop support for Emacs 27.
    mod_name = paths.join(name, "example-module.so")

    # The Emacs Lisp Bazel rules don’t allow multiple libraries with
    # overlapping source files, so make a per-target copy of the test file.
    test_el = "_" + name + "_test.el"
    copy_file(
        name = "_" + name + "_copy",
        src = "//:test.el",
        out = test_el,
    )
    elisp_test(
        name = name + "_elisp_test",
        size = "medium",
        timeout = "short",
        srcs = [test_el],
        deps = [elisp_lib_name],
    )
    elisp_library(
        name = elisp_lib_name,
        srcs = [mod_name],
        load_path = [name],
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
        linkmode = "c-shared",
        deps = [lib_name],
    )

    # We copy the module file so that it’s guaranteed to be in the “bin”
    # directory of the “elisp_library” rule.  “go_binary” seems to add a
    # configuration transition.  This should better be addressed in the
    # implementation of “elisp_library” itself.
    copy_file(
        name = bin_name + "_copy",
        src = bin_name,
        out = mod_name,
    )
