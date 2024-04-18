package user

import "log"

type UserType int

const (
	TYPE_UNKNOWN UserType = iota
	TYPE_BOOL
	TYPE_INT32
	TYPE_UINT32
	TYPE_INT64
	TYPE_UINT64
	TYPE_FLOAT32
	TYPE_FLOAT64
	TYPE_STRING
	TYPE_BYTE
	TYPE_MESSAGE_DECL // Declaration
	TYPE_NESTED_MESSAGE
)

var USERTYPE_TO_STR = [12]string{"unknown", "bool", "int32", "uint32", "int64", "uint64", "float32", "float64", "string", "byte", "message-decl", "message-nested"}

func typeToStr(usertype UserType) string {
	if int(usertype) >= len(USERTYPE_TO_STR) {
		log.Fatalf("Invalid typeToStr. type: %v", usertype)
	}
	return USERTYPE_TO_STR[int(usertype)]
}

func typeToSize(usertype UserType) int {
	switch usertype {
	case TYPE_BOOL, TYPE_BYTE:
		return 1
	case TYPE_INT32, TYPE_UINT32, TYPE_FLOAT32:
		return 32
	case TYPE_INT64, TYPE_UINT64, TYPE_FLOAT64:
		return 64
	}
	log.Fatalf("Error invalid typeToSize. type: %s", typeToStr(usertype))
	return 0
}
