package main

import (
	"fmt"
	"log"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestSplitDigits(t *testing.T) {
	tests := []struct {
		num      int
		expected []int
	}{
		{9162, []int{0, 9, 1, 6, 2}},
		{847, []int{0, 0, 8, 4, 7}},
		{29, []int{0, 0, 0, 2, 9}},
		{5, []int{0, 0, 0, 0, 5}},
	}

	for _, tt := range tests {
		result := splitDigits(tt.num)
		for i := range tt.expected {
			if int(result[i]) != tt.expected[i] {
				fmt.Println("result: ", result)
				fmt.Println("expected: ", tt.expected)
				t.Fatalf("digit should be %d. Got %d", tt.expected[i], result[i])
			}
		}
	}
}

func initTestDisplay(d *Display) {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%v", err)
	}
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	d.Screen = screen
	d.Screen.SetStyle(style)
	d.BufStyle = style
	d.width, d.height = d.Screen.Size()
}

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
