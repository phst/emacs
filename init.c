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

#include "wrappers.h"

#include <assert.h>
#include <stddef.h>
#include <stdint.h>

#include <emacs-module.h>

static_assert(PTRDIFF_MAX <= SIZE_MAX, "unsupported architecture");

int emacs_module_init(struct emacs_runtime *rt) {
  if ((size_t)rt->size < sizeof *rt) {
    return 1;
  }
  emacs_env *env = rt->get_environment(rt);
  if ((size_t)env->size < sizeof(struct emacs_env_26)) {
    return 2;
  }
  struct init_result result = go_emacs_init(env);
  handle_nonlocal_exit(env, result.base);
  // We return 0 even if go_emacs_init exited nonlocally.  See
  // https://phst.eu/emacs-modules#module-loading-and-initialization.
  return 0;
}
