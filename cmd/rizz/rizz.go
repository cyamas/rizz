package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

const (
	Normal = iota
	Insert
	Exit
	Open
	Write
	Delete
	New
)

var modes = map[int]string{
	Normal: "Normal",
	Insert: "Insert",
	Exit:   "Exit",
	Open:   "Open",
	Write:  "Write",
	Delete: "Delete",
	New:    "New",
}

type Display struct {
	Screen    tcell.Screen
	width     int
	height    int
	Style     tcell.Style
	Buffers   []*Buffer
	CurrBuf   *Buffer
	Mode      int
	StatusBar []rune
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
	d.width, d.height = d.Screen.Size()
	screen.SetStyle(style)
	screen.Clear()

}

func (d *Display) setStatusBar() {
	d.clearStatusBar()
	lineCount := d.CurrBuf.content.length
	lineLength := len(d.CurrBuf.content.lines[cur.y].runes)
	text := convertRunesToText(d.CurrBuf.getLine(cur.y).runes)
	status := []rune(fmt.Sprintf("%s Mode			Line: %d		Col: %d		LineCount: %d		LineLength: %d		Text: %s",
		modes[d.Mode],
		cur.y,
		cur.x,
		lineCount,
		lineLength,
		text))
	d.StatusBar = status
	for i, r := range status {
		d.Screen.SetContent(i, d.height-1, r, nil, d.Style)
	}
}

func convertRunesToText(runes []rune) string {
	str := ""
	for _, r := range runes {
		if r == '\t' {
			str += "/t"
			continue
		}
		str += string(r)
	}
	return str
}

func (d *Display) clearStatusBar() {
	for i := range d.StatusBar {
		d.Screen.SetContent(i, d.height-1, ' ', nil, d.Style)
	}
}

func (d *Display) addBuffer(buf *Buffer) {
	d.Buffers = append(d.Buffers, buf)
	d.CurrBuf = buf
}

