package user

import "log"

type SymbolTable struct {
	O             interface{}
	msg_to_id_map map[string]uint32
	id_to_msg     map[uint32]string
	id_to_symbol  map[uint32]*Symbol
}

var sym_tab_intance *SymbolTable = nil

func InitSymTab() *SymbolTable {
	if sym_tab_intance == nil {
		sym_tab_intance = &SymbolTable{msg_to_id_map: make(map[string]uint32), id_to_msg: make(map[uint32]string), id_to_symbol: make(map[uint32]*Symbol)}
	}
	return sym_tab_intance
}

func GetSymTab() *SymbolTable {
	return sym_tab_intance
}

func (sym_tab *SymbolTable) getMessageId(msg string) uint32 {
	return sym_tab.msg_to_id_map[msg]
}

func (sym_tab *SymbolTable) getSymbolFromId(id uint32) *Symbol {
	return sym_tab.id_to_symbol[id]
}

func (sym_tab *SymbolTable) getSymbolFromName(name string) *Symbol {
	return sym_tab.getSymbolFromId(sym_tab.getMessageId(name))
}

func (sym_tab *SymbolTable) addMessageDecl(symbol *Symbol) bool {
	if !symbol.IsDeclaration {
		log.Fatalf("symbol must be a declaration. symbol: %v", symbol)
	}
	if sym, ok := sym_tab.msg_to_id_map[symbol.Name]; ok {
		log.Fatalf("Symbol is already present in the map. PresentSym: %v, sym: %v", sym, symbol)
	}
	if sym, ok := sym_tab.id_to_msg[uint32(symbol.Id)]; ok {
		log.Fatalf("Symbol is already present in the map. PresentSym: %v, sym: %v", sym, symbol)
	}

	sym_tab.msg_to_id_map[symbol.Name] = symbol.Id
	sym_tab.id_to_msg[symbol.Id] = symbol.Name
	sym_tab.id_to_symbol[symbol.Id] = symbol

	return true
}
