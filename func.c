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
#include <stdint.h>

#include "wrappers.h"

static_assert(PTRDIFF_MIN == INT64_MIN, "unsupported architecture");
static_assert(PTRDIFF_MAX == INT64_MAX, "unsupported architecture");
static_assert(UINTPTR_MAX == UINT64_MAX, "unsupported architecture");

emacs_value trampoline(emacs_env *env, ptrdiff_t nargs, emacs_value *args,
                       void *data) {
  return go_emacs_trampoline(env, nargs, args, (uintptr_t)data);
}

emacs_value funcall(emacs_env *env, emacs_value function, int64_t nargs,
                    emacs_value *args) {
  return env->funcall(env, function, nargs, args);
}
