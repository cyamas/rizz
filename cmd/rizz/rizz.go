package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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

const leftMarginSize = 8

var modes = map[int]string{
	Normal: "Normal",
	Insert: "Insert",
	Exit:   "Exit",
	Open:   "Open",
	Write:  "Write",
	Delete: "Delete",
	New:    "New",
}

type cell struct {
	x, y int
}

var cur cell
var bufPos cell

func main() {
	fmt.Println("rizz will be a simple text editor one day!")
	args := os.Args
	d := NewDisplay()
	d.init()
	quit := func() {
		maybePanic := recover()
		d.Screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()
	d.ActiveBuf = newBuffer()
	d.BufWindow = newBuffer()
	if len(args) == 2 {
		d.ActiveBuf.readFile(args[1])
		d.setBufWindow(0)
		d.displayBufWindow()
	}
	d.Mode = Normal
	cur.x = leftMarginSize
	cur.y = 0
	d.run()
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
	content := newLineArray()
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		content.addLineFromFile(text)
	}
	return content
}

func (la *LineArray) addLineFromFile(text string) {
	line := newLine()
	for _, ch := range text {
		switch ch {
		case '\n':
			break
		case '\t':
			line.addTabFromFile()
		default:
			line.runes = append(line.runes, ch)
		}
	}
	if len(la.lines) == 1 && len(la.lines[0].runes) == 0 {
		la.lines[0] = line
		return
	}
	la.lines = append(la.lines, line)
}

func (d *Display) setBufWindow(idx int) {
	d.BufWindow.startIdx = idx
	activeBuf := d.ActiveBuf
	bufHeight := d.height - 1
	if activeBuf.length() < bufHeight {
		d.BufWindow.content.lines = activeBuf.content.lines
		return
	}
	if idx+bufHeight > activeBuf.length() {
		idx = activeBuf.length() - bufHeight
	}
	d.BufWindow.content.lines = activeBuf.content.lines[idx : idx+bufHeight]
}

func (d *Display) displayBufWindow() {
	buf := d.BufWindow
	cur.x = leftMarginSize
	for y, line := range buf.content.lines {
		for j, r := range line.runes {
			x := cur.x + j
			d.Screen.SetContent(x, y, r, nil, d.BufStyle)
		}
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

type Display struct {
	Screen         tcell.Screen
	width          int
	height         int
	BufWindow      *Buffer
	ActiveBuf      *Buffer
	Mode           int
	StatusBar      []rune
	BufStyle       tcell.Style
	LineNoStyle    tcell.Style
	StatusBarStyle tcell.Style
}

func NewDisplay() *Display {
	bufStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	lineNoStyle := tcell.StyleDefault.Foreground(tcell.ColorAqua).Background(tcell.ColorBlack)
	statusBarStyle := tcell.StyleDefault.Foreground(tcell.ColorWhiteSmoke).Background(tcell.ColorBlack)
	return &Display{
		BufStyle:       bufStyle,
		LineNoStyle:    lineNoStyle,
		StatusBarStyle: statusBarStyle,
	}
}

func (d *Display) init() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%v", err)
	}
	d.Screen = screen
	if err := screen.Init(); err != nil {
		log.Fatalf("%v", err)
	}
	d.Screen.SetStyle(d.BufStyle)
	d.width, d.height = d.Screen.Size()
	d.Screen.Clear()
	d.Mode = Normal
}

func (d *Display) run() {
	for {
		if d.Mode == Exit {
			return
		}
		if d.Mode == Write {
			d.ActiveBuf.writeToFile()
			d.Mode = Normal
		}
		setBufPos()
		d.setStatusBar()
		d.setLineNumbers()
		d.Screen.ShowCursor(cur.x, cur.y)
		d.Screen.Show()
		ev := d.Screen.PollEvent()
		switch {
		case d.Mode == Normal:
			d.runNormalMode(ev)
		case d.Mode == Insert:
			d.runInsertMode(ev)
		case d.Mode == New:
			d.runNewMode(ev)
		case d.Mode == Delete:
			d.runDeleteMode(ev)
		}
	}

}

func (d *Display) runDeleteMode(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch {
		case ev.Rune() == 'l':
			d.deleteLine()
			d.Mode = Normal
		}
	}
}
func (d *Display) deleteLine() {
	line := d.BufWindow.currLine()
	d.clearCurrLine()
	line.runes = []rune{}
	if cur.y == d.BufWindow.length()-1 {
		d.deleteLastLine()
		return
	}
	d.shiftLinesUp()
	d.reRenderLinesToEOF()
	cur.x = leftMarginSize
}

func (d *Display) deleteLastLine() {
	content := d.BufWindow.content
	d.clearCurrLine()
	if cur.y == 0 {
		content.lines[0].runes = []rune{}
		return
	}
	content.lines = content.lines[:d.BufWindow.length()-1]
	cur.x = leftMarginSize
	cur.y--
}

func (d *Display) runNewMode(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch {
		case ev.Rune() == 'l':
			d.insertBlankLine()
		}
	}
	d.Mode = Insert
}

func (d *Display) insertBlankLine() {
	line := newLine()
	d.clearLinesToEOF()
	d.BufWindow.content.insertNewLine(line)
	d.reRenderLinesToEOF()
	d.setLineNumbers()
	cur.x = leftMarginSize
	cur.y++
}

func (b *Buffer) writeToFile() {
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

func nextTabStopOffsetFromIndex(idx int) int {
	return 8 - (idx % 8)
}

func (d *Display) runNormalMode(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventResize:
		d.Screen.Sync()
	case *tcell.EventKey:
		switch ev.Rune() {
		case 'Q':
			d.Mode = Exit
		case 'W':
			d.Mode = Write
		case 'j':
			d.moveCursorDown()
		case 'k':
			d.moveCursorUp()
		case 'l':
			d.moveCursorRight()
		case 'h':
			d.moveCursorLeft()
		case 'w':
			d.moveCursorToNextWord(false)
		case 'b':
			d.moveCursorToPrevWord()
		case 'I':
			d.Mode = Insert
		case 'n':
			d.Mode = New
		case 'd':
			d.Mode = Delete
		}
	}
}

var nonSpaceSeparators = map[rune]bool{
	'.':  true,
	':':  true,
	';':  true,
	',':  true,
	'"':  true,
	'{':  true,
	'}':  true,
	'(':  true,
	')':  true,
	'[':  true,
	']':  true,
	'\'': true,
}

func (d *Display) moveCursorToPrevWord() {
	line := d.BufWindow.currLine()
	if bufPos.x == 0 && cur.y == 0 {
		return
	}
	if bufPos.x > 0 {
		if pos, ok := line.prevWordPos(); ok {
			cur.x = pos
			return
		}
	}
	cur.y--
	line = d.BufWindow.currLine()
	if line.length() > 0 {
		cur.x = leftMarginSize + line.length() - 1
		return
	}
	cur.x = leftMarginSize
}

func (l *Line) prevWordPos() (int, bool) {
	if l.prevRuneIsPrevWord() {
		return bufPos.x - 1 + leftMarginSize, true
	}
	for i := bufPos.x - 1; i > 0; i-- {
		curr, next := l.runes[i], l.runes[i-1]
		switch {
		case curr == ' ' || curr == '\t' || isApostrophe(l.runes, i):
			continue
		case l.prevWordFound(curr, next):
			return i + leftMarginSize, true
		}
	}
	if isLetterOrNumber(l.runes[0]) || isNonSpaceSeparator(l.runes[0]) {
		return leftMarginSize, true
	}
	return -1, false
}

func (l *Line) prevRuneIsPrevWord() bool {
	switch {
	case isLetterOrNumber(l.curRune()) && isNonSpaceSeparator(l.prevRune()):
		return true
	case isNonSpaceSeparator(l.curRune()) && isNonSpaceSeparator(l.prevRune()):
		return true
	}
	return false
}

func (l *Line) prevWordFound(char, nextChar rune) bool {
	switch {
	case nextChar == ' ' || nextChar == '\t':
		return true
	case isNonSpaceSeparator(nextChar) && isNonSpaceSeparator(char):
		return true
	case isNonSpaceSeparator(nextChar) && isLetterOrNumber(char):
		return true
	}
	return false
}

func (l *Line) curRune() rune {
	return l.runes[bufPos.x]
}

func (l *Line) prevRune() rune {
	return l.runes[bufPos.x-1]
}

func (d *Display) moveCursorToNextWord(sepFound bool) {
	line := d.BufWindow.currLine()
	if len(line.runes) == 0 {
		cur.y++
		d.moveCursorToNextWord(true)
		return
	}
	if pos, ok := line.nextWordPos(sepFound); ok {
		cur.x = pos
		return
	}
	if d.BufWindow.length() == cur.y+1 {
		return
	}
	cur.x = leftMarginSize
	setBufPos()
	cur.y++
	d.moveCursorToNextWord(true)
}

func (l *Line) nextWordPos(sepFound bool) (int, bool) {
	for i := bufPos.x; i < len(l.runes); i++ {
		r := l.runes[i]
		switch {
		case r == ' ':
			sepFound = true
		case !sepFound && isNonSpaceSeparator(r):
			sepFound = true
		case r == '\t' || isApostrophe(l.runes, i):
			continue
		case sepFound:
			return i + leftMarginSize, true
		}
	}
	return -1, false
}

func isNonSpaceSeparator(r rune) bool {
	_, result := nonSpaceSeparators[r]
	return result
}

func isLetterOrNumber(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
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

func (l *Line) nextTabStopFromIndex(x int) int {
	return x + 8 - (x % 8)
}

func (d *Display) moveCursorDown() {
	if d.cursor75PercentDown() && d.canScrollDown() {
		d.scrollDown()
		return
	}
	if cur.y == d.BufWindow.length()-1 {
		return
	}
	nextLineLength := d.BufWindow.getLine(cur.y+1).length() + leftMarginSize
	if nextLineLength < cur.x {
		cur.x = nextLineLength
	}
	cur.y++
}

func (d *Display) cursor75PercentDown() bool {
	return cur.y > (d.height-1)*3/4
}

func (d *Display) scrollDown() {
	start := d.BufWindow.startIdx + 1
	d.clearBufWindow()
	d.setBufWindow(start)
	d.displayBufWindow()
}

func (d *Display) canScrollDown() bool {
	return d.BufWindow.lastLine() != d.ActiveBuf.lastLine()
}

func (b *Buffer) lastLine() *Line {
	return b.content.lines[b.length()-1]
}

func (d *Display) clearBufWindow() {
	for i := range d.BufWindow.length() {
		d.clearLineByIndex(i)
	}
}

func (d *Display) moveCursorUp() {
	if cur.y == 0 {
		return
	}
	if d.cursor25PercentUp() && d.canScrollUp() {
		d.scrollUp()
		return
	}
	prevLineLength := d.BufWindow.getLine(cur.y-1).length() + leftMarginSize
	if prevLineLength < cur.x {
		cur.x = prevLineLength
	}
	cur.y--
}

func (d *Display) cursor25PercentUp() bool {
	return cur.y < (d.height-1)/4
}

func (d *Display) canScrollUp() bool {
	return d.BufWindow.content.lines[0] != d.ActiveBuf.content.lines[0]
}

func (d *Display) scrollUp() {
	start := d.BufWindow.startIdx - 1
	d.clearBufWindow()
	d.setBufWindow(start)
	d.displayBufWindow()
}

func (d *Display) moveCursorRight() {
	if cur.x == d.BufWindow.currLine().length()+leftMarginSize {
		return
	}
	cur.x++
}

func (d *Display) moveCursorLeft() {
	if cur.x == leftMarginSize {
		return
	}
	cur.x--
}

func (d *Display) runInsertMode(ev tcell.Event) {
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
		default:
			d.setRune(ev.Rune())
		}
	}
}

func (d *Display) handleKeyEnter() {
	d.shiftLinesDown()
	cur.x = leftMarginSize
	cur.y++
}

func (d *Display) shiftLinesDown() {
	content := d.BufWindow.content
	d.clearLinesToEOF()
	newLine := content.newLineFromKeyEnter()
	content.insertNewLine(newLine)
	d.reRenderLinesToEOF()
	d.setLineNumbers()
}

