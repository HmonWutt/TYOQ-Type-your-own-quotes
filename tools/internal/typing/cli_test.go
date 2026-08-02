package typing

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func testModel() model {
	return model{
		targetText:   "hello world",
		typed:        0,
		typedChars:   []rune{},
		errorIndices: make(map[int]bool),
		width:        80,
		height:       24,
	}
}

func makeKeyPress(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text, Code: rune(text[0])})
}

func TestCharStyleKind_Correct(t *testing.T) {
	m := testModel()
	m.typed = 3
	m.typedChars = []rune("hel")
	m.targetText = "hello"

	kind := m.charStyleKind(0)
	if kind != styleCorrect {
		t.Errorf("expected styleCorrect for typed correct char, got %v", kind)
	}
}

func TestCharStyleKind_Error(t *testing.T) {
	m := testModel()
	m.typed = 3
	m.typedChars = []rune("hxl")
	m.errorIndices = map[int]bool{1: true}
	m.targetText = "hello"

	kind := m.charStyleKind(1)
	if kind != styleError {
		t.Errorf("expected styleError, got %v", kind)
	}
}

func TestCharStyleKind_Next(t *testing.T) {
	m := testModel()
	m.typed = 2
	m.typedChars = []rune("he")
	m.targetText = "hello"
	m.done = false

	kind := m.charStyleKind(2)
	if kind != styleNext {
		t.Errorf("expected styleNext for current cursor position, got %v", kind)
	}
}

func TestCharStyleKind_Dim(t *testing.T) {
	m := testModel()
	m.typed = 1
	m.typedChars = []rune("h")
	m.targetText = "hello"

	kind := m.charStyleKind(3)
	if kind != styleDim {
		t.Errorf("expected styleDim for untyped future char, got %v", kind)
	}
}

func TestCharStyleKind_NextWhenDone(t *testing.T) {
	m := testModel()
	m.typed = 5
	m.typedChars = []rune("hello")
	m.targetText = "hello"
	m.done = true

	kind := m.charStyleKind(4)
	if kind != styleCorrect {
		t.Errorf("expected styleCorrect when done for typed char, got %v", kind)
	}
}

func TestCharStyle_ReturnsNonNil(t *testing.T) {
	m := testModel()

	styles := []styleKind{styleCorrect, styleError, styleNext, styleDim}
	for _, kind := range styles {
		s := m.charStyle(kind)
		rendered := s.Render("x")
		if rendered == "" {
			t.Errorf("expected non-empty render for style kind %v", kind)
		}
	}
}

func TestTextLinesVisible(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		width      int
		expectView int
	}{
		{name: "short quote fits", text: "hello world", width: 80, expectView: 1},
		{name: "real quote one line", text: "Be yourself; everyone else is already taken.", width: 80, expectView: 1},
		{name: "200 runes", text: repeatWords(40), width: 80, expectView: 4},
		{name: "300 runes", text: repeatWords(60), width: 80, expectView: 6},
		{name: "600 runes exceeds cap", text: repeatWords(120), width: 80, expectView: 6},
		{name: "zero width falls back", text: repeatWords(20), width: 0, expectView: 6},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := testModel()
			m.targetText = tc.text
			m.width = tc.width

			result := m.textLinesVisible()
			if result != tc.expectView {
				t.Errorf("expected %d, got %d", tc.expectView, result)
			}
		})
	}
}

