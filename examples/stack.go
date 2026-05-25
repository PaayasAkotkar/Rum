package example

import (
	"log"
	"rum/app/stack"
)

func Example() {
	x := stack.NewStack[int]()
	x.Push(1)
	x.Push(2)
	x.Push(3)
	x.Push(4)
	x.Push(5)
	// x.Rearrange(3, 4)
	x.PushLast()
	log.Println("range: ", x.Max())

}
