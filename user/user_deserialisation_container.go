package user

import (
	"encoding/binary"
	"log"
	"unsafe"
)

type DeserialContainers interface {
	GetSymbol() *Symbol
}

/************** SinglePrimitiveContainer **************/

type SinglePrimitiveContainer struct {
	symbol_ *Symbol
	value_  interface{}
}

func (con *SinglePrimitiveContainer) GetSymbol() *Symbol {
	return con.symbol_
}

func NewSinglePrimitiveContainer(symbol *Symbol, value interface{}) *SinglePrimitiveContainer {
	return &SinglePrimitiveContainer{symbol_: symbol, value_: value}
}

func (con *SinglePrimitiveContainer) GetVal() interface{} {
	return con.value_
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

func (con *RepeatedPrimitiveContainer) EndDistanceInBytes() uint64 {
	log.Println("NESTED REPETED CRASH POINT")

	return 4 + uint64(con.size_)*con.symbol_.SymbolType.TypeToSize()
}

/************** SingleComplexContainer **************/

type SingleComplexContainer struct {
	symbol_        *Symbol
	buf_           *[]byte
	header_offset_ uint64
	data_offset_   uint64
	len_           uint32
	data_          *string
}

func (con *SingleComplexContainer) GetSymbol() *Symbol {
	return con.symbol_
}

func NewSingleComplexContainer(symbol *Symbol, buf *[]byte, hdr_offset uint64) *SingleComplexContainer {
	con := &SingleComplexContainer{symbol_: symbol, buf_: buf, header_offset_: hdr_offset, data_offset_: hdr_offset + 4, len_: 0, data_: nil}
	// read size
	con.len_ = binary.LittleEndian.Uint32((*con.buf_)[con.header_offset_ : con.header_offset_+4])

	//bytePtr := unsafe.Pointer(&((*con.buf_)[con.data_offset_]))
	bytePtr := unsafe.Pointer(&((*con.buf_)[con.data_offset_]))
	con.data_ = (*string)(unsafe.Pointer(&struct {
		data unsafe.Pointer
		len  int
	}{bytePtr, int(con.len_)}))
	log.Printf("SIZE::::: %v \"%v\" %v\n", con.len_, *con.data_, con.data_offset_)
	return con
}

func (con *SingleComplexContainer) GetVal() *string {
	return con.data_
}

func (con *SingleComplexContainer) GetStrLen() uint32 {
	return con.len_
}

func (con *SingleComplexContainer) EndDistanceInBytes() uint64 {
	return 4 + uint64(con.len_)
}
