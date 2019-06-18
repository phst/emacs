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
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#include "wrappers.h"

bool should_quit(emacs_env *env) { return env->should_quit(env); }

int process_input(emacs_env *env) {
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
  static_assert(SIZE_MAX >= PTRDIFF_MAX, "unsupported architecture");
  if ((size_t)env->size > offsetof(emacs_env, process_input)) {
    return env->process_input(env);
  }
#endif
  return env->should_quit(env);
}
