// Copyright 2019-2021 Google LLC
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
#include <string.h>
#include <time.h>

#include "emacs-module.h"
#include "wrappers.h"

static_assert(PTRDIFF_MIN == INT64_MIN, "unsupported architecture");
static_assert(PTRDIFF_MAX == INT64_MAX, "unsupported architecture");
static_assert(UINTPTR_MAX == UINT64_MAX, "unsupported architecture");
static_assert(PTRDIFF_MAX <= SIZE_MAX, "unsupported architecture");

// Sets the nonlocal exit state of env according to result.  Call this only
// directly before returning control to Emacs.
static void handle_nonlocal_exit(emacs_env *env,
                                 struct result_base_with_optional_error_info result);

static struct phst_emacs_result_base out_of_memory(emacs_env *env);
static struct phst_emacs_result_base overflow_error(emacs_env *env);
static struct phst_emacs_result_base unimplemented(emacs_env *env);

int emacs_module_init(struct emacs_runtime *rt) {
  if ((size_t)rt->size < sizeof *rt) {
    return 1;
  }
  emacs_env *env = rt->get_environment(rt);
  if ((size_t)env->size < sizeof(struct emacs_env_27)) {
    return 2;
  }
  struct phst_emacs_init_result result = phst_emacs_init(env);
  handle_nonlocal_exit(env, result.base);
  // We return 0 even if phst_emacs_init exited nonlocally.  See
  // https://phst.eu/emacs-modules#module-loading-and-initialization.
  return 0;
}

// Checks for a nonlocal exit in env.  Clears and returns it.
static struct phst_emacs_result_base check(emacs_env *env) {
  struct phst_emacs_result_base result;
  result.exit =
      env->non_local_exit_get(env, &result.error_symbol, &result.error_data);
  env->non_local_exit_clear(env);
  return result;
}

static struct phst_emacs_void_result check_void(emacs_env *env) {
  return (struct phst_emacs_void_result){check(env)};
}

static struct phst_emacs_value_result check_value(emacs_env *env,
                                                  emacs_value value) {
  return (struct phst_emacs_value_result){check(env), value};
}

static emacs_value trampoline(emacs_env *env, ptrdiff_t nargs,
                              emacs_value *args, void *data) {
  struct phst_emacs_trampoline_result result =
      phst_emacs_trampoline(env, nargs, args, (uintptr_t)data);
  handle_nonlocal_exit(env, result.base);
  return result.value;
}

#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 28
static void finalizer(void *data) {
  phst_emacs_function_finalizer((uintptr_t)data);
}
#endif

struct phst_emacs_value_result phst_emacs_funcall(emacs_env *env,
                                                  emacs_value function,
                                                  int64_t nargs,
                                                  emacs_value *args) {
  return check_value(env, env->funcall(env, function, nargs, args));
}

