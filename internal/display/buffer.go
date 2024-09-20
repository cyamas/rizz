package display

import (
	"bufio"
	"log"
	"os"

	"github.com/cyamas/rizz/internal/highlighter"
)

type Buffer struct {
	content     *LineArray
	path        string
	windowStart int
	highlighter *highlighter.Highlighter
}

func NewBuffer(h *highlighter.Highlighter) *Buffer {
	return &Buffer{
		content:     newLineArray(h),
		highlighter: h,
	}
}

func (b *Buffer) getLine(idx int) *Line {
	return b.content.lines[idx]
}

func (b *Buffer) currLine() *Line {
	return b.getLine(bufPos.Y)
}

func (b *Buffer) lastLine() *Line {
	return b.content.lines[b.length()-1]
}

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
		log.Println("could not open file", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	content := newLineArray(b.highlighter)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		content.addLineFromFile(text, b.highlighter)
		Cur.Y++
	}
	return content
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

func (b *Buffer) addRune(r rune) {
	line := b.getLine(bufPos.Y)
	line.runes = append(line.runes, ' ')
	copy(line.runes[bufPos.X+1:], line.runes[bufPos.X:])
	line.runes[bufPos.X] = r
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

func (b *Buffer) length() int {
	return len(b.content.lines)
}

func (b *Buffer) appendLine(line *Line) {
	b.content.lines = append(b.content.lines, line)
}

func (b *Buffer) addClosingRuneLine() {
	tabCount := b.currLine().tabCountForNewLine()
	closingRunes := b.currLine().extractRestOfLine()
	newLine := newLine(b.highlighter)
	newLine.autoIndent(tabCount)
	newLine.runes = append(newLine.runes, closingRunes...)
	currContext := b.content.currLine().Context()
	newLine.highlight(currContext)
	b.content.insertNewLine(newLine)
}
