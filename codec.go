package protolizer

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
)

func EncodeVarint(u uint64) []byte {
	var buf []byte
	for u >= 0x80 {
		buf = append(buf, byte(u)|0x80)
		u >>= 7
	}
	buf = append(buf, byte(u))
	return buf
}

func DecodeVarint(buf []byte, i *int) (uint64, error) {
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

func EncodeTag(fieldNum int, wireType WireType) []byte {
	tag := (fieldNum << 3) | wireTypeNum(wireType)
	return EncodeVarint(uint64(tag))
}

func DecodeTag(buf []byte, i *int) (int, WireType, error) {
	tag, err := DecodeVarint(buf, i)
	if err != nil {
		return 0, "", err
	}
	fieldNum := int(tag >> 3)
	wireType := WireTypeFromNum(int(tag & 0x7))
	return fieldNum, wireType, nil
}

func EncodeFixed32(v uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf
}

func DecodeFixed32(buf []byte, i *int) (uint32, error) {
	if *i+4 > len(buf) {
		return 0, fmt.Errorf("buffer underflow in fixed32")
	}
	result := binary.LittleEndian.Uint32(buf[*i:])
	*i += 4
	return result, nil
}

func EncodeFixed64(v uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, v)
	return buf
}

func DecodeFixed64(buf []byte, i *int) (uint64, error) {
	if *i+8 > len(buf) {
		return 0, fmt.Errorf("buffer underflow in fixed64")
	}
	result := binary.LittleEndian.Uint64(buf[*i:])
	*i += 8
	return result, nil
}

func EncodeLengthDelimited(data []byte) []byte {
	result := EncodeVarint(uint64(len(data)))
	result = append(result, data...)
	return result
}

func DecodeLengthDelimited(buf []byte, i *int) ([]byte, error) {
	length, err := DecodeVarint(buf, i)
	if err != nil {
		return nil, err
	}
	if *i+int(length) > len(buf) {
		return nil, fmt.Errorf("buffer underflow in length-delimited")
	}
	result := make([]byte, length)
	copy(result, buf[*i:*i+int(length)])
	*i += int(length)
	return result, nil
}

func EncodeField(field *Field, val reflect.Value) ([]byte, error) {
	if field.Kind == reflect.Map || (field.Kind == reflect.Slice && val.Type().Elem().Kind() != reflect.Uint8) {
		return EncodeRepeatedOrMap(field, val)
	}

	tag := EncodeTag(field.Tags.Protobuf.FieldNum, field.Tags.Protobuf.WireType)
	var data []byte

	switch field.Tags.Protobuf.WireType {
	case WIRETYPE_VARINT:
		var u uint64
		switch field.Kind {
		case reflect.Bool:
			if val.Bool() {
				u = 1
			}
		case reflect.Int32, reflect.Int64, reflect.Int:
			u = uint64(val.Int())
		case reflect.Uint32, reflect.Uint64, reflect.Uint:
			u = val.Uint()
		default:
			return nil, fmt.Errorf("unsupported varint type: %v", field.Kind)
		}
		data = EncodeVarint(u)

	case WIRETYPE_FIXED_32:
		switch field.Kind {
		case reflect.Float32:
			u := math.Float32bits(float32(val.Float()))
			data = EncodeFixed32(u)
		case reflect.Uint32:
			data = EncodeFixed32(uint32(val.Uint()))
		case reflect.Int32:
			data = EncodeFixed32(uint32(val.Int()))
		default:
			return nil, fmt.Errorf("unsupported fixed32 type: %v", field.Kind)
		}

	case WIRETYPE_FIXED_64:
		switch field.Kind {
		case reflect.Float64:
			u := math.Float64bits(val.Float())
			data = EncodeFixed64(u)
		case reflect.Uint64:
			data = EncodeFixed64(val.Uint())
		case reflect.Int64:
			data = EncodeFixed64(uint64(val.Int()))
		default:
			return nil, fmt.Errorf("unsupported fixed64 type: %v", field.Kind)
		}

	case WIRETYPE_LENGTH_DELIMITED:
		switch field.Kind {
		case reflect.String:
			data = EncodeLengthDelimited([]byte(val.String()))
		case reflect.Slice:
			if val.Type().Elem().Kind() == reflect.Uint8 {
				// []byte
				data = EncodeLengthDelimited(val.Bytes())
			} else {
				return nil, fmt.Errorf("unsupported slice type for length-delimited: %v", val.Type())
			}
		case reflect.Struct:
			// Embedded message
			embedded, err := Marshal(val.Interface())
			if err != nil {
				return nil, err
			}
			data = EncodeLengthDelimited(embedded)
		default:
			return nil, fmt.Errorf("unsupported length-delimited type: %v", field.Kind)
		}

	default:
		return nil, fmt.Errorf("unsupported wire type: %v", field.Tags.Protobuf.WireType)
	}

	result := make([]byte, len(tag)+len(data))
	copy(result, tag)
	copy(result[len(tag):], data)
	return result, nil
}

