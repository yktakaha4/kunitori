package main

import (
	"io"
	"log"
)

func main() {
	log.SetOutput(io.Discard)
	println("hello")
}
