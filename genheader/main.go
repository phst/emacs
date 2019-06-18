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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	flag.Parse()
	snippets, newestVersion := readSnippets()
	t := readTemplate()
	t = strings.ReplaceAll(t, "@emacs_major_version@", strconv.Itoa(newestVersion))
	for v, s := range snippets {
		t = strings.ReplaceAll(t, fmt.Sprintf("@module_env_snippet_%d@", v), s)
	}
	if err := ioutil.WriteFile(*output, []byte(t), 0400); err != nil {
		log.Fatalf("error writing output: %v", err)
	}
}

func readSnippets() (snippets map[int]string, newestVersion int) {
	snippets = make(map[int]string)
	newestVersion = 0
	r := regexp.MustCompile(`^module-env-(\d+)\.h$`)
	for _, a := range flag.Args() {
		m := r.FindStringSubmatch(filepath.Base(a))
		if m == nil {
			log.Fatalf("unexpected argument %s", a)
		}
		v, err := strconv.Atoi(m[1])
		if err != nil {
			log.Fatalf("unexpected argument %s: %v", a, err)
		}
		if v <= 0 {
			log.Fatalf("unexpected argument %s", a)
		}
		if v > newestVersion {
			newestVersion = v
		}
		b, err := ioutil.ReadFile(a)
		if err != nil {
			log.Fatalf("error reading snippet: %v", err)
		}
		snippets[v] = string(b)
	}
	if len(snippets) == 0 {
		log.Fatal("no snippets")
	}
	return
}

func readTemplate() string {
	b, err := ioutil.ReadFile(*template)
	if err != nil {
		log.Fatalf("error reading template: %v", err)
	}
	return string(b)
}

var (
	template = flag.String("template", "", "filename of the header template")
	output   = flag.String("output", "", "filename of header file to generate")
)
