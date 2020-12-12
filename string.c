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
#include <limits.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>

#include "emacs-module.h"
#include "wrappers.h"

static_assert(CHAR_BIT == 8, "unsupported architecture");

struct string_result copy_string_contents(emacs_env *env, emacs_value value) {
  // See https://phst.eu/emacs-modules#copy_string_contents.
  ptrdiff_t size;
  if (!env->copy_string_contents(env, value, NULL, &size)) {
    return (struct string_result){check(env), NULL, 0};
  }
  assert(size >= 0);
  if (size == 0) {
    return (struct string_result){
        {emacs_funcall_exit_return, NULL, NULL}, NULL, 0};
  }
  if (size >= INT_MAX) {
    return (struct string_result){overflow_error(env), NULL, 0};
  }
  static_assert(PTRDIFF_MAX <= SIZE_MAX, "unsupported architecture");
  char *buffer = malloc((size_t)size);
  if (buffer == NULL) {
    return (struct string_result){out_of_memory(env), NULL, 0};
  }
  if (!env->copy_string_contents(env, value, buffer, &size)) {
    free(buffer);
    return (struct string_result){check(env), NULL, 0};
  }
  return (struct string_result){
      {emacs_funcall_exit_return, NULL, NULL}, buffer, (int)size - 1};
}

struct value_result make_string_impl(emacs_env *env, const char *data,
                                     size_t size) {
  if (size > PTRDIFF_MAX) {
    return (struct value_result){overflow_error(env), NULL};
  }
  return check_value(env, env->make_string(env, data, (ptrdiff_t)size));
}
