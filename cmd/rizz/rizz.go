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
	if r == '\t' {
		cur.x += 7
	}
	d.CurrBuf.addRune(r)
	d.reRenderLineAtCursor()
	cur.x++
	d.Screen.ShowCursor(cur.x, cur.y)
	d.Screen.Show()
}

func (d *Display) handleKeyEnter() {
	d.shiftLinesDown()
	cur.x = 0
	cur.y++
	d.CurrBuf.content.IncrementLength()
	d.Screen.ShowCursor(cur.x, cur.y)
	d.Screen.Show()
}

func (d *Display) shiftLinesDown() {
	content := d.CurrBuf.content
	numLines := content.Length()
	line := content.lines[cur.y]
	extracted := line.extractRestOfLine(d)
	d.reRenderLine(cur.y)

	for i := cur.y + 1; i <= numLines; i++ {
		content.lines[i].runes, extracted = extracted, content.lines[i].runes
		d.reRenderLine(i)
		content.lines[i].setLengthAndTabs()
	}
}

func (l *Line) extractRestOfLine(d *Display) []rune {
	currRuneIdx := l.currRuneIndex()
	pushedRunes := []rune{}
	for i := currRuneIdx; i < len(l.runes); i++ {
		pushedRunes = append(pushedRunes, l.runes[i])
	}
	d.clearLineFromCursor(l)
	l.runes = l.runes[:currRuneIdx]
	l.setLengthAndTabs()
	return pushedRunes
}

func (d *Display) handleKeyBackspace() {
	if cur.x == 0 {
		if cur.y == 0 {
			return
		}
		prevLineLength := d.shiftLinesUp()
		d.CurrBuf.content.DecrementLength()
		cur.y--
		cur.x = prevLineLength
		d.Screen.ShowCursor(cur.x, cur.y)
		d.Screen.Show()
		return
	}
	r := d.CurrBuf.extractPrevRune()
	eolIdx := d.CurrBuf.currLineLength()
	d.Screen.SetContent(eolIdx, cur.y, ' ', []rune{}, d.Style)
	if r == '\t' {
		for i := range 8 {
			d.Screen.SetContent(eolIdx+i, cur.y, ' ', []rune{}, d.Style)
		}
		cur.x -= 7
	}
	cur.x--
	d.reRenderLineAtCursor()
	d.Screen.ShowCursor(cur.x, cur.y)
	d.Screen.Show()
}

func (d *Display) shiftLinesUp() int {
	content := d.CurrBuf.content
	numLines := content.length
	prevLineLength := content.lines[cur.y-1].length
	for i := cur.y; i < numLines; i++ {
		currRunes := content.lines[i].runes
		content.lines[i-1].runes = append(content.lines[i-1].runes, currRunes...)
		content.lines[i-1].setLengthAndTabs()
		d.reRenderLine(i - 1)

		content.lines[i].runes = []rune{}
		d.clearLine(i)
		content.lines[i].setLengthAndTabs()
	}
	return prevLineLength
}

func (d *Display) reRenderLine(idx int) {
	line := d.CurrBuf.content.lines[idx]
	d.clearLine(idx)
	for i, r := range line.runes {
		d.Screen.SetContent(i, idx, r, []rune{}, d.Style)
	}
}

func (d *Display) clearLine(idx int) {
	line := d.CurrBuf.content.lines[idx]
	for i := range line.length {
		d.Screen.SetContent(i, idx, ' ', []rune{}, d.Style)
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
	runeIdx := line.currRuneIndex()
	idx := cur.x
	for i := runeIdx; i < len(line.runes); i++ {
		r := line.runes[i]
		if r == '\t' {
			idx += 7
		}
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
	la.length = 1
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
	line.setLengthAndTabs()
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
	return cur.x - (l.tabs * 7)
}

func (l *Line) prevRuneIndex() int {
	return cur.x - 1 - (l.tabs * 7)
}

func (l *Line) setLengthAndTabs() {
	l.length = 0
	l.tabs = 0
	for _, r := range l.runes {
		l.length++
		if r == '\t' {
			l.length += 7
			l.tabs++
		}
	}
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
		d.Screen.Show()
		d.Screen.ShowCursor(cur.x, cur.y)
		ev := d.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			d.Screen.Sync()
		case *tcell.EventKey:
			switch {
			case ev.Rune() == 'j':
				cur.y++
			case ev.Rune() == 'k':
				cur.y--
			case ev.Rune() == 'l':
				cur.x++
			case ev.Rune() == 'h':
				cur.x--
			case ev.Rune() == 'Q':
				return
			case ev.Rune() == 'O':
				d.runOpenMode()
			case ev.Rune() == 'I':
				d.runInsertMode()
			}
		}
		d.Screen.ShowCursor(cur.x, cur.y)
	}
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
			default:
				d.SetRune(ev.Rune())

			}
		}
	}
}

func (b *Buffer) addRune(r rune) {
	line := b.getLine(cur.y)
	x := cur.x
	line.runes = append(line.runes, r)
	line.setLengthAndTabs()
	if line.length == x {
		return
	}
	for i := len(line.runes) - 1; i > line.currRuneIndex(); i-- {
		line.runes[i], line.runes[i-1] = line.runes[i-1], line.runes[i]
	}
}

func (b *Buffer) extractPrevRune() rune {
	r := b.getPrevRune()
	line := b.getLine(cur.y)
	runeIdx := line.prevRuneIndex()
	newRunes := []rune{}
	for i, rune := range line.runes {
		if i == runeIdx {
			continue
		}
		newRunes = append(newRunes, rune)
	}
	line.runes = newRunes
	line.setLengthAndTabs()

	return r
}
