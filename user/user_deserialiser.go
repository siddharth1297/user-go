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
	idx_to_process_ uint32
	parse_offset_   uint64
}

func NewDeserialiser(message_name string, buf *[]byte, buf_len uint64, buf_start uint64) *Deserialiser {
	des := &Deserialiser{message_name_: message_name, message_symbol_: GetSymTab().getSymbolFromName(message_name), buf_: buf, buf_len_: buf_len, buf_start_: buf_start, bitmap_: 0, cache_: make(map[uint32]DeserialContainers), idx_to_process_: 1, parse_offset_: 0}

	if des.message_symbol_ == nil {
		log.Fatalf("Symbol not found for deserialisation. message name: %v\n", message_name)
	}
	log.Printf("-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=Creating NewDeserialiser for \"%v\"\n", message_name)
	des.parse_offset_ = buf_start
	log.Printf("BUF LEN: %v\n", des.buf_len_)
	// set bitmap
	//des.bitmap_ = masktype_t(binary.LittleEndian.Uint32((*des.buf_)[des.buf_start_ : des.buf_start_+BITMAP_SIZE]))
	bitmapval := getPrimitiveValFromByteBuffer(des.buf_, des.buf_start_, BITMAP_SIZE, TYPE_UINT32)
	des.bitmap_ = masktype_t(bitmapval.(uint32))
	des.parse_offset_ += BITMAP_SIZE
	log.Printf("BITMAP: %v\n", des.bitmap_)
	log.Println(buf_len, " ", buf_start)
	if des.bitmap_ == masktype_t(0) {
		des = nil
	}

	return des
}

func (des *Deserialiser) GetVal(msg_name string) interface{} {
	symbol, idx := des.message_symbol_.getMemberSymbolAndIdx(msg_name)

	if !isPresent(des.bitmap_, symbol.Id) {
		return nil
	}
	val, exists := des.cache_[symbol.Id]
	log.Println("Starting: ", symbol, " ", exists)
	if !exists {
		des.calcOffset(symbol.Id)
		log.Println("Calculated: ", symbol)
		val = des.cache_[symbol.Id]
	}
	log.Println("Decodign: ", symbol)
	if des.message_symbol_.Members[idx].SymbolType == UserType(TYPE_NESTED_MESSAGE) {
		//return val.(*SingleComplexContainer).GetNestedMessage()
		a, ok := val.(*SingleComplexContainer)
		if !ok {
			return nil
		}
		//log.Println(msg_name, " A:: ", a)
		return a.GetNestedMessage()
	}
	if des.message_symbol_.Members[idx].SymbolType.IsComplexType() {
		//return val.(*SingleComplexContainer).GetNestedMessage()
		a := val.(*SingleComplexContainer)
		log.Println(msg_name, " A:: ", a)
		return a
	}
	log.Println("DecodignExit: ", symbol)
	return val.(*SinglePrimitiveContainer).GetVal()
}

func (des *Deserialiser) GetSize(msg_name string) uint32 {
	symbol, _ := des.message_symbol_.getMemberSymbolAndIdx(msg_name)

	if !symbol.Repeated {
		log.Fatalf("InvalidAPI. GetSize is only for repeated fields. symbol: %s\n", symbol)
	}
	if !isPresent(des.bitmap_, symbol.Id) {
		return 0
	}

	val, exists := des.cache_[symbol.Id]
	if !exists {
		log.Printf("Not in cache. Computing\n")
		des.calcOffset(symbol.Id)
		val = des.cache_[symbol.Id]
	}

	if symbol.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
		return val.(*RepeatedComplexContainer).Size()
	}
	if symbol.SymbolType.IsComplexType() {
		return val.(*RepeatedComplexContainer).Size()
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
	if symbol.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
		return val.(*RepeatedComplexContainer).GetNestedMessageAt(idx)
	}
	if symbol.SymbolType.IsComplexType() {
		return val.(*RepeatedComplexContainer).GetValAt(idx)
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

func (des *Deserialiser) GetStrAt(msg_name string, idx uint32) *string {
	symbol, _ := des.message_symbol_.getMemberSymbolAndIdx(msg_name)

	if !isPresent(des.bitmap_, symbol.Id) {
		return nil
	}
	val, exists := des.cache_[symbol.Id]
	if !exists {
		des.calcOffset(symbol.Id)
		val = des.cache_[symbol.Id]
	}
	return val.(*RepeatedComplexContainer).GetValAt(idx)
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
		log.Println("[[[[[[[[[[[[[[[[UINT64]]]]]]]]]]]]]]]] ", offset, " ", size)
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

// The id must be present.
func (des *Deserialiser) calcOffset(id uint32) uint64 {
	log.Println("==========================>>>>>>>>>>Inside calcOffset ", des.idx_to_process_, " ", id)
	//for id_iter := uint32(1); id_iter <= MAX_ALLOWED_MEMBERS; id_iter++ {
	for ; des.idx_to_process_ <= id; des.idx_to_process_++ {
		if !isPresent(des.bitmap_, des.idx_to_process_) {
			continue
		}
		mem_sym := des.message_symbol_.getSymbolFromId(des.idx_to_process_)
		log.Println("Symbol: ", mem_sym)
		if mem_sym.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
			if !mem_sym.Repeated {
				log.Println("Single-complex-Nested ", des.idx_to_process_, " ", id, mem_sym)
				con := NewSingleNestedContainer(mem_sym, des.buf_, uint32(des.buf_len_)-uint32(des.parse_offset_), des.parse_offset_)
				if con != nil {
					des.cache_[des.idx_to_process_] = con
					des.parse_offset_ += uint64(con.TotalLenInBuffer())
				} else {
					des.parse_offset_ += 4
				}

				log.Println("Single-complex-Nested END", des.idx_to_process_, " ", id, mem_sym)
			} else {
				log.Printf("Repeated-complex-Nested for %v\n", mem_sym)
				con := NewRepeatedNestedContainer(mem_sym, des.buf_, uint32(des.buf_len_)-uint32(des.parse_offset_), des.parse_offset_)
				des.cache_[des.idx_to_process_] = con
				des.parse_offset_ += uint64(con.TotalLenInBuffer())
			}
			if des.idx_to_process_ == id {
				break
			}
		}

		if !mem_sym.SymbolType.IsComplexType() && mem_sym.SymbolType != UserType(TYPE_NESTED_MESSAGE) && !mem_sym.Repeated {
			log.Println("Single-simple ", mem_sym)
			con := NewSinglePrimitiveContainer(mem_sym, des.buf_, des.parse_offset_)
			des.cache_[des.idx_to_process_] = con
			des.parse_offset_ += uint64(con.TotalLenInBuffer())
			if des.idx_to_process_ == id {
				break
			}
		}

		if !mem_sym.SymbolType.IsComplexType() && mem_sym.SymbolType != UserType(TYPE_NESTED_MESSAGE) && mem_sym.Repeated {
			log.Println("repeated-simple")
			con := NewRepeatedPrimitiveContainer(mem_sym, des.buf_, des.parse_offset_)
			des.cache_[des.idx_to_process_] = con
			des.parse_offset_ += uint64(con.TotalLenInBuffer())
			if des.idx_to_process_ == id {
				break
			}
		}

		if mem_sym.SymbolType.IsComplexType() && mem_sym.SymbolType != UserType(TYPE_NESTED_MESSAGE) && !mem_sym.Repeated {
			log.Println("Single-complex ", des.idx_to_process_, " ", id, " ", des.message_symbol_, mem_sym)
			con := NewSingleComplexContainer(mem_sym, des.buf_, des.parse_offset_)
			des.cache_[des.idx_to_process_] = con
			des.parse_offset_ += uint64(con.TotalLenInBuffer())
			if des.idx_to_process_ == id {
				break
			}
		}

		if mem_sym.SymbolType.IsComplexType() && mem_sym.SymbolType != UserType(TYPE_NESTED_MESSAGE) && mem_sym.Repeated {
			log.Println("Repeated-complex")
			con := NewRepeatedComplexContainer(mem_sym, des.buf_, des.parse_offset_)
			des.cache_[des.idx_to_process_] = con
			des.parse_offset_ += uint64(con.TotalLenInBuffer())
			if des.idx_to_process_ == id {
				break
			}
		}
		log.Println("PASSSSS")
	}
	// Control may reach here because of break. So increment it here.
	des.idx_to_process_++
	return des.parse_offset_
}

func (des *Deserialiser) DeserialiseAllMember() {
	//log.Fatalf("DESERIALISING")
	des.calcOffset(MAX_ALLOWED_MEMBERS)
}

func (des *Deserialiser) TotalLenInBuffer() uint32 {
	des.DeserialiseAllMember()
	return uint32(des.parse_offset_) - uint32(des.buf_start_)
}
