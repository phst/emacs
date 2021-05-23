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
    name = "emacs_module_header_master",
    build_file = "@//:header_master.BUILD",
    sha256 = "0ff2f2008b891943fe65193a15af57384f541ad92b01912fd37045c806a5b3f6",
    strip_prefix = "emacs-1b2547de23ef6bcab9ec791878178f5ade99bd19/src",
    urls = ["https://git.savannah.gnu.org/cgit/emacs.git/snapshot/emacs-1b2547de23ef6bcab9ec791878178f5ade99bd19.tar.gz"],
)

http_archive(
    name = "bazel_skylib",
    sha256 = "28f81e36692e1d87823623a99966b2daf85af3fdc1b40f98e37bd5294f3dd185",
    strip_prefix = "bazel-skylib-1.0.3",
    urls = ["https://github.com/bazelbuild/bazel-skylib/archive/1.0.3.zip"],
)

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "69de5c704a05ff37862f7e0f5534d4f479418afc21806c887db544a316f3cb6b",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.27.0/rules_go-v0.27.0.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.27.0/rules_go-v0.27.0.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(
    nogo = "@//:nogo",
    version = "1.16",
)

http_archive(
    name = "com_github_bazelbuild_buildtools",
    sha256 = "05eb52437fb250c7591dd6cbcfd1f9b5b61d85d6b20f04b041e0830dd1ab39b3",
    strip_prefix = "buildtools-0.29.0",
    urls = ["https://github.com/bazelbuild/buildtools/archive/0.29.0.zip"],
)

http_archive(
    name = "phst_rules_elisp",
    sha256 = "f326b45e846359718c36c6615abb5756e1b468cd8e9267a2ce50d9dc26b94664",
    strip_prefix = "rules_elisp-f9dabffc588f85059721c83c6352b23d5eff5923",
    urls = ["https://github.com/phst/rules_elisp/archive/f9dabffc588f85059721c83c6352b23d5eff5923.zip"],
)

load(
    "@phst_rules_elisp//elisp:repositories.bzl",
    "rules_elisp_dependencies",
    "rules_elisp_toolchains",
)

rules_elisp_dependencies()

rules_elisp_toolchains()

load("@com_github_bazelbuild_buildtools//buildifier:deps.bzl", "buildifier_dependencies")

buildifier_dependencies()

# Local Variables:
# mode: bazel-workspace
# End:
