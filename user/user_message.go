package user

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"
)

type masktype_t uint32

const BITMAP_SIZE = uint64(4)
const MAX_ALLOWED_MEMBERS = uint32(32)

type Message struct {
	message_name_   string
	message_symbol_ *Symbol
	bitmap_         masktype_t
	field_values_   []FieldValue
}

func NewMessage(message_name string) *Message {
	msg := &Message{message_name_: message_name, message_symbol_: GetSymTab().getSymbolFromName(message_name), bitmap_: 0, field_values_: make([]FieldValue, 0)}

	for _, mem_sym := range msg.message_symbol_.Members {
		var field_val FieldValue
		if mem_sym.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
			if mem_sym.Repeated {
				log.Println("NestedType-Repeated ", mem_sym)
				field_val = NewRepeatedComplexFieldValue(mem_sym)
			} else {
				log.Println("NestedType-Single", mem_sym)
				field_val = NewSingleComplexFieldValue(mem_sym)
			}
		} else if !mem_sym.SymbolType.IsComplexType() {
			log.Println("NONComplextype", mem_sym)
			if mem_sym.Repeated {
				field_val = NewRepeatedPrimitiveFieldValue(mem_sym)
			} else {
				field_val = NewSinglePrimitiveFieldValue(mem_sym)
			}
		} else {
			log.Println("Complextype", mem_sym)
			if mem_sym.Repeated {
				field_val = NewRepeatedComplexFieldValue(mem_sym)
			} else {
				field_val = NewSingleComplexFieldValue(mem_sym)
			}
		}
		msg.field_values_ = append(msg.field_values_, field_val)
	}
	return msg
}

// ------------------------ Set ------------------------

// Single, Primitive
func (msg *Message) SetVal(message_name string, value interface{}) {
	symbol, idx := msg.message_symbol_.getMemberSymbolAndIdx(message_name)
	if symbol.Repeated {
		log.Fatalf("InvalidAPI: SetVal. Not a repeated field. Member: [%v], message: [%v] structure: %v", symbol, msg.message_name_, msg.message_symbol_)
	}
	if symbol.SymbolType.IsComplexType() {
		log.Fatalf("InvalidAPI: SetVal. Not complex type. Member: [%v], message: [%v] structure: %v", symbol, msg.message_name_, msg.message_symbol_)
	}
	field := msg.field_values_[idx]
	value_field, _ := field.(*SinglePrimitiveFieldValue)
	value_field.SetValue(value)
	msg.bitmap_ |= getMask(symbol.Id)
}

// Repeat, Primitive
func (msg *Message) AddVal(message_name string, value interface{}) {
	symbol, idx := msg.message_symbol_.getMemberSymbolAndIdx(message_name)
	if !symbol.Repeated {
		log.Fatalf("InvalidAPI: AddVal. Must be a repeated field. Member: [%v], message: [%v] structure: %v", symbol, msg.message_name_, msg.message_symbol_)
	}
	if symbol.SymbolType.IsComplexType() {
		log.Fatalf("InvalidAPI: SetVal. Not complex type. Member: [%v], message: [%v] structure: %v", symbol, msg.message_name_, msg.message_symbol_)
	}
	field := msg.field_values_[idx]
	value_field, _ := field.(*RepeatedPrimitiveFieldValue)
	value_field.AddValue(value)
	msg.bitmap_ |= getMask(symbol.Id)
}

// Single, Complex
func (msg *Message) SetValStr(message_name string, val *string, len uint32) {
	symbol, idx := msg.message_symbol_.getMemberSymbolAndIdx(message_name)
	if symbol.Repeated {
		log.Fatalf("InvalidAPI: SetValStr. Must not a repeated field. Member: [%v], message: [%v] structure: %v", symbol, msg.message_name_, msg.message_symbol_)
	}
	if !symbol.SymbolType.IsComplexType() {
		log.Fatalf("InvalidAPI: SetValStr. Must be a complex type. Member: [%v], message: [%v] structure: %v", symbol, msg.message_name_, msg.message_symbol_)
	}
	field := msg.field_values_[idx]
	log.Println("Set")
	value_field, _ := field.(*SingleComplexFieldValue)
	fmt.Println("VALFIEDL: ", value_field)
	value_field.SetValueStr(val, len)
	msg.bitmap_ |= getMask(symbol.Id)
}

func (msg *Message) SetValMsg(message_name string, val *Message) {
	symbol, idx := msg.message_symbol_.getMemberSymbolAndIdx(message_name)
	if symbol.Repeated {
		log.Fatalf("InvalidAPI: SetValStr. Must not a repeated field. Member: [%v], message: [%v] structure: %v", symbol, msg.message_name_, msg.message_symbol_)
	}
	if !symbol.SymbolType.IsComplexType() {
		log.Fatalf("InvalidAPI: SetValStr. Must be a complex type. Member: [%v], message: [%v] structure: %v", symbol, msg.message_name_, msg.message_symbol_)
	}
	field := msg.field_values_[idx]
	log.Println("Set")
	value_field, _ := field.(*SingleComplexFieldValue)
	value_field.SetNestedMessage(val)
	msg.bitmap_ |= getMask(symbol.Id)
}

