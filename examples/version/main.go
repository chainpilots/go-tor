package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chainpilots/go-tor/process"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	p, err := process.New(context.Background(), "--version")
	if err != nil {
		return err
	}
	fmt.Printf("Starting...\n")
	if err = p.Start(); err != nil {
		return err
	}
	fmt.Printf("Waiting...\n")
	return p.Wait()
}
