package main

import (
	"fmt"
	"log"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestInsertBlankLine(t *testing.T) {
	// test insertion of blank line when cursor is on the first line
	d1 := NewDisplay()
	initTestDisplay(d1)
	d1TestLines := createTestLines(10)
	d1Expected := []string{d1TestLines[0]}
	d1Expected = append(d1Expected, "")
	d1Expected = append(d1Expected, d1TestLines[1:]...)
	d1.ActiveBuf.addTestLines(createTestLines(10))
	d1.bufWindow.update(0)

	// tests insertion of blank line at end of file when file < window size
	d2 := NewDisplay()
	initTestDisplay(d2)
	d2Expected := createTestLines(10)
	d2Expected = append(d2Expected, "")
	d2.ActiveBuf.addTestLines(createTestLines(10))
	d2.bufWindow.update(0)

	// tests insertion of blank line in middle of file
	d3 := NewDisplay()
	initTestDisplay(d3)
	d3TestLines := createTestLines(10)
	d3Expected := append([]string(nil), d3TestLines[:6]...)
	d3Expected = append(d3Expected, "")
	d3Expected = append(d3Expected, d3TestLines[6:]...)
	d3.ActiveBuf.addTestLines(createTestLines(10))
	d3.bufWindow.update(0)

	// tests insertion of blank line at end of file when file > window size
	d4 := NewDisplay()
	initTestDisplay(d4)
	d4ExpBuf := createTestLines(100)
	d4ExpBuf = append(d4ExpBuf, "")
	d4ExpWindow := append([]string(nil), d4ExpBuf[52:]...)
	d4.ActiveBuf.addTestLines(createTestLines(100))
	d4.bufWindow.update(0)

	tests := []struct {
		display        *Display
		idx            int
		expBufLines    []string
		expWindowLines []string
	}{
		{
			d1,
			0,
			d1Expected,
			d1Expected,
		},
		{
			d2,
			9,
			d2Expected,
			d2Expected,
		},
		{
			d3,
			5,
			d3Expected,
			d3Expected,
		},
		{
			d4,
			99,
			d4ExpBuf,
			d4ExpWindow,
		},
	}
	count := 0
	for _, tt := range tests {
		count++
		//fmt.Println("TEST ", count)
		cur.x = leftMarginSize
		tt.display.bufWindow.update(tt.idx)
		cur.y = tt.idx - tt.display.bufWindow.bufIdx
		tt.display.setBufPos()

		tt.display.insertBlankLine()

		bufRes := []string{}
		for _, line := range tt.display.ActiveBuf.content.lines {
			bufRes = append(bufRes, string(line.runes))
		}
		if len(bufRes) != len(tt.expBufLines) {
			fmt.Println("len(bufRes) FAIL TEST ", count)
			t.Fatalf("len should be %d. Got %d", len(tt.expBufLines), len(bufRes))
		}
		for i := range tt.expBufLines {
			if bufRes[i] != tt.expBufLines[i] {
				fmt.Println("bufRes FAIL TEST ", count)
				printLines(tt.expBufLines, bufRes)
				t.Fatalf("line should be: %s. Got %s", tt.expBufLines[i], bufRes[i])
			}
		}

		windowRes := []string{}
		for _, line := range tt.display.bufWindow.lines {
			windowRes = append(windowRes, string(line.runes))
		}
		if len(windowRes) != len(tt.expWindowLines) {
			printLines(tt.expWindowLines, windowRes)
			fmt.Println("len(windowRes) FAIL TEST ", count)
			t.Fatalf("len should be %d. Got %d", len(tt.expWindowLines), len(windowRes))
		}
		for i := range tt.expWindowLines {
			if windowRes[i] != tt.expWindowLines[i] {
				fmt.Println("windowRes FAIL TEST ", count)
				printLines(tt.expWindowLines, windowRes)
				t.Fatalf("line should be: %s. Got %s", tt.expWindowLines[i], windowRes[i])
			}
		}
	}
}

func TestDeleteLine(t *testing.T) {
	// tests deletion of first and last line
	d1 := NewDisplay()
	initTestDisplay(d1)
	d1TestLines := createTestLines(10)
	d1.ActiveBuf.addTestLines(createTestLines(10))
	d1.bufWindow.update(0)

	//tests deletion of some middle line
	d2 := NewDisplay()
	initTestDisplay(d2)
	d2testLines := createTestLines(10)
	d2Expected := append(d2testLines[:4], d2testLines[5:]...)
	d2.ActiveBuf.addTestLines(createTestLines(10))
	d2.bufWindow.update(0)

	// tests deletion of line with an index greater than the window size
	d3 := NewDisplay()
	initTestDisplay(d3)
	d3TestLines := createTestLines(100)
	d3Expected := append(d3TestLines[:82], d3TestLines[83:]...)
	d3.ActiveBuf.addTestLines(createTestLines(100))
	d3.bufWindow.update(0)

	tests := []struct {
		display        *Display
		idx            int
		expBufLines    []string
		expWindowLines []string
	}{
		{
			d1,
			0,
			d1TestLines[1:],
			d1TestLines[1:],
		},
		{
			d1,
			8,
			d1TestLines[1:9],
			d1TestLines[1:9],
		},
		{
			d2,
			4,
			d2Expected,
			d2Expected,
		},
		{
			d3,
			82,
			d3Expected,
			d3Expected[51:],
		},
	}
	count := 0
	for _, tt := range tests {
		count++
		cur.x = leftMarginSize
		tt.display.bufWindow.update(tt.idx)
		cur.y = tt.idx - tt.display.bufWindow.bufIdx
		tt.display.setBufPos()
		tt.display.deleteLine()

		bufRes := []string{}
		for _, line := range tt.display.ActiveBuf.content.lines {
			bufRes = append(bufRes, string(line.runes))
		}
		if len(bufRes) != len(tt.expBufLines) {
			fmt.Println("len(bufRes) FAIL TEST ", count)
			t.Fatalf("len should be %d. Got %d", len(tt.expBufLines), len(bufRes))
		}
		for i := range tt.expBufLines {
			if bufRes[i] != tt.expBufLines[i] {
				fmt.Println("bufRes FAIL TEST ", count)
				printLines(tt.expBufLines, bufRes)
				t.Fatalf("line should be: %s. Got %s", tt.expBufLines[i], bufRes[i])
			}
		}

		windowRes := []string{}
		for _, line := range tt.display.bufWindow.lines {
			windowRes = append(windowRes, string(line.runes))
		}
		if len(windowRes) != len(tt.expWindowLines) {
			printLines(tt.expWindowLines, windowRes)
			fmt.Println("len(windowRes) FAIL TEST ", count)
			t.Fatalf("len should be %d. Got %d", len(tt.expWindowLines), len(windowRes))
		}
		for i := range tt.expWindowLines {
			if windowRes[i] != tt.expWindowLines[i] {
				fmt.Println("windowRes FAIL TEST ", count)
				printLines(tt.expWindowLines, windowRes)
				t.Fatalf("line should be: %s. Got %s", tt.expWindowLines[i], windowRes[i])
			}
		}

	}

}

func printLines(expected, result []string) {
	fmt.Println("EXPECTED VS RESULT")
	for i := range expected {
		fmt.Printf("%s\t\t\t%s\n", expected[i], result[i])
	}
}

func createTestLines(num int) []string {
	testLines := []string{}
	for i := range num {
		line := fmt.Sprintf("%d: this is a test line", i)
		testLines = append(testLines, line)
	}
	return testLines
}

func (b *Buffer) addTestLines(lines []string) {
	for _, l := range lines {
		line := newLine()
		line.runes = []rune(l)
		if b.length() == 1 && b.content.lines[0].length() == 0 {
			b.content.lines[0] = line
			continue
		}
		b.appendLine(line)
	}
}

func testLine(text string) *Line {
	line := newLine()
	line.runes = []rune(text)
	return line
}

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
	d.width, d.height = 200, 50
	d.ActiveBuf = newBuffer()
	d.bufWindow = newBufWindow()
	d.initBufWindow()
	d.setBufWindow()
}

func TestAddKeyTab(t *testing.T) {
	d := NewDisplay()
	initTestDisplay(d)
	cur.x = leftMarginSize
	cur.y = 0
	d.setBufPos()
	buf := d.ActiveBuf
	buf.content.lines[0].runes = []rune("hello")
	buf.addKeyTab()
	if len(buf.content.lines[0].runes) != 13 {
		t.Fatalf("Length should  be %d. Got %d", 13, len(buf.content.lines[0].runes))
	}
	expected := []rune{'\t', ' ', ' ', ' ', ' ', ' ', ' ', '\t', 'h', 'e', 'l', 'l', 'o'}
	for i := range buf.content.lines[0].runes {
		if buf.content.lines[0].runes[i] != expected[i] {
			t.Fatalf("rune should be %s. Got %s", string(expected[i]), string(buf.content.lines[0].runes[i]))
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