func EncodeRepeatedOrMap(field *Field, val reflect.Value) ([]byte, error) {
	var result []byte

	if field.Kind == reflect.Map {
		// Handle maps as repeated key-value pairs
		for _, key := range val.MapKeys() {
			mapVal := val.MapIndex(key)

			// Create a map entry message
			var entryData []byte

			// Key (field 1)
			keyTag := EncodeTag(1, WIRETYPE_LENGTH_DELIMITED)
			keyData := EncodeLengthDelimited([]byte(key.String()))
			entryData = append(entryData, keyTag...)
			entryData = append(entryData, keyData...)

			// Value (field 2)
			valueTag := EncodeTag(2, WIRETYPE_LENGTH_DELIMITED)
			valueData := EncodeLengthDelimited([]byte(mapVal.String()))
			entryData = append(entryData, valueTag...)
			entryData = append(entryData, valueData...)

			// Wrap the entry in length-delimited encoding
			tag := EncodeTag(field.Tags.Protobuf.FieldNum, WIRETYPE_LENGTH_DELIMITED)
			data := EncodeLengthDelimited(entryData)

			result = append(result, tag...)
			result = append(result, data...)
		}
		return result, nil
	}

	// Handle repeated fields (slices)
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		elemKind := elem.Kind()

		var tag []byte
		var data []byte

		switch elemKind {
		case reflect.String:
			tag = EncodeTag(field.Tags.Protobuf.FieldNum, WIRETYPE_LENGTH_DELIMITED)
			data = EncodeLengthDelimited([]byte(elem.String()))

		case reflect.Bool, reflect.Int32, reflect.Int64, reflect.Int, reflect.Uint32, reflect.Uint64, reflect.Uint:
			tag = EncodeTag(field.Tags.Protobuf.FieldNum, WIRETYPE_VARINT)
			var u uint64
			switch elemKind {
			case reflect.Bool:
				if elem.Bool() {
					u = 1
				}
			case reflect.Int32, reflect.Int64, reflect.Int:
				u = uint64(elem.Int())
			case reflect.Uint32, reflect.Uint64, reflect.Uint:
				u = elem.Uint()
			}
			data = EncodeVarint(u)

		case reflect.Float32:
			tag = EncodeTag(field.Tags.Protobuf.FieldNum, WIRETYPE_FIXED_32)
			u := math.Float32bits(float32(elem.Float()))
			data = EncodeFixed32(u)

		case reflect.Float64:
			tag = EncodeTag(field.Tags.Protobuf.FieldNum, WIRETYPE_FIXED_64)
			u := math.Float64bits(elem.Float())
			data = EncodeFixed64(u)

		case reflect.Struct:
			tag = EncodeTag(field.Tags.Protobuf.FieldNum, WIRETYPE_LENGTH_DELIMITED)
			embedded, err := Marshal(elem.Interface())
			if err != nil {
				return nil, err
			}
			data = EncodeLengthDelimited(embedded)

		default:
			return nil, fmt.Errorf("unsupported repeated element type: %v", elemKind)
		}

		result = append(result, tag...)
		result = append(result, data...)
	}

	return result, nil
}

func DecodeField(field *Field, val reflect.Value, buf []byte, i *int, wireType WireType) error {
	// Handle repeated fields (slices) and maps (but not []byte which is handled as length-delimited)
	if field.Kind == reflect.Map || (field.Kind == reflect.Slice && val.Type().Elem().Kind() != reflect.Uint8) {
		return DecodeRepeatedOrMap(field, val, buf, i, wireType)
	}

	if field.IsPointer {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}
		val = val.Elem()
	}

	switch wireType {
	case WIRETYPE_VARINT:
		u, err := DecodeVarint(buf, i)
		if err != nil {
			return err
		}
		switch field.Kind {
		case reflect.Bool:
			val.SetBool(u != 0)
		case reflect.Int32, reflect.Int64, reflect.Int:
			val.SetInt(int64(u))
		case reflect.Uint32, reflect.Uint64, reflect.Uint:
			val.SetUint(u)
		default:
			return fmt.Errorf("unsupported varint type: %v", field.Kind)
		}

	case WIRETYPE_FIXED_32:
		u, err := DecodeFixed32(buf, i)
		if err != nil {
			return err
		}
		switch field.Kind {
		case reflect.Float32:
			val.SetFloat(float64(math.Float32frombits(u)))
		case reflect.Uint32:
			val.SetUint(uint64(u))
		case reflect.Int32:
			val.SetInt(int64(int32(u)))
		default:
			return fmt.Errorf("unsupported fixed32 type: %v", field.Kind)
		}

	case WIRETYPE_FIXED_64:
		u, err := DecodeFixed64(buf, i)
		if err != nil {
			return err
		}
		switch field.Kind {
		case reflect.Float64:
			val.SetFloat(math.Float64frombits(u))
		case reflect.Uint64:
			val.SetUint(u)
		case reflect.Int64:
			val.SetInt(int64(u))
		default:
			return fmt.Errorf("unsupported fixed64 type: %v", field.Kind)
		}

	case WIRETYPE_LENGTH_DELIMITED:
		data, err := DecodeLengthDelimited(buf, i)
		if err != nil {
			return err
		}
		switch field.Kind {
		case reflect.String:
			val.SetString(string(data))
		case reflect.Slice:
			if val.Type().Elem().Kind() == reflect.Uint8 {
				val.SetBytes(data)
			} else {
				return fmt.Errorf("unsupported slice type for length-delimited: %v", val.Type())
			}
		case reflect.Struct:
			// Embedded message
			if err := Unmarshal(data, val.Addr().Interface()); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported length-delimited type: %v", field.Kind)
		}

	default:
		return fmt.Errorf("unsupported wire type: %v", wireType)
	}

	return nil
}

