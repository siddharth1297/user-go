package user

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"testing"
)

func TestSL0(t *testing.T) {
	// prepare message
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
	   message M1 {
	       int32 a = 1;
	       float b = 2;
	   }
	*/
	m1 := NewDeclarationSymbol("M1")
	sym_tab.addMessageDecl(m1)
	m1.addMember(NewMemberSymbol(1, "a", UserType(TYPE_INT32), m1, nil, false, false))
	m1.addMember(NewMemberSymbol(2, "b", UserType(TYPE_FLOAT), m1, nil, false, false))
	fmt.Println(m1)

	port := 8087
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	conn := server.Accept()

	{
		msg_m1 := AllocateMessage("M1")
		msg_m1.SetVal("a", int32(-20))
		msg_m1.SetVal("b", float32(-21.4))

		iovs := msg_m1.InitAndSerialiseToSGBuf(0)
		n, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", n)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}
		m1_des := AllocateDeserialiser("M1", &buf, uint64(n), 0)
		a := m1_des.GetVal("a")
		b := m1_des.GetVal("b")
		log.Printf("%v\n", a.(int32))
		log.Printf("%v\n", b.(float32))
	}
}

func TestSL1(t *testing.T) {
	// prepare message
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
		   message M1 {
		       uint32 a = 1;
		       repeated float b = 2;
			   M1 c = 3;
		   }
	*/
	m1 := NewDeclarationSymbol("M1")
	sym_tab.addMessageDecl(m1)
	m1.addMember(NewMemberSymbol(1, "a", UserType(TYPE_INT32), m1, nil, false, false))
	m1.addMember(NewMemberSymbol(2, "b", UserType(TYPE_FLOAT), m1, nil, true, false))
	m1.addMember(NewMemberSymbol(3, "c", UserType(TYPE_NESTED_MESSAGE), m1, m1, false, false))
	fmt.Println(m1)

	port := 8087
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	conn := server.Accept()

	{
		msg_m1 := AllocateMessage("M1")
		msg_m1.SetVal("a", int32(-20))

		msg_m1c := AllocateMessage("M1")
		msg_m1c.SetVal("a", int32(-40))

		msg_m1.SetNestedMsg("c", msg_m1c)

		iovs := msg_m1.InitAndSerialiseToSGBuf(0)
		n, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", n)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}
		m1_des := AllocateDeserialiser("M1", &buf, uint64(n), 0)
		a := m1_des.GetVal("a")
		log.Printf("%v\n", a.(int32)) // -20

		m1c := m1_des.GetVal("c")
		m1c_des := m1c.(*Deserialiser)
		m1c_a := m1c_des.GetVal("a")
		log.Printf("m1_c_a: %v\n", m1c_a.(int32))

		m1c_a_c := m1c_des.GetVal("c")
		log.Println(m1c_a_c)
	}
}

func TestSL2(t *testing.T) {
	// prepare message
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
		   message M1 {
		       int32 a = 1;
			   M1 b = 2;
		   }
	*/
	m1 := NewDeclarationSymbol("M1")
	sym_tab.addMessageDecl(m1)
	m1.addMember(NewMemberSymbol(1, "a", UserType(TYPE_INT32), m1, nil, false, false))
	m1.addMember(NewMemberSymbol(2, "b", UserType(TYPE_NESTED_MESSAGE), m1, m1, false, false))
	fmt.Println(m1)

	port := 8087
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	conn := server.Accept()

	{
		//arr := []int{1, 2, 3, 4, 5, 10, 9, 8, 7, 6}
		arr := []int{1, 2}

		msg := AllocateMessage("M1")
		p := msg
		for i := range arr {
			m := AllocateMessage("M1")
			p.SetVal("a", int32(arr[i]))
			p.SetNestedMsg("b", m)
			p = m
		}
		iovs := msg.InitAndSerialiseToSGBuf(0)
		n, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", n)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}

		des := AllocateDeserialiser("M1", &buf, uint64(n), 0)

		for {
			a := des.GetVal("a").(int32)
			fmt.Printf(">> %v \n", a)
			log.Println("Getting......b")
			b := des.GetVal("b")
			if b == nil {
				break
			}
			des = b.(*Deserialiser)
		}
	}
}

