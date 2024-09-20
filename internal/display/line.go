package display

import (
	"github.com/cyamas/rizz/internal/highlighter"
	"github.com/cyamas/rizz/internal/highlighter/token"
	"github.com/gdamore/tcell/v2"
)

type Line struct {
	runes       []rune
	context     []token.TokenType
	highlighter *highlighter.Highlighter
	styles      []tcell.Style
}

func newLine(h *highlighter.Highlighter) *Line {
	line := &Line{runes: []rune{}, highlighter: h}
	return line
}

func (l *Line) highlight(ctx []token.TokenType) {
	if len(ctx) == 0 {
		ctx = []token.TokenType{token.TYPE_NONE}
	}
	if l.length() == 0 {
		l.context = ctx
		return
	}
	line := l.convertRunesForParsing()
	l.highlighter.ParseLine(line, ctx)
	l.setStyles(l.highlighter.Tokens())
	l.context = l.highlighter.LineContext()

}

func (l *Line) Context() []token.TokenType {
	return l.context
}

var styles = map[string]tcell.Style{
	token.PARAM_NAME:     tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack),
	token.TYPE:           tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.INT:            tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.FLOAT_32:       tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.FLOAT_64:       tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.STRING:         tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.RUNE:           tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.BYTE:           tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.MAP_DECLARE:    tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.ARRAY_DECLARE:  tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.SLICE_DECLARE:  tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.STRUCT:         tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.FUNC_DECLARE:   tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.RANGE:          tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.FOR:            tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.IF:             tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.ELSE:           tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.LEN:            tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.RETURN:         tcell.StyleDefault.Foreground(tcell.ColorViolet).Background(tcell.ColorBlack),
	token.RETURN_TYPE:    tcell.StyleDefault.Foreground(tcell.NewRGBColor(255, 205, 255)).Background(tcell.ColorBlack),
	token.FUNC_NAME:      tcell.StyleDefault.Foreground(tcell.NewRGBColor(0, 255, 255)).Background(tcell.ColorBlack),
	token.FUNC_CALL:      tcell.StyleDefault.Foreground(tcell.ColorTurquoise).Background(tcell.ColorBlack),
	token.DBL_QUOTE:      tcell.StyleDefault.Foreground(tcell.ColorPaleGreen).Background(tcell.ColorBlack),
	token.STRING_LITERAL: tcell.StyleDefault.Foreground(tcell.ColorPaleGreen).Background(tcell.ColorBlack),
	token.IMPORT_NAME:    tcell.StyleDefault.Foreground(tcell.ColorPaleGreen).Background(tcell.ColorBlack),
	token.IMPORT_CALL:    tcell.StyleDefault.Foreground(tcell.ColorPaleTurquoise).Background(tcell.ColorBlack),
	token.TYPE_NAME:      tcell.StyleDefault.Foreground(tcell.NewRGBColor(0, 255, 255)).Background(tcell.ColorBlack),
	token.INT_LITERAL:    tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack),
	token.IMPORT:         tcell.StyleDefault.Foreground(tcell.ColorMediumTurquoise).Background(tcell.ColorBlack),
	token.PACKAGE:        tcell.StyleDefault.Foreground(tcell.ColorMediumTurquoise).Background(tcell.ColorBlack),
	token.IDENT:          tcell.StyleDefault.Foreground(tcell.NewRGBColor(105, 186, 255)).Background(tcell.ColorBlack),
	token.TYPE_NONE:      tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack),
}

func (l *Line) setStyles(tokens []token.Token) {
	l.styles = []tcell.Style(nil)
	if len(tokens) == 0 {
		return
	}
	tokIdx := 0
	currToken := tokens[0]
	defStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	for i := range l.runes {
		if i >= currToken.StartIndex+currToken.Length {
			if tokIdx == len(tokens)-1 {
				l.styles = append(l.styles, defStyle)
				return
			}
			tokIdx++
			currToken = tokens[tokIdx]
		}
		if style, ok := styles[string(currToken.Type)]; ok {
			l.styles = append(l.styles, style)
			continue
		}
		l.styles = append(l.styles, defStyle)
	}
}

func (l *Line) getRuneStyle(idx int) tcell.Style {
	if idx >= len(l.styles) {
		return styles[token.TYPE_NONE]
	}
	return l.styles[idx]
}

func (l *Line) addRune(r rune) {
	l.runes = append(l.runes, ' ')
	copy(l.runes[bufPos.X+1:], l.runes[bufPos.X:])
	l.runes[bufPos.X] = r

}

func (l *Line) addTabFromFile() {
	l.runes = append(l.runes, '\t')
	offset := nextTabStopOffsetFromIndex(len(l.runes) - 1)
	for i := 1; i < offset-1; i++ {
		l.runes = append(l.runes, ' ')
	}
	l.runes = append(l.runes, '\t')
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

func (l *Line) convertRunesForParsing() string {
	str := ""
	for _, r := range l.runes {
		if r == '\t' {
			str += " "
			continue
		}
		str += string(r)
	}
	return str
}

func (l *Line) curRune() rune {
	if bufPos.X >= l.length() {
		return ' '
	}
	return l.runes[bufPos.X]
}

func (l *Line) prevRune() rune {
	return l.runes[bufPos.X-1]
}

func (l *Line) length() int {
	return len(l.runes)
}

func (l *Line) lastRune() rune {
	return l.runes[l.length()-1]
}

func (l *Line) extractRestOfLine() []rune {
	pushedRunes := append([]rune(nil), l.runes[bufPos.X:]...)
	l.runes = l.runes[:bufPos.X]
	return pushedRunes
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

func (l *Line) setHighlights() {

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
		}
	}
	return -1, false
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