func (d *Display) clearLinesToEOF() {
	for i := cur.y; i < d.BufWindow.length(); i++ {
		d.clearLineByIndex(i)
	}
}

func (d *Display) clearLineByIndex(idx int) {
	buf := d.BufWindow
	line := buf.content.lines[idx]
	displayLineLength := line.length() + leftMarginSize
	for i := leftMarginSize; i < displayLineLength; i++ {
		d.Screen.SetContent(i, idx, ' ', nil, d.BufStyle)
	}
}

func (la *LineArray) newLineFromKeyEnter() *Line {
	newLine := newLine()
	newLine.runes = la.lines[cur.y].extractRestOfLine()
	return newLine
}

func (l *Line) extractRestOfLine() []rune {
	pushedRunes := l.runes[bufPos.x:]
	l.runes = l.runes[:bufPos.x]
	return pushedRunes
}

func (la *LineArray) insertNewLine(line *Line) {
	if cur.y == len(la.lines)-1 {
		la.lines = append(la.lines, line)
		return
	}
	la.lines = append(la.lines, nil)
	copy(la.lines[cur.y+2:], la.lines[cur.y+1:])
	la.lines[cur.y+1] = line
}

func (d *Display) reRenderLinesToEOF() {
	for i := cur.y; i < d.BufWindow.length(); i++ {
		d.reRenderLine(i)
	}
}

func (d *Display) reRenderLine(idx int) {
	line := d.BufWindow.getLine(idx)
	for i, r := range line.runes {
		x := i + leftMarginSize
		d.Screen.SetContent(x, idx, r, nil, d.BufStyle)
	}
}
func (d *Display) handleKeyBackspace() {
	switch {
	case bufPos.x == 0 && bufPos.y == 0:
		return
	case bufPos.x == 0:
		d.backspaceToPrevLine()
	default:
		cur.x--
		setBufPos()
		d.backspaceChar()
	}
}

func (d *Display) backspaceToPrevLine() {
	cur.y--
	prevLineLength := d.shiftLinesUp()
	d.reRenderLinesToEOF()
	d.setLineNumbers()
	cur.x = prevLineLength + leftMarginSize
}

func (d *Display) shiftLinesUp() int {
	content := d.BufWindow.content
	d.clearLinesToEOF()
	ogLineLength := d.BufWindow.getLine(cur.y).length()
	content.lines[cur.y].runes = append(content.lines[cur.y].runes, content.lines[cur.y+1].runes...)
	content.lines = append(content.lines[:cur.y+1], content.lines[cur.y+2:]...)
	return ogLineLength
}

func (d *Display) backspaceChar() {
	d.clearCurrLine()
	d.BufWindow.removeRune()
	d.reRenderLine(cur.y)
}

func (d *Display) clearCurrLine() {
	d.clearLineByIndex(cur.y)
}

func (b *Buffer) removeRune() {
	line := b.getLine(cur.y)
	r := line.runes[bufPos.x]
	if r == '\t' {
		line.removeTabRunes()
		return
	}
	line.runes = append(line.runes[:bufPos.x], line.runes[bufPos.x+1:]...)
}

func (l *Line) removeTabRunes() {
	idx := bufPos.x - 1
	if l.runes[idx] != ' ' && l.runes[idx] != '\t' {
		l.runes = append(l.runes[:bufPos.x], l.runes[bufPos.x+1:]...)
		return
	}
	for {
		if l.runes[idx] == '\t' {
			break
		}
		idx--
	}
	l.runes = append(l.runes[:idx], l.runes[bufPos.x+1:]...)
	cur.x = idx + leftMarginSize
}

func (d *Display) handleKeyTab() {
	d.BufWindow.addKeyTab()
	d.clearCurrLine()
	d.reRenderLine(cur.y)
	cur.x += nextTabStopOffset()
}

