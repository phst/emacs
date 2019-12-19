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

workspace(name = "com_github_phst_emacs")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "emacs_module_header",
    build_file = "@//:header.BUILD",
    sha256 = "4d90e6751ad8967822c6e092db07466b9d383ef1653feb2f95c93e7de66d3485",
    strip_prefix = "emacs-26.3/src/",
    urls = ["https://ftp.gnu.org/gnu/emacs/emacs-26.3.tar.xz"],
)

http_archive(
    name = "emacs_module_header_master",
    build_file = "@//:header_master.BUILD",
    sha256 = "10bd66905d3ca32ab8b7bcbaa5d2037079be8e0bc3f45261696a17a06e57e9c0",
    strip_prefix = "emacs-3f36cab333a01bec3850d27ac0b2383570edb14e/src",
    urls = ["https://git.savannah.gnu.org/cgit/emacs.git/snapshot/emacs-3f36cab333a01bec3850d27ac0b2383570edb14e.tar.gz"],
)

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "e88471aea3a3a4f19ec1310a55ba94772d087e9ce46e41ae38ecebe17935de7b",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/rules_go/releases/download/v0.20.3/rules_go-v0.20.3.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.20.3/rules_go-v0.20.3.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(nogo = "@//:nogo")

http_archive(
    name = "com_github_bazelbuild_buildtools",
    sha256 = "05eb52437fb250c7591dd6cbcfd1f9b5b61d85d6b20f04b041e0830dd1ab39b3",
    strip_prefix = "buildtools-0.29.0",
    urls = ["https://github.com/bazelbuild/buildtools/archive/0.29.0.zip"],
)

load("@com_github_bazelbuild_buildtools//buildifier:deps.bzl", "buildifier_dependencies")

buildifier_dependencies()

# Local Variables:
# mode: bazel
# End:
