// Copyright 2018 Google Inc.
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

package container

import (
	"fmt"
	"image"
	"strings"
	"testing"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/event"
	"github.com/mum4k/termdash/private/event/testevent"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// pointCase is a test case for the pointCont function.
type pointCase struct {
	desc      string
	point     image.Point
	wantNil   bool
	wantColor cell.Color // expected container identified by its border color
}

func TestPointCont(t *testing.T) {
	tests := []struct {
		desc      string
		termSize  image.Point
		container func(ft *faketerm.Terminal) (*Container, error)
		cases     []pointCase
	}{
		{
			desc:     "single container, no border",
			termSize: image.Point{3, 3},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					BorderColor(cell.ColorBlue),
				)
			},
			cases: []pointCase{
				{
					desc:      "inside the container",
					point:     image.Point{1, 1},
					wantColor: cell.ColorBlue,
				},
				{
					desc:      "top left corner",
					point:     image.Point{0, 0},
					wantColor: cell.ColorBlue,
				},
				{
					desc:      "top right corner",
					point:     image.Point{2, 0},
					wantColor: cell.ColorBlue,
				},
				{
					desc:      "bottom left corner",
					point:     image.Point{0, 2},
					wantColor: cell.ColorBlue,
				},
				{
					desc:      "bottom right corner",
					point:     image.Point{2, 2},
					wantColor: cell.ColorBlue,
				},
				{
					desc:    "outside of the container, too large",
					point:   image.Point{3, 3},
					wantNil: true,
				},
				{
					desc:    "outside of the container, too small",
					point:   image.Point{-1, -1},
					wantNil: true,
				},
			},
		},
		{
			desc:     "single container, border",
			termSize: image.Point{3, 3},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(linestyle.Light),
					BorderColor(cell.ColorBlue),
				)
			},
			cases: []pointCase{
				{
					desc:      "inside the container",
					point:     image.Point{1, 1},
					wantColor: cell.ColorBlue,
				},
				{
					desc:      "on the border",
					point:     image.Point{0, 1},
					wantColor: cell.ColorBlue,
				},
			},
		},
		{
			desc:     "split containers, parent has no border",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					BorderColor(cell.ColorBlack),
					SplitVertical(
						Left(
							SplitHorizontal(
								Top(
									BorderColor(cell.ColorGreen),
								),
								Bottom(
									BorderColor(cell.ColorWhite),
								),
							),
						),
						Right(
							BorderColor(cell.ColorRed),
						),
					),
				)
			},
			cases: []pointCase{
				{
					desc:      "right sub container, inside corner",
					point:     image.Point{5, 5},
					wantColor: cell.ColorRed,
				},
				{
					desc:      "right sub container, outside corner",
					point:     image.Point{9, 9},
					wantColor: cell.ColorRed,
				},
				{
					desc:      "top left",
					point:     image.Point{0, 0},
					wantColor: cell.ColorGreen,
				},
				{
					desc:      "bottom left",
					point:     image.Point{0, 9},
					wantColor: cell.ColorWhite,
				},
			},
		},
		{
			desc:     "split containers, parent has border",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(linestyle.Light),
					BorderColor(cell.ColorBlack),
					SplitVertical(
						Left(
							SplitHorizontal(
								Top(
									BorderColor(cell.ColorGreen),
								),
								Bottom(
									BorderColor(cell.ColorWhite),
								),
							),
						),
						Right(
							BorderColor(cell.ColorRed),
						),
					),
				)
			},
			cases: []pointCase{
				{
					desc:      "right sub container, inside corner",
					point:     image.Point{5, 5},
					wantColor: cell.ColorRed,
				},
				{
					desc:      "top right corner focuses parent",
					point:     image.Point{9, 9},
					wantColor: cell.ColorBlack,
				},
				{
					desc:      "right sub container, outside corner",
					point:     image.Point{8, 8},
					wantColor: cell.ColorRed,
				},
				{
					desc:      "top left focuses parent",
					point:     image.Point{0, 0},
					wantColor: cell.ColorBlack,
				},
				{
					desc:      "top left sub container",
					point:     image.Point{1, 1},
					wantColor: cell.ColorGreen,
				},
				{
					desc:      "bottom left focuses parent",
					point:     image.Point{0, 9},
					wantColor: cell.ColorBlack,
				},
				{
					desc:      "bottom left sub container",
					point:     image.Point{1, 8},
					wantColor: cell.ColorWhite,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ft, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			cont, err := tc.container(ft)
			if err != nil {
				t.Fatalf("tc.container => unexpected error: %v", err)
			}
			// Initial draw to determine sizes of containers.
			if err := cont.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}
			for _, pc := range tc.cases {
				gotCont := pointCont(cont, pc.point)
				if (gotCont == nil) != pc.wantNil {
					t.Errorf("%s, pointCont%v => got %v, wantNil: %v", pc.desc, pc.point, gotCont, pc.wantNil)
				}
				if gotCont == nil {
					continue
				}

				gotColor := gotCont.opts.inherited.borderColor
				if gotColor != pc.wantColor {
					t.Errorf("%s, pointCont%v => got container with border color %v, want %v", pc.desc, pc.point, gotColor, pc.wantColor)
				}
			}
		})
	}
}