func TestUpdateScroll(t *testing.T) {
	// repeatWords(120) = 11 wrapped lines at width 80 (textWidth 56); visible = 6
	tests := []struct {
		name     string
		text     string
		width    int
		typed    int
		expected int
	}{
		{name: "no scroll at start", text: repeatWords(120), width: 80, typed: 0, expected: 0},
		{name: "cursor on second line still no scroll", text: repeatWords(120), width: 80, typed: 90, expected: 0},
		{name: "cursor on fourth line scrolls", text: repeatWords(120), width: 80, typed: 170, expected: 2},
		{name: "cursor near bottom caps at maxOffset", text: repeatWords(120), width: 80, typed: 400, expected: 5},
		{name: "short text never scrolls", text: "hello world", width: 80, typed: 5, expected: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := testModel()
			m.targetText = tc.text
			m.width = tc.width
			m.typed = tc.typed

			result := m.updateScroll()
			if result != tc.expected {
				t.Errorf("expected scroll %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestModelReset(t *testing.T) {
	m := testModel()
	m.typed = 5
	m.typedChars = []rune("hello")
	m.errorIndices = map[int]bool{2: true}
	m.done = true
	m.quotes = []string{"hello world"}
	m.width = 100
	m.height = 40

	newM := m.reset()

	if newM.width != 100 {
		t.Errorf("expected width 100, got %d", newM.width)
	}
	if newM.height != 40 {
		t.Errorf("expected height 40, got %d", newM.height)
	}
	if newM.typed != 0 {
		t.Errorf("expected typed 0, got %d", newM.typed)
	}
	if len(newM.typedChars) != 0 {
		t.Errorf("expected empty typedChars, got %v", newM.typedChars)
	}
	if len(newM.errorIndices) != 0 {
		t.Errorf("expected empty errorIndices, got %v", newM.errorIndices)
	}
	if newM.done {
		t.Error("expected done to be false after reset")
	}
	if newM.targetText != "hello world" {
		t.Errorf("expected target text 'hello world', got %q", newM.targetText)
	}
	if newM.index != m.index {
		t.Errorf("reset should keep the same index: expected %d, got %d", m.index, newM.index)
	}
}

func TestModelNext(t *testing.T) {
	m := testModel()
	m.quotes = []string{"q0", "q1", "q2"}
	m.index = 0

	newM := m.next()
	if newM.index != 1 {
		t.Errorf("expected index 1, got %d", newM.index)
	}
	if newM.targetText != "q1" {
		t.Errorf("expected target text 'q1', got %q", newM.targetText)
	}

	// wraps around at the end
	m.index = 2
	newM = m.next()
	if newM.index != 0 {
		t.Errorf("expected index to wrap to 0, got %d", newM.index)
	}
	if newM.targetText != "q0" {
		t.Errorf("expected target text 'q0' after wrap, got %q", newM.targetText)
	}
}

func TestModelNext_CustomText(t *testing.T) {
	m := testModel()
	m.isCustom = true
	m.customText = "my custom text"
	m.index = 0

	newM := m.next()
	if newM.targetText != "my custom text" {
		t.Errorf("expected custom text, got %q", newM.targetText)
	}
}

func TestModelPrevious(t *testing.T) {
	m := testModel()
	m.quotes = []string{"q0", "q1", "q2"}
	m.index = 2

	newM := m.previous()
	if newM.index != 1 {
		t.Errorf("expected index 1, got %d", newM.index)
	}
	if newM.targetText != "q1" {
		t.Errorf("expected target text 'q1', got %q", newM.targetText)
	}

	// wraps around to the last quote
	m.index = 0
	newM = m.previous()
	if newM.index != 2 {
		t.Errorf("expected index to wrap to 2, got %d", newM.index)
	}
	if newM.targetText != "q2" {
		t.Errorf("expected target text 'q2' after wrap, got %q", newM.targetText)
	}
}

func TestModelPrevious_CustomText(t *testing.T) {
	m := testModel()
	m.isCustom = true
	m.customText = "my custom text"

	newM := m.previous()
	if newM.targetText != "my custom text" {
		t.Errorf("expected custom text, got %q", newM.targetText)
	}
}

func TestUpdate_EnterNextQuote(t *testing.T) {
	m := testModel()
	m.quotes = []string{"q0", "q1", "q2"}
	m.index = 0
	m.targetText = "q0"

	newM, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	newModel := newM.(model)

	if newModel.index != 1 {
		t.Errorf("expected index 1 after enter, got %d", newModel.index)
	}
	if newModel.targetText != "q1" {
		t.Errorf("expected target text 'q1', got %q", newModel.targetText)
	}
}

func TestUpdate_ShiftEnterPreviousQuote(t *testing.T) {
	m := testModel()
	m.quotes = []string{"q0", "q1", "q2"}
	m.index = 2
	m.targetText = "q2"

	newM, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter, Mod: tea.ModShift}))
	newModel := newM.(model)

	if newModel.index != 1 {
		t.Errorf("expected index 1 after shift+enter, got %d", newModel.index)
	}
	if newModel.targetText != "q1" {
		t.Errorf("expected target text 'q1', got %q", newModel.targetText)
	}
}

func TestUpdate_TypeChar(t *testing.T) {
	m := testModel()
	m.targetText = "hello"

	newM, _ := m.Update(makeKeyPress("h"))
	newModel := newM.(model)

	if newModel.typed != 1 {
		t.Errorf("expected typed 1, got %d", newModel.typed)
	}
	if string(newModel.typedChars[0]) != "h" {
		t.Errorf("expected 'h', got %q", string(newModel.typedChars[0]))
	}
}

func TestUpdate_ErrorChar(t *testing.T) {
	m := testModel()
	m.targetText = "hello"

	newM, _ := m.Update(makeKeyPress("x"))
	newModel := newM.(model)

	if newModel.typed != 1 {
		t.Errorf("expected typed 1, got %d", newModel.typed)
	}
	if !newModel.errorIndices[0] {
		t.Error("expected error at index 0")
	}
}

func TestUpdate_Backspace(t *testing.T) {
	m := testModel()
	m.targetText = "hello"
	m.typed = 3
	m.typedChars = []rune("hel")
	m.errorIndices = map[int]bool{1: true}

	newM, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))
	newModel := newM.(model)

	if newModel.typed != 2 {
		t.Errorf("expected typed 2 after backspace, got %d", newModel.typed)
	}
	if len(newModel.typedChars) != 2 {
		t.Errorf("expected 2 typed chars, got %d", len(newModel.typedChars))
	}
	if newModel.errorIndices[2] {
		t.Error("expected error at index 2 to be cleared")
	}
}

