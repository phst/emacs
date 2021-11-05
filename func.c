// Copyright 2019, 2021 Google LLC
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
#include <stddef.h>
#include <stdint.h>

#include "emacs-module.h"
#include "wrappers.h"

static_assert(PTRDIFF_MIN == INT64_MIN, "unsupported architecture");
static_assert(PTRDIFF_MAX == INT64_MAX, "unsupported architecture");
static_assert(UINTPTR_MAX == UINT64_MAX, "unsupported architecture");
static_assert(PTRDIFF_MAX <= SIZE_MAX, "unsupported architecture");

static emacs_value trampoline(emacs_env *env, ptrdiff_t nargs,
                              emacs_value *args, void *data) {
  struct trampoline_result result =
      go_emacs_trampoline(env, nargs, args, (uintptr_t)data);
  handle_nonlocal_exit(env, result.base);
  return result.value;
}

#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 28
static void finalizer(void *data) {
  go_emacs_function_finalizer((uintptr_t)data);
}
#endif

struct value_result funcall(emacs_env *env, emacs_value function, int64_t nargs,
                            emacs_value *args) {
  return check_value(env, env->funcall(env, function, nargs, args));
}

struct value_result make_function_impl(emacs_env *env, int64_t min_arity,
                                       int64_t max_arity,
                                       const char *documentation,
                                       uint64_t data) {
  emacs_value value =
      env->make_function(env, min_arity, max_arity, trampoline, documentation,
                         (void *)(uintptr_t)data);
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 28
  if ((size_t)env->size > offsetof(emacs_env, set_function_finalizer))
    env->set_function_finalizer(env, value, finalizer);
#endif
  return check_value(env, value);
}

struct void_result make_interactive(emacs_env *env, emacs_value function,
                                    emacs_value spec) {
  struct void_result result;
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 28
  static_assert(SIZE_MAX >= PTRDIFF_MAX, "unsupported architecture");
  if ((size_t)env->size > offsetof(emacs_env, make_interactive)) {
    env->make_interactive(env, function, spec);
    result.base = check(env);
    return result;
  }
#endif
  result.base = unimplemented(env);
  return result;
}
