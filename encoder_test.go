package phpserialize

import (
	"fmt"
	"testing"
)

func TestSerialize(t *testing.T) {
	a1 := A{1, "2"}
	a2 := A{11, "22"}
	b := B{3, "4"}
	c := C{[]int{1, 2, 3}}
	d := D{a1}
	s, _ := Serialize(nil)
	fmt.Println("nil", s)
	if s != "N;" {
		t.Error("nil serialize error")
	}
	s, _ = Serialize(true)
	fmt.Println("bool", s)
	if s != "b:1;" {
		t.Error("bool serialize error")
	}
	s, _ = Serialize(int(1))
	fmt.Println("int", s)
	if s != "i:1;" {
		t.Error("int serialize error")
	}
	s, _ = Serialize(float64(1.23))
	fmt.Println("float", s)
	if s != "d:1.23;" {
		t.Error("float serialize error")
	}
	s, _ = Serialize("qwe")
	fmt.Println("string", s)
	if s != `s:3:"qwe";` {
		t.Error("string serialize error")
	}
	s, _ = Serialize(a1)
	fmt.Println("object", s)
	if s != `O:1:"A":2:{s:3:"One";i:1;s:3:"Two";s:1:"2";}` {
		t.Error("object serialize error")
	}
	s, _ = Serialize([]bool{true, false})
	fmt.Println("array bool", s)
	if s != `a:2:{i:0;b:1;i:1;b:0;}` {
		t.Error("array bool serialize error")
	}
	s, _ = Serialize([]int{11, 12, 13})
	fmt.Println("array int", s)
	if s != `a:3:{i:0;i:11;i:1;i:12;i:2;i:13;}` {
		t.Error("array int serialize error")
	}
	s, _ = Serialize([]float64{0.1, 0.2, 0.3})
	fmt.Println("array float", s)
	if s != `a:3:{i:0;d:0.1;i:1;d:0.2;i:2;d:0.3;}` {
		t.Error("array float serialize error")
	}
	s, _ = Serialize([]string{"a", "b", "c"})
	fmt.Println("array string", s)
	if s != `a:3:{i:0;s:1:"a";i:1;s:1:"b";i:2;s:1:"c";}` {
		t.Error("array string serialize error")
	}
	s, _ = Serialize([]A{a1, a2})
	fmt.Println("array object", s)
	if s != `a:2:{i:0;O:1:"A":2:{s:3:"One";i:1;s:3:"Two";s:1:"2";}i:1;O:1:"A":2:{s:3:"One";i:11;s:3:"Two";s:2:"22";}}` {
		t.Error("array object serialize error")
	}
	s, _ = Serialize([]interface{}{a1, b})
	fmt.Println("array object", s)
	if s != `a:2:{i:0;O:1:"A":2:{s:3:"One";i:1;s:3:"Two";s:1:"2";}i:1;O:1:"B":2:{s:5:"Three";i:3;s:4:"Four";s:1:"4";}}` {
		t.Error("array object serialize error")
	}
	s, _ = Serialize(c)
	fmt.Println("object array", s)
	if s != `O:1:"C":1:{s:6:"ArrayC";a:3:{i:0;i:1;i:1;i:2;i:2;i:3;}}` {
		t.Error("object array serialize error")
	}
	s, _ = Serialize(d)
	fmt.Println("object object", s)
	if s != `O:1:"D":1:{s:7:"ObjectD";O:1:"A":2:{s:3:"One";i:1;s:3:"Two";s:1:"2";}}` {
		t.Error("object object serialize error")
	}
}
