package display

type Action string

const (
	ADD    = "ADD"
	REMOVE = "REMOVE"
	UNDO   = "UNDO"
	REDO   = "REDO"
)

type History struct {
	undoStack []*Record
	redoStack []*Record
}

func NewHistory() *History {
	return &History{}
}

func (h *History) AddEvent(action Action, ogRunes []rune, line *Line) {
	lastEvent := h.lastUndoRecord()
	switch {
	case action == UNDO:
		if lastEvent == nil {
			return
		}
		lastEvent.lastX = Cur.X
		Cur.X = lastEvent.firstX
		lastEvent.line.SetRunes(lastEvent.prevRunes)
		lastEvent.prevRunes, lastEvent.currRunes = lastEvent.currRunes, lastEvent.prevRunes
		h.PushRedoStack(h.PopUndoStack())
	case action == REDO:
		if len(h.redoStack) == 0 {
			return
		}
		lastUndo := h.PopRedoStack()
		Cur.X = lastUndo.lastX
		lastUndo.line.SetRunes(lastUndo.prevRunes)
		lastUndo.prevRunes, lastUndo.currRunes = lastUndo.currRunes, lastUndo.prevRunes
		h.PushUndoStack(lastUndo)
	case lastEvent != nil && lastEvent.action == action && lastEvent.line == line:
		lastEvent.lastX = Cur.X
		lastEvent.currRunes = line.Runes()
	default:
		h.PushUndoStack(CreateRecord(action, Cur.X, Cur.Y, ogRunes, line))
	}
}

func (h *History) lastUndoRecord() *Record {
	if len(h.undoStack) == 0 {
		return nil
	}
	return h.undoStack[len(h.undoStack)-1]
}

func (h *History) PushUndoStack(record *Record) {
	h.undoStack = append(h.undoStack, record)
}

func (h *History) PopUndoStack() *Record {
	idx := len(h.undoStack) - 1
	record := h.undoStack[idx]
	h.undoStack = h.undoStack[:idx]
	return record
}

func (h *History) PushRedoStack(record *Record) {
	h.redoStack = append(h.redoStack, record)
}

func (h *History) PopRedoStack() *Record {
	idx := len(h.redoStack) - 1
	record := h.redoStack[idx]
	h.redoStack = h.redoStack[:idx]
	return record

}

type Record struct {
	action    Action
	line      *Line
	firstX    int
	lastX     int
	y         int
	prevRunes []rune
	currRunes []rune
}

func CreateRecord(action Action, x, y int, ogRunes []rune, line *Line) *Record {
	return &Record{
		action:    action,
		line:      line,
		firstX:    x,
		lastX:     x,
		y:         y,
		prevRunes: ogRunes,
		currRunes: line.Runes(),
	}
}
