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
    sha256 = "c6966ec828da198c5d9adbaa94c05e3a1c7f21bd012a0b29ba8ddbccb2c93b0d",
    urls = [
        "https://github.com/bazelbuild/bazel-skylib/releases/download/1.1.1/bazel-skylib-1.1.1.tar.gz",
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.1.1/bazel-skylib-1.1.1.tar.gz",
    ],
)

load("@bazel_skylib//:workspace.bzl", "bazel_skylib_workspace")

bazel_skylib_workspace()

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "2b1641428dff9018f9e85c0384f03ec6c10660d935b750e3fa1492a281a53b0f",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.29.0/rules_go-v0.29.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.29.0/rules_go-v0.29.0.zip",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(
    nogo = "@//:nogo",
    version = "1.17.1",
)

http_archive(
    name = "com_google_protobuf",
    sha256 = "9ceef0daf7e8be16cd99ac759271eb08021b53b1c7b6edd399953a76390234cd",
    strip_prefix = "protobuf-3.19.2/",
    urls = [
        "https://github.com/protocolbuffers/protobuf/archive/refs/tags/v3.19.2.zip",  # 2022-01-05
    ],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

http_archive(
    name = "com_github_bazelbuild_buildtools",
    sha256 = "97eb2b241da89b50a0a28a45ae4408661ab740c682db550ea611e8b57334ea9e",
    strip_prefix = "buildtools-4.2.4/",
    urls = [
        "https://github.com/bazelbuild/buildtools/archive/refs/tags/4.2.4.zip",  # 2021-11-29
    ],
)

http_archive(
    name = "phst_rules_elisp",
    sha256 = "56929525dea4a1281ddec49aee6dc4824e2b59d3e3575930babd56186804d29d",
    strip_prefix = "rules_elisp-8eee3f354e77d668d244c41828c055a078c52480",
    urls = [
        "https://github.com/phst/rules_elisp/archive/8eee3f354e77d668d244c41828c055a078c52480.zip",  # 2021-10-20
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
