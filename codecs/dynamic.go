package codecs

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sync"

	aloc "github.com/vedadiyan/protolizer/memory"
	"github.com/vedadiyan/protolizer/metadata"
	"github.com/vedadiyan/protolizer/pdk"
	"github.com/vedadiyan/protolizer/util"
)

type (
	Dynamic struct {
		_builtEncoders map[string]func(any) ([]byte, error)
		_builtDecoders map[string]func(*bytes.Buffer, any) error
	}
)

func NewDynamic() *Dynamic {
	out := new(Dynamic)
	out._builtEncoders = make(map[string]func(any) ([]byte, error))
	out._builtDecoders = make(map[string]func(*bytes.Buffer, any) error)
	return out
}

func (d *Dynamic) Marshal(v any) ([]byte, error) {
	if encoder, ok := d._builtEncoders[metadata.TypeName(reflect.TypeOf(v))]; ok {
		return encoder(v)
	}
	return nil, fmt.Errorf("type %T has not been registered", v)
}

func (d *Dynamic) Unmarshal(data []byte, v any) error {
	if decoder, ok := d._builtDecoders[metadata.TypeName(reflect.TypeOf(v))]; ok {
		return decoder(bytes.NewBuffer(data), v)
	}
	return fmt.Errorf("type %T has not been registered", v)
}

func (d *Dynamic) Register(t reflect.Type) {
	elemType := util.GetElemenType(t)
	if elemType.Kind() != reflect.Struct {
		return
	}
	metadata.RegisterType(elemType)
	_ = d.buildEncoder(elemType)
	_ = d.buildDecoder(elemType)

	for i := range elemType.NumField() {
		f := elemType.Field(i)
		e := util.GetElemenType(f.Type)
		d.Register(e)
	}
}

func (d *Dynamic) buildEncoder(t reflect.Type) func(any) ([]byte, error) {
	typ := metadata.CaptureType(t)
	out := make(map[int]func(reflect.Value, *bytes.Buffer) error)
	for index, field := range typ.FieldsIndexer {
		out[index] = d.encode(field)
	}
	d._builtEncoders[metadata.TypeName(t)] = func(in any) ([]byte, error) {
		buffer := aloc.Alloc(0)
		defer aloc.Dealloc(buffer)
		v := reflect.ValueOf(in)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		for _, field := range typ.Fields {
			value := v.FieldByIndex(field.FieldIndex)
			if value.IsZero() {
				continue
			}
			if field.IsPointer {
				value = value.Elem()
			}
			util.IgnoreReturn(buffer.Write(field.Tag))
			if err := out[field.Tags.Protobuf.FieldNum](value, buffer); err != nil {
				return nil, err
			}
		}
		return bytes.Clone(buffer.Bytes()), nil
	}
	return d._builtEncoders[metadata.TypeName(t)]
}

