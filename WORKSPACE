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
    sha256 = "4a4c128f915fc937d61edfc273c98106711b540c9be3cd5d2e2b9b5b2f172e41",
    strip_prefix = "emacs-27.1/src/",
    urls = [
        "https://ftpmirror.gnu.org/emacs/emacs-27.1.tar.xz",
        "https://ftp.gnu.org/gnu/emacs/emacs-27.1.tar.xz",
    ],
)

http_archive(
    name = "emacs_module_header_master",
    build_file = "@//:header_master.BUILD",
    sha256 = "0ff2f2008b891943fe65193a15af57384f541ad92b01912fd37045c806a5b3f6",
    strip_prefix = "emacs-1b2547de23ef6bcab9ec791878178f5ade99bd19/src",
    urls = ["https://git.savannah.gnu.org/cgit/emacs.git/snapshot/emacs-1b2547de23ef6bcab9ec791878178f5ade99bd19.tar.gz"],
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
# mode: bazel-workspace
# End:
