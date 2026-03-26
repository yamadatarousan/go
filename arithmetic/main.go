package main

import (
	"errors"
	"fmt"
)

func addition(a, b float64) (float64, error) {
	return a + b, nil
}

func subtraction(a, b float64) (float64, error) {
	return a - b, nil
}

func multiplication(a, b float64) (float64, error) {
	return a * b, nil
}

func division(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func main() {
	a, _ := addition(1, 2)
	fmt.Println("addition:", a)

	b, _ := subtraction(1, 2)
	fmt.Println("addition:", b)

	c, _ := multiplication(1, 2)
	fmt.Println("multiplication:", c)

	d, _ := division(1, 2)
	fmt.Println("division:", d)

	e, err := division(1, 0)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("division:", e)
	}
}