func DecodeRepeatedOrMap(field *Field, val reflect.Value, buf []byte, i *int, wireType WireType) error {
	if field.Kind == reflect.Map {
		// Handle map entry
		if wireType != WIRETYPE_LENGTH_DELIMITED {
			return fmt.Errorf("map field must use length-delimited wire type")
		}

		entryData, err := DecodeLengthDelimited(buf, i)
		if err != nil {
			return err
		}

		// Decode key-value pairs from the entry
		var key, value string
		j := 0
		for j < len(entryData) {
			fieldNum, entryWireType, err := DecodeTag(entryData, &j)
			if err != nil {
				return err
			}

			if entryWireType != WIRETYPE_LENGTH_DELIMITED {
				return fmt.Errorf("map entry fields must be length-delimited")
			}

			data, err := DecodeLengthDelimited(entryData, &j)
			if err != nil {
				return err
			}

			if fieldNum == 1 {
				key = string(data)
			} else if fieldNum == 2 {
				value = string(data)
			}
		}

		val.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		return nil
	}

	// Handle repeated fields (slices)
	elemType := val.Type().Elem()

	var newElem reflect.Value
	switch wireType {
	case WIRETYPE_VARINT:
		newElem = reflect.New(elemType).Elem()
		u, err := DecodeVarint(buf, i)
		if err != nil {
			return err
		}

		switch elemType.Kind() {
		case reflect.Bool:
			newElem.SetBool(u != 0)
		case reflect.Int32, reflect.Int64, reflect.Int:
			newElem.SetInt(int64(u))
		case reflect.Uint32, reflect.Uint64, reflect.Uint:
			newElem.SetUint(u)
		default:
			return fmt.Errorf("unsupported repeated varint type: %v", elemType.Kind())
		}

	case WIRETYPE_FIXED_32:
		newElem = reflect.New(elemType).Elem()
		u, err := DecodeFixed32(buf, i)
		if err != nil {
			return err
		}

		switch elemType.Kind() {
		case reflect.Float32:
			newElem.SetFloat(float64(math.Float32frombits(u)))
		case reflect.Uint32:
			newElem.SetUint(uint64(u))
		case reflect.Int32:
			newElem.SetInt(int64(int32(u)))
		default:
			return fmt.Errorf("unsupported repeated fixed32 type: %v", elemType.Kind())
		}

	case WIRETYPE_FIXED_64:
		newElem = reflect.New(elemType).Elem()
		u, err := DecodeFixed64(buf, i)
		if err != nil {
			return err
		}

		switch elemType.Kind() {
		case reflect.Float64:
			newElem.SetFloat(math.Float64frombits(u))
		case reflect.Uint64:
			newElem.SetUint(u)
		case reflect.Int64:
			newElem.SetInt(int64(u))
		default:
			return fmt.Errorf("unsupported repeated fixed64 type: %v", elemType.Kind())
		}

	case WIRETYPE_LENGTH_DELIMITED:
		data, err := DecodeLengthDelimited(buf, i)
		if err != nil {
			return err
		}

		if elemType.Kind() == reflect.String {
			newElem = reflect.ValueOf(string(data))
		} else if elemType.Kind() == reflect.Struct {
			// Repeated embedded message
			newElem = reflect.New(elemType).Elem()
			if err := Unmarshal(data, newElem.Addr().Interface()); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unsupported repeated length-delimited type: %v", elemType.Kind())
		}

	default:
		return fmt.Errorf("unsupported wire type for repeated field: %v", wireType)
	}

	val.Set(reflect.Append(val, newElem))
	return nil
}

func SkipField(buf []byte, i *int, wireType WireType) error {
	switch wireType {
	case WIRETYPE_VARINT:
		_, err := DecodeVarint(buf, i)
		return err
	case WIRETYPE_FIXED_32:
		if *i+4 > len(buf) {
			return fmt.Errorf("buffer underflow skipping fixed32")
		}
		*i += 4
		return nil
	case WIRETYPE_FIXED_64:
		if *i+8 > len(buf) {
			return fmt.Errorf("buffer underflow skipping fixed64")
		}
		*i += 8
		return nil
	case WIRETYPE_LENGTH_DELIMITED:
		_, err := DecodeLengthDelimited(buf, i)
		return err
	default:
		return fmt.Errorf("cannot skip unsupported wire type: %v", wireType)
	}
}
