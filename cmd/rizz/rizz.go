package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
)

type Buffer struct {
	content *LineArray
}

type LineArray struct {
	lines []*Line
	lock  sync.Mutex
}

func NewLineArray() *LineArray {
	return &LineArray{}
}

func (la *LineArray) AppendLine(line []byte) {
	newLine := &Line{}
	newLine.Bytes = line
	la.lines = append(la.lines, newLine)
}

type Line struct {
	Bytes []byte
	lock  sync.Mutex
}

type cell struct {
	x, y  int
	r     rune
	style tcell.Style
}

var screen tcell.Screen
var lastCursor cell
var style = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)

func main() {
	fmt.Println("rizz will be a simple text editor one day!")
	args := os.Args

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%v", err)
	}

	if err := screen.Init(); err != nil {
		log.Fatalf("%v", err)
	}
	screen.SetStyle(style)
	screen.Clear()
	quit := func() {
		maybePanic := recover()
		screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()
	var bufferFromFile *Buffer
	if len(args) == 2 {
		bufferFromFile = createBufferFromFile(args[1])
	}
	if bufferFromFile.content != nil && len(bufferFromFile.content.lines) != 0 {
		displayBufferContent(bufferFromFile, screen, style)
	}

	for {
		screen.Show()
		screen.ShowCursor(lastCursor.x, lastCursor.y)
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventKey:
			switch {
			case ev.Rune() == 'Q':
				return
			case ev.Rune() == 'O':
				openMode(screen, style)
			case ev.Rune() == 'I':
				insertMode(screen, style)
			}
		}
	}
}

func createBufferFromFile(filename string) *Buffer {
	buf := &Buffer{}
	buf.content = getFileContents(filename)
	return buf
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
	content := &LineArray{}
	for {
		bytes, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		content.AppendLine(bytes)
	}
	return content
}

func displayBufferContent(buffer *Buffer, s tcell.Screen, style tcell.Style) {
	for _, line := range buffer.content.lines {
		str := string(line.Bytes)
		for _, r := range str {
			s.SetContent(lastCursor.x, lastCursor.y, r, []rune{}, style)
			lastCursor.x++
		}
		lastCursor.x = 0
		lastCursor.y++
	}
	lastCursor.x = 0
	lastCursor.y = 0
}

func openMode(s tcell.Screen, style tcell.Style) {

}

func insertMode(s tcell.Screen, style tcell.Style) {
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyCtrlN:
				return
			case ev.Key() == tcell.KeyEnter:
				lastCursor.x = 0
				lastCursor.y++
				s.ShowCursor(lastCursor.x, lastCursor.y)
			case ev.Key() == tcell.KeyBackspace2:
				if lastCursor.x == 0 {
					lastCursor.y--
					break
				}
				s.SetContent(lastCursor.x-1, lastCursor.y, ' ', []rune{}, style)
				lastCursor.x--
				s.ShowCursor(lastCursor.x, lastCursor.y)
			default:
				s.SetContent(lastCursor.x, lastCursor.y, ev.Rune(), []rune{}, style)
				lastCursor.x++

			}
		}
		s.Show()
		s.ShowCursor(lastCursor.x, lastCursor.y)
	}
}
