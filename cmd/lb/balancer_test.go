package main

import (
	 "testing"
	 "fmt"
	 "github.com/stretchr/testify/assert"
)

func TestBalancer(t *testing.T) {
	// TODO: Реалізуйте юніт-тест для балансувальникка.
	a := clientHashAddress("8080", 3)
	fmt.Printf("%b\n", a)
}
