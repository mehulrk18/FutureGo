package main

import (
	"fmt"
	"math"

	future "github.com/mehulrk18/FutureGo/future"
)

//Add 2 numbers
func Add() {
	fmt.Println("4 + 6 = ", 4+6)
}

//Sub 2 numbers
func Sub() {
	fmt.Println("64 - 45 = ", 64-45)
}

//Multiply 2 numbers
func Multiply() {
	fmt.Println("58 * 4 = ", 58*4)
}

//Power x^y
func Power() {
	fmt.Println(" 8^5 = ", math.Pow(8, 5))
}

func main() {
	_ = future.TestFuture(Add)
	_ = future.TestFuture(Sub)
	_ = future.TestFuture(Multiply)
	_ = future.TestFuture(Power)

}
