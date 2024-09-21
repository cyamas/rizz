package display

import (
	"fmt"
	"log"
	"unicode"

	"github.com/cyamas/rizz/internal/highlighter"
	"github.com/cyamas/rizz/internal/highlighter/lexer"
	"github.com/cyamas/rizz/internal/highlighter/token"
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

type cell struct {
	X, Y  int
	r     rune
	Style tcell.Style
}

var Cur cell
var bufPos cell

func (d *Display) SetBufWindow() {
	window := d.bufWindow
	Cur.X = LeftMarginSize
	for y, line := range window.lines {
		if line == nil {
			continue
		}
		for j, r := range line.runes {
			x := Cur.X + j
			d.Screen.SetContent(x, y, r, nil, line.getRuneStyle(j))
		}
	}
}

func (d *Display) windowAtBottom() bool {
	return d.bufWindow.lines[Cur.Y] == d.ActiveBuf.lastLine()
}

func (d *Display) InitBufWindow() {
	bw := newBufWindow(d.height - 1)
	d.bufWindow = bw
	bw.buf = d.ActiveBuf
	for i, line := range d.ActiveBuf.content.lines {
		if i >= d.bufWindow.size {
			break
		}
		d.bufWindow.lines[i] = line
	}
}

type Display struct {
	Screen         tcell.Screen
	width          int
	height         int
	bufWindow      *BufWindow
	ActiveBuf      *Buffer
	Highlighter    *highlighter.Highlighter
	Mode           int
	StatusBar      []rune
	BufStyle       tcell.Style
	LineNoStyle    tcell.Style
	StatusBarStyle tcell.Style
}

func NewDisplay() *Display {
	bufStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	lineNoStyle := tcell.StyleDefault.Foreground(tcell.ColorSilver).Background(tcell.ColorBlack)
	statusBarStyle := tcell.StyleDefault.Foreground(tcell.ColorWhiteSmoke).Background(tcell.ColorDarkSlateGray)
	return &Display{
		BufStyle:       bufStyle,
		LineNoStyle:    lineNoStyle,
		StatusBarStyle: statusBarStyle,
		Highlighter:    highlighter.New(lexer.New()),
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
		d.SetBufWindow()
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
	line := newLine(d.Highlighter)
	tabCount := d.ActiveBuf.currLine().tabCountForNewLine()
	line.autoIndent(tabCount)
	currContext := d.currLine().Context()
	line.highlight(currContext)
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

func nextTabStopFromIndex(x int) int {
	return x + 8 - (x % 8)
}

func (d *Display) moveCursorDown() {
	if Cur.Y == d.ActiveBuf.length()-1 {
		return
	}
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
	if d.ActiveBuf.length() < d.bufWindow.size {
		return false
	}
	return d.bufWindow.lastLine() != d.ActiveBuf.lastLine() && d.cursor75PercentDown()
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
	newLine := content.newLineFromKeyEnter(d.Highlighter)
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

func (d *Display) shiftLinesDown() {
	content := d.ActiveBuf.content
	d.clearLinesToEOW()
	newLine := content.newLineFromKeyEnter(d.Highlighter)
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

func (d *Display) reRenderLinesToEOF() {
	for i := Cur.Y; i < d.bufWindow.length(); i++ {
		d.reRenderLine(i)
	}
}

func (d *Display) reRenderLine(y int) {
	line := d.bufWindow.line(y)
	for i, r := range line.runes {
		x := i + LeftMarginSize
		d.Screen.SetContent(x, y, r, nil, line.getRuneStyle(i))
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
		Cur.Y--
		return
	}
	if d.canScrollUp() {
		d.scrollUp()
		return
	}
	Cur.Y--
}

func (d *Display) shiftLinesUp() int {
	if Cur.Y == 0 {
		d.ActiveBuf.content.lines = d.ActiveBuf.content.lines[1:]
		return 0
	}
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

func (d *Display) handleKeyTab() {
	d.ActiveBuf.currLine().addKeyTab()
	d.clearCurrLine()
	d.reRenderLine(Cur.Y)
	Cur.X += nextTabStopOffset()
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
	prevLine, ok := d.prevLine()
	if !ok {
		d.currLine().highlight([]token.TokenType{token.TYPE_NONE})
	} else {
		d.currLine().highlight(prevLine.Context())
	}
	d.reRenderLine(Cur.Y)
	Cur.X++
	if isAutoClosable(r) {
		d.setBufPos()
		d.clearCurrLine()
		d.currLine().autoClose(r)
		if prevLine == nil {
			d.currLine().highlight(d.currLine().Context())
		} else {
			d.currLine().highlight(prevLine.Context())
		}
		d.reRenderLine(Cur.Y)
	}
}

func (d *Display) prevLine() (*Line, bool) {
	if Cur.Y == 0 {
		return nil, false
	}
	return d.ActiveBuf.getLine(Cur.Y - 1), true
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

func (d *Display) setBufPos() {
	bufPos.X = Cur.X - LeftMarginSize
	bufPos.Y = Cur.Y + d.bufWindow.bufIdx
}
