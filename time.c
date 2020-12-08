// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#include <assert.h>
#include <emacs-module.h>
#include <inttypes.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>
#include <time.h>

#include "wrappers.h"

static_assert(PTRDIFF_MAX <= SIZE_MAX, "unsupported architecture");
static_assert((time_t)1.5 == 1, "unsupported architecture");
static_assert(LONG_MAX >= 1000000000, "unsupported architecture");

struct timespec_result extract_time(emacs_env *env, emacs_value value) {
  struct timespec_result result;
  result.value = env->extract_time(env, value);
  result.base = check(env);
  return result;
}

struct value_result make_time(emacs_env *env, struct timespec time) {
  assert(time.tv_nsec >= 0 && time.tv_nsec < 1000000000);
  return check_value(env, env->make_time(env, time));
}
