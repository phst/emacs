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
    sha256 = "151ce69dbe5b809d4492ffae4a4b153b2778459de6deb26f35691e1281a9c58e",
    strip_prefix = "emacs-26.2/src/",
    urls = ["https://ftp.gnu.org/gnu/emacs/emacs-26.2.tar.xz"],
)

http_archive(
    name = "emacs_module_header_master",
    build_file = "@//:header_master.BUILD",
    sha256 = "ecc8f1f0260811c1c82de2c5006aea6d2ed0d9f31b8033437a5a693657d6d63a",
    strip_prefix = "emacs-622bfdffa8b0c830bc6a979a2e9c114bad1ac114/src",
    urls = ["https://git.savannah.gnu.org/cgit/emacs.git/snapshot/emacs-622bfdffa8b0c830bc6a979a2e9c114bad1ac114.tar.gz"],
)

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "f04d2373bcaf8aa09bccb08a98a57e721306c8f6043a2a0ee610fd6853dcde3d",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/rules_go/releases/download/0.18.6/rules_go-0.18.6.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/0.18.6/rules_go-0.18.6.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(nogo = "@//:nogo")

http_archive(
    name = "com_github_bazelbuild_buildtools",
    sha256 = "5fb946659443db737844bfd07fec58acf92a8213567306641b56da2993c50ffa",
    strip_prefix = "buildtools-0.26.0",
    urls = ["https://github.com/bazelbuild/buildtools/archive/0.26.0.zip"],
)

load("@com_github_bazelbuild_buildtools//buildifier:deps.bzl", "buildifier_dependencies")

buildifier_dependencies()

# Local Variables:
# mode: bazel
# End:
