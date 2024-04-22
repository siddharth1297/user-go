package user

import (
	"encoding/binary"
	"log"
	"math"
)

type Deserialiser struct {
	message_name_   string
	message_symbol_ *Symbol
	buf_            *[]byte
	buf_len_        uint64
	buf_start_      uint64
	bitmap_         masktype_t
	cache_          map[uint32]DeserialContainers
}

func NewDeserialiser(message_name string, buf *[]byte, buf_len uint64, buf_start uint64) *Deserialiser {
	des := &Deserialiser{message_name_: message_name, message_symbol_: GetSymTab().getSymbolFromName(message_name), buf_: buf, buf_len_: buf_len, buf_start_: buf_start, bitmap_: 0, cache_: make(map[uint32]DeserialContainers)}
	if des.message_symbol_ == nil {
		log.Fatalf("Symbol not found for deserialisation. message name: %v\n", message_name)
	}
	log.Printf("BUF LEN: %v\n", des.buf_len_)
	// set bitmap
	//des.bitmap_ = masktype_t(binary.LittleEndian.Uint32((*des.buf_)[des.buf_start_ : des.buf_start_+BITMAP_SIZE]))
	bitmapval := getPrimitiveValFromByteBuffer(des.buf_, des.buf_start_, BITMAP_SIZE, TYPE_UINT32)
	des.bitmap_ = masktype_t(bitmapval.(uint32))
	log.Printf("BITMAP: %v\n", des.bitmap_)
	return des
}

func (des *Deserialiser) isPresent(msg_name string) bool {
	symbol, _ := des.message_symbol_.getMemberSymbolAndIdx(msg_name)
	return !isPresent(des.bitmap_, symbol.Id)
}

func (des *Deserialiser) GetVal(msg_name string) interface{} {
	symbol, _ := des.message_symbol_.getMemberSymbolAndIdx(msg_name)

	if !isPresent(des.bitmap_, symbol.Id) {
		return nil
	}
	if val, exists := des.cache_[symbol.Id]; exists {
		return val.(*SinglePrimitiveContainer).GetVal()
	}
	offset := des.calcOffset(symbol.Id)

	if symbol.SymbolType == TYPE_NESTED_MESSAGE {
		des := NewDeserialiser(symbol.NestedType.Name, des.buf_, 8, offset)
		if des.bitmap_ == masktype_t(0) {
			return nil
		}
		return des
	}
	size := symbol.SymbolType.TypeToSize()
	//log.Printf("INT64 %v %v\n", offset, size)
	//log.Printf("Bytes: %v\n", (*des.buf_)[des.buf_start_:des.buf_len_])
	val := getPrimitiveValFromByteBuffer(des.buf_, offset, size, symbol.SymbolType)
	des.cache_[symbol.Id] = NewSinglePrimitiveContainer(symbol, val)
	return val
}

func (des *Deserialiser) GetSize(msg_name string) uint32 {
	symbol, _ := des.message_symbol_.getMemberSymbolAndIdx(msg_name)

	if !isPresent(des.bitmap_, symbol.Id) {
		return 0
	}

	val, exists := des.cache_[symbol.Id]
	if !exists {
		des.calcOffset(symbol.Id)
		val, _ = des.cache_[symbol.Id]
	}
	return val.(*RepeatedPrimitiveContainer).Size()
}

func (des *Deserialiser) GetValAt(msg_name string, idx uint32) interface{} {
	symbol, _ := des.message_symbol_.getMemberSymbolAndIdx(msg_name)

	if !isPresent(des.bitmap_, symbol.Id) {
		return 0
	}
	val, exists := des.cache_[symbol.Id]
	if !exists {
		des.calcOffset(symbol.Id)
		val = des.cache_[symbol.Id]
	}
	return val.(*RepeatedPrimitiveContainer).GetValAt(idx)
}

// Single, Complex
func (des *Deserialiser) GetStrLen(msg_name string) uint32 {
	symbol, _ := des.message_symbol_.getMemberSymbolAndIdx(msg_name)

	if !isPresent(des.bitmap_, symbol.Id) {
		return 0
	}
	val, exists := des.cache_[symbol.Id]
	if !exists {
		des.calcOffset(symbol.Id)
		val = des.cache_[symbol.Id]
	}
	return val.(*SingleComplexContainer).GetStrLen()
}

func (des *Deserialiser) GetStr(msg_name string) *string {
	symbol, _ := des.message_symbol_.getMemberSymbolAndIdx(msg_name)

	if !isPresent(des.bitmap_, symbol.Id) {
		return nil
	}
	val, exists := des.cache_[symbol.Id]
	if !exists {
		des.calcOffset(symbol.Id)
		val = des.cache_[symbol.Id]
	}
	return val.(*SingleComplexContainer).GetVal()
}

func getPrimitiveValFromByteBuffer(bytes *[]byte, offset, size uint64, usertype UserType) interface{} {
	switch usertype {
	case TYPE_BOOL:
		single_byte := (*bytes)[offset : offset+size]
		return single_byte[0] == 0
	case TYPE_INT32:
		return int32(binary.LittleEndian.Uint32((*bytes)[offset : offset+size]))
	case TYPE_UINT32:
		return binary.LittleEndian.Uint32((*bytes)[offset : offset+size])
	case TYPE_INT64:
		return int64(binary.LittleEndian.Uint64((*bytes)[offset : offset+size]))
	case TYPE_UINT64:
		return binary.LittleEndian.Uint64((*bytes)[offset : offset+size])
	case TYPE_FLOAT32:
		bits := binary.LittleEndian.Uint32((*bytes)[offset : offset+size])
		return math.Float32frombits(bits)
	case TYPE_FLOAT64:
		bits := binary.LittleEndian.Uint64((*bytes)[offset : offset+size])
		return math.Float64frombits(bits)
	}
	log.Fatalf("Invalid type. Type: %v\n", usertype.TypeToStr())
	return 0
}

func (des *Deserialiser) calcOffset(id uint32) uint64 {
	offset := uint64(des.buf_start_) + BITMAP_SIZE
	for id_iter := uint32(1); id_iter <= MAX_ALLOWED_MEMBERS; id_iter++ {
		if !isPresent(des.bitmap_, id_iter) {
			continue
		}
		mem_sym := des.message_symbol_.getSymbolFromId(id_iter)
		if !mem_sym.SymbolType.IsComplexType() && !mem_sym.Repeated {
			if id_iter == id {
				return offset
			}
			offset += mem_sym.SymbolType.TypeToSize()
		}
		if !mem_sym.SymbolType.IsComplexType() && mem_sym.Repeated {
			con := NewRepeatedPrimitiveContainer(mem_sym, des.buf_, offset)
			des.cache_[id_iter] = con
			offset += con.EndDistanceInBytes()
			if id_iter == id {
				return offset
			}
		}
		if mem_sym.SymbolType.IsComplexType() && !mem_sym.Repeated {
			con := NewSingleComplexContainer(mem_sym, des.buf_, offset)
			des.cache_[id_iter] = con
			offset += con.EndDistanceInBytes()
			if id_iter == id {
				return offset
			}
		}
	}
	return offset
}
