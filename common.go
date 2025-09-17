package protolizer

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type (
	WireType     string
	Serializable struct {
		Value any
	}
	Tags struct {
		Protobuf *ProtobufInfo
		JsonName string
		MapKey   string
		MapValue string
	}
	ProtobufInfo struct {
		WireType WireType
		FieldNum int
		Label    string
		Name     string
		Syntax   string
		OneOf    bool
	}
	Field struct {
		Name      string
		Kind      reflect.Kind
		IsPointer bool
		Tags      *Tags
		GetValue  func(v reflect.Value) *Serializable
		SetValue  func(v reflect.Value, value any)
	}
	Type struct {
		Fields []*Field
	}
)

const (
	WIRETYPE_VARINT           WireType = "varint"
	WIRETYPE_FIXED_64         WireType = "fixed64"
	WIRETYPE_LENGTH_DELIMITED WireType = "bytes"
	WRIETYPE_START_GROUP
	WIRETYPE_END_GROUP
	WIRETYPE_FIXED_32 WireType = "fixed32"
)

var (
	_registry map[reflect.Type]*Type
	_mut      sync.Mutex
)

func init() {
	_registry = make(map[reflect.Type]*Type)
}

func RegisterType(t reflect.Type) *Type {
	_mut.Lock()
	defer _mut.Unlock()
	if value, ok := _registry[t]; ok {
		return value
	}
	out := new(Type)
	out.Fields = make([]*Field, t.Elem().NumField())
	for i := range t.Elem().NumField() {
		f := NewField(t.Elem().Field(i))
		out.Fields[i] = f
	}
	_registry[t] = out
	return out
}

func NewTags(t reflect.StructTag) *Tags {
	out := new(Tags)
	if tag, ok := t.Lookup("protobuf"); ok {
		out.Protobuf = ParseProtoTag(tag)
	}
	out.JsonName = t.Get("json")
	return out
}

func NewField(f reflect.StructField) *Field {
	out := new(Field)
	out.Name = f.Name
	out.Kind, out.IsPointer = GetKind(f.Type)
	out.Tags = NewTags(f.Tag)

	out.GetValue = func(v reflect.Value) *Serializable {
		fieldVal := v.FieldByIndex(f.Index)
		if out.IsPointer {
			if fieldVal.IsNil() {
				return &Serializable{Value: nil}
			}
			return &Serializable{Value: fieldVal.Elem().Interface()}
		}
		return &Serializable{Value: fieldVal.Interface()}
	}

	out.SetValue = func(v reflect.Value, value any) {
		fieldVal := v.FieldByIndex(f.Index)
		if out.IsPointer {
			ptr := reflect.New(fieldVal.Type().Elem())
			ptr.Elem().Set(reflect.ValueOf(value))
			fieldVal.Set(ptr)
		} else {
			fieldVal.Set(reflect.ValueOf(value))
		}
	}
	return out
}

func ParseProtoTag(tag string) *ProtobufInfo {
	tag = strings.TrimPrefix(tag, "protobuf:\"")
	tag = strings.TrimSuffix(tag, "\"")

	segments := strings.Split(tag, ",")
	l := len(segments)

	fieldNum, err := strconv.Atoi(segments[1])
	if err != nil {
		panic(fmt.Errorf("invalid field number: %w", err))
	}

	out := new(ProtobufInfo)
	if l > 0 {
		out.WireType = WireType(segments[0])
	}
	if l > 1 {
		out.FieldNum = fieldNum
	}
	if l > 2 {
		out.Label = segments[2]
	}
	if l > 3 {
		out.Name = strings.TrimPrefix(segments[3], "name=")
	}
	if l > 4 {
		out.Syntax = segments[4]
	}
	if len(segments) == 5 {
		out.OneOf = true
	}

	return out
}

func GetKind(t reflect.Type) (reflect.Kind, bool) {
	if t.Kind() == reflect.Pointer {
		return t.Elem().Kind(), true
	}
	return t.Kind(), false
}

func (t *Tags) IsProtobuf() bool {
	return t.Protobuf != nil
}

