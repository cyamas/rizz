package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
)

type Display struct {
	Screen  tcell.Screen
	Style   tcell.Style
	Buffers []*Buffer
	CurrBuf *Buffer
}

func NewDisplay() *Display {
	return &Display{Buffers: []*Buffer{}}
}

func (d *Display) Init() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%v", err)
	}
	d.Screen = screen
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	d.Screen.SetStyle(style)
	d.Style = style
	if err := screen.Init(); err != nil {
		log.Fatalf("%v", err)
	}
	screen.SetStyle(style)
	screen.Clear()

}

func (d *Display) addBuffer(buf *Buffer) {
	d.Buffers = append(d.Buffers, buf)
	d.CurrBuf = buf
}

func (d *Display) displayCurrentBuffer() {
	buf := d.CurrBuf
	for _, line := range buf.content.lines {
		for _, r := range line.runes {
			if r == '\t' {
				cur.x += 7
			}
			d.Screen.SetContent(cur.x, cur.y, r, []rune{}, d.Style)
			cur.x++
		}
		cur.x = 0
		cur.y++
	}
	cur.x = 0
	cur.y = 0
}

func (d *Display) SetRune(r rune) {
	d.CurrBuf.addRune(r)
	d.reRenderLineAtCursor()
	cur.x++
}

func (d *Display) handleKeyEnter() {
	d.CurrBuf.content.IncrementLength()
	d.shiftLinesDown()
	cur.x = 0
	cur.y++
}

func (d *Display) shiftLinesDown() {
	content := d.CurrBuf.content
	numLines := content.Length()
	line := content.lines[cur.y]
	d.clearCurrLine()
	extracted := line.extractRestOfLine(d)
	d.reRenderLine(cur.y)

	for i := cur.y + 1; i <= numLines; i++ {
		d.clearLineByIndex(i)
		content.lines[i].runes, extracted = extracted, content.lines[i].runes
		d.reRenderLine(i)
	}
}

func (l *Line) extractRestOfLine(d *Display) []rune {
	pushedRunes := []rune{}
	for i := cur.x; i < len(l.runes); i++ {
		pushedRunes = append(pushedRunes, l.runes[i])
	}
	d.clearLineFromCursor(l)
	l.runes = l.runes[:cur.x]
	return pushedRunes
}

func (d *Display) handleKeyBackspace() {
	switch {
	case cur.x == 0 && cur.y == 0:
		return
	case cur.x == 0:
		d.backspaceToPrevLine()
	default:
		d.backspaceChar()
	}
}

func (d *Display) backspaceToPrevLine() {
	prevLineLength := d.shiftLinesUp()
	d.CurrBuf.content.DecrementLength()
	cur.y--
	cur.x = prevLineLength
}

func (d *Display) backspaceChar() {
	d.clearCurrLine()
	d.CurrBuf.removePrevRune()
	d.reRenderLine(cur.y)
}

func (d *Display) shiftLinesUp() int {
	content := d.CurrBuf.content
	numLines := content.length
	prevLineLength := len(content.lines[cur.y-1].runes)
	for i := cur.y; i <= numLines; i++ {
		d.clearLineByIndex(i - 1)
		currRunes := content.lines[i].runes
		content.lines[i-1].runes = append(content.lines[i-1].runes, currRunes...)
		d.reRenderLine(i - 1)

		d.clearCurrLine()
		content.lines[i].runes = []rune{}
	}
	return prevLineLength
}

func (d *Display) reRenderLine(idx int) {
	line := d.CurrBuf.content.lines[idx]
	x := 0
	for _, r := range line.runes {
		d.Screen.SetContent(x, idx, r, []rune{}, d.Style)
		x++
	}
}

func (d *Display) clearLineByIndex(idx int) {
	line := d.CurrBuf.content.lines[idx]
	for i := 0; i < line.length; i++ {
		d.Screen.SetContent(i, idx, ' ', []rune{}, d.Style)
	}
}

func (d *Display) clearCurrLine() {
	d.clearLineByIndex(cur.y)
}

