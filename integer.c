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
#include <limits.h>
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>

#include "wrappers.h"

static_assert(CHAR_BIT == 8, "unsupported architecture");
static_assert(UCHAR_MAX == 0xFF, "unsupported architecture");
static_assert(sizeof(uint8_t) == 1, "unsupported architecture");
static_assert(UINT8_MAX == 0xFF, "unsupported architecture");
static_assert(PTRDIFF_MAX <= SIZE_MAX, "unsupported architecture");
static_assert(INTMAX_MIN == INT64_MIN, "unsupported architecture");
static_assert(INTMAX_MAX == INT64_MAX, "unsupported architecture");
static_assert(LONG_MIN >= INT64_MIN, "unsupported architecture");
static_assert(LONG_MAX <= INT64_MAX, "unsupported architecture");
static_assert(ULONG_MAX <= UINT64_MAX, "unsupported architecture");

#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
// Rule out padding bits.
static_assert((sizeof(emacs_limb_t) == 4 && EMACS_LIMB_MAX == 0xFFFFFFFF) ||
              (sizeof(emacs_limb_t) == 8 && EMACS_LIMB_MAX == 0xFFFFFFFFFFFFFFFF),
              "unsupported architecture");
static_assert(sizeof(emacs_limb_t) < PTRDIFF_MAX, "unsupported architecture");
#endif

struct integer_result extract_integer(emacs_env *env, emacs_value value) {
  return check_integer(env, env->extract_integer(env, value));
}

struct big_integer_result extract_big_integer(emacs_env *env,
                                              emacs_value value) {
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
  if ((size_t)env->size > offsetof(emacs_env, extract_big_integer)) {
    int sign;
    ptrdiff_t count;
    bool ok = env->extract_big_integer(env, value, &sign, &count, NULL);
    if (!ok || sign == 0) {
      return (struct big_integer_result){check(env), 0, NULL, 0};
    }
    ptrdiff_t limb_size = (ptrdiff_t)sizeof(emacs_limb_t);
    assert(count > 0 && count <= PTRDIFF_MAX / limb_size);
    ptrdiff_t size = count * limb_size;
    if (size > INT_MAX || (size_t)size > SIZE_MAX / 2) {
      return (struct big_integer_result){overflow_error(env), 0, NULL, 0};
    }
    uint8_t *bytes = malloc(2 * (size_t)size);
    if (bytes == NULL) {
      return (struct big_integer_result){out_of_memory(env), 0, NULL, 0};
    }
    emacs_limb_t *magnitude = (emacs_limb_t *)(bytes + size);
    ok = env->extract_big_integer(env, value, NULL, &count, magnitude);
    assert(ok);
    for (ptrdiff_t i = 0; i < count; ++i) {
      emacs_limb_t limb = magnitude[i];
      for (ptrdiff_t j = 0; j < limb_size; ++j) {
        bytes[size - i * limb_size - j - 1] = (uint8_t)(limb >> (j * CHAR_BIT));
      }
    }
    return (struct big_integer_result){
        {emacs_funcall_exit_return, NULL, NULL}, sign, bytes, (int)size};
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
  bytes[0] = (uint8_t)(u >> 56);
  bytes[1] = (uint8_t)(u >> 48);
  bytes[2] = (uint8_t)(u >> 40);
  bytes[3] = (uint8_t)(u >> 32);
  bytes[4] = (uint8_t)(u >> 24);
  bytes[5] = (uint8_t)(u >> 16);
  bytes[6] = (uint8_t)(u >> 8);
  bytes[7] = (uint8_t)u;
  return (struct big_integer_result){
      {emacs_funcall_exit_return, NULL, NULL}, sign, bytes, sizeof u};
}

struct value_result make_integer(emacs_env *env, int64_t value) {
  return check_value(env, env->make_integer(env, value));
}

struct value_result make_big_integer(emacs_env *env, int sign,
                                     const uint8_t *data, int64_t count) {
  assert(sign != 0);
  assert(count > 0);
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
  if ((size_t)env->size > offsetof(emacs_env, make_big_integer)) {
    ptrdiff_t limb_size = (ptrdiff_t)sizeof(emacs_limb_t);
    if (count > INT64_MAX - limb_size || count > PTRDIFF_MAX - limb_size) {
      return (struct value_result){overflow_error(env), NULL};
    }
    ptrdiff_t nlimbs = (count + limb_size - 1) / limb_size;
    assert(nlimbs <= PTRDIFF_MAX / limb_size);
    ptrdiff_t size = nlimbs * limb_size;
    emacs_limb_t *magnitude = malloc((size_t)size);
    if (magnitude == NULL) {
      return (struct value_result){out_of_memory(env), NULL};
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
    struct value_result result =
        check_value(env, env->make_big_integer(env, sign, nlimbs, magnitude));
    free(magnitude);
    return result;
  }
#endif
  // The code below always calls make_integer if possible, so this can only
  // overflow.
  return (struct value_result){overflow_error(env), NULL};
}