func (d *Dynamic) encode(field *metadata.Field) func(v reflect.Value, buffer *bytes.Buffer) error {
	switch k := field.Kind; {
	case k == 1:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				pdk.BoolInlineEncode(v.Bool(), buffer)
				return nil
			}
		}
	case k >= 2 && k <= 6:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				pdk.SignedNumberInlineEncoder(v.Int(), field.Tags.Protobuf.WireType, buffer)
				return nil
			}
		}
	case k >= 7 && k <= 11:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				pdk.UnsignedNumberInlineEncoder(v.Uint(), field.Tags.Protobuf.WireType, buffer)
				return nil
			}
		}
	case k == 13:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				pdk.Float32InlineEncode(float32(v.Float()), buffer)
				return nil
			}
		}
	case k == 14:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				pdk.Float64InlineEncode(v.Float(), buffer)
				return nil
			}
		}
	case k == 17 || k == 23:
		{
			if field.Index == reflect.Uint8 {
				return func(v reflect.Value, buffer *bytes.Buffer) error {
					bytes := pdk.BytesEncode(v.Bytes())
					defer aloc.Dealloc(bytes)
					util.IgnoreReturn(bytes.WriteTo(buffer))
					return nil
				}
			}
			w := field.Tags.Protobuf.WireType
			if w == pdk.WireTypeVarint || w == pdk.WireTypeI32 || w == pdk.WireTypeI64 {
				f := *field
				f.Kind = f.Index
				fn := d.encode(&f)
				return func(v reflect.Value, buffer *bytes.Buffer) error {
					innerBuffer := aloc.Alloc(0)
					defer aloc.Dealloc(innerBuffer)
					for i := range v.Len() {
						x := v.Index(i)
						fn(x, innerBuffer)
					}
					bytes := pdk.BufferEncode(innerBuffer)
					bytes.WriteTo(buffer)
					aloc.Dealloc(bytes)
					return nil
				}
			}
			f := *field
			f.Kind = f.Index
			fn := d.encode(&f)
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				tag, err := pdk.TagEncode(int32(field.Tags.Protobuf.FieldNum), pdk.WireTypeLen)
				defer aloc.Dealloc(tag)
				if err != nil {
					return err
				}
				for i := range v.Len() {
					if i != 0 {
						buffer.Write(tag.Bytes())
					}
					x := v.Index(i)
					fn(x, buffer)
				}
				return nil
			}
		}
	case k == 21:
		{
			kf := *field
			kf.Tags.Protobuf.WireType = kf.Tags.MapKey
			kf.Kind = kf.Key
			kv := *field
			kv.Tags.Protobuf.WireType = kv.Tags.MapValue
			kv.Kind = kv.Index
			kfn := d.encode(&kf)
			vfn := d.encode(&kv)
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				tag, err := pdk.TagEncode(int32(field.Tags.Protobuf.FieldNum), pdk.WireTypeLen)
				defer aloc.Dealloc(tag)
				if err != nil {
					return err
				}
				i := 0
				r := v.MapRange()
				for r.Next() {
					key := r.Key()
					value := r.Value()
					if i != 0 {
						util.IgnoreReturn(buffer.Write(tag.Bytes()))
					}
					i++
					innerBuffer := aloc.Alloc(0)
					util.IgnoreReturn(innerBuffer.Write(field.KeyTag))

					if err := kfn(key, innerBuffer); err != nil {
						return err
					}

					util.IgnoreReturn(innerBuffer.Write(field.ValueTag))

					if err := vfn(value, innerBuffer); err != nil {
						return err
					}

					bytes := pdk.BufferEncode(innerBuffer)
					util.IgnoreReturn(bytes.WriteTo(buffer))
					aloc.Dealloc(innerBuffer)
					aloc.Dealloc(bytes)
				}
				return nil
			}
		}
	case k == 24:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				pdk.StringInlineEncode(v.String(), buffer)
				return nil
			}
		}
	case k == 25:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				out, err := d._builtEncoders[metadata.TypeName(v.Type())](v.Interface())
				if err != nil {
					return err
				}
				bytes := pdk.BytesEncode(out)
				defer aloc.Dealloc(bytes)
				util.IgnoreReturn(bytes.WriteTo(buffer))
				return nil
			}
		}
	}
	return func(v reflect.Value, buffer *bytes.Buffer) error {
		return nil
	}
}

func (d *Dynamic) buildDecoder(t reflect.Type) func(*bytes.Buffer, any) error {
	typ := metadata.CaptureType(t)
	out := make(map[int]func(reflect.Value, *bytes.Buffer) error)
	for index, field := range typ.FieldsIndexer {
		out[index] = d.deode(field)
	}
	d._builtDecoders[metadata.TypeName(t)] = func(data *bytes.Buffer, v any) error {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		for data.Len() != 0 {
			fieldNumber, _, err := pdk.TagDecode(data)
			if err != nil {
				return err
			}
			field := typ.FieldsIndexer[int(fieldNumber)]
			if err := out[int(field.Tags.Protobuf.FieldNum)](rv.FieldByIndex(field.FieldIndex), data); err != nil {
				return err
			}

		}
		return nil
	}
	return d._builtDecoders[metadata.TypeName(t)]
}

