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

#ifndef PHST_EMACS_GO_WRAPPERS_H
#define PHST_EMACS_GO_WRAPPERS_H

#if !defined __STDC_VERSION__ || __STDC_VERSION__ < 201112L
#error "This library requires ISO C11 or later"
#endif

#include <emacs-module.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <time.h>

#if !defined __has_attribute
#error "This library requires __has_attribute"
#endif

#if !__has_attribute(__visibility__)
#error "This library requires __attribute__((__visibility__))"
#endif

__attribute__((__visibility__("default"))) void plugin_is_GPL_compatible(void);

__attribute__((__visibility__("default"))) int
emacs_module_init(struct emacs_runtime *rt);

// Result of non_local_exit_get.
struct result_base {
  enum emacs_funcall_exit exit;
  emacs_value error_symbol;
  emacs_value error_data;
};

// Variant of result_base where error_symbol and error_data may be missing.
struct result_base_with_optional_error_info {
  enum emacs_funcall_exit exit;
  bool has_error_info;
  emacs_value error_symbol;
  emacs_value error_data;
};

// Checks for a nonlocal exit in env.  Clears and returns it.
struct result_base check(emacs_env *env);

// Wrapper types and functions for check for all possible return types.

struct void_result {
  struct result_base base;
};

struct void_result check_void(emacs_env *env);

struct init_result {
  struct result_base_with_optional_error_info base;
};

struct init_result go_emacs_init(emacs_env *env);

bool eq(emacs_env *env, emacs_value a, emacs_value b);

struct value_result {
  struct result_base base;
  emacs_value value;
};

struct value_result check_value(emacs_env *env, emacs_value value);

struct trampoline_result {
  struct result_base_with_optional_error_info base;
  emacs_value value;
};

struct trampoline_result go_emacs_trampoline(emacs_env *env, int64_t nargs,
                                             emacs_value *args, uint64_t data);

struct value_result funcall(emacs_env *env, emacs_value function, int64_t nargs,
                            emacs_value *args);
struct value_result make_function_impl(emacs_env *env, int64_t min_arity,
                                       int64_t max_arity,
                                       const char *documentation,
                                       uint64_t data);

struct integer_result {
  struct result_base base;
  int64_t value;
};

struct integer_result check_integer(emacs_env *env, int64_t value);

struct big_integer_result {
  struct result_base base;
  int sign;            // −1, 0, or +1
  const uint8_t *data; // allocated with malloc iff successful and sign ≠ 0
  int size;            // int because of GoSlice signature
};

struct integer_result extract_integer(emacs_env *env, emacs_value value);
struct big_integer_result extract_big_integer(emacs_env *env,
                                              emacs_value value);
struct value_result make_integer(emacs_env *env, int64_t value);

// The number (and therefore sign) may not be zero.  sign must be −1, 0, or +1.
struct value_result make_big_integer(emacs_env *env, int sign,
                                     const uint8_t *data, int64_t size);

struct float_result {
  struct result_base base;
  double value;
};

struct float_result extract_float(emacs_env *env, emacs_value value);
struct value_result make_float(emacs_env *env, double value);

struct string_result {
  struct result_base base;
  const char *data;  // allocated with malloc iff successful and size > 0
  int size;          // int because of GoStringN signature
};

struct string_result copy_string_contents(emacs_env *env, emacs_value value);

struct value_result make_string_impl(emacs_env *env, const char *data,
                                     size_t size);

// symbol_name must be ASCII-only without embedded null characters.
struct value_result intern_impl(emacs_env *env, const char *symbol_name);

struct value_result vec_get(emacs_env *env, emacs_value vec, int64_t i);
struct void_result vec_set(emacs_env *env, emacs_value vec, int64_t i,
                           emacs_value val);
struct integer_result vec_size(emacs_env *env, emacs_value vec);

struct timespec_result {
  struct result_base base;
  struct timespec value;
};

struct timespec_result extract_time(emacs_env *env, emacs_value value);
struct value_result make_time(emacs_env *env, struct timespec time);

bool should_quit(emacs_env *env);
struct void_result process_input(emacs_env *env);

// Sets the nonlocal exit state of env according to result.  Call this only
// directly before returning control to Emacs.
void handle_nonlocal_exit(emacs_env *env,
                          struct result_base_with_optional_error_info result);

struct result_base out_of_memory(emacs_env *env);
struct result_base overflow_error(emacs_env *env);

#endif
