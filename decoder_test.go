package phpserialize

import (
	"fmt"
	"testing"
)

type A struct {
	One int    `phps:"one"`
	Two string `phps:"two"`
}

type B struct {
	Three int
	Four  string
}

type C struct {
	ArrayC []int
}

type D struct {
	ObjectD A
}

type MyBool bool

func TestUnserialize(t *testing.T) {
	var err error

	nilValuePtr := []int{}
	err = Unserialize("N;", &nilValuePtr)
	if err != nil {
		t.Errorf("%v nilValuePtr unserialize error", err)
	} else if nilValuePtr != nil {
		t.Errorf("nilValuePtr unserialize error")
	}
	fmt.Println("nilValuePtr", nilValuePtr)

	boolValuePtr := new(bool)
	err = Unserialize("b:1;", boolValuePtr)
	if err != nil {
		t.Errorf("%v boolValuePtr unserialize error", err)
	} else if *boolValuePtr != true {
		t.Errorf("boolValuePtr unserialize error")
	}
	fmt.Println("boolValuePtr", *boolValuePtr)

	myboolValuePtr := new(MyBool)
	err = Unserialize("b:1;", myboolValuePtr)
	if err != nil {
		t.Errorf("%v myboolValuePtr unserialize error", err)
	} else if *myboolValuePtr != true {
		t.Errorf("myboolValuePtr unserialize error")
	}
	fmt.Println("myboolValuePtr", *myboolValuePtr)

	intValuePtr := new(int)
	err = Unserialize("i:1;", intValuePtr)
	if err != nil {
		t.Errorf("%v intValuePtr unserialize error", err)
	} else if *intValuePtr != 1 {
		t.Errorf("intValuePtr unserialize error")
	}
	fmt.Println("intValuePtr", *intValuePtr)

	err = Unserialize("i:-1;", intValuePtr)
	if err != nil {
		t.Errorf("%v intValuePtr unserialize error", err)
	} else if *intValuePtr != -1 {
		t.Errorf("intValuePtr unserialize error")
	}
	fmt.Println("intValuePtr", *intValuePtr)

	uintValuePtr := new(uint)
	err = Unserialize("i:1;", uintValuePtr)
	if err != nil {
		t.Errorf("%v uintValuePtr unserialize error", err)
	} else if *uintValuePtr != 1 {
		t.Errorf("uintValuePtr unserialize error")
	}
	fmt.Println("uintValuePtr", *uintValuePtr)

	float32ValuePtr := new(float32)
	err = Unserialize("d:1.23;", float32ValuePtr)
	if err != nil {
		t.Errorf("%v float32ValuePtr unserialize error", err)
	} else if *float32ValuePtr != 1.23 {
		t.Errorf("float32ValuePtr unserialize error")
	}
	fmt.Println("float32ValuePtr", *float32ValuePtr)

	float64ValuePtr := new(float64)
	err = Unserialize("d:1.23;", float64ValuePtr)
	if err != nil {
		t.Errorf("%v float64ValuePtr unserialize error", err)
	} else if *float64ValuePtr != 1.23 {
		t.Errorf("float64ValuePtr unserialize error")
	}
	fmt.Println("float64ValuePtr", *float64ValuePtr)

	stringValuePtr := new(string)
	err = Unserialize(`s:3:"qwe";`, stringValuePtr)
	if err != nil {
		t.Errorf("%v stringValuePtr unserialize error", err)
	} else if *stringValuePtr != "qwe" {
		t.Errorf("stringValuePtr unserialize error")
	}
	fmt.Println("stringValuePtr", *stringValuePtr)

	a := A{1, "2"}
	objectAValuePtr := new(A)
	err = Unserialize(`O:1:"A":2:{s:3:"One";i:1;s:3:"Two";s:1:"2";}`, objectAValuePtr)
	if err != nil {
		t.Errorf("%v objectAValuePtr unserialize error", err)
	} else if *objectAValuePtr != a {
		t.Errorf("objectAValuePtr unserialize error")
	}
	fmt.Println("objectAValuePtr", *objectAValuePtr)

	objectAValueMap := make(map[string]interface{})
	err = Unserialize(`O:1:"A":2:{s:3:"One";i:1;s:3:"Two";s:1:"2";}`, &objectAValueMap)
	if err != nil {
		t.Errorf("%v objectAValueMap unserialize error", err)
	}
	fmt.Println("objectAValueMap", objectAValueMap)

	arrayboolValuePtr := []bool{}
	err = Unserialize(`a:2:{i:0;b:1;i:1;b:0;}`, &arrayboolValuePtr)
	if err != nil {
		t.Errorf("%v arrayboolValuePtr unserialize error", err)
	}
	fmt.Println("arrayboolValuePtr", arrayboolValuePtr)

	arrayfloat64ValuePtr := []float64{}
	err = Unserialize(`a:3:{i:0;d:0.1;i:1;d:0.2;i:2;d:0.3;}`, &arrayfloat64ValuePtr)
	if err != nil {
		t.Errorf("%v arrayfloat64ValuePtr unserialize error", err)
	}
	fmt.Println("arrayfloat64ValuePtr", arrayfloat64ValuePtr)

	arrayobjectAValuePtr := []A{}
	err = Unserialize(`a:2:{i:0;O:1:"A":2:{s:3:"One";i:1;s:3:"Two";s:1:"2";}i:1;O:1:"A":2:{s:3:"One";i:11;s:3:"Two";s:2:"22";}}`, &arrayobjectAValuePtr)
	if err != nil {
		t.Errorf("%v arrayobjectAValuePtr unserialize error", err)
	}
	fmt.Println("arrayobjectAValuePtr", arrayobjectAValuePtr)

	/* 	//这个不支持
	   	arrayobjectValuePtr := []interface{}{}
	   	err = Unserialize(`a:2:{i:0;O:1:"A":2:{s:3:"One";i:1;s:3:"Two";s:1:"2";}i:1;O:1:"B":2:{s:5:"Three";i:3;s:4:"Four";s:1:"4";}}`, &arrayobjectValuePtr)
	   	if err != nil {
	   		t.Errorf("%v arrayobjectValuePtr unserialize error", err)
	   	}
	   	fmt.Println("arrayobjectValuePtr", arrayobjectValuePtr) */

	objectDValuePtr := new(D)
	err = Unserialize(`O:1:"D":1:{s:7:"ObjectD";O:1:"A":2:{s:3:"One";i:1;s:3:"Two";s:1:"2";}}`, objectDValuePtr)
	if err != nil {
		t.Errorf("%v objectDValuePtr unserialize error", err)
	}
	fmt.Println("objectDValuePtr", *objectDValuePtr)

}
