package credentials

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"reflect"
)

func randomString(requestedLength int) string {
	encoding := base64.RawURLEncoding
	rawLength := encoding.DecodedLen(requestedLength)
	raw := make([]byte, rawLength)
	_, err := rand.Read(raw)
	if err != nil {
		panic(err)
	}
	return encoding.EncodeToString(raw)[:requestedLength]
}

type Generator struct {
	Length int
}

func (g Generator) Fill(toFill interface{}) error {
	if g.Length < 1 {
		return errors.New("length must be positive")
	}
	t := reflect.TypeOf(toFill)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return errors.New("expecting a pointer to a struct")
	}
	ptrValue := reflect.ValueOf(toFill)
	if ptrValue.IsNil() {
		return errors.New("pointer must not be nil")
	}
	v := reflect.ValueOf(toFill).Elem()
	numFields := v.NumField()
	for i := 0; i < numFields; i++ {
		v.Field(i).Set(reflect.ValueOf(randomString(g.Length)))
	}
	return nil
}
