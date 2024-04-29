package user

import (
	"log"
	"syscall"
	"unsafe"
)

type FieldValue interface {
	GetSymbol() *Symbol
}

/************** SingleSimpleFieldValue **************/

type SinglePrimitiveFieldValue struct {
	symbol_      *Symbol
	value_bytes_ []byte // 8 byte always
	nested_msg_  *Message
}

func NewSinglePrimitiveFieldValue(symbol *Symbol) *SinglePrimitiveFieldValue {
	field := &SinglePrimitiveFieldValue{symbol_: symbol, value_bytes_: make([]byte, 8), nested_msg_: nil}
	return field
}

func (field *SinglePrimitiveFieldValue) GetSymbol() *Symbol {
	return field.symbol_
}

func (field *SinglePrimitiveFieldValue) SetValue(value interface{}) {
	if field.symbol_.SymbolType == TYPE_NESTED_MESSAGE {
		field.nested_msg_ = value.(*Message)
		if field.nested_msg_ == nil {
			log.Fatalf("Nested message is null. Symbol: %v\n", field.symbol_)
		}
		if field.nested_msg_.message_name_ != field.symbol_.NestedType.Name {
			log.Fatalf("NestedType and declaration type mismatch. Setting %v. Declared %v", field.nested_msg_, field.symbol_.NestedType.Name)
		}
		return
	}
	if !field.symbol_.SymbolType.IsCompatibleType(value, &field.value_bytes_) {
		log.Fatalf("SetValue Incomaptible type. Member: %v", field.symbol_)
	}
}

func (field *SinglePrimitiveFieldValue) SerialiseToIoVec(iovs *[]syscall.Iovec, iov_idx *int) {
	if field.symbol_.SymbolType == TYPE_NESTED_MESSAGE {
		log.Fatalf("Should not follow this path")
		if field.nested_msg_ == nil {
			return
		}
		field.nested_msg_.SerialisaeToSGBuf(iovs, iov_idx)
		return
	}
	(*iovs)[*iov_idx] = syscall.Iovec{Base: (*byte)(unsafe.Pointer(&field.value_bytes_[0])), Len: field.symbol_.SymbolType.TypeToSize()}
	(*iov_idx)++
}

/************** RepeatedSimpleFieldValue **************/
// TODO: Optimisation: Set size at the starting, then allocate contigous byte array
type RepeatedPrimitiveFieldValue struct {
	symbol_      *Symbol
	value_bytes_ [][]byte // 8 byte always
	size_        uint32   // size of the value_bytes_
	nested_msgs_ []*Message
}

func NewRepeatedPrimitiveFieldValue(symbol *Symbol) *RepeatedPrimitiveFieldValue {
	field := &RepeatedPrimitiveFieldValue{symbol_: symbol, value_bytes_: make([][]byte, 0), size_: 0, nested_msgs_: make([]*Message, 0)}
	return field
}

func (field *RepeatedPrimitiveFieldValue) GetSymbol() *Symbol {
	return field.symbol_
}
func (field *RepeatedPrimitiveFieldValue) Size() uint32 {
	return field.size_
}
func (field *RepeatedPrimitiveFieldValue) AddValue(value interface{}) {
	field.size_++
	if field.symbol_.SymbolType == TYPE_NESTED_MESSAGE {
		field.nested_msgs_ = append(field.nested_msgs_, value.(*Message))
		if field.nested_msgs_[len(field.nested_msgs_)-1] == nil {
			log.Fatalf("Nested message is null. Symbol: %v\n", field.symbol_)
		}
		if field.nested_msgs_[len(field.nested_msgs_)-1].message_name_ != field.symbol_.NestedType.Name {
			log.Fatalf("NestedType and declaration type mismatch. Setting %v. Declared %v", field.nested_msgs_, field.symbol_.NestedType.Name)
		}
		return
	}
	field.value_bytes_ = append(field.value_bytes_, make([]byte, 8))
	if !field.symbol_.SymbolType.IsCompatibleType(value, &(field.value_bytes_[len(field.value_bytes_)-1])) {
		log.Fatalf("AddValue Incomaptible type. Member: %v", field.symbol_)
	}
}

func (field *RepeatedPrimitiveFieldValue) SerialiseToIoVec(iovs *[]syscall.Iovec, iov_idx *int) {
	(*iovs)[*iov_idx] = syscall.Iovec{Base: (*byte)(unsafe.Pointer(&field.size_)), Len: 4}
	(*iov_idx)++
	log.Println("Bfr----========, ", field.size_)
	for i := uint32(0); i < field.size_; i++ {
		log.Println("Bfr---")
		if field.symbol_.SymbolType == TYPE_NESTED_MESSAGE {
			if field.nested_msgs_[i] == nil {
				return
			}
			field.nested_msgs_[i].SerialisaeToSGBuf(iovs, iov_idx)
		} else {
			(*iovs)[*iov_idx] = syscall.Iovec{Base: (*byte)(unsafe.Pointer(&field.value_bytes_[i][0])), Len: field.symbol_.SymbolType.TypeToSize()}
			(*iov_idx)++
		}

		log.Println("After---")
	}
}