// Repeated, Complex
func (msg *Message) AddValStr(message_name string, val *string, len uint32) {
	symbol, idx := msg.message_symbol_.getMemberSymbolAndIdx(message_name)
	if !symbol.Repeated {
		log.Fatalf("InvalidAPI: AddValStr. Must be a repeated field. Member: [%v], message: [%v] structure: %v", symbol, msg.message_name_, msg.message_symbol_)
	}
	if !symbol.SymbolType.IsComplexType() {
		log.Fatalf("InvalidAPI: AddValStr. Must be a complex type. Member: [%v], message: [%v] structure: %v", symbol, msg.message_name_, msg.message_symbol_)
	}
	field := msg.field_values_[idx]

	str_val := val
	value_field, _ := field.(*RepeatedComplexFieldValue)
	value_field.AddValueStr(str_val, len)

	msg.bitmap_ |= getMask(symbol.Id)
}

// For all type of nested message
func (msg *Message) SetNestedMsg(message_name string, nested_msg *Message) {
	symbol, idx := msg.message_symbol_.getMemberSymbolAndIdx(message_name)
	if symbol.SymbolType != UserType(TYPE_NESTED_MESSAGE) || symbol.Repeated {
		log.Fatalf("InvalidAPI: SetNestedMsg. Not a nested field or not single. Member: [%v], message: [%v] structure: %v", symbol, msg.message_name_, msg.message_symbol_)
	}
	if nested_msg == nil {
		log.Fatalf("Nested message is null. Symbol: %v\n", symbol)
	}
	if nested_msg.message_name_ != symbol.NestedType.Name {
		log.Fatalf("NestedType and declaration type mismatch. Setting %v. Declared %v", nested_msg.message_symbol_.Name, symbol.NestedType.Name)
	}
	field := msg.field_values_[idx]
	log.Println("SetNestedMsg-Single")
	typed_field, _ := field.(*SingleComplexFieldValue)
	typed_field.SetNestedMessage(nested_msg)

	msg.bitmap_ |= getMask(symbol.Id)
}

func (msg *Message) AddNestedMsg(message_name string, nested_msg *Message) {
	symbol, idx := msg.message_symbol_.getMemberSymbolAndIdx(message_name)
	if symbol.SymbolType != UserType(TYPE_NESTED_MESSAGE) || !symbol.Repeated {
		log.Fatalf("InvalidAPI: AddNestedMsg. Not a nested field or not repeated. Member: [%v], message: [%v] structure: %v", symbol, msg.message_name_, msg.message_symbol_)
	}
	if nested_msg == nil {
		log.Fatalf("Nested message is null. Symbol: %v\n", symbol)
	}
	if nested_msg.message_name_ != symbol.NestedType.Name {
		log.Fatalf("NestedType and declaration type mismatch. Setting %v. Declared %v", nested_msg.message_symbol_.Name, symbol.NestedType.Name)
	}
	field := msg.field_values_[idx]

	log.Println("SetNestedMsg-Repeated")
	typed_field, _ := field.(*RepeatedComplexFieldValue)
	typed_field.AddNestedMessage(nested_msg)

	msg.bitmap_ |= getMask(symbol.Id)
}

// -------------------------- Serialise using SG --------

func (msg *Message) InitAndSerialiseToSGBuf(reserve int) []syscall.Iovec {
	iovec_len := uint32(reserve) + msg.CountIoVecLen()
	fmt.Printf("IOVECLEN: %v\n", iovec_len)
	iov_idx := reserve
	var iovs []syscall.Iovec = make([]syscall.Iovec, iovec_len)
	log.Printf("IOVLen: %v\n", iovec_len)
	msg.SerialisaeToSGBuf(&iovs, &iov_idx)
	return iovs
}

