package user

import (
	"encoding/binary"
	"log"
	"math"
)

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

type byte_t struct {
	Buf_   []*byte
	Offset uint64
	Len_   uint64
}

// func typeToStr(usertype UserType) string {
func (usertype UserType) TypeToStr() string {
	if int(usertype) >= len(USERTYPE_TO_STR) {
		log.Fatalf("Invalid typeToStr. type: %v", usertype)
	}
	return USERTYPE_TO_STR[int(usertype)]
}

func (usertype UserType) TypeToSize() uint64 {
	switch usertype {
	case TYPE_BOOL, TYPE_BYTE:
		return 1
	case TYPE_INT32, TYPE_UINT32, TYPE_FLOAT32:
		return 4
	case TYPE_INT64, TYPE_UINT64, TYPE_FLOAT64:
		return 8
	}
	log.Fatalf("Error invalid typeToSize. type: %s", usertype.TypeToStr())
	return 0
}

func (usertype UserType) IsComplexType() bool {
	return (usertype == UserType(TYPE_BYTE)) || (usertype == UserType(TYPE_STRING))
}

func (usertype UserType) IsCompatibleType(value interface{}, bytes *[]byte) bool {

	switch usertype {
	case TYPE_BOOL:
		{
			switch value.(type) {
			case bool:
				return true
			}
		}
	case TYPE_INT32:
		{
			switch value.(type) {
			case int32:
				int64ToBytes(int64(value.(int32)), bytes)
				return true
			}
		}
	case TYPE_UINT32:
		{
			switch value.(type) {
			case uint32:
				uint64ToBytes(uint64(value.(uint32)), bytes)
				return true
			}
		}
	case TYPE_INT64:
		{
			switch value.(type) {
			case int64:
				int64ToBytes(value.(int64), bytes)
				return true
			}
		}
	case TYPE_UINT64:
		{
			switch value.(type) {
			case uint64:
				uint64ToBytes(value.(uint64), bytes)
				return true
			}
		}

	case TYPE_FLOAT32:
		{
			switch value.(type) {
			case float32:
				float32ToBytes(value.(float32), bytes)
				return true
			}
		}
	case TYPE_FLOAT64:
		{
			switch value.(type) {
			case float64:
				float64ToBytes(value.(float64), bytes)
				return true
			}
		}
	}
	return false
}

func int64ToBytes(n int64, bytes *[]byte) {
	binary.LittleEndian.PutUint64(*bytes, uint64(n))
}

func uint64ToBytes(n uint64, bytes *[]byte) {
	binary.LittleEndian.PutUint64(*bytes, n)
}

func float32ToBytes(float float32, bytes *[]byte) {
	// https://stackoverflow.com/questions/22491876/convert-byte-slice-uint8-to-float64-in-golang
	bits := math.Float32bits(float)
	binary.LittleEndian.PutUint32(*bytes, bits)
}

func float64ToBytes(float float64, bytes *[]byte) {
	// https://stackoverflow.com/questions/22491876/convert-byte-slice-uint8-to-float64-in-golang
	bits := math.Float64bits(float)
	binary.LittleEndian.PutUint64(*bytes, bits)
}

/*
func (usertype UserType) IsCompatibleType(value interface{}) bool {
	switch usertype {
	case TYPE_BOOL:
		{
			switch value.(type) {
			case bool:
				return true
			}
		}
	case TYPE_INT32:
		{
			switch value.(type) {
			case int, int32:
				return true
			}
		}
	case TYPE_UINT32:
		{
			switch value.(type) {
			case uint, uint32:
				return true
			}
		}
	case TYPE_INT64:
		{
			switch value.(type) {
			case int, int32, int64:
				return true
			}
		}
	case TYPE_UINT64:
		{
			switch value.(type) {
			case uint32, uint64:
				return true
			}
		}

	case TYPE_FLOAT32:
		{
			switch value.(type) {
			case float32:
				return true
			}
		}
	case TYPE_FLOAT64:
		{
			switch value.(type) {
			case float32, float64:
				return true
			}
		}
	}
	return false
}
*/
