// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trace_test

import (
	"internal/trace"
	"internal/trace/testtrace"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestOldtrace(t *testing.T) {
	traces, err := filepath.Glob("./internal/oldtrace/testdata/*_good")
	if err != nil {
		t.Fatalf("failed to glob for tests: %s", err)
	}
	var testedUserRegions bool
	for _, p := range traces {
		p := p
		testName, err := filepath.Rel("./internal/oldtrace/testdata", p)
		if err != nil {
			t.Fatalf("failed to relativize testdata path: %s", err)
		}
		t.Run(testName, func(t *testing.T) {
			f, err := os.Open(p)
			if err != nil {
				t.Fatalf("failed to open test %q: %s", p, err)
			}
			defer f.Close()

			tr, err := trace.NewReader(f)
			if err != nil {
				t.Fatalf("failed to create reader: %s", err)
			}

			v := testtrace.NewValidator()
			v.Go121 = true
			for {
				ev, err := tr.ReadEvent()
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Fatalf("couldn't read converted event: %s", err)
				}
				if err := v.Event(ev); err != nil {
					t.Fatalf("converted event did not validate; event: \n%s\nerror: %s", ev, err)
				}

				if testName == "user_task_region_1_21_good" {
					testedUserRegions = true
					validRegions := map[string]struct{}{
						"post-existing region": struct{}{},
						"region0":              struct{}{},
						"region1":              struct{}{},
					}
					// Check that we correctly convert user regions. These
					// strings were generated by
					// runtime/trace.TestUserTaskRegion, which is the basis for
					// the user_task_region_* test cases. We only check for the
					// Go 1.21 traces because earlier traces used different
					// strings.
					switch ev.Kind() {
					case trace.EventRegionBegin, trace.EventRegionEnd:
						if _, ok := validRegions[ev.Region().Type]; !ok {
							t.Fatalf("converted event has unexpected region type:\n%s", ev)
						}
					case trace.EventTaskBegin, trace.EventTaskEnd:
						if ev.Task().Type != "task0" {
							t.Fatalf("converted event has unexpected task type name:\n%s", ev)
						}
					case trace.EventLog:
						l := ev.Log()
						if l.Task != 1 || l.Category != "key0" || l.Message != "0123456789abcdef" {
							t.Fatalf("converted event has unexpected user log:\n%s", ev)
						}
					}
				}
			}
		})
	}
	if !testedUserRegions {
		t.Fatal("didn't see expected test case user_task_region_1_21_good")
	}
}