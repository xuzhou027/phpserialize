package phpserialize

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
)

func Unserialize(data string, v interface{}) error {
	var d decodeState
	/* 	err := checkValid([]byte(data), &d.scan)
	   	if err != nil {
	   		return err
	   	} */

	d.init([]byte(data))
	return d.unmarshal(v)
}

func (d *decodeState) init(data []byte) *decodeState {
	d.data = data
	d.off = 0
	d.savedError = nil
	d.errorContext.Struct = ""
	d.errorContext.Field = ""
	return d
}

func (d *decodeState) unmarshal(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	// We decode rv not rv.Elem because the Unmarshaler interface
	// test must be applied at the top level of the value.
	d.value(rv)
	return d.savedError
}

// decodeState represents the state while decoding a JSON value.
type decodeState struct {
	data         []byte
	off          int      // read offset in data
	errorContext struct { // provides context for type errors
		Struct string
		Field  string
	}
	savedError            error
	useNumber             bool
	disallowUnknownFields bool
}

type UnmarshalTypeError struct {
	Value  string       // description of JSON value - "bool", "array", "number -5"
	Type   reflect.Type // type of Go value it could not be assigned to
	Offset int64        // error occurred after reading Offset bytes
	Struct string       // name of the struct type containing the field
	Field  string       // name of the field holding the Go value
}

type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "phps: Unserialize(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "phps: Unserialize(non-pointer " + e.Type.String() + ")"
	}
	return "phps: Unserialize(nil " + e.Type.String() + ")"
}

