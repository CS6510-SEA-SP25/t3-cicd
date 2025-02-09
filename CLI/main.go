/*
Copyright © 2025 Minh Nguyen minh160302@gmail.com
*/
package main

import (
	"cicd/gocc/cmd"
	"log"
)

func main() {
	log.SetPrefix("GoCC: ")
	log.SetFlags(0)

	err := cmd.Execute()
	if err != nil {
		log.Fatalf("%v", err)
	}
}