func (ser *Serializable) Serialize(wireType WireType) ([]byte, error) {
	switch wireType {
	case WIRETYPE_VARINT:
		var u uint64
		switch v := ser.Value.(type) {
		case int:
			u = uint64(v)
		case int32:
			u = uint64(v)
		case int64:
			u = uint64(v)
		case uint32:
			u = uint64(v)
		case uint64:
			u = v
		case bool:
			if v {
				u = 1
			} else {
				u = 0
			}
		default:
			return nil, fmt.Errorf("varint expects int/bool, got %T", ser.Value)
		}
		return encodeVarint(u), nil

	case WIRETYPE_LENGTH_DELIMITED:
		switch v := ser.Value.(type) {
		case []byte:
			lenBuf := encodeVarint(uint64(len(v)))
			return append(lenBuf, v...), nil
		case string:
			data := []byte(v)
			lenBuf := encodeVarint(uint64(len(data)))
			return append(lenBuf, data...), nil
		default:
			rv := reflect.ValueOf(ser.Value)
			if rv.Kind() == reflect.Struct || (rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Struct) {
				b, err := SerializeStruct(ser.Value)
				if err != nil {
					return nil, err
				}
				lenBuf := encodeVarint(uint64(len(b)))
				return append(lenBuf, b...), nil
			}
			return nil, fmt.Errorf("length-delimited expects []byte/string/struct, got %T", ser.Value)
		}

	case WIRETYPE_FIXED_64:
		v, ok := ser.Value.(uint64)
		if !ok {
			return nil, fmt.Errorf("fixed64 expects uint64, got %T", ser.Value)
		}
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, v)
		return buf, nil

	case WIRETYPE_FIXED_32:
		v, ok := ser.Value.(uint32)
		if !ok {
			return nil, fmt.Errorf("fixed32 expects uint32, got %T", ser.Value)
		}
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, v)
		return buf, nil

	default:
		return nil, fmt.Errorf("unexpected wire type %s", wireType)
	}
}

func (ser *Serializable) SerializeRepeated(wireType WireType, packed bool) ([][]byte, error) {
	values, ok := ser.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("repeated expects []any, got %T", ser.Value)
	}

	if packed {
		var packedBuf []byte
		for _, v := range values {
			elem := &Serializable{Value: v}
			b, err := elem.Serialize(wireType)
			if err != nil {
				return nil, err
			}
			packedBuf = append(packedBuf, b...)
		}
		lenBuf := encodeVarint(uint64(len(packedBuf)))
		return [][]byte{append(lenBuf, packedBuf...)}, nil
	}

	var out [][]byte
	for _, v := range values {
		elem := &Serializable{Value: v}
		b, err := elem.Serialize(wireType)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

func wireTypeNum(wt WireType) int {
	switch wt {
	case WIRETYPE_VARINT:
		return 0
	case WIRETYPE_FIXED_64:
		return 1
	case WIRETYPE_LENGTH_DELIMITED:
		return 2
	case WIRETYPE_FIXED_32:
		return 5
	default:
		return -1
	}
}

func encodeVarint(u uint64) []byte {
	var buf []byte
	for u >= 0x80 {
		buf = append(buf, byte(u)|0x80)
		u >>= 7
	}
	buf = append(buf, byte(u))
	return buf
}

func SerializeStruct(v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("SerializeStruct expects struct or *struct, got %T", v)
	}

	rt := rv.Type()
	t := RegisterType(reflect.PointerTo(rt))
	var out []byte

	for _, f := range t.Fields {
		if f.Tags == nil || !f.Tags.IsProtobuf() {
			continue
		}
		info := f.Tags.Protobuf
		fieldVal := rv.FieldByName(f.Name)
		if f.IsPointer && fieldVal.IsNil() {
			continue
		}
		ser := f.GetValue(rv)
		wireNum := wireTypeNum(info.WireType)
		if wireNum < 0 {
			return nil, fmt.Errorf("invalid wire type %s", info.WireType)
		}
		tag := uint64(info.FieldNum<<3 | wireNum)

		if strings.HasPrefix(info.Label, "repeated") {
			packed := (info.WireType == WIRETYPE_VARINT ||
				info.WireType == WIRETYPE_FIXED_32 ||
				info.WireType == WIRETYPE_FIXED_64)
			bufs, err := ser.SerializeRepeated(info.WireType, packed)
			if err != nil {
				return nil, err
			}
			for _, b := range bufs {
				out = append(out, encodeVarint(tag)...)
				out = append(out, b...)
			}
		} else {
			valBytes, err := ser.Serialize(info.WireType)
			if err != nil {
				return nil, err
			}
			out = append(out, encodeVarint(tag)...)
			out = append(out, valBytes...)
		}
	}
	return out, nil
}

