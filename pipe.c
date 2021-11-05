// Copyright 2020, 2021 Google LLC
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
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "emacs-module.h"
#include "wrappers.h"

static_assert(UINTPTR_MAX == UINT64_MAX, "unsupported architecture");

struct integer_result open_channel(emacs_env *env, emacs_value value) {
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 28
  static_assert(SIZE_MAX >= PTRDIFF_MAX, "unsupported architecture");
  if ((size_t)env->size > offsetof(emacs_env, open_channel)) {
    static_assert(INT64_MIN <= INT_MIN, "unsupported architecture");
    static_assert(INT64_MAX >= INT_MAX, "unsupported architecture");
    return check_integer(env, env->open_channel(env, value));
  }
#endif
  return (struct integer_result){unimplemented(env), -1};
}
