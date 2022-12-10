# Copyright 2019, 2021, 2022 Google LLC
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
    name = "emacs_module_header_master",
    build_file = "@//:header_master.BUILD",
    sha256 = "0ff2f2008b891943fe65193a15af57384f541ad92b01912fd37045c806a5b3f6",
    strip_prefix = "emacs-1b2547de23ef6bcab9ec791878178f5ade99bd19/src",
    urls = ["https://git.savannah.gnu.org/cgit/emacs.git/snapshot/emacs-1b2547de23ef6bcab9ec791878178f5ade99bd19.tar.gz"],
)

http_archive(
    name = "bazel_skylib",
    sha256 = "74d544d96f4a5bb630d465ca8bbcfe231e3594e5aae57e1edbf17a6eb3ca2506",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
        "https://github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
    ],
)

load("@bazel_skylib//:workspace.bzl", "bazel_skylib_workspace")

bazel_skylib_workspace()

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "ae013bf35bd23234d1dea46b079f1e05ba74ac0321423830119d3e787ec73483",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.36.0/rules_go-v0.36.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.36.0/rules_go-v0.36.0.zip",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(
    nogo = "@//:nogo",
    version = "1.19.3",
)

http_archive(
    name = "com_google_protobuf",
    sha256 = "0faa3f28b150efcaf044d33a05f613e3b052b852f41d2d420eb7ad49d11a06df",
    strip_prefix = "protobuf-21.11/",
    urls = [
        "https://github.com/protocolbuffers/protobuf/archive/refs/tags/v21.11.zip",  # 2022-12-08
    ],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

http_archive(
    name = "com_github_bazelbuild_buildtools",
    sha256 = "e3bb0dc8b0274ea1aca75f1f8c0c835adbe589708ea89bf698069d0790701ea3",
    strip_prefix = "buildtools-5.1.0/",
    urls = [
        "https://github.com/bazelbuild/buildtools/archive/refs/tags/5.1.0.tar.gz",  # 2022-04-13
    ],
)

http_archive(
    name = "phst_rules_elisp",
    sha256 = "ca04b2dc7dc345d9746c6dc723879fb4a110e62dc17ebe7a4bbc982d76e60bc2",
    strip_prefix = "rules_elisp-4482e8e3fe4dc5b93aad597342d70ef4762f93f0",
    urls = [
        "https://github.com/phst/rules_elisp/archive/4482e8e3fe4dc5b93aad597342d70ef4762f93f0.zip",  # 2022-11-28
    ],
)

load(
    "@phst_rules_elisp//elisp:repositories.bzl",
    "rules_elisp_dependencies",
    "rules_elisp_toolchains",
)

rules_elisp_dependencies()

rules_elisp_toolchains()

http_archive(
    name = "aio",
    build_file = "//:aio.BUILD",
    sha256 = "63e3170f2d720995b318bc1feb8414fca3dea16cb707f24a981f3f0cade3fcbf",
    strip_prefix = "emacs-aio-da93523e235529fa97d6f251319d9e1d6fc24a41/",
    urls = [
        "https://github.com/skeeto/emacs-aio/archive/da93523e235529fa97d6f251319d9e1d6fc24a41.zip",  # 2020-06-10
    ],
)

load("@com_github_bazelbuild_buildtools//buildifier:deps.bzl", "buildifier_dependencies")

buildifier_dependencies()

# Local Variables:
# mode: bazel-workspace
# End:
