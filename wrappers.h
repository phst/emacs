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

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <time.h>

#include "emacs-module.h"

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
struct phst_emacs_result_base {
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

// Wrapper types for all possible return types.

struct phst_emacs_void_result {
  struct phst_emacs_result_base base;
};

struct phst_emacs_init_result {
  struct result_base_with_optional_error_info base;
};

struct phst_emacs_init_result phst_emacs_init(emacs_env *env);

struct phst_emacs_value_result {
  struct phst_emacs_result_base base;
  emacs_value value;
};

struct phst_emacs_trampoline_result {
  struct result_base_with_optional_error_info base;
  emacs_value value;
};

struct phst_emacs_trampoline_result phst_emacs_trampoline(emacs_env *env,
                                                          int64_t nargs,
                                                          emacs_value *args,
                                                          uint64_t data);
void phst_emacs_function_finalizer(uint64_t data);

struct phst_emacs_value_result phst_emacs_funcall(emacs_env *env,
                                                  emacs_value function,
                                                  int64_t nargs,
                                                  emacs_value *args);
struct phst_emacs_value_result phst_emacs_make_function_impl(emacs_env *env,
                                                             int64_t min_arity,
                                                             int64_t max_arity,
                                                             const char *documentation,
                                                             uint64_t data);

struct phst_emacs_integer_result {
  struct phst_emacs_result_base base;
  int64_t value;
};

struct phst_emacs_big_integer_result {
  struct phst_emacs_result_base base;
  int sign;            // −1, 0, or +1
  const uint8_t *data; // allocated with malloc iff successful and sign ≠ 0
  int size;            // int because of GoSlice signature
};

struct phst_emacs_integer_result phst_emacs_extract_integer(emacs_env *env,
                                                            emacs_value value);
struct phst_emacs_big_integer_result phst_emacs_extract_big_integer(emacs_env *env,
                                                                    emacs_value value);
struct phst_emacs_value_result phst_emacs_make_integer(emacs_env *env,
                                                       int64_t value);

// The number (and therefore sign) may not be zero.  sign must be −1 or +1.
struct phst_emacs_value_result phst_emacs_make_big_integer(emacs_env *env,
                                                           int sign,
                                                           const uint8_t *data,
                                                           int64_t size);

struct phst_emacs_float_result {
  struct phst_emacs_result_base base;
  double value;
};

struct phst_emacs_float_result phst_emacs_extract_float(emacs_env *env,
                                                        emacs_value value);
struct phst_emacs_value_result phst_emacs_make_float(emacs_env *env,
                                                     double value);

struct phst_emacs_string_result {
  struct phst_emacs_result_base base;
  const char *data;  // allocated with malloc iff successful and size > 0
  int size;          // int because of GoStringN signature
};

struct phst_emacs_string_result phst_emacs_copy_string_contents(emacs_env *env,
                                                                emacs_value value);

struct phst_emacs_value_result phst_emacs_make_string_impl(emacs_env *env,
                                                           const char *data,
                                                           size_t size);
struct phst_emacs_value_result phst_emacs_make_unibyte_string(emacs_env *env,
                                                              const void *data,
                                                              int64_t size);

// symbol_name must be ASCII-only without embedded null characters.
struct phst_emacs_value_result phst_emacs_intern_impl(emacs_env *env,
                                                      const char *symbol_name);

struct phst_emacs_value_result phst_emacs_vec_get(emacs_env *env,
                                                  emacs_value vec,
                                                  int64_t i);
struct phst_emacs_void_result phst_emacs_vec_set(emacs_env *env,
                                                 emacs_value vec,
                                                 int64_t i,
                                                 emacs_value val);
struct phst_emacs_integer_result phst_emacs_vec_size(emacs_env *env,
                                                     emacs_value vec);

struct phst_emacs_timespec_result {
  struct phst_emacs_result_base base;
  struct timespec value;
};

struct phst_emacs_timespec_result phst_emacs_extract_time(emacs_env *env,
                                                          emacs_value value);
struct phst_emacs_value_result phst_emacs_make_time(emacs_env *env,
                                                    struct timespec time);

bool phst_emacs_should_quit(emacs_env *env);
struct phst_emacs_void_result phst_emacs_process_input(emacs_env *env);

struct phst_emacs_integer_result phst_emacs_open_channel(emacs_env *env,
                                                         emacs_value value);

struct phst_emacs_void_result phst_emacs_make_interactive(emacs_env *env,
                                                          emacs_value function,
                                                          emacs_value spec);

#endif
