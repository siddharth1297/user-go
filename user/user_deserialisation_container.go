package user

import (
	"encoding/binary"
	"log"
	"unsafe"
)

type DeserialContainers interface {
	GetSymbol() *Symbol
	TotalLenInBuffer() uint32
}

/************** SinglePrimitiveContainer **************/

type SinglePrimitiveContainer struct {
	symbol_ *Symbol
	value_  interface{}
}

func (con *SinglePrimitiveContainer) GetSymbol() *Symbol {
	return con.symbol_
}

/*
func NewSinglePrimitiveContainer(symbol *Symbol, value interface{}) *SinglePrimitiveContainer {
	return &SinglePrimitiveContainer{symbol_: symbol, value_: value}
}
*/

func NewSinglePrimitiveContainer(symbol *Symbol, buf *[]byte, hdr_offset uint64) *SinglePrimitiveContainer {
	con := &SinglePrimitiveContainer{symbol_: symbol, value_: getPrimitiveValFromByteBuffer(buf, hdr_offset, symbol.SymbolType.TypeToSize(), symbol.SymbolType)}
	return con
}

func (con *SinglePrimitiveContainer) GetVal() interface{} {
	return con.value_
}

func (con *SinglePrimitiveContainer) TotalLenInBuffer() uint32 {
	return uint32(con.symbol_.SymbolType.TypeToSize())
}

/************** RepeatedPrimitiveContainer **************/

type RepeatedPrimitiveContainer struct {
	symbol_        *Symbol
	size_          uint32
	buf_           *[]byte
	header_offset_ uint64
}

func (con *RepeatedPrimitiveContainer) GetSymbol() *Symbol {
	return con.symbol_
}

func NewRepeatedPrimitiveContainer(symbol *Symbol, buf *[]byte, hdr_offset uint64) *RepeatedPrimitiveContainer {
	con := &RepeatedPrimitiveContainer{symbol_: symbol, size_: 0, buf_: buf, header_offset_: hdr_offset}
	con.size_ = binary.LittleEndian.Uint32((*con.buf_)[con.header_offset_ : con.header_offset_+4])
	return con
}

func (con *RepeatedPrimitiveContainer) Size() uint32 {
	return con.size_
}

func (con *RepeatedPrimitiveContainer) GetValAt(idx uint32) interface{} {
	if con.size_ <= idx {
		return nil
	}
	off := con.header_offset_ + 4 + uint64(idx)*con.symbol_.SymbolType.TypeToSize()
	return getPrimitiveValFromByteBuffer(con.buf_, off, con.symbol_.SymbolType.TypeToSize(), con.symbol_.SymbolType)
}

func (con *RepeatedPrimitiveContainer) TotalLenInBuffer() uint32 {
	log.Println("NESTED REPETED CRASH POINT")

	return 4 + (con.size_)*uint32(con.symbol_.SymbolType.TypeToSize())
}

/************** SingleComplexContainer **************/

type SingleComplexContainer struct {
	symbol_        *Symbol
	buf_           *[]byte
	header_offset_ uint64
	data_offset_   uint64
	len_           uint32
	data_          *string

	nested_msg_ *Deserialiser
}

func (con *SingleComplexContainer) GetSymbol() *Symbol {
	return con.symbol_
}

func NewSingleComplexContainer(symbol *Symbol, buf *[]byte, hdr_offset uint64) *SingleComplexContainer {
	con := &SingleComplexContainer{symbol_: symbol, buf_: buf, header_offset_: hdr_offset, data_offset_: hdr_offset + 4, len_: 0, data_: nil, nested_msg_: nil}
	if symbol.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
		log.Fatalf("InvalidAPI. NewSingleComplexContainer is only for nested messages. Symbol: %v\n", con.symbol_)
	}
	// read size
	con.len_ = binary.LittleEndian.Uint32((*con.buf_)[con.header_offset_ : con.header_offset_+4])

	//bytePtr := unsafe.Pointer(&((*con.buf_)[con.data_offset_]))

	bytePtr := unsafe.Pointer(&((*con.buf_)[con.data_offset_]))
	con.data_ = (*string)(unsafe.Pointer(&struct {
		data unsafe.Pointer
		len  int
	}{bytePtr, int(con.len_)}))
	data := string((*con.buf_)[con.data_offset_ : con.data_offset_+uint64(con.len_)])
	// TODO: Set it while returning
	con.data_ = &data

	//log.Printf("SIZE::::: %v \"%v\" %v\n", con.len_, *con.data_, con.data_offset_)
	return con
}

func NewSingleNestedContainer(symbol *Symbol, buf *[]byte, buf_len uint32, hdr_offset uint64) *SingleComplexContainer {
	con := &SingleComplexContainer{symbol_: symbol, buf_: buf, header_offset_: hdr_offset, data_offset_: hdr_offset + 4, len_: 0, data_: nil, nested_msg_: nil}
	if symbol.SymbolType != UserType(TYPE_NESTED_MESSAGE) {
		log.Fatalf("InvalidAPI. NewSingleNestedContainer is only for nested messages. Symbol: %v\n", con.symbol_)
	}
	con.nested_msg_ = NewDeserialiser(con.symbol_.NestedType.Name, con.buf_, uint64(buf_len), hdr_offset)
	if con.nested_msg_ == nil {
		return nil
	}
	return con
}

func (con *SingleComplexContainer) GetVal() *string {
	return con.data_
}

func (con *SingleComplexContainer) GetStrLen() uint32 {
	return con.len_
}

func (con *SingleComplexContainer) TotalLenInBuffer() uint32 {
	if con.symbol_.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
		return con.nested_msg_.TotalLenInBuffer()
	}
	return 4 + con.len_
}

func (con *SingleComplexContainer) GetNestedMessage() *Deserialiser {
	return con.nested_msg_
}

/************** RepeatedComplexContainer **************/

type RepeatedComplexContainer struct {
	symbol_        *Symbol
	buf_           *[]byte
	header_offset_ uint64
	data_offset_   uint64
	size_          uint32
	data_          []*SingleComplexContainer
	total_len_     uint32 // including size of each field

	nested_msgs_ []*Deserialiser
}

func (con *RepeatedComplexContainer) GetSymbol() *Symbol {
	return con.symbol_
}

func NewRepeatedComplexContainer(symbol *Symbol, buf *[]byte, hdr_offset uint64) *RepeatedComplexContainer {
	con := &RepeatedComplexContainer{symbol_: symbol, buf_: buf, header_offset_: hdr_offset, data_offset_: hdr_offset + 4, size_: 0, data_: make([]*SingleComplexContainer, 0), nested_msgs_: nil}
	// read size
	con.size_ = binary.LittleEndian.Uint32((*con.buf_)[con.header_offset_ : con.header_offset_+4])
	hdr_offset = con.header_offset_ + uint64(4) // 4B for the size

	for i := 0; i < int(con.size_); i++ {
		single_cmpl_obj := NewSingleComplexContainer(con.symbol_, con.buf_, hdr_offset)
		con.data_ = append(con.data_, single_cmpl_obj)
		hdr_offset += uint64(single_cmpl_obj.TotalLenInBuffer())
	}
	con.total_len_ = uint32(hdr_offset) - uint32(con.header_offset_)
	return con
}

func NewRepeatedNestedContainer(symbol *Symbol, buf *[]byte, buf_len uint32, hdr_offset uint64) *RepeatedComplexContainer {
	con := &RepeatedComplexContainer{symbol_: symbol, buf_: buf, header_offset_: hdr_offset, data_offset_: hdr_offset + 4, size_: 0, data_: nil, total_len_: 0, nested_msgs_: make([]*Deserialiser, 0)}
	if symbol.SymbolType != UserType(TYPE_NESTED_MESSAGE) || !symbol.Repeated {
		log.Fatalf("InvalidAPI. NewRepeatedNestedContainer is only for repeated nested messages. Symbol: %v\n", con.symbol_)
	}
	// read size
	con.size_ = binary.LittleEndian.Uint32((*con.buf_)[con.header_offset_ : con.header_offset_+4])
	log.Printf("Reading Size: %v\n", con.size_)
	hdr_offset = con.header_offset_ + uint64(4) // 4B for the size

	for i := 0; i < int(con.size_); i++ {
		des_obj := NewDeserialiser(con.symbol_.NestedType.Name, con.buf_, uint64(buf_len), hdr_offset)
		con.nested_msgs_ = append(con.nested_msgs_, des_obj)
		hdr_offset += uint64(des_obj.TotalLenInBuffer())
	}
	con.total_len_ = uint32(hdr_offset) - uint32(con.header_offset_)
	return con
}

func (con *RepeatedComplexContainer) Size() uint32 {
	return con.size_
}

func (con *RepeatedComplexContainer) GetValAt(idx uint32) *string {
	if con.size_ <= idx {
		return nil
	}
	return con.data_[idx].GetVal()
}

func (con *RepeatedComplexContainer) GetStrLen() uint32 {
	return con.size_
}

func (con *RepeatedComplexContainer) TotalLenInBuffer() uint32 {
	return con.total_len_
}

func (con *RepeatedComplexContainer) GetNestedMessageAt(idx uint32) *Deserialiser {
	if con.size_ <= idx {
		return nil
	}
	return con.nested_msgs_[idx]
}
