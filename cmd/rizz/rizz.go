package main

import (
	"fmt"
	"os"

	display "github.com/cyamas/rizz/internal/display"
)

func main() {
	fmt.Println("rizz will be a simple text editor one day!")
	args := os.Args
	d := display.NewDisplay()
	d.Init()
	quit := func() {
		maybePanic := recover()
		d.Screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()
	d.ActiveBuf = display.NewBuffer(d.Highlighter)
	if len(args) == 2 {
		d.ActiveBuf.ReadFile(args[1])

	}
	d.InitBufWindow()
	d.SetBufWindow()
	d.Mode = display.Normal
	display.Cur.X = display.LeftMarginSize
	display.Cur.Y = 0
	d.Run()
}