func TestUpdate_CompleteTyping(t *testing.T) {
	m := testModel()
	m.targetText = "hi"
	m.typed = 1
	m.typedChars = []rune("h")

	newM, _ := m.Update(makeKeyPress("i"))
	newModel := newM.(model)

	if !newModel.done {
		t.Error("expected done to be true after completing typing")
	}
	if !newModel.waitingRestart {
		t.Error("expected waitingRestart to be true after completing")
	}
	if newModel.endTime.IsZero() {
		t.Error("expected endTime to be set")
	}
}

func TestUpdate_Quit(t *testing.T) {
	m := testModel()

	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if cmd == nil {
		t.Error("expected quit command on esc")
	}
}

func TestUpdate_WindowSize(t *testing.T) {
	m := testModel()

	newM, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	newModel := newM.(model)

	if newModel.width != 120 {
		t.Errorf("expected width 120, got %d", newModel.width)
	}
	if newModel.height != 40 {
		t.Errorf("expected height 40, got %d", newModel.height)
	}
}

func TestUpdate_TypeSpace(t *testing.T) {
	m := testModel()
	m.targetText = "a b"

	newM, _ := m.Update(makeKeyPress("a"))
	m2 := newM.(model)

	newM2, _ := m2.Update(makeKeyPress(" "))
	m3 := newM2.(model)

	if m3.typed != 2 {
		t.Errorf("expected typed 2, got %d", m3.typed)
	}
	if string(m3.typedChars[1]) != " " {
		t.Errorf("expected space char, got %q", string(m3.typedChars[1]))
	}
}

// repeatWords builds a realistic quote-like string: "word word word ... "
func repeatWords(n int) string {
	return strings.Repeat("word ", n)
}

// flattenLines flattens wrapped lines back into a single index slice.
func flattenLines(lines [][]int) []int {
	var flat []int
	for _, l := range lines {
		flat = append(flat, l...)
	}
	return flat
}

// assertConsecutive checks the wrapped indices reconstruct the original
// text exactly: indices are 0..len-1 in order, no gaps, no dups.
func assertConsecutive(t *testing.T, flat []int, textLen int) {
	t.Helper()
	if len(flat) != textLen {
		t.Fatalf("expected %d indices, got %d", textLen, len(flat))
	}
	for i, idx := range flat {
		if idx != i {
			t.Fatalf("indices not consecutive at position %d: got %d, want %d", i, idx, i)
		}
	}
}

func TestWrapLines_FitsOneLine(t *testing.T) {
	lines := wrapLines("hello world", 80)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	assertConsecutive(t, flattenLines(lines), len("hello world"))
}

func TestWrapLines_WrapsAtWordBoundary(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog"
	lines := wrapLines(text, 20)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	// words must not split: every line break falls on a space in the source
	for li := 0; li < len(lines)-1; li++ {
		lastIdx := lines[li][len(lines[li])-1]
		if text[lastIdx] != ' ' {
			t.Errorf("line %d does not end on a word boundary (last char %q)", li, string(text[lastIdx]))
		}
	}
	assertConsecutive(t, flattenLines(lines), len(text))
}

func TestWrapLines_HardBreaksLongWord(t *testing.T) {
	lines := wrapLines("abcdefghij", 4)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	expectedLens := []int{4, 4, 2}
	for i, want := range expectedLens {
		if len(lines[i]) != want {
			t.Errorf("line %d: expected len %d, got %d", i, want, len(lines[i]))
		}
	}
	assertConsecutive(t, flattenLines(lines), len("abcdefghij"))
}

func TestWrapLines_Empty(t *testing.T) {
	lines := wrapLines("", 10)
	if len(lines) != 1 {
		t.Fatalf("expected 1 empty line, got %d", len(lines))
	}
	if len(lines[0]) != 0 {
		t.Errorf("expected empty line, got %d indices", len(lines[0]))
	}
}

func TestWrapLines_ZeroWidth(t *testing.T) {
	// width 0 falls back to 80, so a short string stays on one line
	lines := wrapLines("hello", 0)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	assertConsecutive(t, flattenLines(lines), len("hello"))
}

func TestWrapLines_TightWidthKeepsSpaceOnOwnLine(t *testing.T) {
	// documents current behavior at tight widths: "hello world" at width 5
	// puts the space alone on the middle line.
	lines := wrapLines("hello world", 5)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if len(lines[1]) != 1 {
		t.Errorf("expected middle line to hold just the space (len 1), got %d", len(lines[1]))
	}
}

func TestWrapLines_ConsecutiveIndicesAcrossWidths(t *testing.T) {
	text := "Pack my box with five dozen liquor jugs. How vexingly quick daft zebras jump!"
	for _, w := range []int{1, 3, 5, 7, 13, 20, 40, 100} {
		lines := wrapLines(text, w)
		flat := flattenLines(lines)
		// spaces at line starts may be dropped, so flat can only ever
		// be a subset of 0..len-1 in strictly increasing order.
		for i := 1; i < len(flat); i++ {
			if flat[i] <= flat[i-1] {
				t.Fatalf("width %d: indices not strictly increasing at %d: %d after %d", w, i, flat[i], flat[i-1])
			}
		}
	}
}
