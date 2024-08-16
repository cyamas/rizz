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
				readKeyTab()
			}
			d.Screen.SetContent(lastCursor.x, lastCursor.y, r, []rune{}, d.Style)
			lastCursor.x++
		}
		lastCursor.x = 0
		lastCursor.y++
	}
	lastCursor.x = 0
	lastCursor.y = 0
}

func (d *Display) SetRune(r rune) {
	if r == '\t' {
		lastCursor.x += 7
	}
	d.Screen.SetContent(lastCursor.x, lastCursor.y, r, []rune{}, d.Style)
	d.Screen.Show()
	d.CurrBuf.addRune(r)
	lastCursor.x++
	d.Screen.ShowCursor(lastCursor.x, lastCursor.y)
}

func (d *Display) reRenderLine() {
	line := d.CurrBuf.getLine(lastCursor.y)
	d.clearLine(line)
	d.resetContent(line)
}

func (d *Display) clearLine(line *Line) {
	for i := range line.length {
		d.Screen.SetContent(i, lastCursor.y, ' ', []rune{}, d.Style)
		d.Screen.Show()
	}
}

func (d *Display) resetContent(line *Line) {
	idx := 0
	for _, r := range line.runes {
		if r == '\t' {
			idx += 7
		}
		d.Screen.SetContent(idx, lastCursor.y, r, []rune{}, d.Style)
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
	line := b.content.lines[lastCursor.y]
	runeIdx := line.prevRuneIndex()
	return line.runes[runeIdx]
}

func (b *Buffer) currLineLength() int {
	return b.content.lines[lastCursor.y].length
}

type LineArray struct {
	lines []*Line
	lock  sync.Mutex
}

func NewLineArray() *LineArray {
	la := &LineArray{}
	for i := 0; i < 10000; i++ {
		la.lines = append(la.lines, newLine())
	}
	return la
}

func (la *LineArray) AppendLine(text string, lineNum int) {
	line := la.lines[lineNum]
	line.runes = []rune(text)
	line.setLength()
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
	return lastCursor.x - (l.tabs * 7)
}

func (l *Line) prevRuneIndex() int {
	return lastCursor.x - 1 - (l.tabs * 7)
}

func (l *Line) setLength() {
	l.length = 0
	for _, r := range l.runes {
		l.length++
		if r == '\t' {
			l.length += 7
		}
	}
}

type cell struct {
	x, y  int
	r     rune
	style tcell.Style
}

var lastCursor cell

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
		d.Screen.ShowCursor(lastCursor.x, lastCursor.y)
		ev := d.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			d.Screen.Sync()
		case *tcell.EventKey:
			switch {
			case ev.Rune() == 'j':
				lastCursor.y++
			case ev.Rune() == 'k':
				lastCursor.y--
			case ev.Rune() == 'l':
				lastCursor.x++
			case ev.Rune() == 'h':
				lastCursor.x--
			case ev.Rune() == 'Q':
				return
			case ev.Rune() == 'O':
				d.runOpenMode()
			case ev.Rune() == 'I':
				d.runInsertMode()
			}
		}
		d.Screen.ShowCursor(lastCursor.x, lastCursor.y)
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
		content.AppendLine(text, lineNum)
		lineNum++
	}
	return content
}

// Skips 7 spaces when the tab character is reached. Not sure if I am able to change tcell's tab = 1 space myself
// so this is my hack around it
func readKeyTab() {
	lastCursor.x += 7
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
				lastCursor.x = 0
				lastCursor.y++
				d.Screen.ShowCursor(lastCursor.x, lastCursor.y)
			case ev.Key() == tcell.KeyBackspace2:
				d.handleBackspace()
				d.reRenderLine()
				d.Screen.ShowCursor(lastCursor.x, lastCursor.y)
			default:
				d.SetRune(ev.Rune())
				d.Screen.ShowCursor(lastCursor.x, lastCursor.y)

			}
		}
		d.Screen.Show()
		d.Screen.ShowCursor(lastCursor.x, lastCursor.y)
	}
}

func (b *Buffer) addRune(r rune) {
	line := b.getLine(lastCursor.y)
	x := lastCursor.x
	line.runes = append(line.runes, r)
	line.setLength()
	if r == '\t' {
		line.tabs++
	}
	if line.length == x {
		return
	}
	for i := len(line.runes) - 1; i > line.currRuneIndex(); i-- {
		line.runes[i], line.runes[i-1] = line.runes[i-1], line.runes[i]
	}
}

func (d *Display) handleBackspace() {
	if lastCursor.x == 0 {
		if lastCursor.y == 0 {
			return
		}
		lastCursor.y--
		lastCursor.x = d.CurrBuf.currLineLength()
		return
	}
	r := d.CurrBuf.extractPrevRune()
	d.Screen.SetContent(lastCursor.x-1, lastCursor.y, ' ', []rune{}, d.Style)
	if r == '\t' {
		lastCursor.x -= 7
	}
	lastCursor.x--
}

func (b *Buffer) extractPrevRune() rune {
	r := b.getPrevRune()
	line := b.getLine(lastCursor.y)
	runeIdx := line.prevRuneIndex()
	if r == '\t' {
		line.tabs--
	}

	newRunes := []rune{}
	for i, rune := range line.runes {
		if i == runeIdx {
			continue
		}
		newRunes = append(newRunes, rune)
	}
	line.runes = newRunes
	line.setLength()

	return r
}
