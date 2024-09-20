package main

import "fmt"

type TestStruct struct {
	x, y int
}

func testPrint() {
	fmt.Println("ok what the hell")
}

func testFunc0() bool {
	fmt.Println("split display.go file")
	fmt.Println("Ok well there are some problems here!")
	fmt.Println("Fixed the tests...for now")
	fmt.Println("test again")
	return true
}

func testFunc1() {
	fmt.Println("this is a test")
	fmt.Println("ok what if we add a line here!")
	fmt.Println("well i'm going to add a line here too!")
	fmt.Println("added an insert new line function!")
}

func testFunc2() {
	fmt.Println("hello there this is another test!")
	fmt.Println("hello again")
}

func testFunc3() {
	fmt.Println("Editing from rizz. We'll see what happens")
}

func testFunc4() {
	fmt.Println("Hey man, I think i've finally added a working left margin.")
	fmt.Println("And I've refactored the code a bit to make it nicer")
}

func testFunc5() {
	fmt.Println("another manual test on the mac")
}

func testFunc6() {
	x := 0
	for {
		if x > 10 {
			break
		} else {
			fmt.Println("x is less than 10")
		}
		fmt.Println("this will print 10 times I believe")
		fmt.Println("here is antoher test on the old macbook air.")
		fmt.Println("probably should use printf but oh well")
		fmt.Println("this works fine too I think.")
		fmt.Println("Wow this will all also print 10 times.")
	}
}

func testFunc7() {
	fmt.Println("Ok well now we should be past the initial window!")
}

func testFunc8() {
	fmt.Println("manually testing on mac to make sure it still works!")
}

func testFunc9() {
	fmt.Println("We're just adding some more lines to look at the scroll feature more closely")
}

func testFunc10() {
	fmt.Println("It seems like i've finally fixed this bug hooray!")
	fmt.Println("Well the scroll up and down half a window thing is not really working very well ha")
}

func testFunc11() {
	fmt.Println("testing again... oh dear")
}

func testFunc12() {
	fmt.Println("another test just to see if we can append to the end of the file nicely")
}

func testFunc13() {
	fmt.Println("moved some files around. Testing to see if this still works")
}

func testFunc14() {
	fmt.Println("test for auto indent")
}

func testFunc15(x int) {
	fmt.Println("hello ", x)
}

func testFunc16(x, y string) {
	if x == y {
		for _, ch := range x {
			fmt.Println(ch)
		}
	}
}