struct phst_emacs_value_result phst_emacs_make_function_impl(emacs_env *env,
                                                             int64_t min_arity,
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

static_assert(CHAR_BIT == 8, "unsupported architecture");
static_assert(UCHAR_MAX == 0xFF, "unsupported architecture");
static_assert(sizeof(uint8_t) == 1, "unsupported architecture");
static_assert(UINT8_MAX == 0xFF, "unsupported architecture");
static_assert(INTMAX_MIN == INT64_MIN, "unsupported architecture");
static_assert(INTMAX_MAX == INT64_MAX, "unsupported architecture");
static_assert(LONG_MIN >= INT64_MIN, "unsupported architecture");
static_assert(LONG_MAX <= INT64_MAX, "unsupported architecture");
static_assert(ULONG_MAX <= UINT64_MAX, "unsupported architecture");

// Rule out padding bits.
static_assert((sizeof(emacs_limb_t) == 4 && EMACS_LIMB_MAX == 0xFFFFFFFF) ||
              (sizeof(emacs_limb_t) == 8 && EMACS_LIMB_MAX == 0xFFFFFFFFFFFFFFFF),
              "unsupported architecture");
static_assert(sizeof(emacs_limb_t) < PTRDIFF_MAX, "unsupported architecture");

static struct phst_emacs_integer_result check_integer(emacs_env *env,
                                                      int64_t value) {
  return (struct phst_emacs_integer_result){check(env), value};
}

struct phst_emacs_integer_result phst_emacs_extract_integer(emacs_env *env,
                                                            emacs_value value) {
  return check_integer(env, env->extract_integer(env, value));
}

struct phst_emacs_big_integer_result phst_emacs_extract_big_integer(emacs_env *env,
                                                                    emacs_value value) {
  int sign;
  ptrdiff_t count;
  bool ok = env->extract_big_integer(env, value, &sign, &count, NULL);
  if (!ok || sign == 0) {
    return (struct phst_emacs_big_integer_result){check(env), 0, NULL, 0};
  }
  ptrdiff_t limb_size = (ptrdiff_t)sizeof(emacs_limb_t);
  assert(count > 0 && count <= PTRDIFF_MAX / limb_size);
  ptrdiff_t size = count * limb_size;
  if (size > INT_MAX) {
    return (struct phst_emacs_big_integer_result){
      overflow_error(env), 0, NULL, 0
    };
  }
  emacs_limb_t *magnitude = malloc((size_t)size);
  if (magnitude == NULL) {
    return (struct phst_emacs_big_integer_result){
      out_of_memory(env), 0, NULL, 0
    };
  }
  ptrdiff_t temp_count = count;
  ok = env->extract_big_integer(env, value, NULL, &temp_count, magnitude);
  assert(ok && count == temp_count);
  for (ptrdiff_t i = 0; i < count / 2; ++i) {
    emacs_limb_t temp = magnitude[i];
    magnitude[i] = magnitude[count - i - 1];
    magnitude[count - i - 1] = temp;
  }
  unsigned char *bytes = (unsigned char *)magnitude;
  for (ptrdiff_t i = 0; i < count; ++i) {
    emacs_limb_t limb = magnitude[i];
    unsigned char *ptr = &bytes[i * limb_size];
    for (ptrdiff_t j = 0; j < limb_size; ++j) {
      ptr[limb_size - j - 1] = (unsigned char)(limb >> (j * CHAR_BIT));
    }
  }
  return (struct phst_emacs_big_integer_result){
    {emacs_funcall_exit_return, NULL, NULL}, sign, bytes, (int)size};
}

struct phst_emacs_value_result phst_emacs_make_integer(emacs_env *env,
                                                       int64_t value) {
  return check_value(env, env->make_integer(env, value));
}

struct phst_emacs_value_result phst_emacs_make_big_integer(emacs_env *env,
                                                           int sign,
                                                           const uint8_t *data,
                                                           int64_t count) {
  assert(sign != 0);
  assert(count > 0);
  ptrdiff_t limb_size = (ptrdiff_t)sizeof(emacs_limb_t);
  if (count > INT64_MAX - limb_size || count > PTRDIFF_MAX - limb_size) {
    return (struct phst_emacs_value_result){overflow_error(env), NULL};
  }
  ptrdiff_t nlimbs = (count + limb_size - 1) / limb_size;
  assert(nlimbs <= PTRDIFF_MAX / limb_size);
  ptrdiff_t size = nlimbs * limb_size;
  emacs_limb_t *magnitude = malloc((size_t)size);
  if (magnitude == NULL) {
    return (struct phst_emacs_value_result){out_of_memory(env), NULL};
  }
  assert(size >= count && size - count < limb_size);
  for (ptrdiff_t i = 0; i < nlimbs; ++i) {
    emacs_limb_t limb = 0;
    for (ptrdiff_t j = 0; j < limb_size && i * limb_size + j < count; ++j) {
      limb += (emacs_limb_t)data[count - i * limb_size - j - 1]
              << (j * CHAR_BIT);
    }
    magnitude[i] = limb;
  }
  struct phst_emacs_value_result result =
      check_value(env, env->make_big_integer(env, sign, nlimbs, magnitude));
  free(magnitude);
  return result;
}

struct phst_emacs_float_result phst_emacs_extract_float(emacs_env *env,
                                                        emacs_value value) {
  struct phst_emacs_float_result result;
  result.value = env->extract_float(env, value);
  result.base = check(env);
  return result;
}

struct phst_emacs_value_result phst_emacs_make_float(emacs_env *env,
                                                     double value) {
  return check_value(env, env->make_float(env, value));
}

struct phst_emacs_string_result phst_emacs_copy_string_contents(emacs_env *env,
                                                                emacs_value value) {
  // See https://phst.eu/emacs-modules#copy_string_contents.
  ptrdiff_t size;
  if (!env->copy_string_contents(env, value, NULL, &size)) {
    return (struct phst_emacs_string_result){check(env), NULL, 0};
  }
  assert(size >= 0);
  if (size == 0) {
    return (struct phst_emacs_string_result){
        {emacs_funcall_exit_return, NULL, NULL}, NULL, 0};
  }
  if (size >= INT_MAX) {
    return (struct phst_emacs_string_result){overflow_error(env), NULL, 0};
  }
  static_assert(PTRDIFF_MAX <= SIZE_MAX, "unsupported architecture");
  char *buffer = malloc((size_t)size);
  if (buffer == NULL) {
    return (struct phst_emacs_string_result){out_of_memory(env), NULL, 0};
  }
  if (!env->copy_string_contents(env, value, buffer, &size)) {
    free(buffer);
    return (struct phst_emacs_string_result){check(env), NULL, 0};
  }
  return (struct phst_emacs_string_result){
      {emacs_funcall_exit_return, NULL, NULL}, buffer, (int)size - 1};
}

struct phst_emacs_value_result phst_emacs_make_string_impl(emacs_env *env,
                                                           const char *data,
                                                           size_t size) {
  if (size > PTRDIFF_MAX) {
    return (struct phst_emacs_value_result){overflow_error(env), NULL};
  }
  return check_value(env, env->make_string(env, data, (ptrdiff_t)size));
}

struct phst_emacs_value_result phst_emacs_make_unibyte_string(emacs_env *env,
                                                              const void *data,
                                                              int64_t size) {
  static_assert(CHAR_BIT == 8, "unsupported architecture");
  static_assert(SIZE_MAX >= PTRDIFF_MAX, "unsupported architecture");
  assert(size >= 0);
  if (size > PTRDIFF_MAX) {
    return (struct phst_emacs_value_result){overflow_error(env), NULL};
  }
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 28
  if ((size_t)env->size > offsetof(emacs_env, make_unibyte_string)) {
    static_assert(CHAR_MAX - CHAR_MIN + 1 == 0x100, "unsupported architecture");
    return check_value(env, env->make_unibyte_string(env, data, size));
  }
#endif
  static_assert(UCHAR_MAX == 0xFF, "unsupported architecture");
  const unsigned char *bytes = data;
  emacs_value *args = calloc(size, sizeof *args);
  if (args == NULL && size > 0) {
    return (struct phst_emacs_value_result){out_of_memory(env), NULL};
  }
  for (ptrdiff_t i = 0; i < size; ++i) {
    static_assert(INTMAX_MIN <= 0, "unsupported architecture");
    static_assert(INTMAX_MAX >= UCHAR_MAX, "unsupported architecture");
    args[i] = env->make_integer(env, bytes[i]);
  }
  emacs_value result = env->funcall(env, env->intern(env, "unibyte-string"),
                                    size, args);
  free(args);
  return check_value(env, result);
}

struct phst_emacs_value_result phst_emacs_intern_impl(emacs_env *env,
                                                      const char* data) {
  return check_value(env, env->intern(env, data));
}

struct phst_emacs_value_result phst_emacs_vec_get(emacs_env *env,
                                                  emacs_value vec,
                                                  int64_t i) {
  return check_value(env, env->vec_get(env, vec, i));
}

struct phst_emacs_void_result phst_emacs_vec_set(emacs_env *env,
                                                 emacs_value vec,
                                                 int64_t i,
                                                 emacs_value val) {
  env->vec_set(env, vec, i, val);
  return check_void(env);
}

struct phst_emacs_integer_result phst_emacs_vec_size(emacs_env *env,
                                                     emacs_value vec) {
  return check_integer(env, env->vec_size(env, vec));
}

static_assert((time_t)1.5 == 1, "unsupported architecture");
static_assert(LONG_MAX >= 1000000000, "unsupported architecture");

struct phst_emacs_timespec_result phst_emacs_extract_time(emacs_env *env,
                                                          emacs_value value) {
  struct phst_emacs_timespec_result result;
  result.value = env->extract_time(env, value);
  result.base = check(env);
  return result;
}

struct phst_emacs_value_result phst_emacs_make_time(emacs_env *env,
                                                    struct timespec time) {
  assert(time.tv_nsec >= 0 && time.tv_nsec < 1000000000);
  return check_value(env, env->make_time(env, time));
}

bool phst_emacs_should_quit(emacs_env *env) { return env->should_quit(env); }

struct phst_emacs_void_result phst_emacs_process_input(emacs_env *env) {
  env->process_input(env);
  return check_void(env);
}

struct phst_emacs_integer_result phst_emacs_open_channel(emacs_env *env,
                                                         emacs_value value) {
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 28
  static_assert(SIZE_MAX >= PTRDIFF_MAX, "unsupported architecture");
  if ((size_t)env->size > offsetof(emacs_env, open_channel)) {
    static_assert(INT64_MIN <= INT_MIN, "unsupported architecture");
    static_assert(INT64_MAX >= INT_MAX, "unsupported architecture");
    return check_integer(env, env->open_channel(env, value));
  }
#endif
  return (struct phst_emacs_integer_result){unimplemented(env), -1};
}

struct phst_emacs_void_result phst_emacs_make_interactive(emacs_env *env,
                                                          emacs_value function,
                                                          emacs_value spec) {
  struct phst_emacs_void_result result;
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

static void handle_nonlocal_exit(emacs_env *env,
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

static struct phst_emacs_result_base out_of_memory(emacs_env *env) {
  const char *message = "Out of memory";
  size_t length = strlen(message);
  assert(length < PTRDIFF_MAX);
  emacs_value temp = env->make_string(env, message, (ptrdiff_t)length);
  env->non_local_exit_signal(
      env, env->intern(env, "error"),
      env->funcall(env, env->intern(env, "list"), 1, &temp));
  return check(env);
}

static struct phst_emacs_result_base overflow_error(emacs_env *env) {
  env->non_local_exit_signal(env, env->intern(env, "overflow-error"),
                             env->intern(env, "nil"));
  return check(env);
}

static struct phst_emacs_result_base unimplemented(emacs_env *env) {
  env->non_local_exit_signal(env, env->intern(env, "go-unimplemented-error"),
                             env->intern(env, "nil"));
  return check(env);
}