/************** SingleComplexFieldValue **************/

type SingleComplexFieldValue struct {
	symbol_      *Symbol
	value_bytes_ *string
	len_         uint32 // Length of the string
	nested_msg_  *Message
}

func NewSingleComplexFieldValue(symbol *Symbol) *SingleComplexFieldValue {
	field := &SingleComplexFieldValue{symbol_: symbol, value_bytes_: nil, len_: 0, nested_msg_: nil}
	return field
}

func (field *SingleComplexFieldValue) GetSymbol() *Symbol {
	return field.symbol_
}

func (field *SingleComplexFieldValue) SetValueStr(value *string, len uint32) {
	if field.symbol_.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
		log.Fatalf("INVALID API. SetValueStr is used only when the member is non-nested type. member: %v\n", field.symbol_)
	}
	field.value_bytes_ = value
	field.len_ = len
}

func (field *SingleComplexFieldValue) SetNestedMessage(msg *Message) {
	if field.symbol_.SymbolType != UserType(TYPE_NESTED_MESSAGE) {
		log.Fatalf("INVALID API. SetNestedMessage is used only when the member is nested type. member: %v\n", field.symbol_)
	}
	if msg.message_name_ != field.symbol_.NestedType.Name {
		log.Fatalf("NestedType and declaration type mismatch. Member: %v. Setting %v. Declared %v", msg.message_symbol_, msg.message_symbol_.Name, field.symbol_.NestedType.Name)
	}
	field.nested_msg_ = msg
}

func (field *SingleComplexFieldValue) SerialiseToIoVec(iovs *[]syscall.Iovec, iov_idx *int) {
	if field.symbol_.SymbolType == TYPE_NESTED_MESSAGE {
		if field.nested_msg_ == nil {
			return
		}
		field.nested_msg_.SerialisaeToSGBuf(iovs, iov_idx)
		return
	}

	//(*iovs)[*iov_idx] = syscall.Iovec{Base: (*byte)(unsafe.Pointer(&field.len_[0])), Len: field.symbol_.SymbolType.TypeToSize()}
	(*iovs)[*iov_idx] = syscall.Iovec{Base: (*byte)(unsafe.Pointer(&field.len_)), Len: 4}
	(*iov_idx)++

	// Working but it is converting it to byte
	//(*iovs)[*iov_idx] = syscall.Iovec{Base: (*byte)(unsafe.Pointer(&[]byte(*field.value_bytes_)[0])), Len: uint64(field.len_)}

	(*iovs)[*iov_idx] = syscall.Iovec{Base: unsafe.StringData(*field.value_bytes_), Len: uint64(field.len_)}
	(*iov_idx)++

	log.Println("-----DUMPED----- ", *iov_idx)
}

/************** RepeatedComplexFieldValue **************/

type RepeatedComplexFieldValue struct {
	symbol_         *Symbol
	values_         []*SingleComplexFieldValue
	size_           uint32
	nesated_values_ []*Message
}

func NewRepeatedComplexFieldValue(symbol *Symbol) *RepeatedComplexFieldValue {
	field := &RepeatedComplexFieldValue{symbol_: symbol, values_: nil, size_: 0, nesated_values_: nil}
	if symbol.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
		field.nesated_values_ = make([]*Message, 0)
	} else {
		field.values_ = make([]*SingleComplexFieldValue, 0)
	}
	return field
}

func (field *RepeatedComplexFieldValue) GetSymbol() *Symbol {
	return field.symbol_
}

func (field *RepeatedComplexFieldValue) Size() uint32 {
	return field.size_
}

func (field *RepeatedComplexFieldValue) AddValueStr(value *string, len uint32) {
	single_cmplx_obj := NewSingleComplexFieldValue(field.symbol_)
	single_cmplx_obj.SetValueStr(value, len)
	field.values_ = append(field.values_, single_cmplx_obj)
	field.size_++
}

func (field *RepeatedComplexFieldValue) AddNestedMessage(value *Message) {
	field.nesated_values_ = append(field.nesated_values_, value)
	log.Printf("Appended %v\n", len(field.nesated_values_))
	field.size_++
}

func (field *RepeatedComplexFieldValue) SerialiseToIoVec(iovs *[]syscall.Iovec, iov_idx *int) {
	// Write size
	(*iovs)[*iov_idx] = syscall.Iovec{Base: (*byte)(unsafe.Pointer(&field.size_)), Len: 4}
	(*iov_idx)++

	if field.symbol_.SymbolType == TYPE_NESTED_MESSAGE {
		if field.nesated_values_ == nil {
			return
		}
		log.Println("SIZEEEEE: ", field.size_)
		for _, value := range field.nesated_values_ {
			value.SerialisaeToSGBuf(iovs, iov_idx)
		}
		return
	}

	for _, value := range field.values_ {
		value.SerialiseToIoVec(iovs, iov_idx)
	}
}