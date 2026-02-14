/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

package filesystem

import (
	"context"
	"iter"
	"slices"
	"testing"
)

func TestFilePathsIteratorForDirectMode_All(t *testing.T) {

	collectSeq := func(seq iter.Seq[string]) []string {
		t.Helper() // marking helper function to get more clear logs in case of error here

		var res []string
		for v := range seq {
			res = append(res, v)
		}

		return res
	}

	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "nil as paths",
			input: nil,
			want:  []string{},
		},
		{
			name:  "no paths",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "single path",
			input: []string{"01.jpg"},
			want:  []string{"01.jpg"},
		},
		{
			name:  "multiple paths",
			input: []string{"01.jpg", "02.jpg", "dir/03.jpg", "dir/04.jpg"},
			want:  []string{"01.jpg", "02.jpg", "dir/03.jpg", "dir/04.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := NewFilePathsIteratorForDirectMode(tt.input)

			seq := it.All(context.Background())

			got := collectSeq(seq)

			slices.Sort(got)
			slices.Sort(tt.want)

			if !slices.Equal(got, tt.want) {
				t.Fatalf("got = %v, want %v", got, tt.want)
			}
		})
	}
}