func (l *Line) Clear(d *Display) {
	for i := 0; i < len(l.runes); i++ {
		d.Screen.SetContent(i, cur.y, ' ', []rune{}, d.Style)
	}
}

func (d *Display) reRenderLineAtCursor() {
	line := d.CurrBuf.getLine(cur.y)
	d.clearLineFromCursor(line)
	d.resetContentFromCursor(line)
}

func (d *Display) clearLineFromCursor(line *Line) {
	for i := cur.x; i < line.length; i++ {
		d.Screen.SetContent(i, cur.y, ' ', []rune{}, d.Style)
	}
}

func (d *Display) resetContentFromCursor(line *Line) {
	idx := cur.x
	for i := idx; i < line.Length(); i++ {
		r := line.runes[i]
		d.Screen.SetContent(idx, cur.y, r, []rune{}, d.Style)
		idx++
	}
}

type Buffer struct {
	content *LineArray
}

func newBuffer() *Buffer {
	return &Buffer{content: NewLineArray()}
}

func (b *Buffer) getLine(idx int) *Line {
	return b.content.lines[idx]
}

func (b *Buffer) getPrevRune() rune {
	line := b.content.lines[cur.y]
	runeIdx := line.prevRuneIndex()
	return line.runes[runeIdx]
}

func (b *Buffer) currLineLength() int {
	return b.content.lines[cur.y].length
}

type LineArray struct {
	lines  []*Line
	length int
	lock   sync.Mutex
}

func NewLineArray() *LineArray {
	la := &LineArray{}
	for i := 0; i < 10000; i++ {
		la.lines = append(la.lines, newLine())
	}
	return la
}

func (la *LineArray) addLineContent(text string, lineNum int) {
	line := la.lines[lineNum]
	for _, r := range text {
		line.runes = append(line.runes, r)
	}
	la.IncrementLength()
}

func (la *LineArray) IncrementLength() {
	la.length++
}

func (la *LineArray) DecrementLength() {
	la.length--
}

func (la *LineArray) Length() int {
	return la.length
}

type Line struct {
	runes  []rune
	length int
	tabs   int
	lock   sync.Mutex
}

func newLine() *Line {
	return &Line{runes: []rune{}}
}

func (l *Line) currRuneIndex() int {
	idx := 0
	for i, r := range l.runes {
		if idx == cur.x {
			return i
		}
		if r == '\t' {
			idx = l.nextTabStopFromIndex(idx)
			continue
		}
		idx++
	}
	return len(l.runes) - 1
}

func (l *Line) prevRuneIndex() int {
	return cur.x - 1 - (l.tabs * 7)
}

func (l *Line) Length() int {
	return len(l.runes)
}

type cell struct {
	x, y  int
	r     rune
	style tcell.Style
}

var cur cell

func main() {
	fmt.Println("rizz will be a simple text editor one day!")
	args := os.Args

	d := NewDisplay()
	d.Init()

	quit := func() {
		maybePanic := recover()
		d.Screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	d.addBuffer(newBuffer())
	if len(args) == 2 {
		d.CurrBuf.readFile(args[1])
		d.displayCurrentBuffer()
	}

	d.runNormalMode()
}

func (d *Display) runNormalMode() {
	for {
		d.Screen.ShowCursor(cur.x, cur.y)
		d.Screen.Show()
		ev := d.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			d.Screen.Sync()
		case *tcell.EventKey:
			switch {
			case ev.Rune() == 'Q':
				return
			case ev.Rune() == 'j':
				d.moveCursorDown()
			case ev.Rune() == 'k':
				d.moveCursorUp()
			case ev.Rune() == 'l':
				d.moveCursorRight()
			case ev.Rune() == 'h':
				d.moveCursorLeft()
			case ev.Rune() == 'w':
				d.moveCursorToNextWord()
			case ev.Rune() == 'O':
				d.runOpenMode()
			case ev.Rune() == 'I':
				d.runInsertMode()
			}
		}
	}
}

func (d *Display) moveCursorToNextWord() {

}

func (d *Display) moveCursorDown() {
	if cur.y == d.CurrBuf.content.Length() {
		return
	}
	nextLineLength := d.CurrBuf.getLine(cur.y + 1).length
	if nextLineLength < cur.x {
		cur.x = nextLineLength
	}
	cur.y++
}

