package store

import (
	"fmt"
	"reflect"
)

func reflectDeepCopyInto(source, target interface{}) error {
	sv := reflect.TypeOf(source)
	tv := reflect.TypeOf(target)
	if tv != sv {
		return fmt.Errorf("can not deep copy different types: from %T to %T ", source, target)
	}

	st := reflect.TypeOf(source)
	method, ok := st.MethodByName("DeepCopyInto")
	if !ok {
		return fmt.Errorf("type %T does not have a DeepCopyInto method", sv)
	}

	if numIn := method.Type.NumIn(); numIn != 2 {
		return fmt.Errorf("%T's DeepCopyInto requires more arguments than expected. expected 2, requiring %v", source, numIn)
	}

	if argT := method.Type.In(0); argT != sv {
		return fmt.Errorf("%T's DeepCopyInto requires unexpected type. expected %s, requiring %s", source, argT.Name(), sv.Name())
	}

	return doReflectDeepCopyInto(source, target)
}

func doReflectDeepCopyInto(source, target interface{}) error {
	i := reflectDeepCopyIntoInvoker{}
	i.doReflectDeepCopyInto(source, target)
	return i.err
}

type reflectDeepCopyIntoInvoker struct {
	err error
}

func (i *reflectDeepCopyIntoInvoker) doReflectDeepCopyInto(source, target interface{}) {
	defer func() {
		if err := recover(); err != nil {
			i.err = fmt.Errorf("error invoking DeepCopyInto from %T to %T: %v", source, target, err)
		}
	}()

	reflect.
		ValueOf(source).
		MethodByName("DeepCopyInto").
		Call([]reflect.Value{reflect.ValueOf(target)})
}
