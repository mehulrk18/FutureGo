package main

import (
	"fmt"
	"math"

	future "github.com/mehulrk18/FutureGo/future"
)

//Add 2 numbers
func Add() int64 {
	fmt.Println("4 + 6 = ")
	return 4 + 6
}

//Sub 2 numbers
func Sub() int64 {
	fmt.Println("64 - 45 = ")
	return 64 - 45
}

//Multiply 2 numbers
func Multiply() int64 {
	fmt.Println("58 * 4 = ")
	return 58 * 4
}

//Power x^y
func Power() int64 {
	fmt.Println(" 8^5 = ")
	return int64(math.Pow(8, 5))
}

func main() {
	listFuncs := [...]future.Fn{
		Add,
		Power,
		Sub,
		Multiply,
	}

	var f *future.Future
	f = new(future.Future)

	for _, fun := range listFuncs {
		f = future.Submit(fun)
		res, e := f.GetResult()
		// res, e := f.FinalResult(2)
		fmt.Println("Future result: ", res)
		fmt.Println("Error in future: ", e)
	}
}
