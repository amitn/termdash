// Copyright 2019 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package termbox

import (
	"fmt"
	"testing"

	"github.com/mum4k/termdash/cell"
	tbx "github.com/nsf/termbox-go"
)

func TestCellColor(t *testing.T) {
	tests := []struct {
		color cell.Color
		want  tbx.Attribute
	}{
		{cell.ColorDefault, tbx.ColorDefault},
		{cell.ColorBlack, tbx.ColorBlack},
		{cell.ColorRed, tbx.ColorRed},
		{cell.ColorGreen, tbx.ColorGreen},
		{cell.ColorYellow, tbx.ColorYellow},
		{cell.ColorBlue, tbx.ColorBlue},
		{cell.ColorMagenta, tbx.ColorMagenta},
		{cell.ColorCyan, tbx.ColorCyan},
		{cell.ColorWhite, tbx.ColorWhite},
		{cell.Color(42), tbx.Attribute(42)},
	}

	for _, tc := range tests {
		t.Run(tc.color.String(), func(t *testing.T) {
			got := cellColor(tc.color)
			if got != tc.want {
				t.Errorf("cellColor(%v) => got %v, want %v", tc.color, got, tc.want)
			}

		})
	}
}

func TestCellFontModifier(t *testing.T) {
	tests := []struct {
		opt     cell.Options
		want    tbx.Attribute
		wantErr bool
	}{
		{cell.Options{Bold: true}, tbx.AttrBold, false},
		{cell.Options{Underline: true}, tbx.AttrUnderline, false},
		{cell.Options{Italic: true}, 0, true},
		{cell.Options{Strikethrough: true}, 0, true},
		{cell.Options{Inverse: true}, tbx.AttrReverse, false},
		{cell.Options{Blink: true}, 0, true},
		{cell.Options{Dim: true}, 0, true},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc.opt), func(t *testing.T) {
			got, err := cellOptsToFg(&tc.opt)
			if (err != nil) != tc.wantErr {
				t.Errorf("cellOptsToFg(%v) => unexpected error: %v, wantErr: %v", tc.opt, err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if got != tc.want {
				t.Errorf("cellOptsToFg(%v) => got %v, want %v", tc.opt, got, tc.want)
			}
		})
	}
}