func TestSL3(t *testing.T) {
	// prepare message
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
		   message M1 {
		       repeated int32 a = 1;
			   repeated double b = 2;
		   }
	*/
	m1 := NewDeclarationSymbol("M1")
	sym_tab.addMessageDecl(m1)
	m1.addMember(NewMemberSymbol(1, "a", UserType(TYPE_INT32), m1, nil, true, false))
	m1.addMember(NewMemberSymbol(2, "b", UserType(TYPE_DOUBLE), m1, nil, true, false))
	fmt.Println(m1)

	port := 8087
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	conn := server.Accept()

	{
		msg := AllocateMessage("M1")
		for i := 0; i < 10; i++ {
			msg.AddVal("a", int32(i))
			msg.AddVal("b", float64(i)/float64(2))
		}
		iovs := msg.InitAndSerialiseToSGBuf(0)
		n, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", n)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}
		des := AllocateDeserialiser("M1", &buf, uint64(n), 0)
		size_a := des.GetSize("a")
		size_b := des.GetSize("b")
		log.Printf("Size a: %v b: %v\n", size_a, size_b)
		for i := uint32(0); i < size_a; i++ {
			if i == 0 {
				fmt.Printf("\na: [")
			}
			a_i := des.GetValAt("a", (i))
			fmt.Printf("%v ", a_i.(int32))

			if i == size_a-1 {
				fmt.Printf("]\n")
			}
		}

		for i := uint32(0); i < size_b; i++ {
			if i == 0 {
				fmt.Printf("\nb: [")
			}
			b_i := des.GetValAt("b", (i))
			fmt.Printf("%v ", b_i.(float64))

			if i == size_b-1 {
				fmt.Printf("]\n")
			}
		}
	}
}

func TestSL4(t *testing.T) {
	// prepare message
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
			message Point {
				uint64 X = 1;
				uint64 Y = 2;
			}
		   message Points {
		       repeated Point points = 1;
		   }
	*/
	point := NewDeclarationSymbol("Point")
	sym_tab.addMessageDecl(point)
	point.addMember(NewMemberSymbol(1, "X", UserType(TYPE_UINT64), point, nil, false, false))
	point.addMember(NewMemberSymbol(2, "Y", UserType(TYPE_UINT64), point, nil, false, false))
	fmt.Println(point)

	points_msg := NewDeclarationSymbol("Points")
	sym_tab.addMessageDecl(points_msg)
	points_msg.addMember(NewMemberSymbol(1, "points", UserType(TYPE_NESTED_MESSAGE), points_msg, point, true, false))
	fmt.Println(points_msg)

	port := 8087
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	conn := server.Accept()

	{
		point0 := AllocateMessage("Point")
		point0.SetVal("X", uint64(1))
		point0.SetVal("Y", uint64(2))

		point1 := AllocateMessage("Point")
		point1.SetVal("X", uint64(3))
		point1.SetVal("Y", uint64(4))

		points := AllocateMessage("Points")
		points.AddNestedMsg("points", point0)
		points.AddNestedMsg("points", point1)

		iovs := points.InitAndSerialiseToSGBuf(0)
		n, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", n)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}
		points_des := AllocateDeserialiser("Points", &buf, uint64(n), 0)
		size_points := points_des.GetSize("points")
		log.Printf("points size: %v \n", size_points)

		for i := uint32(0); i < size_points; i++ {
			p := points_des.GetValAt("points", i)
			p_des := p.(*Deserialiser)
			p_x := p_des.GetVal("X")
			log.Printf(">>point%v : %v\n", i, p_x.(uint64))
			p_y := p_des.GetVal("Y")
			log.Printf(">>point%v : %v\n", i, p_y.(uint64))
			log.Printf(">>point%v : %v %v\n", i, p_x.(uint64), p_y.(uint64))

		}
	}
}