func (d *Display) displayCurrentBuffer() {
	buf := d.CurrBuf
	for _, line := range buf.content.lines {
		for _, r := range line.runes {
			d.Screen.SetContent(cur.x, cur.y, r, nil, d.Style)
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
	d.shiftLinesDown()
	cur.x = 0
	cur.y++
}

func (d *Display) shiftLinesDown() {
	content := d.CurrBuf.content
	d.clearLinesToEOF()
	newLine := content.newLineFromKeyEnter()
	content.insertNewLine(newLine)
	content.length++
	d.reRenderLinesToEOF()
}

func (la *LineArray) insertNewLine(line *Line) {
	if cur.y == la.length-1 {
		la.lines = append(la.lines, line)
		return
	}
	la.lines = append(la.lines, nil)
	copy(la.lines[cur.y+2:], la.lines[cur.y+1:])
	la.lines[cur.y+1] = line
}

func (d *Display) clearLinesToEOF() {
	content := d.CurrBuf.content
	for i := cur.y; i < content.length; i++ {
		d.clearLineByIndex(i)
	}
}

func (d *Display) reRenderLinesToEOF() {
	content := d.CurrBuf.content
	for i := cur.y; i < content.length; i++ {
		d.reRenderLine(i)
	}
}

func (la *LineArray) newLineFromKeyEnter() *Line {
	newLine := newLine()
	newLine.runes = la.lines[cur.y].extractRestOfLine()
	newLine.length = len(newLine.runes)
	return newLine
}

func (l *Line) extractRestOfLine() []rune {
	pushedRunes := l.runes[cur.x:]
	l.runes = l.runes[:cur.x]
	l.length = len(l.runes)
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
	cur.y--
	prevLineLength := d.shiftLinesUp()
	d.reRenderLinesToEOF()
	cur.x = prevLineLength
}

func (d *Display) backspaceChar() {
	d.clearCurrLine()
	d.CurrBuf.removePrevRune()
	d.reRenderLine(cur.y)
}

func (d *Display) shiftLinesUp() int {
	content := d.CurrBuf.content
	d.clearLinesToEOF()
	ogLineLength := d.CurrBuf.getLine(cur.y).length
	content.lines[cur.y].runes = append(content.lines[cur.y].runes, content.lines[cur.y+1].runes...)
	content.lines = append(content.lines[:cur.y+1], content.lines[cur.y+2:]...)
	content.length--
	return ogLineLength
}

func (d *Display) reRenderLine(idx int) {
	line := d.CurrBuf.getLine(idx)
	line.length = len(line.runes)
	x := 0
	for i := range line.runes {
		d.Screen.SetContent(x, idx, line.runes[i], nil, d.Style)
		x++
	}
}

func (d *Display) clearLineByIndex(idx int) {
	line := d.CurrBuf.content.lines[idx]
	for i := 0; i < line.length; i++ {
		d.Screen.SetContent(i, idx, ' ', nil, d.Style)
	}
}

func (d *Display) clearCurrLine() {
	d.clearLineByIndex(cur.y)
}

func (d *Display) reRenderLineAtCursor() {
	line := d.CurrBuf.getLine(cur.y)
	d.clearLineFromCursor(line)
	d.resetContentFromCursor(line)
}

func (d *Display) clearLineFromCursor(line *Line) {
	for i := cur.x; i < line.length; i++ {
		d.Screen.SetContent(i, cur.y, ' ', nil, d.Style)
	}
}

func (d *Display) resetContentFromCursor(line *Line) {
	idx := cur.x
	for i := idx; i < line.Length(); i++ {
		r := line.runes[i]
		d.Screen.SetContent(idx, cur.y, r, nil, d.Style)
		idx++
	}
}

type Buffer struct {
	content *LineArray
	path    string
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
	return b.currLine().length
}

type LineArray struct {
	lines  []*Line
	length int
	lock   sync.Mutex
}

func NewLineArray() *LineArray {
	la := &LineArray{}
	la.lines = make([]*Line, 0, 10000)
	la.lines = append(la.lines, newLine())
	la.length = 1
	return la
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
	d.Mode = Normal
	d.run()
}

func (d *Display) run() {
	for {
		if d.Mode == Exit {
			return
		}
		if d.Mode == Write {
			d.CurrBuf.writeFile()
			d.Mode = Normal
		}
		d.setStatusBar()
		d.Screen.ShowCursor(cur.x, cur.y)
		d.Screen.Show()
		ev := d.Screen.PollEvent()
		switch {
		case d.Mode == Insert:
			d.runInsertMode(ev)
		case d.Mode == Normal:
			d.runNormalMode(ev)
		case d.Mode == Open:
			d.runOpenMode(ev)
		case d.Mode == Delete:
			d.runDeleteMode(ev)
		case d.Mode == New:
			d.runNewMode(ev)
		}

	}
}

func (d *Display) runNormalMode(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventResize:
		d.Screen.Sync()
	case *tcell.EventKey:
		switch {
		case ev.Rune() == 'Q':
			d.Mode = Exit
		case ev.Rune() == 'W':
			d.Mode = Write
		case ev.Rune() == 'j':
			d.moveCursorDown()
		case ev.Rune() == 'k':
			d.moveCursorUp()
		case ev.Rune() == 'l':
			d.moveCursorRight()
		case ev.Rune() == 'h':
			d.moveCursorLeft()
		case ev.Rune() == 'w':
			d.moveCursorToNextWord(false)
		case ev.Rune() == 'b':
			d.moveCursorToPreviousWord()
		case ev.Rune() == 'O':
			d.Mode = Open
		case ev.Rune() == 'I':
			d.Mode = Insert
		case ev.Rune() == 'd':
			d.Mode = Delete
		case ev.Rune() == 'n':
			d.Mode = New
		}
	}

}

func (d *Display) runInsertMode(ev tcell.Event) {
	d.setStatusBar()
	d.Screen.ShowCursor(cur.x, cur.y)
	d.Screen.Show()
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch {
		case ev.Key() == tcell.KeyCtrlN:
			d.Mode = Normal
		case ev.Key() == tcell.KeyEnter:
			d.handleKeyEnter()
		case ev.Key() == tcell.KeyBackspace2:
			d.handleKeyBackspace()
		case ev.Key() == tcell.KeyTab:
			d.handleKeyTab()
			cur.x += nextTabStopOffset()
		default:
			d.SetRune(ev.Rune())

		}
	}
}

