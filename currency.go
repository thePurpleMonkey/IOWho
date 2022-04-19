package main

import "fmt"

// Code borrowed from StackExchange user Niko Kovacevic
// Link: https://stackoverflow.com/a/45472066/697674

type Currency int64

func ToCurrency(f float64) Currency {
	return Currency((f * 100) + 0.5)
}

func (m Currency) Float64() float64 {
	return float64(m) / 100
}

func (m Currency) Multiply(f float64) Currency {
	return Currency((float64(m) * f) + 0.5)
}

func (m Currency) String() string {
	return fmt.Sprintf("$%.2f", float64(m)/100)
}
