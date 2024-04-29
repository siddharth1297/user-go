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

	for _, msg := range allMsgs {
		fmt.Printf("Name: %s\n", msg.Name)
		fmt.Printf("Len: %d\n", len(msg.Members))
		for _, f := range msg.Members {
			fmt.Printf("%s %s\n", f.Name, f.SymbolType.TypeToStr())
			if f.SymbolType == TYPE_NESTED_MESSAGE {
				fmt.Println("\tNested Message", f.NestedType.Name)
			}
		}
	}
}