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
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>

#include "emacs-module.h"
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

// Rule out padding bits.
static_assert((sizeof(emacs_limb_t) == 4 && EMACS_LIMB_MAX == 0xFFFFFFFF) ||
              (sizeof(emacs_limb_t) == 8 && EMACS_LIMB_MAX == 0xFFFFFFFFFFFFFFFF),
              "unsupported architecture");
static_assert(sizeof(emacs_limb_t) < PTRDIFF_MAX, "unsupported architecture");

struct integer_result extract_integer(emacs_env *env, emacs_value value) {
  return check_integer(env, env->extract_integer(env, value));
}

struct big_integer_result extract_big_integer(emacs_env *env,
                                              emacs_value value) {
  int sign;
  ptrdiff_t count;
  bool ok = env->extract_big_integer(env, value, &sign, &count, NULL);
  if (!ok || sign == 0) {
    return (struct big_integer_result){check(env), 0, NULL, 0};
  }
  ptrdiff_t limb_size = (ptrdiff_t)sizeof(emacs_limb_t);
  assert(count > 0 && count <= PTRDIFF_MAX / limb_size);
  ptrdiff_t size = count * limb_size;
  if (size > INT_MAX) {
    return (struct big_integer_result){overflow_error(env), 0, NULL, 0};
  }
  emacs_limb_t *magnitude = malloc((size_t)size);
  if (magnitude == NULL) {
    return (struct big_integer_result){out_of_memory(env), 0, NULL, 0};
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
  return (struct big_integer_result){
    {emacs_funcall_exit_return, NULL, NULL}, sign, bytes, (int)size};
}

struct value_result make_integer(emacs_env *env, int64_t value) {
  return check_value(env, env->make_integer(env, value));
}

struct value_result make_big_integer(emacs_env *env, int sign,
                                     const uint8_t *data, int64_t count) {
  assert(sign != 0);
  assert(count > 0);
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