func (d *Display) runDeleteMode(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch {
		case ev.Rune() == 'd':
			d.deleteLine()
			d.Mode = Normal
		}
	}
}

func (d *Display) runNewMode(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch {
		case ev.Rune() == 'l':
			d.insertBlankLine()
		}
	}
}

func (d *Display) insertBlankLine() {
	line := newLine()
	d.clearLinesToEOF()
	d.CurrBuf.content.insertNewLine(line)
	d.reRenderLinesToEOF()
	cur.x = 0
	cur.y++
	d.Mode = Insert

}

func (d *Display) deleteLine() {
	line := d.CurrBuf.currLine()
	line.runes = []rune{}
	d.shiftLinesUp()
	d.reRenderLinesToEOF()
	cur.x = 0
}

func (d *Display) runOpenMode(ev tcell.Event) {

}

var wordSeparators = map[rune]bool{
	' ': true,
	'.': true,
	':': true,
	';': true,
	',': true,
	'"': true,
	'{': true,
	'}': true,
	'(': true,
	')': true,
	'[': true,
	']': true,
}

func (d *Display) moveCursorToPreviousWord() {
	wordFound := false
	line := d.CurrBuf.currLine()
	if cur.x == 0 && cur.y == 0 {
		return
	}
	if cur.x == 0 {
		cur.y--
		if d.CurrBuf.currLineLength() == 0 {
			cur.x = 0
		} else {
			cur.x = d.CurrBuf.currLineLength() - 1
		}
		return
	}
	if isLetterOrNumber(line.runes[cur.x]) && isLetterOrNumber(line.runes[cur.x-1]) {
		wordFound = true
	}
	for i := cur.x - 1; i >= 0; i-- {
		r := line.runes[i]
		switch {
		case i == 0:
			cur.x = 0
			return
		case r == '\t':
			if wordFound && (r != ' ' || r != '\t') {
				cur.x = i + 1
				return
			}
		case wordFound && isSeparator(r):
			cur.x = i + 1
			return
		case !wordFound && isSeparator(r):
			if r == ' ' {
				continue
			}
			cur.x = i
			return
		default:
			wordFound = true
		}
	}
}

func isLetterOrNumber(r rune) bool {
	return unicode.IsDigit(r) || unicode.IsLetter(r)
}

func isSeparator(r rune) bool {
	_, ok := wordSeparators[r]
	return ok
}

func (d *Display) moveCursorToNextWord(sepFound bool) {
	line := d.CurrBuf.currLine()
	for i := cur.x; i < len(line.runes); i++ {
		if _, ok := wordSeparators[line.runes[i]]; ok {
			sepFound = true
			continue
		}
		switch line.runes[i] {
		case '\t':
			i = line.nextTabStopFromIndex(i) - 1
			continue
		case '\'':
			if !isApostrophe(line.runes, i) {
				sepFound = true
			}
		default:
			if sepFound == true {
				cur.x = i
				return
			}
		}
	}
	if d.CurrBuf.content.length == cur.y+1 {
		return
	}
	cur.x = 0
	cur.y++
	d.moveCursorToNextWord(true)
}

func isApostrophe(runes []rune, idx int) bool {
	if idx == 0 || idx == len(runes)-1 {
		return false
	}
	left, right := runes[idx-1], runes[idx+1]
	if unicode.IsLetter(left) && unicode.IsLetter(right) {
		return true
	}
	return false
}

