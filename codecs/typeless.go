package codecs

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	aloc "github.com/vedadiyan/protolizer/memory"
	"github.com/vedadiyan/protolizer/metadata"
	"github.com/vedadiyan/protolizer/pdk"
	"github.com/vedadiyan/protolizer/util"
)

type (
	Typeless struct {
		_builtEncoders map[string]func(map[string]any) ([]byte, error)
		_builtDecoders map[string]func(*bytes.Buffer) (any, error)
	}
)

func NewTypeless() *Typeless {
	out := new(Typeless)
	out._builtEncoders = make(map[string]func(map[string]any) ([]byte, error))
	out._builtDecoders = make(map[string]func(*bytes.Buffer) (any, error))
	return out
}

func (tl *Typeless) Marshal(v map[string]any, t metadata.Type) ([]byte, error) {
	if encoder, ok := tl._builtEncoders[t.Name]; ok {
		return encoder(v)
	}
	return nil, fmt.Errorf("type %T has not been registered", v)
}

func (tl *Typeless) Unmarshal(data []byte, t metadata.Type) (any, error) {
	if decoder, ok := tl._builtDecoders[t.Name]; ok {
		return decoder(bytes.NewBuffer(data))
	}
	return nil, fmt.Errorf("type %s has not been registered", t.Name)
}

func (tl *Typeless) Register(t metadata.Type) {
	_ = tl.buildEncoder(t)
	_ = tl.buildDecoder(t)
}

func (tl *Typeless) buildEncoder(t metadata.Type) func(map[string]any) ([]byte, error) {
	out := make(map[int]func(reflect.Value, *bytes.Buffer) error)
	for index, field := range t.FieldsIndexer {
		out[index] = tl.encode(field)
	}
	tl._builtEncoders[t.Name] = func(in map[string]any) ([]byte, error) {
		buffer := aloc.Alloc(0)
		defer aloc.Dealloc(buffer)
		for _, field := range t.Fields {
			v, ok := in[field.Name]
			if !ok {
				continue
			}
			value := reflect.ValueOf(v)
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
	return tl._builtEncoders[t.Name]
}

func (tl *Typeless) encode(field *metadata.Field) func(v reflect.Value, buffer *bytes.Buffer) error {
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
				fn := tl.encode(&f)
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
			fn := tl.encode(&f)
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
			kfn := tl.encode(&kf)
			vfn := tl.encode(&kv)
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
				return fmt.Errorf("typeless coded does not support structs")
			}
		}
	}
	return func(v reflect.Value, buffer *bytes.Buffer) error {
		return nil
	}
}

func (tl *Typeless) buildDecoder(t metadata.Type) func(*bytes.Buffer) (any, error) {
	out := make(map[int]func(*bytes.Buffer) (any, error))
	for index, field := range t.FieldsIndexer {
		out[index] = tl.deode(field)
	}
	tl._builtDecoders[t.Name] = func(data *bytes.Buffer) (any, error) {
		mapper := make(map[string]any)
		for data.Len() != 0 {
			fieldNumber, _, err := pdk.TagDecode(data)
			if err != nil {
				return nil, err
			}
			field := t.FieldsIndexer[int(fieldNumber)]
			value, err := out[int(field.Tags.Protobuf.FieldNum)](data)
			if err != nil {
				return nil, err
			}
			mapper[field.Name] = value

		}
		return mapper, nil
	}
	return tl._builtDecoders[t.Name]
}