func decodeVarint(buf []byte, i *int) (uint64, error) {
	var result uint64
	var shift uint
	for {
		if *i >= len(buf) {
			return 0, fmt.Errorf("buffer underflow in varint")
		}
		b := buf[*i]
		*i++
		result |= uint64(b&0x7F) << shift
		if b < 0x80 {
			break
		}
		shift += 7
		if shift >= 64 {
			return 0, fmt.Errorf("varint overflow")
		}
	}
	return result, nil
}

func wireTypeFromNum(n int) WireType {
	switch n {
	case 0:
		return WIRETYPE_VARINT
	case 1:
		return WIRETYPE_FIXED_64
	case 2:
		return WIRETYPE_LENGTH_DELIMITED
	case 5:
		return WIRETYPE_FIXED_32
	default:
		return ""
	}
}

func DeserializeStruct(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("DeserializeStruct expects pointer to struct, got %T", v)
	}
	rv = rv.Elem()
	rt := rv.Type()

	t := RegisterType(reflect.PtrTo(rt))
	i := 0
	for i < len(data) {
		tagVal, err := decodeVarint(data, &i)
		if err != nil {
			return err
		}
		fieldNum := int(tagVal >> 3)
		wireNum := int(tagVal & 0x7)

		var f *Field
		for _, fld := range t.Fields {
			if fld.Tags != nil && fld.Tags.IsProtobuf() && fld.Tags.Protobuf.FieldNum == fieldNum {
				f = fld
				break
			}
		}
		if f == nil {
			return fmt.Errorf("unknown field number %d", fieldNum)
		}

		wt := wireTypeFromNum(wireNum)
		if wt == "" {
			return fmt.Errorf("unsupported wire type %d", wireNum)
		}

		switch wt {
		case WIRETYPE_VARINT:
			val, err := decodeVarint(data, &i)
			if err != nil {
				return err
			}
			switch f.Kind {
			case reflect.Int, reflect.Int32, reflect.Int64:
				f.SetValue(rv, int64(val))
			case reflect.Uint32, reflect.Uint64:
				f.SetValue(rv, val)
			case reflect.Bool:
				f.SetValue(rv, val != 0)
			}

		case WIRETYPE_FIXED_32:
			if i+4 > len(data) {
				return fmt.Errorf("buffer underflow for fixed32")
			}
			val := binary.LittleEndian.Uint32(data[i:])
			i += 4
			f.SetValue(rv, val)

		case WIRETYPE_FIXED_64:
			if i+8 > len(data) {
				return fmt.Errorf("buffer underflow for fixed64")
			}
			val := binary.LittleEndian.Uint64(data[i:])
			i += 8
			f.SetValue(rv, val)

		case WIRETYPE_LENGTH_DELIMITED:
			length, err := decodeVarint(data, &i)
			if err != nil {
				return err
			}
			if i+int(length) > len(data) {
				return fmt.Errorf("buffer underflow for length-delimited")
			}
			fieldBytes := data[i : i+int(length)]
			i += int(length)

			switch f.Kind {
			case reflect.String:
				f.SetValue(rv, string(fieldBytes))
			case reflect.Slice:
				if f.Kind == reflect.Slice && rv.FieldByName(f.Name).Type().Elem().Kind() == reflect.Uint8 {
					f.SetValue(rv, fieldBytes)
				}
			case reflect.Struct:
				nestedPtr := reflect.New(rv.FieldByName(f.Name).Type())
				if err := DeserializeStruct(fieldBytes, nestedPtr.Interface()); err != nil {
					return err
				}
				if f.IsPointer {
					rv.FieldByName(f.Name).Set(nestedPtr)
				} else {
					rv.FieldByName(f.Name).Set(nestedPtr.Elem())
				}
			}
		}
	}
	return nil
}
