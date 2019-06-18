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
#include <inttypes.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <time.h>

#include "wrappers.h"

static_assert(PTRDIFF_MAX <= SIZE_MAX, "unsupported architecture");
static_assert((time_t)1.5 == 1, "unsupported architecture");
static_assert(LONG_MAX >= 1000000000, "unsupported architecture");

struct timespec extract_time(emacs_env *env, emacs_value value) {
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
  if ((size_t)env->size > offsetof(emacs_env, extract_time)) {
    return env->extract_time(env, value);
  }
#endif
  emacs_value list =
      env->funcall(env, env->intern(env, "seconds-to-time"), 1, &value);
  emacs_value car = env->intern(env, "car");
  emacs_value cdr = env->intern(env, "cdr");
  intmax_t parts[4] = {0};
  for (int i = 0; i < 4; ++i) {
    parts[i] = env->extract_integer(env, env->funcall(env, car, 1, &list));
    list = env->funcall(env, cdr, 1, &list);
    if (!env->is_not_nil(env, list)) break;
  }
  struct timespec result;
  assert(parts[1] >= 0 && parts[1] <= 0x10000);
  if (__builtin_mul_overflow(parts[0], 0x10000, &result.tv_sec) ||
      __builtin_add_overflow(result.tv_sec, parts[1], &result.tv_sec)) {
    env->non_local_exit_signal(env, env->intern(env, "overflow-error"),
                               env->intern(env, "nil"));
    return result;
  }
  assert(parts[2] >= 0 && parts[2] < 1000000);
  assert(parts[3] >= 0 && parts[3] < 1000000);
  result.tv_nsec = (long)parts[2] * 1000 + (long)parts[3] / 1000;
  return result;
}

emacs_value make_time(emacs_env *env, struct timespec time) {
  assert(time.tv_nsec >= 0 && time.tv_nsec < 1000000000);
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
  if ((size_t)env->size > offsetof(emacs_env, make_time)) {
    return env->make_time(env, time);
  }
#endif
  imaxdiv_t seconds = imaxdiv(time.tv_sec, 0x10000);
  if (seconds.rem < 0) {
    --seconds.quot;
    seconds.rem += 0x10000;
  }
  assert(seconds.rem >= 0 && seconds.rem <= 0x10000);
  ldiv_t nanos = ldiv(time.tv_nsec, 1000);
  assert(nanos.quot >= 0 && nanos.quot < 1000000);
  assert(nanos.rem >= 0 && nanos.rem < 1000000);
  emacs_value args[] = {env->make_integer(env, seconds.quot),
                        env->make_integer(env, seconds.rem),
                        env->make_integer(env, nanos.quot),
                        env->make_integer(env, nanos.rem * 1000)};
  return env->funcall(env, env->intern(env, "list"), 4, args);
}
