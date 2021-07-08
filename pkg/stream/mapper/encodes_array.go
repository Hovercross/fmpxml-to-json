package mapper

import "github.com/francoispqt/gojay"

type encodableArrayItem func(enc *gojay.Encoder)

type encodableArray []encodableArrayItem

func (ds encodableArray) IsNil() bool {
	return false
}

func (ds encodableArray) MarshalJSONArray(enc *gojay.Encoder) {
	for _, e := range ds {
		e(enc)
	}
}
