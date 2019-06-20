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
#include <stdlib.h>

#include "wrappers.h"

static_assert(CHAR_BIT == 8, "unsupported architecture");
static_assert(sizeof(uint8_t) == 1, "unsupported architecture");
static_assert(PTRDIFF_MAX <= SIZE_MAX, "unsupported architecture");
static_assert(INTMAX_MIN == INT64_MIN, "unsupported architecture");
static_assert(INTMAX_MAX == INT64_MAX, "unsupported architecture");
static_assert(LONG_MIN >= INT64_MIN, "unsupported architecture");
static_assert(LONG_MAX <= INT64_MAX, "unsupported architecture");
static_assert(ULONG_MAX <= UINT64_MAX, "unsupported architecture");

struct integer_result extract_integer(emacs_env *env, emacs_value value) {
  return check_integer(env, env->extract_integer(env, value));
}

struct big_integer_result extract_big_integer(emacs_env *env,
                                              emacs_value value) {
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
  if ((size_t)env->size > offsetof(emacs_env, extract_big_integer)) {
    struct emacs_mpz temp;
    mpz_init(temp.value);
    env->extract_big_integer(env, value, &temp);
    struct big_integer_result result = {check(env), 0, NULL, 0};
    if (result.base.exit != emacs_funcall_exit_return) {
      mpz_clear(temp.value);
      return result;
    }
    result.sign = mpz_sgn(temp.value);
    if (result.sign == 0) {
      mpz_clear(temp.value);
      return result;
    }
    // See
    // https://gmplib.org/manual/Integer-Import-and-Export.html#index-Export.
    enum {
      order = 1,
      size = 1,
      endian = 0,
      nails = 0,
      numb = 8 * size - nails
    };
    size_t count = (mpz_sizeinbase(temp.value, 2) + numb - 1) / numb;
    if (count > INT_MAX) {
      mpz_clear(temp.value);
      result.base = overflow_error(env);
      return result;
    }
    uint8_t *data = malloc(size);
    if (data == NULL) {
      mpz_clear(temp.value);
      result.base = out_of_memory(env);
      return result;
    }
    size_t written;
    mpz_export(data, &written, order, size, endian, nails, temp.value);
    assert(written == count);
    mpz_clear(temp.value);
    result.size = (int)count;
    result.data = data;
    return result;
  }
#endif
  struct integer_result i = extract_integer(env, value);
  if (i.base.exit != emacs_funcall_exit_return || i.value == 0) {
    return (struct big_integer_result){i.base, 0, NULL, 0};
  }
  uint64_t u;
  // Set u = abs(i.value).  See https://stackoverflow.com/a/17313717.
  if (i.value > 0) {
    u = (uint64_t)i.value;
  } else {
    u = -(uint64_t)i.value;
  }
  uint8_t *bytes = malloc(sizeof u);
  if (bytes == NULL) {
    return (struct big_integer_result){out_of_memory(env), 0, NULL, 0};
  }
  int sign = i.value > 0 ? 1 : -1;
  static_assert(sizeof u == 8, "unsupported architecture");
  bytes[0] = (u >> 56) & 0xFFu;
  bytes[1] = (u >> 48) & 0xFFu;
  bytes[2] = (u >> 40) & 0xFFu;
  bytes[3] = (u >> 32) & 0xFFu;
  bytes[4] = (u >> 24) & 0xFFu;
  bytes[5] = (u >> 16) & 0xFFu;
  bytes[6] = (u >> 8) & 0xFFu;
  bytes[7] = u & 0xFFu;
  return (struct big_integer_result){
      {emacs_funcall_exit_return, NULL, NULL}, sign, bytes, sizeof u};
}

struct value_result make_integer(emacs_env *env, int64_t value) {
  return check_value(env, env->make_integer(env, value));
}

struct value_result make_big_integer(emacs_env *env, int sign,
                                     const uint8_t *data, int64_t count) {
  assert(sign != 0);
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
  if ((size_t)env->size > offsetof(emacs_env, make_big_integer)) {
    struct emacs_mpz temp;
    mpz_init(temp.value);
    enum { order = 1, size = 1, endian = 0, nails = 0 };
    mpz_import(temp.value, count, order, size, endian, nails, data);
    if (sign == -1) {
      mpz_neg(temp.value, temp.value);
    }
    struct value_result result =
        check_value(env, env->make_big_integer(env, &temp));
    mpz_clear(temp.value);
    return result;
  }
#endif
  // The code below always calls make_integer if possible, so this can only
  // overflow.
  return (struct value_result){overflow_error(env), NULL};
}

// This wrapper function is needed because mpz_sgn is a macro.
int emacs_mpz_sgn(const mpz_t value) { return mpz_sgn(value); }