func (d *Dynamic) deode(field *metadata.Field) func(v reflect.Value, buffer *bytes.Buffer) error {
	switch k := field.Kind; {
	case k == 1:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				out, err := pdk.BoolDecode(buffer)
				if err != nil {
					return err
				}
				v.SetBool(out)
				return nil
			}
		}
	case k >= 2 && k <= 6:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				out, err := pdk.SignedNumberDecoder(field.Tags.Protobuf.WireType, buffer)
				if err != nil {
					return err
				}
				v.SetInt(out)
				return nil
			}
		}
	case k >= 7 && k <= 11:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				out, err := pdk.UnsignedNumberDecoder(field.Tags.Protobuf.WireType, buffer)
				if err != nil {
					return err
				}
				v.SetUint(out)
				return nil
			}
		}
	case k == 13:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				out, err := pdk.Float32Decode(buffer)
				if err != nil {
					return err
				}
				v.SetFloat(float64(out))
				return nil
			}
		}
	case k == 14:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				out, err := pdk.Float64Decode(buffer)
				if err != nil {
					return err
				}
				v.SetFloat(out)
				return nil
			}
		}
	case k == 17 || k == 23:
		{
			if field.Index == reflect.Uint8 {
				return func(v reflect.Value, buffer *bytes.Buffer) error {
					bytes, err := pdk.BytesDecode(buffer)
					if err != nil {
						return err
					}
					v.SetBytes(bytes)
					return nil
				}
			}
			w := field.Tags.Protobuf.WireType
			if w == pdk.WireTypeVarint || w == pdk.WireTypeI32 || w == pdk.WireTypeI64 {
				var arrayType reflect.Type
				var elemType reflect.Type
				var once sync.Once
				f := *field
				f.Kind = f.Index
				fn := d.deode(&f)
				return func(v reflect.Value, buffer *bytes.Buffer) error {
					once.Do(func() {
						arrayType = v.Type()
						elemType = arrayType.Elem()
					})
					value := reflect.New(elemType).Elem()
					bytes, err := pdk.BytesDecode(buffer)
					if err != nil {
						return err
					}
					innerBuffer := aloc.Alloc(0)
					innerBuffer.Write(bytes)
					defer aloc.Dealloc(innerBuffer)
					for innerBuffer.Len() != 0 {
						if err := fn(value, innerBuffer); err != nil {
							return err
						}
						v.Set(reflect.Append(v, value))
					}
					return nil
				}
			}
			var arrayType reflect.Type
			var elemType reflect.Type
			var once sync.Once
			f := *field
			f.Kind = f.Index
			fn := d.deode(&f)
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				once.Do(func() {
					arrayType = v.Type()
					elemType = arrayType.Elem()
				})
				value := reflect.New(elemType).Elem()
				i := 0
				for {
					if i != 0 {
						i, _, read, err := pdk.TagPeek(buffer)
						if err != nil {
							if err == io.EOF {
								return nil
							}
							return err
						}
						if i != int32(field.Tags.Protobuf.FieldNum) {
							break
						}
						read()
					}
					i++

					if err := fn(value, buffer); err != nil {
						return nil
					}
					v.Set(reflect.Append(v, value))
				}
				return nil
			}
		}
	case k == 21:
		{
			var kt reflect.Type
			var vt reflect.Type
			var mapType reflect.Type
			var mapper reflect.Value
			var once sync.Once
			kf := *field
			kf.Kind = kf.Key
			vf := *field
			vf.Kind = vf.Index
			kfn := d.deode(&kf)
			vfn := d.deode(&vf)
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				once.Do(func() {
					t := v.Type()
					kt = t.Key()
					vt = t.Elem()
					mapType = reflect.MapOf(kt, vt)
				})
				key := reflect.New(kt).Elem()
				value := reflect.New(vt).Elem()
				mapper = reflect.MakeMap(mapType)
				i := 0
				for {
					if i != 0 {
						i, _, read, err := pdk.TagPeek(buffer)
						if err != nil {
							if err == io.EOF {
								return nil
							}
							return err
						}
						if i != int32(field.Tags.Protobuf.FieldNum) {
							break
						}
						read()
					}
					i++
					bytes, err := pdk.BytesDecode(buffer)
					if err != nil {
						return err
					}
					innerBuffer := aloc.Alloc(0)
					innerBuffer.Write(bytes)
					_, _, err = pdk.TagDecode(innerBuffer)
					if err != nil {
						aloc.Dealloc(innerBuffer)
						return err
					}

					if err := kfn(key, innerBuffer); err != nil {
						aloc.Dealloc(innerBuffer)
						return err
					}

					_, _, err = pdk.TagDecode(innerBuffer)
					if err != nil {
						aloc.Dealloc(innerBuffer)
						return err
					}

					if err := vfn(value, innerBuffer); err != nil {
						aloc.Dealloc(innerBuffer)
						return err
					}

					mapper.SetMapIndex(key, value)
					aloc.Dealloc(innerBuffer)
				}
				v.Set(mapper)
				return nil
			}
		}
	case k == 24:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				out, err := pdk.StringDecode(buffer)
				if err != nil {
					return err
				}
				v.SetString(out)
				return nil
			}
		}
	case k == 25:
		{
			return func(v reflect.Value, buffer *bytes.Buffer) error {
				value := reflect.New(v.Type())
				err := d._builtDecoders[metadata.TypeName(v.Type())](buffer, value.Interface())
				if err != nil {
					return err
				}
				v.Set(value)
				return nil
			}
		}
	}
	return func(v reflect.Value, buffer *bytes.Buffer) error {
		return nil
	}
}