// value decodes a JSON value from d.data[d.off:] into the value.
// it updates d.off to point past the decoded value.
func (d *decodeState) value(v reflect.Value) {

	var err error
	switch op := d.nextByte(); op {
	default:
		d.error(fmt.Errorf("unsuppoorted type %v", string(op)))

	case 'N':
		err = d.decodeNil(v)
		if err != nil {
			fmt.Println(err)
		}

	case 'b':
		err = d.decodeBool(v)
		if err != nil {
			fmt.Println(err)
		}
	case 'i':
		err = d.decodeInt(v)
		if err != nil {
			fmt.Println(err)
		}
	case 'd':
		err = d.decodeFloat(v)
		if err != nil {
			fmt.Println(err)
		}
	case 's':
		err = d.decodeString(v)
		if err != nil {
			fmt.Println(err)
		}

	case 'O':
		err = d.decodeObject(v)
		if err != nil {
			fmt.Println(err)
		}

	case 'a':
		err = d.decodeArray(v)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// error aborts the decoding by panicking with err.
func (d *decodeState) error(err error) {
	panic(err)
}

func (d *decodeState) decodeNil(v reflect.Value) error {
	err := d.expect('N')
	if err != nil {
		return errors.New("expect N for nil value")
	}
	pv := d.indirect(v, false)
	v = pv
	switch v.Kind() {
	case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice:
		v.Set(reflect.Zero(v.Type()))
		// otherwise, ignore null for primitives/string
	}
	err = d.expect(VALUES_SEPARATOR)
	return err
}

func (d *decodeState) decodeBool(v reflect.Value) error {
	err := d.expect('b')
	if err != nil {
		return errors.New("expect b for bool value")
	}
	err = d.expect(TYPE_VALUE_SEPARATOR)
	item := d.readByte()
	value := item == '1'
	pv := d.indirect(v, false)
	v = pv
	switch v.Kind() {
	default:
		d.saveError(fmt.Errorf("phps trying to unmarshal %q into %v", item, v.Type()))
	case reflect.Bool:
		v.SetBool(value)
	case reflect.Interface:
		if v.NumMethod() == 0 {
			v.Set(reflect.ValueOf(value))
		} else {
			d.saveError(fmt.Errorf(`Value: "bool", Type: %v, Offset: %d}`, v.Type(), d.off))
		}
	}
	err = d.expect(VALUES_SEPARATOR)
	return err
}

func (d *decodeState) decodeInt(v reflect.Value) error {
	err := d.expect('i')
	if err != nil {
		return errors.New("expect i for int value")
	}
	err = d.expect(TYPE_VALUE_SEPARATOR)
	start := d.off
	d.findNext(VALUES_SEPARATOR)
	item := d.data[start : d.off-1]
	s := string(item)
	pv := d.indirect(v, false)
	v = pv
	switch v.Kind() {
	default:
		d.saveError(fmt.Errorf("phps trying to unmarshal %q into %v", s, v.Type()))
	case reflect.Interface:
		n, err := d.convertNumber(s)
		if err != nil {
			d.saveError(err)
			break
		}
		if v.NumMethod() != 0 {

			d.saveError(fmt.Errorf(`Value: "int", Type: %v, Offset: %d}`, v.Type(), d.off))
			break
		}
		v.Set(reflect.ValueOf(n))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || v.OverflowInt(n) {
			d.saveError(fmt.Errorf(`Value: "int", Type: %v, Offset: %d}`, v.Type(), d.off))
			break
		}
		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil || v.OverflowUint(n) {
			d.saveError(fmt.Errorf(`Value: "int", Type: %v, Offset: %d}`, v.Type(), d.off))
			break
		}
		v.SetUint(n)
	}
	return err
}

func (d *decodeState) decodeFloat(v reflect.Value) error {
	err := d.expect('d')
	if err != nil {
		return errors.New("expect d for float value")
	}
	err = d.expect(TYPE_VALUE_SEPARATOR)
	start := d.off
	d.findNext(VALUES_SEPARATOR)
	item := d.data[start : d.off-1]
	s := string(item)
	pv := d.indirect(v, false)
	v = pv
	switch v.Kind() {
	default:
		d.saveError(fmt.Errorf("phps trying to unmarshal %q into %v", s, v.Type()))
	case reflect.Interface:
		n, err := d.convertNumber(s)
		if err != nil {
			d.saveError(err)
			break
		}
		if v.NumMethod() != 0 {

			d.saveError(fmt.Errorf(`Value: "int", Type: %v, Offset: %d}`, v.Type(), d.off))
			break
		}
		v.Set(reflect.ValueOf(n))

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil || v.OverflowFloat(n) {
			d.saveError(fmt.Errorf(`Value: "float", Type: %v, Offset: %d}`, v.Type(), d.off))
			break
		}
		v.SetFloat(n)
	}
	return err
}

func (d *decodeState) decodeString(v reflect.Value) error {
	err := d.expect('s')
	if err != nil {
		return errors.New("expect s for string value")
	}

	//获取字符串长度
	err = d.expect(TYPE_VALUE_SEPARATOR)
	start := d.off
	d.findNext(TYPE_VALUE_SEPARATOR)
	item := d.data[start : d.off-1]
	strLen := string(item)
	length, err := strconv.ParseInt(strLen, 10, 64)
	if err != nil {
		return fmt.Errorf("decodeString false strLen %s", strLen)
	}

	//获取字符串主体部分
	err = d.expect(STRING_SEPARATOR)
	start = d.off
	d.findNext(STRING_SEPARATOR)
	item = d.data[start : d.off-1]
	s := string(item)
	if length != int64(len(s)) {
		return fmt.Errorf("str length do not match %s", strLen)
	}

	pv := d.indirect(v, false)
	v = pv
	switch v.Kind() {
	default:
		d.saveError(fmt.Errorf(`Value: "string", Type: %v, Offset: %d}`, v.Type(), d.off))
	case reflect.String:
		v.SetString(string(s))
	case reflect.Interface:
		if v.NumMethod() == 0 {
			v.Set(reflect.ValueOf(string(s)))
		} else {
			d.saveError(fmt.Errorf(`Value: "string", Type: %v, Offset: %d}`, v.Type(), d.off))
		}
	}
	err = d.expect(VALUES_SEPARATOR)
	return err
}

func (d *decodeState) decodeArray(v reflect.Value) error {
	err := d.expect('a')
	if err != nil {
		return errors.New("expect a for array value")
	}
	//获取array长度
	err = d.expect(TYPE_VALUE_SEPARATOR)
	start := d.off
	d.findNext(TYPE_VALUE_SEPARATOR)
	item := d.data[start : d.off-1]
	strLen := string(item)
	arraryLen, err := strconv.ParseInt(strLen, 10, 64)
	if err != nil {
		return fmt.Errorf("decodeArray false strLen %s", strLen)
	}

	pv := d.indirect(v, false)
	v = pv
	switch v.Kind() {
	default:
		return fmt.Errorf(`Value: "string", Type: %v, Offset: %d}`, v.Type(), d.off)
	case reflect.Array:
	case reflect.Slice:
		break
	}

	i := 0
	d.expect('{')
	for idx := int64(0); idx < arraryLen; idx++ {
		index := 0
		d.value(reflect.ValueOf(&index))

		// Get element of array, growing if necessary.
		if v.Kind() == reflect.Slice {
			// Grow slice if necessary
			if i >= v.Cap() {
				newcap := v.Cap() + v.Cap()/2
				if newcap < 4 {
					newcap = 4
				}
				newv := reflect.MakeSlice(v.Type(), v.Len(), newcap)
				reflect.Copy(newv, v)
				v.Set(newv)
			}
			if i >= v.Len() {
				v.SetLen(i + 1)
			}
		}

		if i < v.Len() {
			// Decode into element.
			d.value(v.Index(i))
		} else {
			// Ran out of fixed array: skip.
			d.value(reflect.Value{})
		}
		i++

		// Next token must be , or ].
	}
	d.expect('}')

	if i < v.Len() {
		if v.Kind() == reflect.Array {
			// Array. Zero the rest.
			z := reflect.Zero(v.Type().Elem())
			for ; i < v.Len(); i++ {
				v.Index(i).Set(z)
			}
		} else {
			v.SetLen(i)
		}
	}
	if i == 0 && v.Kind() == reflect.Slice {
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}
	return err
}

func (d *decodeState) decodeObject(v reflect.Value) error {
	err := d.expect('O')
	if err != nil {
		return errors.New("expect O for object value")
	}
	//获取object name
	d.expect(TYPE_VALUE_SEPARATOR)
	d.expect('1')
	d.expect(TYPE_VALUE_SEPARATOR)
	err = d.expect(STRING_SEPARATOR)
	start := d.off
	d.findNext(STRING_SEPARATOR)
	item := d.data[start : d.off-1]
	strObjectName := string(item)

	//获取object field个数
	d.expect(TYPE_VALUE_SEPARATOR)
	start = d.off
	d.findNext(TYPE_VALUE_SEPARATOR)
	item = d.data[start : d.off-1]
	strLen := string(item)
	objectLen, err := strconv.ParseInt(strLen, 10, 64)
	if err != nil {
		return fmt.Errorf("decodeObject false strLen %s", strLen)
	}

	// Check for unmarshaler.
	pv := d.indirect(v, false)
	v = pv

	// Decoding into nil interface? Switch to non-reflect code.
	if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
		return fmt.Errorf("do not support type interface{}")
	}

	// Check type of target:
	//   struct or
	//   map[T1]T2 where T1 is string, an integer type,
	//             or an encoding.TextUnmarshaler
	switch v.Kind() {
	case reflect.Map:
		// Map key must either have string kind, have an integer kind,
		// or be an encoding.TextUnmarshaler.
		t := v.Type()
		switch t.Key().Kind() {
		case reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		default:
			return fmt.Errorf(`Value: "object", Type: %v, Offset: %d}`, v.Type(), d.off)
		}
		if v.IsNil() {
			v.Set(reflect.MakeMap(t))
		}
	case reflect.Struct:
		if v.Type().Name() != strObjectName {
			return fmt.Errorf("struct name do not match %s, %s", v.Type().Name(), strObjectName)
		}
	default:
		return fmt.Errorf(`Value: "object", Type: %v, Offset: %d}`, v.Type(), d.off)
	}

	var mapElem reflect.Value

	d.expect('{')
	for idx := int64(0); idx < objectLen; idx++ {
		// Read key.
		key := ""
		d.value(reflect.ValueOf(&key))

		// Figure out field corresponding to key.
		var subv reflect.Value

		if v.Kind() == reflect.Map {
			elemType := v.Type().Elem()
			if !mapElem.IsValid() {
				mapElem = reflect.New(elemType).Elem()
			} else {
				mapElem.Set(reflect.Zero(elemType))
			}
			subv = mapElem
		} else {
			var f *field
			fields := cachedTypeFields(v.Type())
			for i := range fields {
				ff := &fields[i]
				if bytes.Equal(ff.nameBytes, []byte(key)) {
					fmt.Println("namebytes", string(ff.nameBytes), "name", ff.name, "key", key)
					f = ff
					break
				}
				if f == nil && ff.equalFold(ff.nameBytes, []byte(key)) {
					fmt.Println("namebytes", string(ff.nameBytes), "name", ff.name, "key", key)
					f = ff
				}
			}
			if f != nil {
				subv = v
				for _, i := range f.index {
					if subv.Kind() == reflect.Ptr {
						if subv.IsNil() {
							// If a struct embeds a pointer to an unexported type,
							// it is not possible to set a newly allocated value
							// since the field is unexported.
							//
							// See https://golang.org/issue/21357
							if !subv.CanSet() {
								d.saveError(fmt.Errorf("phps: cannot set embedded pointer to unexported struct: %v", subv.Type().Elem()))
								// Invalidate subv to ensure d.value(subv) skips over
								// the JSON value without assigning it to subv.
								subv = reflect.Value{}
								break
							}
							subv.Set(reflect.New(subv.Type().Elem()))
						}
						subv = subv.Elem()
					}
					subv = subv.Field(i)
				}
				d.errorContext.Field = f.name
				d.errorContext.Struct = v.Type().Name()
			} else if d.disallowUnknownFields {
				d.saveError(fmt.Errorf("phps: unknown field %q", key))
			}
		}

		// Read  value.
		d.value(subv)

		// Write value back to map;
		// if using struct, subv points into struct already.
		if v.Kind() == reflect.Map {
			kt := v.Type().Key()
			var kv reflect.Value
			switch {
			case kt.Kind() == reflect.String:
				kv = reflect.ValueOf(key).Convert(kt)
			default:
				switch kt.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					s := string(key)
					n, err := strconv.ParseInt(s, 10, 64)
					if err != nil || reflect.Zero(kt).OverflowInt(n) {
						return fmt.Errorf(`Value: "number", Type: %v, Offset: %d}`, kt, d.off)
					}
					kv = reflect.ValueOf(n).Convert(kt)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					s := string(key)
					n, err := strconv.ParseUint(s, 10, 64)
					if err != nil || reflect.Zero(kt).OverflowUint(n) {
						return fmt.Errorf(`Value: "object", Type: %v, Offset: %d}`, kt, d.off)
					}
					kv = reflect.ValueOf(n).Convert(kt)
				default:
					panic("phps: Unexpected key type") // should never occur
				}
			}
			v.SetMapIndex(kv, subv)
		}
	}
	d.expect('}')
	return err
}

