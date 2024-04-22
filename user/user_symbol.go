package user

import (
	"fmt"
	"log"
)

const (
	DEFAULT_ID = uint32(0)
)

var declared_message_id = uint32(0)

type Symbol struct {
	IsDeclaration    bool              // Message declaration
	Id               uint32            // Id of the symbol
	Name             string            // Name of the symbol
	Repeated         bool              // Is it repeated
	Required         bool              // Is it required
	ParentSymbol     *Symbol           // Pointer to parent symbol. Parent symbol IsMessage must be true
	SymbolType       UserType          // Type of the symbol
	NestedType       *Symbol           // If it is a nested type, points to the Symbol
	Members          []*Symbol         // Members of the message. Non-nil in case of declration(IsMessage=1)
	MemberToIdMap    map[string]uint32 // Map contains member name to id of the member
	MemberIdToIdxMap map[uint32]uint32 // Map contains member id to index in the Members array
	MaxId            uint32            // Maximum id present in this Message
}

func (symbol Symbol) String() string {
	if !symbol.IsDeclaration {
		repeated := ""
		if symbol.Repeated {
			repeated = "repeated "
		}
		required := ""
		if symbol.Required {
			required = "required "
		}
		symbol_type_str := symbol.SymbolType.TypeToStr()
		if symbol.SymbolType == UserType(TYPE_NESTED_MESSAGE) {
			symbol_type_str = symbol.NestedType.Name
		}
		return required + repeated + fmt.Sprintf("%v %v = %v;", symbol_type_str, symbol.Name, symbol.Id)
	}
	members_str := ""
	for _, member := range symbol.Members {
		members_str += fmt.Sprintf("\t%v\n", member)
	}
	return fmt.Sprintf("message %v {\n", symbol.Name) + members_str + "}"
}

func (sym *Symbol) IsEqual(other *Symbol) bool {
	return sym.IsDeclaration == other.IsDeclaration && sym.Id == other.Id && sym.Name == other.Name
}

func (sym *Symbol) ValidateSymbol() {
	if sym.IsDeclaration {
		if /*sym.Id != DEFAULT_ID ||*/ sym.Name == "" || sym.ParentSymbol != nil || sym.SymbolType != TYPE_MESSAGE_DECL || sym.Members == nil || sym.MemberToIdMap == nil {
			log.Fatalf("Invalid Symbol %v => %v %v %v %v %v %v", sym.Name, sym.Id != DEFAULT_ID, sym.Name == "", sym.ParentSymbol != nil, sym.SymbolType != TYPE_MESSAGE_DECL, sym.Members == nil, sym.MemberToIdMap == nil)
		}
	} else {
		if sym.Id == DEFAULT_ID || sym.Name == "" || sym.ParentSymbol == nil || sym.SymbolType == TYPE_MESSAGE_DECL || sym.Members != nil || sym.MemberToIdMap != nil {
			log.Fatalf("Invalid Symbol %v", sym)
		}
		if sym.SymbolType == TYPE_NESTED_MESSAGE && sym.NestedType == nil {
			log.Fatalf("Invalid Symbol %v", sym)
		}
	}
}

// Use it for creating a symbol for declaration of a message
func NewDeclarationSymbol(name string) *Symbol {
	declared_message_id++
	symbol := &Symbol{IsDeclaration: true, Id: declared_message_id, Name: name, Repeated: false, Required: false, ParentSymbol: nil, SymbolType: TYPE_MESSAGE_DECL, NestedType: nil,
		Members: make([]*Symbol, 0), MemberToIdMap: make(map[string]uint32), MemberIdToIdxMap: make(map[uint32]uint32), MaxId: DEFAULT_ID}
	return symbol
}

// use it for creating a symbol which is a member of a declared message
func NewMemberSymbol(id uint32, name string, symboltype UserType, parent *Symbol, nestedtype *Symbol, repeated bool, required bool) *Symbol {
	symbol := &Symbol{IsDeclaration: false, Id: id, Name: name, Repeated: repeated, Required: required, ParentSymbol: parent, SymbolType: symboltype,
		NestedType: nestedtype, Members: nil, MemberToIdMap: nil, MemberIdToIdxMap: nil, MaxId: DEFAULT_ID}
	parent.ValidateSymbol()
	return symbol
}

func (sym *Symbol) addMember(member *Symbol) {
	if member.IsDeclaration {
		log.Fatalf("member is expected. Got declaration. member: %v\n", member)
	}
	if !sym.IsDeclaration {
		log.Fatalf("Parent symbol must be a declaration. Parent: %v\n", sym)
	}
	// Check if same name or same id already exists
	for _, existing_member := range sym.Members {
		if member.Id == existing_member.Id {
			log.Fatalf("Member with same ID exists. member: %v, exisitng member: %v\n", member, existing_member)
		}
		if member.Name == existing_member.Name {
			log.Fatalf("Member with same NAME exists. member: %v, exisitng member: %v\n", member, existing_member)
		}
	}
	if len(sym.Members) > 0 && sym.Members[len(sym.Members)-1].Id > member.Id {
		log.Fatalf("Members inside a message must be in increasing order. member: %v\n", member)
	}

	if sym.MaxId < member.Id {
		sym.MaxId = member.Id
	}
	sym.Members = append(sym.Members, member)
	sym.MemberToIdMap[member.Name] = member.Id
	sym.MemberIdToIdxMap[member.Id] = uint32(len(sym.Members)) - 1
}

func (sym *Symbol) getMemberIdAndIdx(member_name string) (uint32, uint32) {
	id := sym.MemberToIdMap[member_name]
	return id, sym.MemberIdToIdxMap[id]
}

func (sym *Symbol) getMemberSymbolAndIdx(member_name string) (*Symbol, uint32) {
	idx := sym.MemberIdToIdxMap[sym.MemberToIdMap[member_name]]
	return sym.Members[idx], idx
}

func (parent *Symbol) getSymbolFromId(id uint32) *Symbol {
	return parent.Members[parent.MemberIdToIdxMap[id]]
}