func TestSL5(t *testing.T) {
	// prepare message
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
	   message M1 {
	       byte x = 1;
	   }
	*/
	m1 := NewDeclarationSymbol("M1")
	sym_tab.addMessageDecl(m1)
	m1.addMember(NewMemberSymbol(1, "x", UserType(TYPE_BYTE), m1, nil, false, false))
	fmt.Println(m1)

	port := 8087
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	conn := server.Accept()

	{
		data := "abcdff ddfghh"
		m1_msg := AllocateMessage("M1")
		m1_msg.SetValStr("x", &data, uint32(len(data)))

		iovs := m1_msg.InitAndSerialiseToSGBuf(0)
		bytes, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", bytes)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}

		m1 := AllocateDeserialiser("M1", &buf, uint64(n), 0)
		str_len := m1.GetStrLen("x")
		str := m1.GetStr("x")
		log.Printf("strig len: %v \"%v\"\n", str_len, *str)
	}
}

func TestSL6(t *testing.T) {
	// prepare message
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
		   message M1 {
		       repeated string x = 1;
			   repeated string y = 2
		   }
	*/
	m1 := NewDeclarationSymbol("M1")
	sym_tab.addMessageDecl(m1)
	m1.addMember(NewMemberSymbol(1, "x", UserType(TYPE_STRING), m1, nil, true, false))
	m1.addMember(NewMemberSymbol(2, "y", UserType(TYPE_STRING), m1, nil, true, false))
	fmt.Println(m1)

	port := 8087
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	conn := server.Accept()

	{
		m1_msg := AllocateMessage("M1")

		data := []string{"a1", "a22", "a333"}

		m1_msg.AddValStr("x", &data[0], uint32(len(data[0])))
		m1_msg.AddValStr("x", &data[1], uint32(len(data[1])))
		m1_msg.AddValStr("x", &data[2], uint32(len(data[2])))

		m1_msg.AddValStr("y", &data[2], uint32(len(data[2])))
		m1_msg.AddValStr("y", &data[1], uint32(len(data[1])))
		m1_msg.AddValStr("y", &data[0], uint32(len(data[0])))

		iovs := m1_msg.InitAndSerialiseToSGBuf(0)
		bytes, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", bytes)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}

		m1 := AllocateDeserialiser("M1", &buf, uint64(n), 0)
		size_y := m1.GetSize("y")
		size_x := m1.GetSize("x")

		log.Printf("Size: x: %v y: %v\n", size_x, size_y)
		for i := uint32(0); i < size_x; i++ {
			if i == 0 {
				fmt.Printf("\n[")
			}
			fmt.Printf("\"%v\" ", *m1.GetStrAt("x", i))
			if i == size_x-1 {
				fmt.Printf("]\n")
			}
		}
		for i := uint32(0); i < size_y; i++ {
			if i == 0 {
				fmt.Printf("\n[")
			}
			fmt.Printf("\"%v\" ", *m1.GetStrAt("y", i))
			if i == size_y-1 {
				fmt.Printf("]\n")
			}
		}

	}
}