// A Number represents a phps number literal.
type Number string

// String returns the literal text of the number.
func (n Number) String() string { return string(n) }

// Float64 returns the number as a float64.
func (n Number) Float64() (float64, error) {
	return strconv.ParseFloat(string(n), 64)
}

// Int64 returns the number as an int64.
func (n Number) Int64() (int64, error) {
	return strconv.ParseInt(string(n), 10, 64)
}

func (d *decodeState) convertNumber(s string) (interface{}, error) {
	if d.useNumber {
		return Number(s), nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		d.saveError(fmt.Errorf(`Value: "number", Type: %v, Offset: %d}`, reflect.TypeOf(0.0), d.off))
	}
	return f, nil
}

// saveError saves the first err it is called with,
// for reporting at the end of the unmarshal.
func (d *decodeState) saveError(err error) {
	if d.savedError == nil {
		d.savedError = err
	}
}

func (d *decodeState) nextByte() byte {
	return d.data[d.off]
}

func (d *decodeState) readByte() byte {
	b := d.data[d.off]
	d.off++
	return b
}

func (d *decodeState) expect(expect byte) error {
	if d.off >= len(d.data) {
		return fmt.Errorf("end of data")
	}
	token := d.data[d.off]
	if token != expect {
		return fmt.Errorf("Read %v, but expected: %v", token, expect)
	}
	d.off++
	return nil
}

