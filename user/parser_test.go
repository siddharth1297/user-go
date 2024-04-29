package user

import (
	"fmt"
	"log"
	"testing"
)

func TestParser1(t *testing.T) {
	allMsgs, err := ParseProtoFile("employee.proto")
	if err != nil {
		log.Fatalf("Error Parsing: %v", err)
	}

	fmt.Printf("Length: %d\n", len(allMsgs))
	for _, msg := range allMsgs {
		fmt.Printf("Name: %s\n", msg.Name)
	}
}