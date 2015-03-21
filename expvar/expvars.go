package expvar

import "reflect"

type WalkFunc func(path string, value float32)

type Expvars map[string]interface{}

func (e Expvars) Size() int {
	size := 0

	e.Walk(func(path string, value float32) {
		size++
	})

	return size
}

func (e Expvars) Walk(fn WalkFunc) {
	for p, v := range e {
		e.walkRec(p, v, fn)
	}
}

func (e Expvars) walkRec(path string, value interface{}, fn WalkFunc) {
	val := reflect.ValueOf(value)

	switch val.Kind() {
	case reflect.Int:
		fn(path, float32(val.Int()))
	case reflect.Uint:
		fn(path, float32(val.Uint()))
	case reflect.Float32:
		fn(path, float32(val.Float()))
	case reflect.Float64:
		fn(path, float32(val.Float()))
	case reflect.Map:
		for _, key := range val.MapKeys() {
			if key.Kind() == reflect.String {
				keyStr := key.String()
				e.walkRec(path+"."+keyStr, val.MapIndex(key).Interface(), fn)
			}
		}
	}
}
