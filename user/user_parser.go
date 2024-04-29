package user

import (
	"bufio"
	"log"
	"os"

	"strconv"
	"strings"
)

var MsgsMap = make(map[string]*Symbol)

func StrToType(typeName string) (UserType) {
	switch typeName {
	case "bool":
		return TYPE_BOOL
	case "int32":
		return TYPE_INT32
	case "uint32":
		return TYPE_UINT32
	case "int64":
		return TYPE_INT64
	case "uint64":
		return TYPE_UINT64
	case "float32":
		return TYPE_FLOAT32
	case "float64":
		return TYPE_FLOAT64
	case "string":
		return TYPE_STRING
	case "byte":
		return TYPE_BYTE
	default:
		_, ok := MsgsMap[typeName]; if ok {
			return TYPE_NESTED_MESSAGE
		} else {
			return TYPE_UNKNOWN
		}
	}
}

func ParseProtoFile(fileName string) ([]*Symbol, error) {
	var allMsgs []*Symbol

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var currMsg *Symbol = nil
	inMsg := false

	for scanner.Scan() {
		wordsList := []string{}
		line := scanner.Text()
		line = strings.TrimSpace(line)
		words := strings.Split(line, " ")
		wordsList = append(wordsList, words...)

		if wordsList[0] == "message" {
			currMsg = NewDeclarationSymbol(wordsList[1])
			inMsg = true
		} else if wordsList[0] == "}" {
			allMsgs = append(allMsgs, currMsg)
			MsgsMap[currMsg.Name] = currMsg
			inMsg = false
			currMsg = nil
		} else if inMsg && len(wordsList) > 0{
			if wordsList[0] == "//" || wordsList[0] == "/*" {
				continue
			}	

			idx := 0
			var id uint32
			var name string
			var symboltype UserType
			var nestedtype *Symbol = nil
			isRepeated := false
			var required bool = false

			if wordsList[0] == "repeated" {
				isRepeated = true
				idx++
			} 

			typeName := wordsList[idx]
			symboltype = StrToType(typeName)
			name = wordsList[idx+1]
			str := wordsList[idx+3]
			idStr := str[:len(str)-1]
			idInt, err := strconv.Atoi(idStr)
			if err != nil {
				log.Fatalf("Error converting ID to int: %v", err)
				return nil, err
			}
			id = uint32(idInt)
			
			if symboltype == TYPE_NESTED_MESSAGE {
				nestedtype = MsgsMap[typeName]
			}
			field := NewMemberSymbol(id, name, symboltype, currMsg, nestedtype, isRepeated, required)
			currMsg.addMember(field)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
		return nil, err
	}

	return allMsgs, nil
}

// func PrintProtoMsg