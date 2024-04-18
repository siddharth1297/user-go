package user

import (
	"fmt"
	"log"
	"testing"
)

func Test1(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
	   message M1 {
	       int32 a = 1;
	       repeated uint32 b = 2;
	   }
	*/
	m1 := NewDeclarationSymbol("M1")
	m1_a := NewMemberSymbol(1, "a", UserType(TYPE_INT32), m1, nil, false, false)
	m1_b := NewMemberSymbol(2, "b", UserType(TYPE_UINT32), m1, nil, true, false)

	sym_tab.addMessageDecl(m1)
	m1.addMember(m1_a)
	m1.addMember(m1_b)

	fmt.Println(m1)

	fmt.Println(m1_a)
	fmt.Println(m1_b)
}

// func SymTabTestNested1(t *testing.T) {
func Test2(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
		   message M1 {
		       int32 a = 1;
		       repeated uint32 b = 2;
		   }
		   message M2 {
				int32 a = 1;
				M1 b = 2;
		   }
	*/
	m1 := NewDeclarationSymbol("M1")
	m1_a := NewMemberSymbol(1, "a", UserType(TYPE_INT32), m1, nil, false, false)
	m1_b := NewMemberSymbol(2, "b", UserType(TYPE_UINT32), m1, nil, true, false)

	sym_tab.addMessageDecl(m1)
	m1.addMember(m1_a)
	m1.addMember(m1_b)

	m2 := NewDeclarationSymbol("M2")
	m2_a := NewMemberSymbol(1, "a", UserType(TYPE_INT32), m2, nil, false, false)
	m2_b := NewMemberSymbol(2, "b", UserType(TYPE_NESTED_MESSAGE), m2, m1, false, false)
	sym_tab.addMessageDecl(m2)
	m2.addMember(m2_a)
	m2.addMember(m2_b)
	fmt.Println(m2)
}
