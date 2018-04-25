package grb

import (
	"reflect"
	"github.com/mitchellh/go-mruby"
	"errors"
	"fmt"
)

func decodeMrbValueHash(v *mruby.MrbValue) (interface{}, error) {
	h := v.Hash()
	obj := make(map[string]interface{})
	keys, err := h.Keys()
	if err != nil {
		return nil, err
	}
	keyArr := keys.Array()
	for i := 0; i < keyArr.Len(); i++ {
		k, err := keyArr.Get(i)
		if err != nil {
			return nil, err
		}
		switch k.Type() {
		case mruby.TypeString:
			hv, err := h.Get(k)
			if err != nil {
				return nil, err
			}
			hvg, err := DecodeMrbValue(hv)
			if err != nil {
				return nil, err
			}
			obj[k.String()] = hvg

		default:
			return nil, errors.New("invalid key type") // TODO: more meaningful err msg
		}
	}
	return obj, nil
}

func decodeMrbValueArray(v *mruby.MrbValue) (interface{}, error) {
	a := v.Array()
	arr := make([]interface{}, a.Len())
	for i := 0; i < a.Len(); i++ {
		av0, err := a.Get(i)
		if err != nil {
			return nil, err
		}
		av1, err := DecodeMrbValue(av0)
		if err != nil {
			return nil, err
		}
		arr[i] = av1
	}
	return arr, nil
}

func DecodeMrbValue(v *mruby.MrbValue) (interface{}, error) {
	switch v.Type() {
	case mruby.TypeFalse:
		return false, nil
	case mruby.TypeTrue:
		return true, nil
	case mruby.TypeFixnum:
		return v.Fixnum(), nil
	case mruby.TypeFloat:
		return v.Float(), nil
	case mruby.TypeString:
		return v.String(), nil
	case mruby.TypeNil:
		return nil, nil
	case mruby.TypeHash:
		return decodeMrbValueHash(v)
	case mruby.TypeArray:
		return decodeMrbValueArray(v)
	default:
		// TODO: omit warning
	}
	return nil, errors.New("non json compliant type found")
}

