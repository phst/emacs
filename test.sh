#!/bin/bash

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

shopt -s -o errexit noclobber noglob nounset pipefail
shopt -u -o braceexpand history
shopt -s failglob

IFS=''

if (($# != 2)); then
  echo "wrong number of arguments, got $#, want two" >&2
  exit 2
fi

declare -r EXAMPLE_MODULE="$1"
declare -r TEST_EL="$2"

shopt -s -o xtrace

"${EMACS:-emacs}" --quick --batch --module-assertions \
  --load=ert --load="${EXAMPLE_MODULE:?}" --load="${TEST_EL:?}" \
  --funcall=ert-run-tests-batch-and-exit