func (d *Display) moveCursorDown() {
	if cur.y == d.CurrBuf.content.length-1 {
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

func (b *Buffer) writeFile() {
	file, err := os.Create(b.path)
	if err != nil {
		log.Fatalln("Could not create file: ", err)
	}
	defer file.Close()

	newContent := ""
	for _, line := range b.content.lines {
		newContent += line.convertRunesForWrite()
	}
	_, err = file.WriteString(newContent)
	if err != nil {
		log.Fatalln("could not write to file")
	}
}

func (l *Line) convertRunesForWrite() string {
	str := ""
	for i := 0; i < len(l.runes); i++ {
		str += string(l.runes[i])
		if l.runes[i] == '\t' {
			i += nextTabStopOffsetFromIndex(i) - 1
		}
	}
	str += "\n"
	return str
}

func (b *Buffer) readFile(filename string) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalln("Could not get working directory", err)
	}
	b.path = dir + "/" + filename
	b.content = b.setContentFromFile()
}

func (b *Buffer) setContentFromFile() *LineArray {
	file, err := os.Open(b.path)
	if err != nil {
		fmt.Println("could not open file", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	content := NewLineArray()
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		content.addLineFromFile(text)
		content.length++
	}
	content.length--
	return content
}

func (la *LineArray) addLineFromFile(text string) {
	line := newLine()
	for _, ch := range text {
		if ch == '\n' {
			break
		}
		if ch == '\t' {
			line.addTabFromFile()
			continue
		}
		line.runes = append(line.runes, ch)
	}
	line.length = len(line.runes)
	if la.length == 1 {
		la.lines[0] = line
	} else {
		la.lines = append(la.lines, line)
	}
}

func (l *Line) addTabFromFile() {
	l.runes = append(l.runes, '\t')
	offset := nextTabStopOffsetFromIndex(len(l.runes) - 1)
	for i := 1; i < offset-1; i++ {
		l.runes = append(l.runes, ' ')
	}
	l.runes = append(l.runes, '\t')
}

func (d *Display) handleKeyTab() {
	d.CurrBuf.addKeyTab()
	d.clearCurrLine()
	d.reRenderLine(cur.y)
}

func (b *Buffer) addKeyTab() {
	tab := createTabRunes()
	line := b.getLine(cur.y)

	for i := 0; i < len(tab); i++ {
		line.runes = append(line.runes, ' ')
	}
	copy(line.runes[cur.x+len(tab):], line.runes[cur.x:])
	for i := range tab {
		line.runes[cur.x+i] = tab[i]
	}
	line.length = len(line.runes)
}

func createTabRunes() []rune {
	tab := []rune{}
	tab = append(tab, '\t')
	offset := nextTabStopOffset()
	if offset == 1 {
		return tab
	}

	for i := 1; i < offset-1; i++ {
		tab = append(tab, ' ')
	}
	tab = append(tab, '\t')
	return tab
}

func nextTabStopOffset() int {
	return 8 - (cur.x % 8)
}

func nextTabStopOffsetFromIndex(idx int) int {
	return 8 - (idx % 8)
}

func nextTabStopIdx() int {
	return cur.x + nextTabStopOffset()
}

func (l *Line) nextTabStopFromIndex(x int) int {
	return x + 8 - (x % 8)
}

func (b *Buffer) addRune(r rune) {
	line := b.getLine(cur.y)
	line.runes = append(line.runes, ' ')
	copy(line.runes[cur.x+1:], line.runes[cur.x:])
	line.runes[cur.x] = r
	line.length++
}

func (b *Buffer) removePrevRune() {
	line := b.getLine(cur.y)
	r := line.runes[cur.x-1]
	if r == '\t' {
		line.removeTabRunes()
		return
	}
	line.runes = append(line.runes[:cur.x-1], line.runes[cur.x:]...)
	line.length--
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
	l.runes = append(l.runes[:idx], l.runes[cur.x:]...)
	l.length = len(l.runes)
	cur.x = idx
}
