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

#include <emacs-module.h>
#include <gmp.h>
#include <stdbool.h>
#include <stdint.h>
#include <time.h>

bool eq(emacs_env *env, emacs_value a, emacs_value b);

emacs_value trampoline(emacs_env *env, ptrdiff_t nargs, emacs_value *args,
                       void *data);
emacs_value funcall(emacs_env *env, emacs_value function, int64_t nargs,
                    emacs_value *args);

int64_t extract_integer(emacs_env *env, emacs_value value);
void extract_big_integer(emacs_env *env, emacs_value value, mpz_t result);
emacs_value make_integer(emacs_env *env, int64_t value);
emacs_value make_big_integer(emacs_env *env, const mpz_t value);
int emacs_mpz_sgn(const mpz_t value);

bool copy_string_contents(emacs_env *env, emacs_value value, uint8_t *buffer,
                          int64_t *size);

emacs_value vec_get(emacs_env *env, emacs_value vec, int64_t i);
void vec_set(emacs_env *env, emacs_value vec, int64_t i, emacs_value val);
int64_t vec_size(emacs_env *env, emacs_value vec);

struct timespec extract_time(emacs_env *env, emacs_value value);
emacs_value make_time(emacs_env *env, struct timespec time);

bool should_quit(emacs_env *env);
int process_input(emacs_env *env);

#endif
