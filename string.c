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
#include <stdbool.h>
#include <stdint.h>

#include "wrappers.h"

static_assert(CHAR_BIT == 8, "unsupported architecture");
static_assert(PTRDIFF_MIN == INT64_MIN, "unsupported architecture");
static_assert(PTRDIFF_MAX == INT64_MAX, "unsupported architecture");

bool copy_string_contents(emacs_env *env, emacs_value value, uint8_t *buffer,
                          int64_t *size) {
  // It’s fine to cast uint8_t * to char *.  See
  // https://en.cppreference.com/w/c/language/object#Strict_aliasing.
  ptrdiff_t size_ptrdiff = *size;
  bool success =
      env->copy_string_contents(env, value, (char *)buffer, &size_ptrdiff);
  *size = size_ptrdiff;
  return success;
}
