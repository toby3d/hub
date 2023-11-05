package urlutil_test

import (
	"testing"

	"source.toby3d.me/toby3d/hub/internal/urlutil"
)

func TestShiftPath(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		input  string
		expect [2]string
	}{
		"empty":  {input: "", expect: [2]string{"", "/"}},
		"root":   {input: "/", expect: [2]string{"", "/"}},
		"page":   {input: "/foo", expect: [2]string{"foo", "/"}},
		"folder": {input: "/foo/bar", expect: [2]string{"foo", "/bar"}},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			head, tail := urlutil.ShiftPath(tc.input)

			if head != tc.expect[0] {
				t.Errorf("want '%s', got '%s'", tc.expect[0], head)
			}

			if tail != tc.expect[1] {
				t.Errorf("want '%s', got '%s'", tc.expect[1], tail)
			}
		})
	}
}