func (msg *Message) SerialisaeToSGBuf(iovs *[]syscall.Iovec, iov_idx *int) {
	if msg.bitmap_ == masktype_t(0) {
		return
	}
	(*iovs)[*iov_idx] = syscall.Iovec{Base: (*byte)(unsafe.Pointer(&msg.bitmap_)), Len: BITMAP_SIZE}
	(*iov_idx)++
	log.Printf("BITMAP: %v\n", msg.bitmap_)

	for _, field := range msg.field_values_ {
		field_symbol := field.GetSymbol()
		if !isPresent(msg.bitmap_, field_symbol.Id) {
			continue
		}
		log.Printf("------------==============================SET: %v type: %v\n", field_symbol, field_symbol.SymbolType)

		if field_symbol.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
			if !field_symbol.Repeated {
				log.Println("Nested-Single-Complex")
				value_field, _ := field.(*SingleComplexFieldValue)
				log.Println("Nested-Single-Complex ", value_field)
				value_field.SerialiseToIoVec(iovs, iov_idx)
			} else {
				log.Println("Nested-Repeated-Complex")
				value_field, _ := field.(*RepeatedComplexFieldValue)
				value_field.SerialiseToIoVec(iovs, iov_idx)
			}
			continue
		}
		log.Println("Non-nested")
		if !field_symbol.Repeated && (!field_symbol.SymbolType.IsComplexType()) {
			log.Println("Single-Primitive")
			value_field, _ := field.(*SinglePrimitiveFieldValue)
			value_field.SerialiseToIoVec(iovs, iov_idx)
		}
		if field_symbol.Repeated && !field_symbol.SymbolType.IsComplexType() {
			log.Println("Complex-Primitive")
			value_field, _ := field.(*RepeatedPrimitiveFieldValue)
			value_field.SerialiseToIoVec(iovs, iov_idx)
		}
		if !field_symbol.Repeated && field_symbol.SymbolType.IsComplexType() {
			log.Println("Single-Complex")
			value_field, _ := field.(*SingleComplexFieldValue)
			value_field.SerialiseToIoVec(iovs, iov_idx)
		}
		if field_symbol.Repeated && field_symbol.SymbolType.IsComplexType() {
			log.Println("Repeated-Complex")
			value_field, _ := field.(*RepeatedComplexFieldValue)
			value_field.SerialiseToIoVec(iovs, iov_idx)
		}
	}
}

func (msg *Message) CountIoVecLen() uint32 {
	if msg.bitmap_ == masktype_t(0) {
		return 0
	}
	n := uint32(1) // bitmap

	for _, field := range msg.field_values_ {
		field_symbol := field.GetSymbol()
		if !isPresent(msg.bitmap_, field_symbol.Id) {
			continue
		}
		log.Printf("---->>> %v present\n", field_symbol.Name)
		if field_symbol.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
			log.Printf("Len---------------------------->>>>>: %v\n", field_symbol)
			if field_symbol.Repeated {
				value_field, _ := field.(*RepeatedComplexFieldValue)
				for _, value := range value_field.nesated_values_ {
					a := value.CountIoVecLen()
					log.Printf("Len---------------------------->>>>>: %v\n", a)
					n += 1 + a // 1 is for list len
				}
			} else {
				value_field, _ := field.(*SingleComplexFieldValue)
				if value_field.nested_msg_ != nil {
					n += value_field.nested_msg_.CountIoVecLen()
				}
			}
		} else {
			if !field_symbol.Repeated && !field_symbol.SymbolType.IsComplexType() {
				if field_symbol.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
					log.Printf("----<<<<<<<<>>> %v present\n", field_symbol.Name)
					//value_field, _ := field.(*SinglePrimitiveFieldValue)
					value_field, _ := field.(*SinglePrimitiveFieldValue)
					if value_field.nested_msg_ != nil {
						n += value_field.nested_msg_.CountIoVecLen()
					}
				} else {
					log.Printf("ELSE\n")
					n++
				}
			}
			if field_symbol.Repeated && !field_symbol.SymbolType.IsComplexType() {
				if field_symbol.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
					log.Printf("----<<<<<<<<>>> %v present\n", field_symbol.Name)
					value_field, _ := field.(*RepeatedPrimitiveFieldValue)
					for _, nested_msg := range value_field.nested_msgs_ {
						if nested_msg != nil {
							n += 1 + nested_msg.CountIoVecLen()
						}
					}
				} else {
					value_field, _ := field.(*RepeatedPrimitiveFieldValue)
					n += 1 + value_field.Size()
				}
			}
			if !field_symbol.Repeated && field_symbol.SymbolType.IsComplexType() {
				if field_symbol.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
					log.Printf("----<<<<<<<<>>> %v present\n", field_symbol.Name)
					value_field, _ := field.(*RepeatedPrimitiveFieldValue)
					for _, nested_msg := range value_field.nested_msgs_ {
						if nested_msg != nil {
							n += 1 + nested_msg.CountIoVecLen()
						}
					}
				} else {
					//value_field, _ := field.(*SingleComplexFieldValue)
					n += 1 + 1
				}
			}
			if field_symbol.Repeated && field_symbol.SymbolType.IsComplexType() {
				if field_symbol.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
					log.Printf("----<<<<<<<<>>> %v present\n", field_symbol.Name)
					value_field, _ := field.(*RepeatedPrimitiveFieldValue)
					for _, nested_msg := range value_field.nested_msgs_ {
						if nested_msg != nil {
							n += 1 + nested_msg.CountIoVecLen()
						}
					}
				} else {
					value_field, _ := field.(*RepeatedComplexFieldValue)
					n += 1 + value_field.Size()*3
				}
			}
		}
	}
	log.Printf("======================================RETURNING: %v\n", n)
	return n
}

func getMask(n uint32) masktype_t {
	if !(0 < n && n <= MAX_ALLOWED_MEMBERS) {
		log.Fatalf("Member id is more than allowed. id: %v\n", n)
	}
	return 1 << (n - 1)
}

func isPresent(bitmap masktype_t, n uint32) bool {
	return (bitmap & getMask(n)) > 0
}
