package main

import "fmt"

func testFunc() {
	fmt.Println("this is a test")
	fmt.Println("ok what if we add a line here!")
	fmt.Println("well i'm going to add a line here too!")
	fmt.Println("added an insert new line function!")
}

func testFunc3() {
	for i := range 69 {
		fmt.Println("Let's count up to 69! we're on number:", i)
	}
}

func testFunc4() {
	fmt.Println("Hey man, I think i've finally added a working left margin.")
	fmt.Println("And I've refactored the code a bit to make it nicer")
}

func testFunc5() {
	fmt.Println("Now we will try to introduce scrolling please!")
	fmt.Println("Yes this is the way")
}

func testFunc6() {
	x := 0
	for {
		if x > 10 {
			break
		}
		fmt.Println("this will print 10 times I believe")
		fmt.Println("probably should use printf but oh well")
		fmt.Println("this works fine too I think.")
		fmt.Println("Wow this will all also print 10 times.")
	}
}

func testFunc7() {
	fmt.Println("Ok well now we should be past the initial window!")
}

func testFunc8() {
	fmt.Println("Ok this should REALLY be enough now")
}
