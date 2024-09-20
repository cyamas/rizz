package display

import (
	"github.com/cyamas/rizz/internal/highlighter"
	"github.com/cyamas/rizz/internal/highlighter/token"
)

type LineArray struct {
	lines []*Line
}

func newLineArray(h *highlighter.Highlighter) *LineArray {
	arr := &LineArray{}
	line := newLine(h)
	arr.lines = append(arr.lines, line)
	return arr
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

func (la *LineArray) newLineFromKeyEnter(h *highlighter.Highlighter) *Line {
	newLine := newLine(h)
	runes := la.lines[bufPos.Y].extractRestOfLine()
	tabCount := la.lines[bufPos.Y].tabCountForNewLine()
	newLine.autoIndent(tabCount)
	currContext := la.currLine().Context()
	newLine.highlight(currContext)
	newLine.runes = append(newLine.runes, runes...)
	return newLine
}

func (la *LineArray) currLine() *Line {
	return la.lines[bufPos.Y]
}

func (la *LineArray) addLineFromFile(text string, h *highlighter.Highlighter) {
	line := newLine(h)
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
		line.highlight([]token.TokenType{token.TYPE_NONE})
		return
	}
	currContext := la.currLine().Context()
	line.highlight(currContext)
	la.lines = append(la.lines, line)
}
