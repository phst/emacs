;;; test.el --- unit tests -*- lexical-binding: t; -*-

;; Copyright 2019 Google LLC
;;
;; Licensed under the Apache License, Version 2.0 (the "License");
;; you may not use this file except in compliance with the License.
;; You may obtain a copy of the License at
;;
;;     https://www.apache.org/licenses/LICENSE-2.0
;;
;; Unless required by applicable law or agreed to in writing, software
;; distributed under the License is distributed on an "AS IS" BASIS,
;; WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
;; See the License for the specific language governing permissions and
;; limitations under the License.

;;; Commentary:

;; Unit tests for Go Emacs module bindings.

;;; Code:

(require 'example-module)

(require 'ert)
(require 'help)

(ert-deftest go-uppercase ()
  (should (equal (go-uppercase "hello" "world") "HELLO WORLD"))
  (should (equal
           (documentation #'go-uppercase)
           "Concatenate STRINGS and return the uppercase version of the result.

\(fn strings)"))
  (should (equal (help-function-arglist #'go-uppercase :preserve-names)
                 '(strings))))

(ert-deftest go-print-now ()
  (should (string-prefix-p "It is " (go-print-now "It is %F %T %Z"))))

(ert-deftest go-var ()
  (should (equal go-var "hi"))
  (should (equal (documentation-property 'go-var 'variable-documentation)
                 "Example variable.")))

(ert-deftest go-error ()
  (should (equal (error-message-string '(example-error)) "Example error"))
  (should-error (go-error 123.45 [8 7 6]) :type 'example-error)
  (should (equal (documentation #'go-error)
                 "Signal an error of type ‘example-error’.

\(fn int float vec)"))
  (should (equal (help-function-arglist #'go-error :preserve-names)
                 '(int float vec))))

(ert-deftest mersenne-prime-p ()
  ;; 2⁴⁴²³ − 1 is a Mersenne prime, see https://oeis.org/A000043.
  (should (mersenne-prime-p 4423))
  (should (equal (documentation #'mersenne-prime-p)
                 "Return whether 2^N − 1 is probably prime.

\(fn N)"))
  (should (equal (help-function-arglist #'mersenne-prime-p :preserve-names)
                 '(n))))


;;; test.el ends here
