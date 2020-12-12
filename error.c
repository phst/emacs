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
#include <string.h>

#include "emacs-module.h"

struct result_base check(emacs_env *env) {
  struct result_base result;
  result.exit =
      env->non_local_exit_get(env, &result.error_symbol, &result.error_data);
  env->non_local_exit_clear(env);
  return result;
}

struct void_result check_void(emacs_env *env) {
  return (struct void_result){check(env)};
}

struct value_result check_value(emacs_env *env, emacs_value value) {
  return (struct value_result){check(env), value};
}

struct integer_result check_integer(emacs_env *env, int64_t value) {
  return (struct integer_result){check(env), value};
}

void handle_nonlocal_exit(emacs_env *env,
                          struct result_base_with_optional_error_info result) {
  if (result.exit == emacs_funcall_exit_return) {
    return;
  }
  if (!result.has_error_info) {
    env->non_local_exit_signal(env, env->intern(env, "go-error"),
                               env->intern(env, "nil"));
    return;
  }
  switch (result.exit) {
  case emacs_funcall_exit_return:
    assert(false);  // handled above
    break;
  case emacs_funcall_exit_signal:
    env->non_local_exit_signal(env, result.error_symbol, result.error_data);
    break;
  case emacs_funcall_exit_throw:
    env->non_local_exit_throw(env, result.error_symbol, result.error_data);
    break;
  }
}

struct result_base out_of_memory(emacs_env *env) {
  const char *message = "Out of memory";
  size_t length = strlen(message);
  assert(length < PTRDIFF_MAX);
  emacs_value temp = env->make_string(env, message, (ptrdiff_t)length);
  env->non_local_exit_signal(
      env, env->intern(env, "error"),
      env->funcall(env, env->intern(env, "list"), 1, &temp));
  return check(env);
}

struct result_base overflow_error(emacs_env *env) {
  env->non_local_exit_signal(env, env->intern(env, "overflow-error"),
                             env->intern(env, "nil"));
  return check(env);
}

struct result_base unimplemented(emacs_env *env) {
  env->non_local_exit_signal(env, env->intern(env, "go-unimplemented-error"),
                             env->intern(env, "nil"));
  return check(env);
}
