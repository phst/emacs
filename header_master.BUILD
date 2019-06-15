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

cc_library(
    name = "header",
    hdrs = ["emacs-module.h"],
    visibility = ["//visibility:public"],
)

VERSIONS = [
    25,
    26,
    27,
]

genrule(
    name = "gen_header",
    srcs = ["emacs-module.h.in"] + ["module-env-{}.h".format(ver) for ver in VERSIONS],
    outs = ["emacs-module.h"],
    cmd = (
        "sed --regexp-extended --expression='s/@emacs_major_version@/{}/'".format(VERSIONS[-1]) +
        " ".join([
            " --expression='/@module_env_snippet_{ver}@/{{r $(location module-env-{ver}.h)\nd}}'".format(ver = ver)
            for ver in VERSIONS
        ]) +
        " -- $(location emacs-module.h.in) > $@"
    ),
)
# Local Variables:
# mode: bazel
# End:
