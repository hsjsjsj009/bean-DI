package beans

import (
	"fmt"
	"reflect"
)

type bean struct {
	dependency []reflect.Type
	initType reflect.Type
	initFun interface{}
	container *ProviderContainer
	haveError bool
	singletonObjectValue []reflect.Value
}

func newBean(initFun interface{},container *ProviderContainer,singleton bool) (reflect.Type,*bean,error) {
	var dependencyList []reflect.Type
	funType := reflect.TypeOf(initFun)
	kind := funType.Kind()
	if kind != reflect.Func {
		return nil, nil, fmt.Errorf(mustBeFunc)
	}
	numOut := funType.NumOut()
	if numOut < 1 || numOut > 2 {
		return nil, nil, fmt.Errorf(outputRestriction)
	}
	numIn := funType.NumIn()
	for i := 0;i<numIn;i++ {
		inType := funType.In(i)
		dependencyList = append(dependencyList,inType)
	}
	returnType := funType.Out(0)
	outputBean := &bean{
		dependency: dependencyList,
		initFun: initFun,
		container: container,
		initType: returnType,
	}
	if numOut == 2 {
		if secType := funType.Out(1);secType.String() != "error" {
			return nil, nil, fmt.Errorf(outputRestriction)
		}
		outputBean.haveError = true
	}
	if singleton {
		val,err := outputBean.call()
		if err != nil {
			return nil, nil, err
		}
		outputBean.singletonObjectValue = val
	}

	return returnType,outputBean,nil
}

func(b *bean) call() ([]reflect.Value,error) {
	if len(b.singletonObjectValue) > 0 {
		return b.singletonObjectValue,nil
	}
	var listDependency []reflect.Value
	for _,dep := range b.dependency{
		depBean,ok := b.container.dependencyInitiator[dep]
		if !ok {
			return nil, errorDepNotFound(dep)
		}
		depVal,err := depBean.call()
		if err != nil {
			return nil, err
		}
		listDependency = append(listDependency,depVal[0])
	}
	retVal := reflect.ValueOf(b.initFun).Call(listDependency)
	if b.haveError {
		if errData := retVal[1];!retVal[1].IsNil() {
			return nil,errorDepReturnError(b.initType,errData.Interface().(error))
		}
	}
	return retVal,nil
}
