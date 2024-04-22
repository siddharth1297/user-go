package user

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"testing"
)

func TestSL1(t *testing.T) {
	// prepare message
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sym_tab := InitSymTab()
	/*
		   message M1 {
		       uint32 a = 1;
		       repeated float32 b = 2;
			   M1 c = 3;
		   }
	*/
	m1 := NewDeclarationSymbol("M1")
	sym_tab.addMessageDecl(m1)
	m1.addMember(NewMemberSymbol(1, "a", UserType(TYPE_INT32), m1, nil, false, false))
	m1.addMember(NewMemberSymbol(2, "b", UserType(TYPE_FLOAT32), m1, nil, true, false))
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
		///////
		msg_m1c := AllocateMessage("M1")
		msg_m1c.SetVal("a", int32(-40))

		msg_m1 := AllocateMessage("M1")
		msg_m1.SetVal("a", int32(-20))
		msg_m1.SetVal("c", msg_m1c)

		iovs := msg_m1.SerialisaeToSGBuf(0)
		n, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", n)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}
		//log.Printf("Client read %v bytes.\n", n)
		//log.Printf("BUFffffff: %v\n", buf[0:n])
		deserialiser := AllocateDeserialiser("M1", &buf, uint64(n), 0)
		a := deserialiser.GetVal("a")
		fmt.Printf("%v\n", a.(int32))

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
		arr := []int{1, 2, 3, 4, 5, 10, 9, 8, 7, 6}

		msg := AllocateMessage("M1")
		p := msg
		for i := range arr {
			m := AllocateMessage("M1")
			p.SetVal("a", int32(arr[i]))
			log.Println("Done a")
			p.SetVal("b", m)
			log.Println("Done b")
			p = m
		}
		iovs := msg.SerialisaeToSGBuf(0)
		n, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", n)

		/*
				msg_m1c := AllocateMessage("M1")
				msg_m1c.SetVal("a", int32(-30))

				msg_m1 := AllocateMessage("M1")
				msg_m1.SetVal("a", int32(20))
				msg_m1.SetVal("b", msg_m1c)


			iovs := msg_m1.SerialisaeToSGBuf(0)
			n, _ := conn.WritevRaw(iovs)
			log.Printf("Server writev %v byten", n)
		*/
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}
		//log.Printf("Client read %v bytes.\n", n)
		/*
			log.Printf("BUFffffff: %v\n", buf[0:n])
			m1_des := AllocateDeserialiser("M1", &buf, uint64(n), 0)
			m1_a := m1_des.getSinglePrimitive("a")
			m1_b := m1_des.getSinglePrimitive("b")
			fmt.Printf("a:: %v \n", m1_a.(int32))
			m1_b_a := m1_b.(*Deserialiser).getSinglePrimitive("a")
			m1_b_b := m1_b.(*Deserialiser).getSinglePrimitive("b")
			fmt.Printf("ba: %v\n", m1_b_a)
			fmt.Printf("ba: %v\n", m1_b_b == nil)
		*/
		des := AllocateDeserialiser("M1", &buf, uint64(n), 0)

		for {
			a := des.GetVal("a").(int32)
			b := des.GetVal("b")

			fmt.Printf(">> %v \n", a)

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
			   repeated float64 b = 2;
		   }
	*/
	m1 := NewDeclarationSymbol("M1")
	sym_tab.addMessageDecl(m1)
	m1.addMember(NewMemberSymbol(1, "a", UserType(TYPE_INT32), m1, nil, true, false))
	m1.addMember(NewMemberSymbol(2, "b", UserType(TYPE_FLOAT64), m1, nil, true, false))
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
			//msg.AddVal("a", int32(i))
			msg.AddVal("b", float64(10))
		}
		iovs := msg.SerialisaeToSGBuf(0)
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
		point1 := AllocateMessage("Point")
		point1.SetVal("X", uint64(1))
		point1.SetVal("Y", uint64(2))

		points := AllocateMessage("Points")
		points.AddVal("points", point1)

		iovs := points.SerialisaeToSGBuf(0)
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
		size_points := points_des.GetSize("a")
		log.Printf("points size: %v \n", size_points)
		//size_b := des.GetSize("b")

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
		data := "abcd de"
		m1 := AllocateMessage("M1")
		m1.SetValStr("x", &data, uint32(len(data)))

		iovs := m1.SerialisaeToSGBuf(0)
		n, _ := conn.WritevRaw(iovs)
		log.Printf("Server writev %v byten", n)
	}

	{
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			log.Fatalf("Error In reading  at client. Err: %s\n", err.Error())
		}
		m1 := AllocateDeserialiser("M1", &buf, uint64(n), 0)
		str_len := m1.GetStrLen("a")
		str := m1.GetStr("a")
		log.Printf("strig len: %v \"%v\"\n", str_len, *str)
		//size_b := des.GetSize("b")

	}
}