func encodeMrbFromReflectValue(mrb *mruby.Mrb, x reflect.Value) (*mruby.MrbValue, error) {
	switch x.Kind() {
	case reflect.Interface:
		return EncodeMrbValue(mrb, x.Interface())
	case reflect.String:
		return mrb.StringValue(x.String()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return mrb.LoadString(fmt.Sprintf("%d", x.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return mrb.LoadString(fmt.Sprintf("%d", x.Uint()))
	case reflect.Float32, reflect.Float64:
		return mrb.LoadString(fmt.Sprintf("%f", x.Float()))
	case reflect.Bool:
		if x.Bool() {
			return mrb.TrueValue(), nil
		} else {
			return mrb.FalseValue(), nil
		}
	case reflect.Slice:
		if a, err := mrb.LoadString(`Array.new`); err != nil {
			return nil, err
		} else {
			for i := 0; i < x.Len(); i++ {
				av := x.Index(i)
				if av0, err := EncodeMrbValue(mrb, av.Interface()); err != nil {
					return nil, err
				} else {
					a.Call("push", av0)
				}
			}
			return a, nil
		}

	case reflect.Map:
		if h, err := mrb.LoadString(`Hash.new`); err != nil {
			return nil, err
		} else {
			for _, hk := range x.MapKeys() {
				if hk.Kind() != reflect.String {
					return nil, errors.New("key of map must be string")
				}
				hv := x.MapIndex(hk)
				if hv0, err := encodeMrbFromReflectValue(mrb, hv); err != nil {
					return nil, err
				} else {
					h.Call("store", mrb.StringValue(hk.String()), hv0)
				}
			}
			return h, nil
		}
	}
	//TODO: debug log
	return nil, errors.New("non json compliant type")
}

func EncodeMrbValue(mrb *mruby.Mrb, v interface{}) (*mruby.MrbValue, error) {
	if v == nil {
		return mrb.NilValue(), nil
	}
	return encodeMrbFromReflectValue(mrb, reflect.ValueOf(v))
}

func NewRuntimeError(mrb *mruby.Mrb, err error) *mruby.MrbValue {
	clsRe := mrb.Class("RuntimeError", nil)
	re, err := clsRe.New(mrb.StringValue(err.Error()))
	if err != nil {
		panic("failed to create RuntimeError")
	}

	// TODO: MRUBY1.2: `#set_backtrace` is not available
	//goPc, goFile, goLine, _ := runtime.Caller(1)
	//goFunc := runtime.FuncForPC(goPc)
	//
	//bt, err := mrb.LoadString(`Array.new`)
	//if err != nil {
	//	panic("Failed to create RuntimeError")
	//}
	//trace := fmt.Sprintf("%s:%d:in `%s'", goFile, goLine, goFunc.Name())
	//bt.Call("push", mrb.StringValue(trace))
	//_, err = re.Call("set_backtrace", bt)
	//if err != nil {
	//	panic("Failed to set backtrace into RuntimeError")
	//}

	return re
}

func unmarshalElem(fk reflect.Kind, fe reflect.Value, rv *mruby.MrbValue) error {
	switch fk {
	case reflect.String:
		fe.SetString(rv.String())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		fe.SetInt(int64(rv.Fixnum()))

	case reflect.Float32, reflect.Float64:
		fe.SetFloat(rv.Float())

	case reflect.Bool:
		// From Ruby language spec
		if rv.Type() == mruby.TypeNil || rv.Type() == mruby.TypeFalse {
			fe.SetBool(false)
		} else {
			fe.SetBool(true)
		}

	case reflect.Slice:
		if rv.Type() != mruby.TypeArray {
			fe.Set(reflect.MakeSlice(fe.Elem().Type(), 0, 0))
			return nil
		}

		ra := rv.Array()
		slice := reflect.MakeSlice(fe.Type(), ra.Len(), ra.Len())

		for i := 0; i < ra.Len(); i++ {
			ri, err := ra.Get(i)
			if err != nil {
				return err
			}
			si := slice.Index(i)
			unmarshalElem(si.Kind(), si, ri)
		}
		fe.Set(slice)

	case reflect.Map:
		if rv.Type() != mruby.TypeHash {
			fe.Set(reflect.MakeMap(fe.Type()))
			return nil
		}

		rh := rv.Hash()
		m := reflect.MakeMap(fe.Type())

		keys, err := rh.Keys()
		if err != nil {
			return err
		}
		ka := keys.Array()
		for i := 0; i < ka.Len(); i++ {
			rvk, err := ka.Get(i)
			if err != nil {
				return err
			}
			rvv, err := rh.Get(rvk)
			if err != nil {
				return err
			}
			k := reflect.New(fe.Type().Key()).Elem()
			v := reflect.New(fe.Type().Elem()).Elem()

			unmarshalElem(fe.Type().Key().Kind(), k, rvk)
			unmarshalElem(fe.Type().Elem().Kind(), v, rvv)

			m.SetMapIndex(k, v)
		}
		fe.Set(m)

	}
	return nil
}

func Unmarshal(v *mruby.MrbValue, o interface{}) error {
	rt := reflect.TypeOf(o)
	if rt.Kind() != reflect.Ptr {
		return errors.New("value must be pointer of struct")
	}

	ov := reflect.ValueOf(o)
	oe := ov.Elem()
	if oe.Kind() != reflect.Struct {
		return errors.New("value must be pointer of struct")
	}
	ot := oe.Type()
	for i := 0; i < ot.NumField(); i++ {
		ft := ot.Field(i)
		fe := oe.Field(i)
		fm := ft.Tag.Get("mruby")

		if fm != "" {
			rv, err := v.Call(fm)
			if err != nil {
				return err
			}
			if err = unmarshalElem(ft.Type.Kind(), fe, rv); err != nil {
				return err
			}
		}
	}

	return nil
}