func TestSL7(t *testing.T) {
	// prepare message
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
		   message Addr {
				string address_line = 1;
		   }
		   message Person {
		    	Addr addresses = 1;
		   }
	*/

	addr := NewDeclarationSymbol("Addr")
	sym_tab.addMessageDecl(addr)
	addr.addMember(NewMemberSymbol(1, "address_line", UserType(TYPE_STRING), addr, nil, false, false))
	fmt.Println(addr)
	person := NewDeclarationSymbol("Person")
	sym_tab.addMessageDecl(person)
	person.addMember(NewMemberSymbol(1, "addresses", UserType(TYPE_NESTED_MESSAGE), person, addr, false, false))

	fmt.Println(person)

	port := 8087
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	conn := server.Accept()

	{
		person := AllocateMessage("Person")

		data := []string{"a1", "a22", "a333"}
		addr1 := AllocateMessage("Addr")
		addr1.SetValStr("address_line", &data[0], uint32(len(data[0])))
		log.Printf("======Sttting nested==========\n")
		person.SetNestedMsg("addresses", addr1)

		iovs := person.InitAndSerialiseToSGBuf(0)
		bytes, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", bytes)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}

		person := AllocateDeserialiser("Person", &buf, uint64(n), 0)
		addr1 := person.GetVal("addresses")
		addr1_des := addr1.(*Deserialiser)
		addr_line1 := addr1_des.GetStr("address_line")
		log.Printf("Address: \"%v\"\n", *addr_line1)
	}
}

func TestSL8(t *testing.T) {
	// prepare message
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
		   message Addr {
				string address_line = 1;
		   }
		   message Person {
		    	repeated Addr addresses = 1;
		   }
	*/

	addr := NewDeclarationSymbol("Addr")
	sym_tab.addMessageDecl(addr)
	addr.addMember(NewMemberSymbol(1, "address_line", UserType(TYPE_STRING), addr, nil, false, false))
	fmt.Println(addr)
	person := NewDeclarationSymbol("Person")
	sym_tab.addMessageDecl(person)
	person.addMember(NewMemberSymbol(1, "addresses", UserType(TYPE_NESTED_MESSAGE), person, addr, true, false))

	fmt.Println(person)

	port := 8087
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	conn := server.Accept()

	{
		person := AllocateMessage("Person")

		data := []string{"a1", "a22", "a333"}
		addr1 := AllocateMessage("Addr")
		addr1.SetValStr("address_line", &data[0], uint32(len(data[0])))
		log.Printf("======Sttting nested==========\n")
		person.AddNestedMsg("addresses", addr1)

		addr2 := AllocateMessage("Addr")
		addr2.SetValStr("address_line", &data[2], uint32(len(data[2])))
		person.AddNestedMsg("addresses", addr2)

		addr3 := AllocateMessage("Addr")
		addr3.SetValStr("address_line", &data[1], uint32(len(data[1])))
		person.AddNestedMsg("addresses", addr3)

		iovs := person.InitAndSerialiseToSGBuf(0)
		bytes, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", bytes)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}

		person := AllocateDeserialiser("Person", &buf, uint64(n), 0)
		size := person.GetSize("addresses")
		log.Printf("Size: %v\n", size)

		for i := uint32(0); i < size; i++ {
			addr := person.GetValAt("addresses", i)
			addr_des := addr.(*Deserialiser)
			addr_line := addr_des.GetStr("address_line")
			log.Printf("Address: \"%v\"\n", *addr_line)
		}
	}
}

func TestSL9(t *testing.T) {
	// prepare message
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
		   message Addr {
				repeated string address_lines = 1;
		   }
		   message Person {
		    	Addr addresses = 1;
		   }
	*/

	addr := NewDeclarationSymbol("Addr")
	sym_tab.addMessageDecl(addr)
	addr.addMember(NewMemberSymbol(1, "address_lines", UserType(TYPE_STRING), addr, nil, true, false))
	fmt.Println(addr)
	person := NewDeclarationSymbol("Person")
	sym_tab.addMessageDecl(person)
	person.addMember(NewMemberSymbol(1, "addresses", UserType(TYPE_NESTED_MESSAGE), person, addr, false, false))

	fmt.Println(person)

	port := 8087
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	conn := server.Accept()

	{
		person := AllocateMessage("Person")

		data := []string{"a1", "a22", "a333"}

		addr1 := AllocateMessage("Addr")
		addr1.AddValStr("address_lines", &data[0], uint32(len(data[0])))
		addr1.AddValStr("address_lines", &data[2], uint32(len(data[2])))
		addr1.AddValStr("address_lines", &data[1], uint32(len(data[1])))
		log.Printf("======Sttting nested==========\n")
		person.SetNestedMsg("addresses", addr1)

		iovs := person.InitAndSerialiseToSGBuf(0)
		bytes, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", bytes)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}

		person := AllocateDeserialiser("Person", &buf, uint64(n), 0)
		addr := person.GetVal("addresses")

		addr_des := addr.(*Deserialiser)
		size := addr_des.GetSize("address_lines")
		log.Printf("Size: %v\n", size)

		for i := uint32(0); i < size; i++ {
			addr := addr_des.GetStrAt("address_lines", i)
			log.Printf("Address: \"%v\"\n", *addr)
		}
	}
}