func (d *decodeState) findNext(expect byte) error {
	var b byte
	for d.off < len(d.data) {
		b = d.data[d.off]
		d.off++
		if b == expect {
			return nil
		}
	}
	return fmt.Errorf("not find %v", expect)
}

func (d *decodeState) indirect(v reflect.Value, decodingNull bool) reflect.Value {
	// Issue #24153 indicates that it is generally not a guaranteed property
	// that you may round-trip a reflect.Value by calling Value.Addr().Elem()
	// and expect the value to still be settable for values derived from
	// unexported embedded struct fields.
	//
	// The logic below effectively does this when it first addresses the value
	// (to satisfy possible pointer methods) and continues to dereference
	// subsequent pointers as necessary.
	//
	// After the first round-trip, we set v back to the original value to
	// preserve the original RW flags contained in reflect.Value.
	v0 := v
	haveAddr := false

	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		haveAddr = true
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodingNull || e.Elem().Kind() == reflect.Ptr) {
				haveAddr = false
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.Elem().Kind() != reflect.Ptr && decodingNull && v.CanSet() {
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if v.Type().NumMethod() > 0 {
			return reflect.Value{}
		}

		if haveAddr {
			v = v0 // restore original value after round-trip Value.Addr().Elem()
			haveAddr = false
		} else {
			v = v.Elem()
		}
	}
	return v
}
