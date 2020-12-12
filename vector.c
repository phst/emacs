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
#include <stdint.h>

#include "emacs-module.h"
#include "wrappers.h"

static_assert(PTRDIFF_MIN == INT64_MIN, "unsupported architecture");
static_assert(PTRDIFF_MAX == INT64_MAX, "unsupported architecture");

struct value_result vec_get(emacs_env *env, emacs_value vec, int64_t i) {
  return check_value(env, env->vec_get(env, vec, i));
}

struct void_result vec_set(emacs_env *env, emacs_value vec, int64_t i,
                           emacs_value val) {
  env->vec_set(env, vec, i, val);
  return check_void(env);
}

struct integer_result vec_size(emacs_env *env, emacs_value vec) {
  return check_integer(env, env->vec_size(env, vec));
}
