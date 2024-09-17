package display

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

const LeftMarginSize = 8

var modes = map[int]string{
	Normal: "Normal",
	Insert: "Insert",
	Exit:   "Exit",
	Open:   "Open",
	Write:  "Write",
	Delete: "Delete",
	New:    "New",
}

var styles = map[string]tcell.Style{
	"(package_clause)":       tcell.StyleDefault.Foreground(tcell.ColorLimeGreen).Background(tcell.ColorBlack),
	"(import_declaration)":   tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorBlack),
	"(function_declaration)": tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlack),
}

type cell struct {
	X, Y int
}

var Cur cell
var bufPos cell

func (b *Buffer) ReadFile(filename string) {
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

func (d *Display) SetBufWindow() {
	window := d.bufWindow
	Cur.X = LeftMarginSize
	for y, line := range window.lines {
		for j, r := range line.runes {
			x := Cur.X + j
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

type BufWindow struct {
	buf    *Buffer
	lines  []*Line
	bufIdx int
	size   int
}

func newBufWindow() *BufWindow {
	return &BufWindow{
		lines: []*Line{},
	}
}

func (bw *BufWindow) update(idx int) {
	switch {
	case bw.buf.length() < bw.size:
		bw.bufIdx = 0
		bw.buf.windowStart = 0
		bw.lines = append([]*Line(nil), bw.buf.content.lines...)
	case idx > bw.buf.length()-bw.size:
		bw.bufIdx = bw.buf.length() - bw.size
		bw.buf.windowStart = bw.bufIdx
		bw.lines = append([]*Line(nil), bw.buf.content.lines[bw.bufIdx:bw.bufIdx+bw.size]...)
	default:
		bw.bufIdx = idx
		bw.buf.windowStart = idx
		bw.lines = append([]*Line(nil), bw.buf.content.lines[bw.bufIdx:bw.bufIdx+bw.size]...)
	}
}

func (bw *BufWindow) resetLines() {
	start := bw.bufIdx
	end := bw.bufIdx + bw.length()
	bw.lines = append([]*Line(nil), bw.buf.content.lines[start:end]...)
}

func (d *Display) windowAtBottom() bool {
	return d.bufWindow.lastLine() == d.ActiveBuf.lastLine()
}

func (bw *BufWindow) line(idx int) *Line {
	return bw.lines[idx]
}

func (bw *BufWindow) lastLine() *Line {
	return bw.lines[bw.size-1]
}

func (d *Display) InitBufWindow() {
	bw := newBufWindow()
	d.bufWindow = bw
	bw.buf = d.ActiveBuf
	bw.size = d.height - 1
	if bw.buf.length() == 1 && bw.buf.getLine(0).length() == 0 {
		return
	}
	bw.lines = append([]*Line(nil), d.ActiveBuf.content.lines[:bw.size]...)
}

type Display struct {
	Screen         tcell.Screen
	width          int
	height         int
	bufWindow      *BufWindow
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
	statusBarStyle := tcell.StyleDefault.Foreground(tcell.ColorWhiteSmoke).Background(tcell.ColorDarkSlateGray)
	return &Display{
		BufStyle:       bufStyle,
		LineNoStyle:    lineNoStyle,
		StatusBarStyle: statusBarStyle,
	}
}

func (d *Display) Init() {
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

func (d *Display) Run() {
	for {
		if d.Mode == Exit {
			return
		}
		if d.Mode == Write {
			d.ActiveBuf.writeToFile()
			d.Mode = Normal
		}
		d.setBufPos()
		d.setStatusBar()
		d.setLineNumbers()
		d.Screen.ShowCursor(Cur.X, Cur.Y)
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

func (d *Display) currLine() *Line {
	return d.ActiveBuf.currLine()
}
func (d *Display) deleteLine() {
	d.clearBufWindow()
	line := d.ActiveBuf.currLine()
	d.clearCurrLine()
	line.runes = []rune{}
	if Cur.Y == d.bufWindow.length()-1 {
		d.deleteLastLine()
		d.bufWindow.update(d.bufWindow.bufIdx)
		return
	}
	if d.bufWindow.bufIdx == d.ActiveBuf.length()-d.bufWindow.size {
		d.SetBufWindow()
		return
	}
	d.shiftLinesUp()
	d.bufWindow.update(d.bufWindow.bufIdx)
	d.SetBufWindow()
	Cur.X = LeftMarginSize
}

func (bw *BufWindow) length() int {
	return len(bw.lines)
}

func (d *Display) deleteLastLine() {
	content := d.ActiveBuf.content
	if Cur.Y == 0 {
		return
	}
	content.lines = content.lines[:len(content.lines)-1]
	if len(content.lines) > d.bufWindow.size {
		d.scrollUp()
		return
	}
	Cur.X = LeftMarginSize
	Cur.Y--
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
	tabCount := d.ActiveBuf.currLine().tabCountForNewLine()
	line.autoIndent(tabCount)
	d.clearLinesToEOW()
	d.ActiveBuf.content.insertNewLine(line)
	if d.cursor75PercentDown() {
		d.scrollDown()
		d.setLineNumbers()
		Cur.X = LeftMarginSize + line.length()
		return
	}
	d.bufWindow.update(d.bufWindow.bufIdx)
	d.reRenderLinesToEOF()
	d.setLineNumbers()
	Cur.X = LeftMarginSize + line.length()
	Cur.Y++
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
		case 'J':
			d.moveCursorHalfWindowDown()
		case 'k':
			d.moveCursorUp()
		case 'K':
			d.moveCursorHalfWindowUp()
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

func (d *Display) moveCursorHalfWindowDown() {
	endBuf := d.ActiveBuf.length() - 1
	if bufPos.Y >= endBuf-(d.bufWindow.size/2) {
		for i := bufPos.Y; i <= endBuf; i++ {
			d.moveCursorDown()
		}
		return
	}
	for i := 0; i < d.bufWindow.size/2; i++ {
		d.moveCursorDown()
	}
}

func (d *Display) moveCursorHalfWindowUp() {
	if bufPos.Y <= d.bufWindow.size/2 {
		for i := bufPos.Y; i >= 0; i-- {
			d.moveCursorUp()
		}
		return
	}
	for i := 0; i < d.bufWindow.size/2; i++ {
		d.moveCursorUp()
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
	line := d.ActiveBuf.currLine()
	if bufPos.X == 0 && Cur.Y == 0 {
		return
	}
	if bufPos.X > 0 {
		if pos, ok := line.prevWordPos(); ok {
			Cur.X = pos
			return
		}
	}
	if d.canScrollUp() {
		d.scrollUp()
		line = d.bufWindow.line(Cur.Y)
		if line.length() > 0 {
			Cur.X = LeftMarginSize + line.length() - 1
			return
		}
		Cur.X = LeftMarginSize
		return
	}
	Cur.Y--
	d.setBufPos()
	line = d.ActiveBuf.currLine()
	if line.length() > 0 {
		Cur.X = LeftMarginSize + line.length() - 1
		return
	}
	Cur.X = LeftMarginSize
}

func (l *Line) prevWordPos() (int, bool) {
	if l.prevRuneIsPrevWord() {
		return bufPos.X - 1 + LeftMarginSize, true
	}
	for i := bufPos.X - 1; i > 0; i-- {
		curr, next := l.runes[i], l.runes[i-1]
		switch {
		case curr == ' ' || curr == '\t' || isApostrophe(l.runes, i):
			continue
		case l.prevWordFound(curr, next):
			return i + LeftMarginSize, true
		}
	}
	if isLetterOrNumber(l.runes[0]) || isNonSpaceSeparator(l.runes[0]) {
		return LeftMarginSize, true
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
	return l.runes[bufPos.X]
}

func (l *Line) prevRune() rune {
	return l.runes[bufPos.X-1]
}

func (d *Display) moveCursorToNextWord(sepFound bool) {
	line := d.ActiveBuf.currLine()
	if len(line.runes) == 0 {
		Cur.Y++
		d.setBufPos()
		d.moveCursorToNextWord(true)
		return
	}
	if pos, ok := line.nextWordPos(sepFound); ok {
		Cur.X = pos
		return
	}
	if d.ActiveBuf.length() == bufPos.Y {
		return
	}
	Cur.X = LeftMarginSize
	d.setBufPos()
	if d.canScrollDown() {
		d.scrollDown()
		d.moveCursorToNextWord(true)
		return
	}
	Cur.Y++
	d.setBufPos()
	sepFound = true
	d.moveCursorToNextWord(sepFound)
}

func (l *Line) nextWordPos(sepFound bool) (int, bool) {
	if sepFound && l.runes[bufPos.X] != ' ' && l.runes[bufPos.X] != '\t' {
		return bufPos.X + LeftMarginSize, true
	}
	for i := bufPos.X + 1; i < len(l.runes); i++ {
		curr := l.runes[i]
		prev := l.runes[i-1]
		switch {
		case curr == ' ' || curr == '\t' || isApostrophe(l.runes, i):
			continue
		case isLetterOrNumber(curr) && isLetterOrNumber(prev):
			continue
		case isLetterOrNumber(prev) && isNonSpaceSeparator(curr):
			return i + LeftMarginSize, true
		default:
			return i + LeftMarginSize, true
			/*
				case prev == ' ' || prev == '\t':
					return i + LeftMarginSize, true
				case isNonSpaceSeparator(curr) && isLetterOrNumber(prev):
					return i + LeftMarginSize, true
				case isNonSpaceSeparator(prev):
					return i + LeftMarginSize, true
			*/
		}
	}
	return -1, false
}

func (l *Line) nnextWordPos(sepFound bool) (int, bool) {
	for i := bufPos.X; i < len(l.runes); i++ {
		r := l.runes[i]
		switch {
		case r == ' ':
			sepFound = true
		case !sepFound && isNonSpaceSeparator(r):
			sepFound = true
		case r == '\t' || isApostrophe(l.runes, i):
			continue
		case sepFound:
			return i + LeftMarginSize, true
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
	if d.canScrollDown() {
		d.scrollDown()
		return
	}
	if Cur.Y == d.bufWindow.size-1 {
		return
	}
	nextLineLength := d.bufWindow.line(Cur.Y+1).length() + LeftMarginSize
	if nextLineLength < Cur.X {
		Cur.X = nextLineLength
	}
	Cur.Y++
}

func (d *Display) cursor75PercentDown() bool {
	return Cur.Y > (d.height-1)*3/4
}

func (d *Display) scrollDown() {
	d.clearBufWindow()
	d.bufWindow.update(d.bufWindow.bufIdx + 1)
	d.SetBufWindow()
}

func (d *Display) canScrollDown() bool {
	return d.bufWindow.lastLine() != d.ActiveBuf.lastLine() && d.cursor75PercentDown()
}

func (b *Buffer) lastLine() *Line {
	return b.content.lines[b.length()-1]
}

func (d *Display) clearBufWindow() {
	for i := range len(d.bufWindow.lines) {
		d.clearLineByIndex(i)
	}
}

func (d *Display) moveCursorUp() {
	if Cur.Y == 0 {
		return
	}
	if d.canScrollUp() {
		d.scrollUp()
		return
	}
	prevLineLength := d.bufWindow.line(Cur.Y-1).length() + LeftMarginSize
	if prevLineLength < Cur.X {
		Cur.X = prevLineLength
	}
	Cur.Y--
}

func (d *Display) cursor25PercentUp() bool {
	return Cur.Y < (d.height-1)/4
}

func (d *Display) canScrollUp() bool {
	return d.bufWindow.lines[0] != d.ActiveBuf.content.lines[0] && d.cursor25PercentUp()
}

func (d *Display) scrollUp() {
	d.clearBufWindow()
	d.bufWindow.update(d.bufWindow.bufIdx - 1)
	d.SetBufWindow()
}

func (d *Display) moveCursorRight() {
	if Cur.X == d.ActiveBuf.currLine().length()+LeftMarginSize {
		return
	}
	Cur.X++
}

func (d *Display) moveCursorLeft() {
	if Cur.X == LeftMarginSize {
		return
	}
	Cur.X--
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
	d.clearBufWindow()
	buf := d.ActiveBuf
	content := buf.content
	if buf.currLine().length() > 0 && isClosingRune(buf.currLine().curRune()) {
		buf.addClosingRuneLine()
		if d.cursor75PercentDown() {
			d.scrollDown()
			Cur.Y--
		}
	}
	newLine := content.newLineFromKeyEnter()
	content.insertNewLine(newLine)
	if d.cursor75PercentDown() {
		d.scrollDown()
		Cur.X = LeftMarginSize + newLine.firstWordIndex()
		return
	}
	d.bufWindow.update(d.bufWindow.bufIdx)
	d.SetBufWindow()
	Cur.X = LeftMarginSize + newLine.firstWordIndex()
	Cur.Y++
}

func (b *Buffer) addClosingRuneLine() {
	tabCount := b.currLine().tabCountForNewLine()
	closingRunes := b.currLine().extractRestOfLine()
	newLine := newLine()
	newLine.autoIndent(tabCount)
	newLine.runes = append(newLine.runes, closingRunes...)
	b.content.insertNewLine(newLine)
}

func (d *Display) shiftLinesDown() {
	content := d.ActiveBuf.content
	d.clearLinesToEOW()
	newLine := content.newLineFromKeyEnter()
	content.insertNewLine(newLine)
	d.reRenderLinesToEOF()
	d.setLineNumbers()
	d.bufWindow.resetLines()
}

func (d *Display) clearLinesToEOW() {
	for i := Cur.Y; i < len(d.bufWindow.lines); i++ {
		d.clearLineByIndex(i)
	}
}

func (d *Display) clearLineByIndex(idx int) {
	line := d.bufWindow.lines[idx]
	displayLineLength := line.length() + LeftMarginSize
	for i := LeftMarginSize; i <= displayLineLength; i++ {
		d.Screen.SetContent(i, idx, ' ', nil, d.BufStyle)
	}
}

func (la *LineArray) newLineFromKeyEnter() *Line {
	newLine := newLine()
	runes := la.lines[bufPos.Y].extractRestOfLine()
	tabCount := la.lines[bufPos.Y].tabCountForNewLine()
	newLine.autoIndent(tabCount)
	newLine.runes = append(newLine.runes, runes...)
	return newLine
}

func (l *Line) extractRestOfLine() []rune {
	pushedRunes := append([]rune(nil), l.runes[bufPos.X:]...)
	l.runes = l.runes[:bufPos.X]
	return pushedRunes
}

func (la *LineArray) insertNewLine(line *Line) {
	if bufPos.Y == len(la.lines)-1 {
		la.lines = append(la.lines, line)
		return
	}
	la.lines = append(la.lines, nil)
	copy(la.lines[bufPos.Y+2:], la.lines[bufPos.Y+1:])
	la.lines[bufPos.Y+1] = line
}

func (d *Display) reRenderLinesToEOF() {
	for i := Cur.Y; i < d.bufWindow.length(); i++ {
		d.reRenderLine(i)
	}
}

func (d *Display) reRenderLine(y int) {
	line := d.bufWindow.line(y)
	for i, r := range line.runes {
		x := i + LeftMarginSize
		d.Screen.SetContent(x, y, r, nil, d.BufStyle)
	}
}
func (d *Display) handleKeyBackspace() {
	switch {
	case bufPos.X == 0 && bufPos.Y == 0:
		return
	case bufPos.X == 0:
		d.backspaceToPrevLine()
	default:
		Cur.X--
		d.setBufPos()
		d.backspaceChar()
	}
}

func (d *Display) backspaceToPrevLine() {
	d.clearBufWindow()
	prevLineLength := d.shiftLinesUp()
	d.bufWindow.update(d.bufWindow.bufIdx)
	d.SetBufWindow()
	d.setLineNumbers()
	Cur.X = prevLineLength + LeftMarginSize
	if d.windowAtBottom() {
		return
	}
	if d.canScrollUp() {
		d.scrollUp()
		return
	}
	Cur.Y--
}

func (d *Display) shiftLinesUp() int {
	Cur.Y--
	d.setBufPos()
	buf := d.ActiveBuf
	lines := buf.content.lines
	line := d.ActiveBuf.currLine()
	idx := bufPos.Y
	ogLineLength := line.length()
	line.runes = append(line.runes, lines[idx+1].runes...)
	buf.content.lines = append(lines[:idx+1], lines[idx+2:]...)
	Cur.Y++
	return ogLineLength
}

func (d *Display) backspaceChar() {
	d.clearCurrLine()
	d.ActiveBuf.removeRune()
	d.reRenderLine(Cur.Y)
}

func (d *Display) clearCurrLine() {
	d.clearLineByIndex(Cur.Y)
}

func (b *Buffer) removeRune() {
	line := b.getLine(bufPos.Y)
	r := line.runes[bufPos.X]
	switch {
	case r == '\t':
		line.removeTabRunes()
	case bufPos.X < line.length()-1 && isAutoClosable(r) && isClosingRune(line.runes[bufPos.X+1]):
		line.runes = append(line.runes[:bufPos.X], line.runes[bufPos.X+2:]...)
	default:
		line.runes = append(line.runes[:bufPos.X], line.runes[bufPos.X+1:]...)
	}
}

func (l *Line) removeTabRunes() {
	idx := bufPos.X - 1
	if l.runes[idx] != ' ' && l.runes[idx] != '\t' {
		l.runes = append(l.runes[:bufPos.X], l.runes[bufPos.X+1:]...)
		return
	}
	for {
		if l.runes[idx] == '\t' {
			break
		}
		idx--
	}
	l.runes = append(l.runes[:idx], l.runes[bufPos.X+1:]...)
	Cur.X = idx + LeftMarginSize
}

func (d *Display) handleKeyTab() {
	d.ActiveBuf.currLine().addKeyTab()
	d.clearCurrLine()
	d.reRenderLine(Cur.Y)
	Cur.X += nextTabStopOffset()
}

func (l *Line) addKeyTab() {
	tab := createTabRunes()

	for i := 0; i < len(tab); i++ {
		l.runes = append(l.runes, ' ')
	}
	copy(l.runes[bufPos.X+len(tab):], l.runes[bufPos.X:])
	for i := range tab {
		l.runes[bufPos.X+i] = tab[i]
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
	return 8 - (bufPos.X % 8)
}

func (d *Display) setRune(r rune) {
	if isClosingRune(r) && r == d.currRune() {
		Cur.X++
		return
	}

	d.clearCurrLine()
	d.currLine().addRune(r)
	d.reRenderLine(Cur.Y)
	Cur.X++
	if isAutoClosable(r) {
		d.setBufPos()
		d.clearCurrLine()
		d.currLine().autoClose(r)
		d.reRenderLine(Cur.Y)
	}
}

func (d *Display) currRune() rune {
	return d.ActiveBuf.currLine().curRune()
}

func isClosingRune(r rune) bool {
	return r == '}' || r == ']' || r == '"' || r == ')'
}

func isAutoClosable(r rune) bool {
	return r == '{' || r == '[' || r == '"' || r == '('
}

func (l *Line) addRune(r rune) {
	l.runes = append(l.runes, ' ')
	copy(l.runes[bufPos.X+1:], l.runes[bufPos.X:])
	l.runes[bufPos.X] = r

}

func (b *Buffer) addRune(r rune) {
	line := b.getLine(bufPos.Y)
	line.runes = append(line.runes, ' ')
	copy(line.runes[bufPos.X+1:], line.runes[bufPos.X:])
	line.runes[bufPos.X] = r
}

func (d *Display) reRenderLineAtCursor() {
	line := d.bufWindow.line(Cur.Y)
	d.clearLineFromCursor(line)
	d.resetContentFromCursor(line)
}

func (d *Display) clearLineFromCursor(line *Line) {
	for i := Cur.X; i < line.length(); i++ {
		d.Screen.SetContent(i, Cur.Y, ' ', nil, d.BufStyle)
	}
}

func (d *Display) resetContentFromCursor(line *Line) {
	displayLineLength := LeftMarginSize + line.length()
	for i := Cur.X; i < displayLineLength; i++ {
		r := line.runes[Cur.X-LeftMarginSize]
		d.Screen.SetContent(i, Cur.Y, r, nil, d.BufStyle)
	}
}

func (b *Buffer) getLine(idx int) *Line {
	return b.content.lines[idx]
}

func (b *Buffer) currLine() *Line {
	return b.getLine(bufPos.Y)
}

func (d *Display) setLineNumbers() {
	d.clearLineNumbers()
	start := d.bufWindow.bufIdx
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
		for j := range LeftMarginSize {
			d.Screen.SetContent(j, i, ' ', nil, d.LineNoStyle)
		}
	}
}

func (d *Display) setStatusBar() {
	d.clearStatusBar()
	line := d.bufWindow.line(Cur.Y)
	currLineNo := bufPos.Y + 1
	lineCount := d.ActiveBuf.length()
	char := ""
	if bufPos.X < len(line.runes) {
		char = string(line.runes[bufPos.X])
	}
	status := []rune(fmt.Sprintf("%s Mode\t\t\tLine: %d\t\tCol: %d\t\tLineCount: %d\t\tChar: %s",
		modes[d.Mode],
		currLineNo,
		bufPos.X+1,
		lineCount,
		char,
	))
	for i, r := range status {
		d.Screen.SetContent(i, d.height-1, r, nil, d.StatusBarStyle)
	}
	for i := len(status); i < d.width; i++ {
		d.Screen.SetContent(i, d.height-1, ' ', nil, d.StatusBarStyle)
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
	content     *LineArray
	path        string
	windowStart int
}

func NewBuffer() *Buffer {
	return &Buffer{
		content: newLineArray(),
	}
}

func (b *Buffer) length() int {
	return len(b.content.lines)
}

func (b *Buffer) appendLine(line *Line) {
	b.content.lines = append(b.content.lines, line)
}

func (d *Display) setBufPos() {
	bufPos.X = Cur.X - LeftMarginSize
	bufPos.Y = Cur.Y + d.bufWindow.bufIdx
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
	runes      []rune
	highlights map[int]tcell.Style
}

func newLine() *Line {
	line := &Line{runes: []rune{}}
	return line
}

func (l *Line) length() int {
	return len(l.runes)
}

func (l *Line) lastRune() rune {
	return l.runes[l.length()-1]
}

func (l *Line) setHighlights() {

}