func (d *Display) moveCursorUp() {
	if cur.y == 0 {
		return
	}
	prevLineLength := d.CurrBuf.getLine(cur.y - 1).length
	if prevLineLength < cur.x {
		cur.x = prevLineLength
	}
	cur.y--
}

func (d *Display) moveCursorLeft() {
	if cur.x == 0 {
		return
	}
	cur.x--
}

func (d *Display) moveCursorRight() {
	if cur.x == d.CurrBuf.currLine().length {
		return
	}
	cur.x++
}

func (b *Buffer) currLine() *Line {
	return b.getLine(cur.y)
}

func (b *Buffer) readFile(filename string) {
	b.content = getFileContents(filename)
}

func getFileContents(filename string) *LineArray {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("%v", err)
	}
	path := dir + "/" + filename

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("could not open file", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	content := NewLineArray()
	lineNum := 0
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		content.addLineContent(text, lineNum)
		lineNum++
	}
	return content
}

// Skips 7 spaces when the tab character is reached. Not sure if I am able to change tcell's tab = 1 space myself
// so this is my hack around it
func readKeyTab() {
	cur.x += 7
}

func (d *Display) runOpenMode() {

}

func (d *Display) runInsertMode() {
	for {
		d.Screen.ShowCursor(cur.x, cur.y)
		d.Screen.Show()
		ev := d.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyCtrlN:
				return
			case ev.Key() == tcell.KeyEnter:
				d.handleKeyEnter()
			case ev.Key() == tcell.KeyBackspace2:
				d.handleKeyBackspace()
			case ev.Key() == tcell.KeyTab:
				d.handleKeyTab()
			default:
				d.SetRune(ev.Rune())

			}
		}
	}
}

func (d *Display) handleKeyTab() {
	d.CurrBuf.addKeyTab()
	d.reRenderLine(cur.y)
}

func (b *Buffer) addKeyTab() {
	line := b.getLine(cur.y)
	tab := createTabRunes()
	newRunes := line.runes[:cur.x]
	newRunes = append(newRunes, tab...)
	newRunes = append(newRunes, line.runes[cur.x:]...)
	line.runes = newRunes
	line.length = len(line.runes)
	cur.x = nextTabStopIdx()
}

func createTabRunes() []rune {
	tab := []rune{}
	tab = append(tab, '\t')
	for i := 1; i < nextTabStopOffset()-1; i++ {
		tab = append(tab, ' ')
	}
	tab = append(tab, '\t')
	return tab
}

func nextTabStopOffset() int {
	return 8 - (cur.x % 8)
}

func nextTabStopIdx() int {
	return cur.x + nextTabStopOffset()
}

func (l *Line) nextTabStopFromIndex(x int) int {
	return x + 8 - (x % 8)
}

func (d *Display) clearContentForTab() {
	for i := 0; i < 7; i++ {
		d.Screen.SetContent(cur.x, cur.y, ' ', []rune{}, d.Style)
		cur.x++
	}
}

func (b *Buffer) addRune(r rune) {
	line := b.getLine(cur.y)
	line.runes = append(line.runes, r)
	line.length = len(line.runes)
	for i := line.Length() - 1; i > cur.x; i-- {
		line.runes[i], line.runes[i-1] = line.runes[i-1], line.runes[i]
	}
}

func (b *Buffer) removePrevRune() {
	line := b.getLine(cur.y)
	r := line.runes[cur.x-1]
	if r == '\t' {
		line.removeTabRunes()
		return
	}
	newRunes := line.runes[:cur.x-1]
	newRunes = append(newRunes, line.runes[cur.x:]...)
	line.runes = newRunes
	line.length = len(line.runes)
	cur.x--
}

func (l *Line) removeTabRunes() {
	idx := cur.x - 2
	for {
		if l.runes[idx] == '\t' {
			break
		}
		idx--
	}
	postTab := l.runes[cur.x:]
	l.runes = l.runes[:idx]
	l.runes = append(l.runes, postTab...)
	cur.x = idx
}