func (tl *Typeless) deode(field *metadata.Field) func(buffer *bytes.Buffer) (any, error) {
	switch k := field.Kind; {
	case k == 1:
		{
			return func(buffer *bytes.Buffer) (any, error) {
				return pdk.BoolDecode(buffer)
			}
		}
	case k >= 2 && k <= 6:
		{
			return func(buffer *bytes.Buffer) (any, error) {
				return pdk.SignedNumberDecoder(field.Tags.Protobuf.WireType, buffer)

			}
		}
	case k >= 7 && k <= 11:
		{
			return func(buffer *bytes.Buffer) (any, error) {
				return pdk.UnsignedNumberDecoder(field.Tags.Protobuf.WireType, buffer)

			}
		}
	case k == 13:
		{
			return func(buffer *bytes.Buffer) (any, error) {
				return pdk.Float32Decode(buffer)
			}
		}
	case k == 14:
		{
			return func(buffer *bytes.Buffer) (any, error) {
				return pdk.Float64Decode(buffer)
			}
		}
	case k == 17 || k == 23:
		{
			if field.Index == reflect.Uint8 {
				return func(buffer *bytes.Buffer) (any, error) {
					return pdk.BytesDecode(buffer)
				}
			}
			w := field.Tags.Protobuf.WireType
			if w == pdk.WireTypeVarint || w == pdk.WireTypeI32 || w == pdk.WireTypeI64 {
				f := *field
				f.Kind = f.Index
				fn := tl.deode(&f)
				return func(buffer *bytes.Buffer) (any, error) {
					out := make([]any, 0)
					bytes, err := pdk.BytesDecode(buffer)
					if err != nil {
						return nil, err
					}
					innerBuffer := aloc.Alloc(0)
					innerBuffer.Write(bytes)
					defer aloc.Dealloc(innerBuffer)
					for innerBuffer.Len() != 0 {
						iv, err := fn(innerBuffer)
						if err != nil {
							return nil, err
						}
						out = append(out, iv)
					}
					return out, nil
				}
			}
			f := *field
			f.Kind = f.Index
			fn := tl.deode(&f)
			return func(buffer *bytes.Buffer) (any, error) {
				out := make([]any, 0)
				i := 0
				for {
					if i != 0 {
						i, _, read, err := pdk.TagPeek(buffer)
						if err != nil {
							if err == io.EOF {
								return out, nil
							}
							return nil, err
						}
						if i != int32(field.Tags.Protobuf.FieldNum) {
							break
						}
						read()
					}
					i++

					iv, err := fn(buffer)
					if err != nil {
						return nil, err
					}
					out = append(out, iv)
				}
				return out, nil
			}
		}
	case k == 21:
		{
			kf := *field
			kf.Kind = kf.Key
			vf := *field
			vf.Kind = vf.Index
			kfn := tl.deode(&kf)
			vfn := tl.deode(&vf)
			return func(buffer *bytes.Buffer) (any, error) {
				mapper := make(map[any]any)
				i := 0
				for {
					if i != 0 {
						i, _, read, err := pdk.TagPeek(buffer)
						if err != nil {
							if err == io.EOF {
								return mapper, nil
							}
							return nil, err
						}
						if i != int32(field.Tags.Protobuf.FieldNum) {
							break
						}
						read()
					}
					i++
					bytes, err := pdk.BytesDecode(buffer)
					if err != nil {
						return nil, err
					}
					innerBuffer := aloc.Alloc(0)
					innerBuffer.Write(bytes)
					_, _, err = pdk.TagDecode(innerBuffer)
					if err != nil {
						aloc.Dealloc(innerBuffer)
						return nil, err
					}
					ikv, err := kfn(innerBuffer)
					if err != nil {
						aloc.Dealloc(innerBuffer)
						return nil, err
					}

					_, _, err = pdk.TagDecode(innerBuffer)
					if err != nil {
						aloc.Dealloc(innerBuffer)
						return nil, err
					}

					ivv, err := vfn(innerBuffer)
					if err != nil {
						aloc.Dealloc(innerBuffer)
						return nil, err
					}

					mapper[ikv] = ivv
					aloc.Dealloc(innerBuffer)
				}
				return mapper, nil
			}
		}
	case k == 24:
		{
			return func(buffer *bytes.Buffer) (any, error) {
				return pdk.StringDecode(buffer)
			}
		}
	case k == 25:
		{
			return func(buffer *bytes.Buffer) (any, error) {
				return nil, fmt.Errorf("typeless coded does not support structs")
			}
		}
	}
	return func(buffer *bytes.Buffer) (any, error) {
		return nil, nil
	}
}
