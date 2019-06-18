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
#include <gmp.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "wrappers.h"

static_assert(PTRDIFF_MAX <= SIZE_MAX, "unsupported architecture");
static_assert(INTMAX_MIN == INT64_MIN, "unsupported architecture");
static_assert(INTMAX_MAX == INT64_MAX, "unsupported architecture");
static_assert(LONG_MIN >= INT64_MIN, "unsupported architecture");
static_assert(LONG_MAX <= INT64_MAX, "unsupported architecture");
static_assert(ULONG_MAX <= UINT64_MAX, "unsupported architecture");

int64_t extract_integer(emacs_env *env, emacs_value value) {
  return env->extract_integer(env, value);
}

void extract_big_integer(emacs_env *env, emacs_value value, mpz_t result) {
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
  if ((size_t)env->size > offsetof(emacs_env, extract_big_integer)) {
    struct emacs_mpz temp = {{*result}};
    env->extract_big_integer(env, value, &temp);
    *result = *temp.value;
    return;
  }
#endif
  int64_t i = env->extract_integer(env, value);
  if (i >= 0 && (uint64_t)i <= ULONG_MAX) {
    mpz_set_ui(result, i);
    return;
  }
  if (i >= LONG_MIN && i <= LONG_MAX) {
    mpz_set_si(result, i);
    return;
  }
  uint64_t u;
  // Set u = abs(i).  See https://stackoverflow.com/a/17313717.
  if (i >= 0) {
    u = i;
  } else {
    u = -(uint64_t)i;
  }
  enum { count = 1, order = 1, size = sizeof u, endian = 0, nails = 0 };
  mpz_import(result, count, order, size, endian, nails, &u);
  if (i < 0) mpz_neg(result, result);
}

emacs_value make_integer(emacs_env *env, int64_t value) {
  return env->make_integer(env, value);
}

emacs_value make_big_integer(emacs_env *env, const mpz_t value) {
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
  if ((size_t)env->size > offsetof(emacs_env, make_big_integer)) {
    struct emacs_mpz temp = {{*value}};
    return env->make_big_integer(env, &temp);
  }
#endif
  // The code below always calls make_integer if possible,
  // so this can only overflow.
  env->non_local_exit_signal(env, env->intern(env, "overflow-error"),
                             env->intern(env, "nil"));
  return NULL;
}

// This wrapper function is needed because mpz_sgn is a macro.
int emacs_mpz_sgn(const mpz_t value) { return mpz_sgn(value); }
