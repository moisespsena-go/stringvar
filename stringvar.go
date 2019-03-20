package stringvar

import (
	"fmt"
	"path/filepath"
	"strings"
)

type StringVar struct {
	Open   string
	Close  string
	Data   map[string]interface{}
	parent *StringVar
}

func New(data ...interface{}) *StringVar {
	v := &StringVar{"{", "}", map[string]interface{}{}, nil}
	v.Merge(data...)
	return v
}

func (v *StringVar) Merge(data ...interface{}) *StringVar {
	for i, l := 0, len(data); i < l; i++ {
		if vs := data[i]; vs != nil {
			switch vt := vs.(type) {
			case string:
				v.Data[vt] = data[i+1]
				i++
			case [2]string:
				v.Data[vt[0]] = vt[1]
			case map[string]interface{}:
				if vt != nil {
					v = &StringVar{v.Open, v.Close, vt, v}
				}
			case map[string]string:
				if vt != nil {
					d := map[string]interface{}{}
					for k, v := range vt {
						d[k] = v
					}
					v = &StringVar{v.Open, v.Close, d, v}
				}
			}
		}
	}
	return v
}

func (v *StringVar) Child(data ...interface{}) *StringVar {
	vs := New(data...)
	vs.parent = v
	return vs
}

func (v *StringVar) Pairs(cb func(k string, v interface{})) {
	for key, value := range v.Data {
		cb(key, value)
	}
}

func (v *StringVar) Walk(cb func(k string, v interface{})) {
	for key, value := range v.Data {
		cb(key, value)
	}
	if v.parent != nil {
		v.parent.Walk(cb)
	}
}

func (v *StringVar) Priority() (ld []*StringVar) {
	for v != nil {
		ld = append(ld, v)
		v = v.parent
	}
	return ld
}

func (v StringVar) Format(s string) string {
	for _, d := range v.Priority() {
		d.Walk(func(key string, value interface{}) {
			s = strings.Replace(s, v.Open+key+v.Close, fmt.Sprint(value), -1)
		})
	}

	return s
}

func (v StringVar) FormatPath(s string) string {
	return filepath.Clean(v.Format(s))
}

func (v *StringVar) FormatPtr(sptrs ...*string) *StringVar {
	for _, s := range sptrs {
		*s = v.Format(*s)
	}
	return v
}

func (v *StringVar) FormatPathPtr(sptrs ...*string) *StringVar {
	for _, s := range sptrs {
		*s = filepath.Clean(v.Format(*s))
	}
	return v
}

func (v *StringVar) Get(key string) (r interface{}, ok bool) {
	for v != nil {
		if r, ok = v.Data[key]; ok {
			return
		}
		v = v.parent
	}
	return
}

func (v *StringVar) GetData() map[string]interface{} {
	if v.parent == nil {
		return v.Data
	}

	d := map[string]interface{}{}
	for v != nil {
		v.Pairs(func(k string, v interface{}) {
			if _, ok := d[k]; !ok {
				d[k] = v
			}
		})
		v = v.parent
	}
	return d
}

func (v *StringVar) Clone() *StringVar {
	sv := &StringVar{v.Open, v.Close, map[string]interface{}{}, nil}
	v.Walk(func(k string, v interface{}) {
		sv.Data[k] = v
	})
	return sv
}