func TestSL10(t *testing.T) {
	// prepare message
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
		   message Addr {
				repeated string address_lines = 1;
		   }
		   message Person {
		    	repeated Addr addresses = 1;
		   }
	*/

	addr := NewDeclarationSymbol("Addr")
	sym_tab.addMessageDecl(addr)
	addr.addMember(NewMemberSymbol(1, "address_lines", UserType(TYPE_STRING), addr, nil, true, false))
	fmt.Println(addr)
	person := NewDeclarationSymbol("Person")
	sym_tab.addMessageDecl(person)
	person.addMember(NewMemberSymbol(1, "addresses", UserType(TYPE_NESTED_MESSAGE), person, addr, true, false))

	fmt.Println(person)

	port := 8087
	config := TCPServerConfig{Port: port, Backlog: 5, ReuseAddr: true, ReusePort: true}
	server := CreateServerTCP(&config)
	server.StartListen()

	client, err := net.Dial("tcp", "127.0.0.1"+":"+strconv.Itoa(port))

	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	conn := server.Accept()

	{
		person := AllocateMessage("Person")

		person1_addrs := []string{"a11", "a12", "a133"}
		person2_addrs := []string{"a21", "a22", "a233"}

		addr1 := AllocateMessage("Addr")
		addr1.AddValStr("address_lines", &person1_addrs[0], uint32(len(person1_addrs[0])))
		addr1.AddValStr("address_lines", &person1_addrs[2], uint32(len(person1_addrs[2])))
		addr1.AddValStr("address_lines", &person1_addrs[1], uint32(len(person1_addrs[1])))

		addr2 := AllocateMessage("Addr")
		addr2.AddValStr("address_lines", &person2_addrs[0], uint32(len(person2_addrs[0])))
		addr2.AddValStr("address_lines", &person2_addrs[2], uint32(len(person2_addrs[2])))
		addr2.AddValStr("address_lines", &person2_addrs[1], uint32(len(person2_addrs[1])))

		log.Printf("======Sttting nested==========\n")
		person.AddNestedMsg("addresses", addr1)
		person.AddNestedMsg("addresses", addr2)

		iovs := person.InitAndSerialiseToSGBuf(0)
		bytes, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", bytes)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}

		person := AllocateDeserialiser("Person", &buf, uint64(n), 0)
		person_size := person.GetSize("addresses")
		log.Printf("Person size: %v\n", person_size)
		for i := uint32(0); i < person_size; i++ {
			addr := person.GetValAt("addresses", i)
			addr_des := addr.(*Deserialiser)
			addr_size := addr_des.GetSize("address_lines")
			log.Printf("Person%v size: %v\n", i, addr_size)
			for j := uint32(0); j < addr_size; j++ {
				addr := addr_des.GetStrAt("address_lines", j)
				log.Printf("Person%v Address%v: \"%v\"\n", i, j, *addr)
			}
		}
		// addr := person.GetVal("addresses")

		// addr_des := addr.(*Deserialiser)
		// size := addr_des.GetSize("address_lines")
		// log.Printf("Size: %v\n", size)

		// for i := uint32(0); i < size; i++ {
		// 	addr := addr_des.GetStrAt("address_lines", i)
		// 	log.Printf("Address: \"%v\"\n", *addr)
		// }

	}
}