func (b *Buffer) addKeyTab() {
	tab := createTabRunes()
	line := b.getLine(cur.y)

	for i := 0; i < len(tab); i++ {
		line.runes = append(line.runes, ' ')
	}
	copy(line.runes[bufPos.x+len(tab):], line.runes[bufPos.x:])
	for i := range tab {
		line.runes[bufPos.x+i] = tab[i]
	}
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

func (d *Display) setRune(r rune) {
	d.BufWindow.addRune(r)
	d.reRenderLineAtCursor()
	cur.x++
}

func (b *Buffer) addRune(r rune) {
	line := b.getLine(cur.y)
	line.runes = append(line.runes, ' ')
	copy(line.runes[bufPos.x+1:], line.runes[bufPos.x:])
	line.runes[bufPos.x] = r
}

func (d *Display) reRenderLineAtCursor() {
	line := d.BufWindow.getLine(cur.y)
	d.clearLineFromCursor(line)
	d.resetContentFromCursor(line)
}

func (d *Display) clearLineFromCursor(line *Line) {
	for i := cur.x; i < line.length(); i++ {
		d.Screen.SetContent(i, cur.y, ' ', nil, d.BufStyle)
	}
}

func (d *Display) resetContentFromCursor(line *Line) {
	displayLineLength := leftMarginSize + line.length()
	for i := cur.x; i < displayLineLength; i++ {
		r := line.runes[cur.x-leftMarginSize]
		d.Screen.SetContent(i, cur.y, r, nil, d.BufStyle)
	}
}

func (b *Buffer) getLine(idx int) *Line {
	return b.content.lines[idx]
}

func (b *Buffer) currLine() *Line {
	return b.getLine(cur.y)
}

func (d *Display) setLineNumbers() {
	d.clearLineNumbers()
	start := d.BufWindow.startIdx
	for i := 0; i < d.height-1; i++ {
		lineNum := i + start + 1
		digits := splitDigits(lineNum)
	inner:
		for j, digit := range digits {
			switch {
			case lineNum < 10 && j < 4:
				continue inner
			case lineNum < 100 && j < 3:
				continue inner
			case lineNum < 1000 && j < 2:
				continue inner
			case lineNum < 10000 && j < 1:
				continue inner
			}
			d.Screen.SetContent(j, i, 48+digit, nil, d.LineNoStyle)
		}
	}
}

func (d *Display) clearLineNumbers() {
	for i := range d.height - 2 {
		for j := range leftMarginSize {
			d.Screen.SetContent(j, i, ' ', nil, d.LineNoStyle)
		}
	}
}

func (d *Display) setStatusBar() {
	d.clearStatusBar()
	line := d.BufWindow.currLine()
	currLineNo := d.BufWindow.startIdx + cur.y + 1
	lineCount := d.ActiveBuf.length()
	char := ""
	if bufPos.x < len(line.runes) {
		char = string(line.runes[bufPos.x])
	}
	status := []rune(fmt.Sprintf("%s Mode\t\t\tLine: %d\t\tCol: %d\t\tLineCount: %d\t\tChar: %s",
		modes[d.Mode],
		currLineNo,
		bufPos.x+1,
		lineCount,
		char,
	))
	for i, r := range status {
		d.Screen.SetContent(i, d.height-1, r, nil, d.StatusBarStyle)
	}
	d.StatusBar = status
}

func (d *Display) clearStatusBar() {
	for i := range d.StatusBar {
		d.Screen.SetContent(i, d.height-1, ' ', nil, d.LineNoStyle)
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

func splitDigits(count int) []rune {
	divisor := 10000
	digits := []rune{rune(count / divisor)}
	divisor /= 10
	for i := divisor; i > 0; i /= 10 {
		digit := (count / i) % 10
		digits = append(digits, rune(digit))
	}
	return digits
}

type Buffer struct {
	content  *LineArray
	path     string
	startIdx int
}

func newBuffer() *Buffer {
	return &Buffer{
		content: newLineArray(),
	}
}

func (b *Buffer) length() int {
	return len(b.content.lines)
}

func setBufPos() {
	bufPos.x = cur.x - leftMarginSize
	bufPos.y = cur.y
}

type LineArray struct {
	lines []*Line
}

func newLineArray() *LineArray {
	arr := &LineArray{}
	line := newLine()
	arr.lines = append(arr.lines, line)
	return arr
}

type Line struct {
	runes []rune
}

func newLine() *Line {
	line := &Line{runes: []rune{}}
	return line
}

func (l *Line) length() int {
	return len(l.runes)
}
