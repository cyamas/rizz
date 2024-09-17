package display

import (
	"fmt"
	"log"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestAutoIndentFromAutoClose(t *testing.T) {
	d1 := NewDisplay()
	initTestDisplay(d1)
	line1 := "{}"
	d1.ActiveBuf.content.lines[0].runes = []rune(line1)
	expBuf1 := []string{
		"{",
		"/t      /t",
		"}",
	}

	tests := []struct {
		display *Display
		x, y    int
		expBuf  []string
	}{
		{
			d1,
			LeftMarginSize + 1, 0,
			expBuf1,
		},
	}

	for _, tt := range tests {
		Cur.X = tt.x
		Cur.Y = tt.y
		tt.display.setBufPos()
		tt.display.handleKeyEnter()

		if d1.ActiveBuf.length() != len(expBuf1) {
			t.Fatalf("length should be %d. Got %d", len(expBuf1), d1.ActiveBuf.length())
		}
	}

}

func TestAutoIndent(t *testing.T) {
	d1 := NewDisplay()
	initTestDisplay(d1)
	d1.ActiveBuf.addTestLines(createTestLines(1))
	d1.ActiveBuf.currLine().addKeyTab()
	x1 := LeftMarginSize + d1.ActiveBuf.currLine().length()
	exp1 := createTestLines(1)
	exp1[0] = "/t      /t" + exp1[0]
	exp1 = append(exp1, "/t      /t")

	d2 := NewDisplay()
	initTestDisplay(d2)
	d2.ActiveBuf.addTestLines(createTestLines(1))
	d2.ActiveBuf.content.lines[0].runes = append(d2.ActiveBuf.content.lines[0].runes, '{')
	x2 := LeftMarginSize + d2.ActiveBuf.content.lines[0].length()
	exp2 := createTestLines(1)
	exp2[0] += "{"
	exp2 = append(exp2, "/t      /t")

	tests := []struct {
		display   *Display
		x, y      int
		expBuf    []string
		expWindow []string
		expCurX   int
	}{
		{
			d1,
			x1, 0,
			exp1,
			exp1,
			16,
		},
		{
			d2,
			x2, 0,
			exp2,
			exp2,
			16,
		},
	}
	count := 0
	for _, tt := range tests {
		count++
		Cur.X = tt.x
		Cur.Y = tt.y
		tt.display.setBufPos()
		tt.display.handleKeyEnter()

		if tt.display.ActiveBuf.length() != len(tt.expBuf) {
			t.Fatalf("ActiveBuf length should be %d. Got %d", tt.display.ActiveBuf.length(), len(tt.expBuf))
		}

		for i := range tt.expBuf {
			res := convertRunesToText(tt.display.ActiveBuf.getLine(i).runes)
			if res != tt.expBuf[i] {
				t.Fatalf("Line should be %s. Got %s", tt.expBuf[i], res)
			}
		}

		if Cur.X != tt.expCurX {
			t.Fatalf("Cur.X should be %d. Got %d", tt.expCurX, Cur.X)
		}

	}
}

func TestHandleKeyEnter(t *testing.T) {
	// tests pressing Key Enter at line 0 at bufCur.X = 0
	d1 := NewDisplay()
	initTestDisplay(d1)
	d1.ActiveBuf.addTestLines(createTestLines(10))
	exp1 := []string{""}
	exp1 = append(exp1, createTestLines(10)...)

	//tests pressing Key Enter at last line at bufCur.X = 0
	d2 := NewDisplay()
	initTestDisplay(d2)
	d2.ActiveBuf.addTestLines(createTestLines(10))
	exp2 := createTestLines(10)
	exp2 = append(exp2, exp2[9])
	exp2[9] = ""

	// tests pressing Key Enter in a middle line at bufCur.X = 0
	d3 := NewDisplay()
	initTestDisplay(d3)
	d3.ActiveBuf.addTestLines(createTestLines(10))
	exp3 := createTestLines(11)
	copy(exp3[6:], exp3[5:])
	exp3[5] = ""

	// tests pressing Key Enter at end of buffer where index > bufWindow size at bufCur.X = 0
	d4 := NewDisplay()
	initTestDisplay(d4)
	d4.ActiveBuf.addTestLines(createTestLines(99))
	expBuf4 := createTestLines(99)
	expBuf4 = append(expBuf4, expBuf4[98])
	expBuf4[98] = ""
	expWindow4 := append([]string(nil), expBuf4[51:]...)

	// tests pressing Key Enter at end of buffer where index > bufWindow size at bufCur.X = len(line.runes)
	d5 := NewDisplay()
	initTestDisplay(d5)
	d5.ActiveBuf.addTestLines(createTestLines(99))
	expBuf5 := createTestLines(99)
	expBuf5 = append(expBuf5, "")
	expWindow5 := append([]string(nil), expBuf5[51:]...)
	test5CurX := LeftMarginSize + len(expBuf5[98])

	// tests pressing Key Enter in the middle of a file in the middle of a line
	d6 := NewDisplay()
	initTestDisplay(d6)
	d6.ActiveBuf.addTestLines(createTestLines(99))
	expBuf6 := createTestLines(99)
	expBuf6 = append(expBuf6, "")
	copy(expBuf6[76:], expBuf6[75:])
	expBuf6[75] = expBuf6[75][:5]
	expBuf6[76] = expBuf6[76][5:]
	expWindow6 := append([]string(nil), expBuf6[50:len(expBuf6)-1]...)
	test6CurX := LeftMarginSize + 5

	tests := []struct {
		display   *Display
		x, y      int
		expBuf    []string
		expWindow []string
	}{
		{
			d1,
			LeftMarginSize, 0,
			exp1,
			exp1,
		},
		{
			d2,
			LeftMarginSize, 9,
			exp2,
			exp2,
		},
		{
			d3,
			LeftMarginSize, 5,
			exp3,
			exp3,
		},
		{
			d4,
			LeftMarginSize, 98,
			expBuf4,
			expWindow4,
		},
		{
			d5,
			test5CurX, 98,
			expBuf5,
			expWindow5,
		},
		{
			d6,
			test6CurX, 75,
			expBuf6,
			expWindow6,
		},
	}
	count := 0
	for _, tt := range tests {
		count++
		tt.display.bufWindow.update(tt.y)
		Cur.Y = tt.y - tt.display.bufWindow.bufIdx
		Cur.X = tt.x
		tt.display.setBufPos()

		tt.display.handleKeyEnter()
		if tt.display.ActiveBuf.length() != len(tt.expBuf) {
			t.Fatalf("length should be %d. Got %d", len(tt.expBuf), tt.display.ActiveBuf.length())
		}

		for i := range tt.expBuf {
			exp := tt.expBuf[i]
			res := string(tt.display.ActiveBuf.content.lines[i].runes)
			if exp != res {
				t.Fatalf("FAIL expBuf: line should be: '%s'. Got'%s'", exp, res)
			}
		}

		for i := range tt.expWindow {
			exp := tt.expWindow[i]
			res := string(tt.display.bufWindow.lines[i].runes)
			if exp != res {
				t.Fatalf("FAIL windowBuf: line should be: '%s'. Got'%s'", exp, res)
			}
		}
	}
}

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
		Cur.X = LeftMarginSize
		tt.display.bufWindow.update(tt.idx)
		Cur.Y = tt.idx - tt.display.bufWindow.bufIdx
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

	// tests deletion of line where the bufWindow cannot be scrolled down anymore
	d3 := NewDisplay()
	initTestDisplay(d3)
	d3Expected := createTestLines(100)
	d3Expected[82] = ""
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
		Cur.X = LeftMarginSize
		tt.display.bufWindow.update(tt.idx)
		Cur.Y = tt.idx - tt.display.bufWindow.bufIdx
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
	Cur.X = 0
	Cur.Y = 0
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%v", err)
	}
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	d.Screen = screen
	d.Screen.SetStyle(style)
	d.BufStyle = style
	d.width, d.height = 200, 50
	d.ActiveBuf = NewBuffer()
	d.bufWindow = newBufWindow()
	d.InitBufWindow()
	d.SetBufWindow()
	d.setBufPos()
}

func TestAddKeyTab(t *testing.T) {
	d := NewDisplay()
	initTestDisplay(d)
	Cur.X = LeftMarginSize
	Cur.Y = 0
	d.setBufPos()
	buf := d.ActiveBuf
	line := buf.getLine(0)
	line.runes = []rune("hello")
	line.addKeyTab()
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
	Cur.X = 0
	Cur.Y = 0
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
