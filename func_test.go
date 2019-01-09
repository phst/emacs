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

package emacs

import "testing"

func TestDoc(t *testing.T) {
	for _, tc := range []struct {
		doc       Doc
		actualDoc Doc
		hasUsage  bool
		usage     Usage
	}{
		{"No usage.", "No usage.", false, ""},
		{"Empty usage.\n\n(fn)", "Empty usage.", true, ""},
		{"Nonempty usage.\n\n(fn foo &key bar)", "Nonempty usage.", true, "foo &key bar"},
	} {
		t.Run(string(tc.doc), func(t *testing.T) {
			gotActualDoc, gotHasUsage, gotUsage := tc.doc.SplitUsage()
			if gotActualDoc != tc.actualDoc {
				t.Errorf("actual doc: got %q, want %q", gotActualDoc, tc.actualDoc)
			}
			if gotHasUsage != tc.hasUsage {
				t.Errorf("has usage: got %v, want %v", gotHasUsage, tc.hasUsage)
			}
			if tc.hasUsage {
				if gotUsage != tc.usage {
					t.Errorf("usage: got %q, want %q", gotUsage, tc.usage)
				}
				if got := tc.doc.WithUsage(tc.usage); got != tc.doc {
					t.Errorf("doc with usage: got %q, want %q", got, tc.doc)
				}
				if got := tc.actualDoc.WithUsage(tc.usage); got != tc.doc {
					t.Errorf("actual doc with usage: got %q, want %q", got, tc.doc)
				}
			}
		})
	}
}