// contLocIntro prints out an introduction explaining the used container
// locations on test failures.
func contLocIntro() string {
	var s strings.Builder
	s.WriteString("Container locations refer to containers in the following tree, i.e. contLocA is the root container:\n")
	s.WriteString(`
    A
   / \
  B   C
 / \
D   E
`)
	return s.String()
}

// contLoc is used in tests to indicate the desired location of a container.
type contLoc int

// String implements fmt.Stringer()
func (cl contLoc) String() string {
	if n, ok := contLocNames[cl]; ok {
		return n
	}
	return "contLocUnknown"
}

// contLocNames maps contLoc values to human readable names.
var contLocNames = map[contLoc]string{
	contLocA: "contLocA",
	contLocB: "contLocB",
	contLocC: "contLocC",
}

const (
	contLocUnknown contLoc = iota
	contLocA
	contLocB
	contLocC
)

func TestFocusTrackerMouse(t *testing.T) {
	t.Log(contLocIntro())

	ft, err := faketerm.New(image.Point{10, 10})
	if err != nil {
		t.Fatalf("faketerm.New => unexpected error: %v", err)
	}

	var (
		insideB = image.Point{1, 1}
		insideC = image.Point{6, 6}
	)

	tests := []struct {
		desc string
		// Can be either the mouse event or a time.Duration to pause for.
		events        []*terminalapi.Mouse
		wantFocused   contLoc
		wantProcessed int
	}{
		{
			desc:        "initially the root is focused",
			wantFocused: contLocA,
		},
		{
			desc: "click and release moves focus to the left",
			events: []*terminalapi.Mouse{
				{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				{Position: image.Point{1, 1}, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocB,
			wantProcessed: 2,
		},
		{
			desc: "click and release moves focus to the right",
			events: []*terminalapi.Mouse{
				{Position: image.Point{5, 5}, Button: mouse.ButtonLeft},
				{Position: image.Point{6, 6}, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocC,
			wantProcessed: 2,
		},
		{
			desc: "click in the same container is a no-op",
			events: []*terminalapi.Mouse{
				{Position: insideC, Button: mouse.ButtonLeft},
				{Position: insideC, Button: mouse.ButtonRelease},
				{Position: insideC, Button: mouse.ButtonLeft},
				{Position: insideC, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocC,
			wantProcessed: 4,
		},
		{
			desc: "click in the same container and release never happens",
			events: []*terminalapi.Mouse{
				{Position: insideC, Button: mouse.ButtonLeft},
				{Position: insideB, Button: mouse.ButtonLeft},
				{Position: insideB, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocB,
			wantProcessed: 3,
		},
		{
			desc: "click in the same container, release elsewhere",
			events: []*terminalapi.Mouse{
				{Position: insideC, Button: mouse.ButtonLeft},
				{Position: insideB, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocA,
			wantProcessed: 2,
		},
		{
			desc: "other buttons are ignored",
			events: []*terminalapi.Mouse{
				{Position: insideB, Button: mouse.ButtonMiddle},
				{Position: insideB, Button: mouse.ButtonRelease},
				{Position: insideB, Button: mouse.ButtonRight},
				{Position: insideB, Button: mouse.ButtonRelease},
				{Position: insideB, Button: mouse.ButtonWheelUp},
				{Position: insideB, Button: mouse.ButtonWheelDown},
			},
			wantFocused:   contLocA,
			wantProcessed: 6,
		},
		{
			desc: "moving mouse with pressed button and then releasing moves focus",
			events: []*terminalapi.Mouse{
				{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				{Position: image.Point{1, 1}, Button: mouse.ButtonLeft},
				{Position: image.Point{2, 2}, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocB,
			wantProcessed: 3,
		},
		{
			desc: "click ignored if followed by another click of the same button elsewhere",
			events: []*terminalapi.Mouse{
				{Position: insideC, Button: mouse.ButtonLeft},
				{Position: insideB, Button: mouse.ButtonLeft},
				{Position: insideC, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocA,
			wantProcessed: 3,
		},
		{
			desc: "click ignored if followed by another click of a different button",
			events: []*terminalapi.Mouse{
				{Position: insideC, Button: mouse.ButtonLeft},
				{Position: insideC, Button: mouse.ButtonMiddle},
				{Position: insideC, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocA,
			wantProcessed: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			root, err := New(
				ft,
				SplitVertical(
					Left(),
					Right(),
				),
			)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			eds := event.NewDistributionSystem()
			root.Subscribe(eds)
			// Initial draw to determine sizes of containers.
			if err := root.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}
			for _, ev := range tc.events {
				eds.Event(ev)
			}
			if err := testevent.WaitFor(5*time.Second, func() error {
				if got, want := eds.Processed(), tc.wantProcessed; got != want {
					return fmt.Errorf("the event distribution system processed %d events, want %d", got, want)
				}
				return nil
			}); err != nil {
				t.Fatalf("testevent.WaitFor => %v", err)
			}

			var wantFocused *Container
			switch wf := tc.wantFocused; wf {
			case contLocA:
				wantFocused = root
			case contLocB:
				wantFocused = root.first
			case contLocC:
				wantFocused = root.second
			default:
				t.Fatalf("unsupported wantFocused value => %v", wf)
			}

			if !root.focusTracker.isActive(wantFocused) {
				t.Errorf("isActive(%v) => false, want true, status: contLocA(%v):%v, contLocB(%v):%v, contLocC(%v):%v",
					tc.wantFocused,
					contLocA, root.focusTracker.isActive(root),
					contLocB, root.focusTracker.isActive(root.first),
					contLocC, root.focusTracker.isActive(root.second),
				)
			}
		})
	}
}

// contDir represents a direction in which we want to change container focus.
type contDir int

// String implements fmt.Stringer()
func (cd contDir) String() string {
	if n, ok := contDirNames[cd]; ok {
		return n
	}
	return "contDirUnknown"
}

// contDirNames maps contDir values to human readable names.
var contDirNames = map[contDir]string{
	contDirNext:     "contDirNext",
	contDirPrevious: "contDirPrevious",
}

const (
	contDirUnknown contDir = iota
	contDirNext
	contDirPrevious
)

func TestFocusTrackerNextAndPrevious(t *testing.T) {
	t.Log(contLocIntro())

	ft, err := faketerm.New(image.Point{10, 10})
	if err != nil {
		t.Fatalf("faketerm.New => unexpected error: %v", err)
	}

	const (
		keyNext     keyboard.Key = keyboard.KeyTab
		keyPrevious keyboard.Key = '~'
	)

	tests := []struct {
		desc          string
		container     func(ft *faketerm.Terminal) (*Container, error)
		events        []*terminalapi.Keyboard
		wantFocused   contLoc
		wantProcessed int
	}{
		{
			desc: "initially the root is focused",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeyFocusNext(keyNext),
				)
			},
			wantFocused: contLocA,
		},
		{
			desc: "keyNext does nothing when only root exists",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					KeyFocusNext(keyNext),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyNext},
			},
			wantFocused:   contLocA,
			wantProcessed: 1,
		},
		{
			desc: "keyNext focuses the first container",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeyFocusNext(keyNext),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyNext},
			},
			wantFocused:   contLocB,
			wantProcessed: 1,
		},
		{
			desc: "two keyNext presses focuses the second container",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeyFocusNext(keyNext),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyNext},
				{Key: keyNext},
			},
			wantFocused:   contLocC,
			wantProcessed: 2,
		},
		{
			desc: "three keyNext presses focuses the first container again",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeyFocusNext(keyNext),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyNext},
				{Key: keyNext},
				{Key: keyNext},
			},
			wantFocused:   contLocB,
			wantProcessed: 3,
		},
		{
			desc: "four keyNext presses focuses the second container again",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeyFocusNext(keyNext),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyNext},
				{Key: keyNext},
				{Key: keyNext},
				{Key: keyNext},
			},
			wantFocused:   contLocC,
			wantProcessed: 4,
		},
		{
			desc: "five keyNext presses focuses the first container again",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeyFocusNext(keyNext),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyNext},
				{Key: keyNext},
				{Key: keyNext},
				{Key: keyNext},
				{Key: keyNext},
			},
			wantFocused:   contLocB,
			wantProcessed: 5,
		},
		{
			desc: "keyPrevious does nothing when only root exists",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					KeyFocusPrevious(keyPrevious),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyPrevious},
			},
			wantFocused:   contLocA,
			wantProcessed: 1,
		},
		{
			desc: "keyPrevious focuses the last container",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeyFocusPrevious(keyPrevious),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyPrevious},
			},
			wantFocused:   contLocC,
			wantProcessed: 1,
		},
		{
			desc: "two keyPrevious presses focuses the first container",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeyFocusPrevious(keyPrevious),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyPrevious},
				{Key: keyPrevious},
			},
			wantFocused:   contLocB,
			wantProcessed: 2,
		},
		{
			desc: "three keyPrevious presses focuses the second container again",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeyFocusPrevious(keyPrevious),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyPrevious},
				{Key: keyPrevious},
				{Key: keyPrevious},
			},
			wantFocused:   contLocC,
			wantProcessed: 3,
		},
		{
			desc: "four keyPrevious presses focuses the first container again",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeyFocusPrevious(keyPrevious),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyPrevious},
				{Key: keyPrevious},
				{Key: keyPrevious},
				{Key: keyPrevious},
			},
			wantFocused:   contLocB,
			wantProcessed: 4,
		},
		{
			desc: "five keyPrevious presses focuses the second container again",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeyFocusPrevious(keyPrevious),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyPrevious},
				{Key: keyPrevious},
				{Key: keyPrevious},
				{Key: keyPrevious},
				{Key: keyPrevious},
			},
			wantFocused:   contLocC,
			wantProcessed: 5,
		},
		{
			desc: "first container requests to be skipped on key based focus changes, using next",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							KeyFocusSkip(),
						),
						Right(),
					),
					KeyFocusNext(keyNext),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyNext},
			},
			wantFocused:   contLocC,
			wantProcessed: 1,
		},
		{
			desc: "last container requests to be skipped on key based focus changes, using next",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(
							KeyFocusSkip(),
						),
					),
					KeyFocusNext(keyNext),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyNext},
				{Key: keyNext},
			},
			wantFocused:   contLocB,
			wantProcessed: 2,
		},
		{
			desc: "all containers request to be skipped on key based focus changes, using next",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							KeyFocusSkip(),
						),
						Right(
							KeyFocusSkip(),
						),
					),
					KeyFocusNext(keyNext),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyNext},
			},
			wantFocused:   contLocA,
			wantProcessed: 1,
		},
		{
			desc: "first container requests to be skipped on key based focus changes, using previous",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							KeyFocusSkip(),
						),
						Right(),
					),
					KeyFocusPrevious(keyPrevious),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyPrevious},
				{Key: keyPrevious},
			},
			wantFocused:   contLocC,
			wantProcessed: 2,
		},
		{
			desc: "last container requests to be skipped on key based focus changes, using previous",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(
							KeyFocusSkip(),
						),
					),
					KeyFocusPrevious(keyPrevious),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyPrevious},
			},
			wantFocused:   contLocB,
			wantProcessed: 1,
		},
		{
			desc: "all containers request to be skipped on key based focus changes, using previous",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							KeyFocusSkip(),
						),
						Right(
							KeyFocusSkip(),
						),
					),
					KeyFocusPrevious(keyPrevious),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: keyPrevious},
			},
			wantFocused:   contLocA,
			wantProcessed: 1,
		},
		{
			desc: "all containers are in focus group zero by default, pressing KeysFocusGroupNext once focuses the first container",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeysFocusGroupNext(0, []keyboard.Key{'n'}),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: 'n'},
			},
			wantFocused:   contLocB,
			wantProcessed: 1,
		},
		{
			desc: "all containers are in focus group zero by default, pressing KeysFocusGroupNext twice focuses the second container",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeysFocusGroupNext(0, []keyboard.Key{'n'}),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: 'n'},
				{Key: 'n'},
			},
			wantFocused:   contLocC,
			wantProcessed: 2,
		},
		{
			desc: "all containers are in focus group zero by default, pressing KeysFocusGroupNext three times focuses the first container again",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeysFocusGroupNext(0, []keyboard.Key{'n'}),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: 'n'},
				{Key: 'n'},
				{Key: 'n'},
			},
			wantFocused:   contLocB,
			wantProcessed: 3,
		},
		{
			desc: "all containers are in focus group zero by default, pressing KeysFocusGroupPrevious once focuses the second container",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeysFocusGroupPrevious(0, []keyboard.Key{'p'}),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: 'p'},
			},
			wantFocused:   contLocC,
			wantProcessed: 1,
		},
		{
			desc: "all containers are in focus group zero by default, pressing KeysFocusGroupPrevious twice focuses the first container",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeysFocusGroupPrevious(0, []keyboard.Key{'p'}),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: 'p'},
				{Key: 'p'},
			},
			wantFocused:   contLocB,
			wantProcessed: 2,
		},
		{
			desc: "all containers are in focus group zero by default, pressing KeysFocusGroupPrevious three times focuses the second container again",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(),
					),
					KeysFocusGroupPrevious(0, []keyboard.Key{'p'}),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: 'p'},
				{Key: 'p'},
				{Key: 'p'},
			},
			wantFocused:   contLocC,
			wantProcessed: 3,
		},
		{
			desc: "configuring container with KeyFocusSkip has no effect on a closed focus group",
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							KeyFocusSkip(),
						),
						Right(
							KeyFocusSkip(),
						),
					),
					KeysFocusGroupNext(0, []keyboard.Key{'n'}),
				)
			},
			events: []*terminalapi.Keyboard{
				{Key: 'n'},
			},
			wantFocused:   contLocB,
			wantProcessed: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			root, err := tc.container(ft)
			if err != nil {
				t.Fatalf("tc.container => unexpected error: %v", err)
			}

			eds := event.NewDistributionSystem()
			root.Subscribe(eds)
			for _, ev := range tc.events {
				eds.Event(ev)
			}
			if err := testevent.WaitFor(5*time.Second, func() error {
				if got, want := eds.Processed(), tc.wantProcessed; got != want {
					return fmt.Errorf("the event distribution system processed %d events, want %d", got, want)
				}
				return nil
			}); err != nil {
				t.Fatalf("testevent.WaitFor => %v", err)
			}

			var wantFocused *Container
			switch wf := tc.wantFocused; wf {
			case contLocA:
				wantFocused = root
			case contLocB:
				wantFocused = root.first
			case contLocC:
				wantFocused = root.second
			default:
				t.Fatalf("unsupported wantFocused value => %v", wf)
			}

			if !root.focusTracker.isActive(wantFocused) {
				t.Errorf("isActive(%v) => false, want true, status: contLocA(%v):%v, contLocB(%v):%v, contLocC(%v):%v",
					tc.wantFocused,
					contLocA, root.focusTracker.isActive(root),
					contLocB, root.focusTracker.isActive(root.first),
					contLocC, root.focusTracker.isActive(root.second),
				)
			}
		})
	}
}
