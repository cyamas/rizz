package main

import (
	"testing"
)

func TestAddKeyTab(t *testing.T) {
	cur.x = 0
	cur.y = 0
	buf := newBuffer()
	buf.content.lines[0].runes = []rune("hello")
	buf.addKeyTab()
	if len(buf.content.lines[0].runes) != 13 {
		t.Fatalf("Length should  be %d. Got %d", 13, len(buf.content.lines[0].runes))
	}
}

func TestRemovePrevRune(t *testing.T) {
	runes1 := []rune{'a', 'b', 'c'}
	runes2 := []rune{'\t', ' ', ' ', ' ', ' ', ' ', ' ', '\t'}
	runes3 := []rune{'\t', ' ', ' ', ' ', ' ', ' ', ' ', '\t', '\t', ' ', ' ', ' ', ' ', ' ', ' ', '\t'}
	runes4 := []rune{'\t', ' ', ' ', ' ', ' ', ' ', ' ', '\t', '\t', ' ', ' ', ' ', ' ', ' ', ' ', '\t', 'a', 'b', 'c'}

	tests := []struct {
		runes    []rune
		x, y     int
		expected []rune
	}{
		{
			runes1,
			3, 0,
			[]rune{'a', 'b'},
		},
		{
			runes2,
			8, 0,
			[]rune{},
		},
		{
			runes3,
			16, 0,
			[]rune{'\t', ' ', ' ', ' ', ' ', ' ', ' ', '\t'},
		},
		{
			runes4,
			19, 0,
			[]rune{'\t', ' ', ' ', ' ', ' ', ' ', ' ', '\t', '\t', ' ', ' ', ' ', ' ', ' ', ' ', '\t', 'a', 'b'},
		},
	}

	for _, tt := range tests {
		cur.x = tt.x
		cur.y = tt.y
		buf := newBuffer()
		line := buf.content.lines[tt.y]
		line.runes = append(line.runes, tt.runes...)
		buf.removePrevRune()
		if len(line.runes) != len(tt.expected) {
			t.Fatalf("line should have length %d. Got %d", len(tt.expected), len(line.runes))
		}
		for i := range line.runes {
			if line.runes[i] != tt.expected[i] {
				t.Fatalf("rune should be %s. Got %s", string(tt.expected[i]), string(line.runes[i]))
			}
		}
	}
}

func TestCreateTabRunes(t *testing.T) {
	cur.x = 0
	cur.y = 0
	test1 := createTabRunes()
	test2 := createTabRunes()
	test2 = append(test2, createTabRunes()...)
	tests := []struct {
		result   []rune
		expected []rune
	}{
		{
			test1,
			[]rune{'\t', ' ', ' ', ' ', ' ', ' ', ' ', '\t'},
		},
		{
			test2,
			[]rune{'\t', ' ', ' ', ' ', ' ', ' ', ' ', '\t', '\t', ' ', ' ', ' ', ' ', ' ', ' ', '\t'},
		},
	}

	for _, tt := range tests {
		if len(tt.result) != len(tt.expected) {
			t.Fatalf("len should be %d. Got %d", len(tt.expected), len(tt.result))
		}
		for i := range tt.expected {
			if tt.result[i] != tt.expected[i] {
				t.Fatalf("expected %v\n Got %v", tt.expected, tt.result)
			}
		}
	}
}

func TestContentLength(t *testing.T) {
	buf := newBuffer()
	buf.content.lines = append(buf.content.lines, newLine())
	buf.content.lines = append(buf.content.lines, newLine())
	buf.content.length = len(buf.content.lines)
	buf.content.addLineContent("hello", 0)
	buf.content.addLineContent("oh hello", 1)
	buf.content.addLineContent("hey there friend!", 2)

	if buf.content.length != 3 {
		t.Fatalf("length should be 3. Got %d", buf.content.length)
	}
}
