package display

import (
	"github.com/cyamas/rizz/internal/highlighter"
	"github.com/cyamas/rizz/internal/highlighter/lexer"
)

type BufWindow struct {
	buf         *Buffer
	lines       []*Line
	bufIdx      int
	size        int
	highlighter *highlighter.Highlighter
}

func newBufWindow(size int) *BufWindow {
	return &BufWindow{
		lines:       []*Line{},
		highlighter: highlighter.New(lexer.New()),
		size:        size,
	}
}

func (bw *BufWindow) length() int {
	return len(bw.lines)
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

func (bw *BufWindow) line(idx int) *Line {
	return bw.lines[idx]
}

func (bw *BufWindow) lastLine() *Line {
	return bw.lines[bw.size-1]
}
