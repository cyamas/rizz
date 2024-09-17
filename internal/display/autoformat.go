package display

func (l *Line) autoIndent(tabs int) {
	ogX := bufPos.X
	bufPos.X = 0
	for i := 0; i < tabs; i++ {
		l.addKeyTab()
	}
	bufPos.X = ogX
}

func (l *Line) autoClose(r rune) {
	switch r {
	case '{':
		l.addRune('}')
	case '[':
		l.addRune(']')
	case '(':
		l.addRune(')')
	case '"':
		l.addRune('"')
	case '\'':
		l.addRune('\'')
	}
}

func (l *Line) firstWordIndex() int {
	for i, r := range l.runes {
		if r != '\t' && r != ' ' {
			return i
		}
	}
	return len(l.runes)
}

func (l *Line) tabCountForNewLine() int {
	count := 0
	if l.length() == 0 {
		return 0
	}
	lastRune := l.runes[l.length()-1]
	if lastRune == ':' || isOpenBracket(lastRune) {
		count++
	}
	for i, r := range l.runes {
		if r != '\t' && r != ' ' {
			count += i / LeftMarginSize
			break
		}

	}
	return count
}

func isOpenBracket(r rune) bool {
	return r == '(' || r == '[' || r == '{'
}
