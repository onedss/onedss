// Copyright 2019, Chef.  All rights reserved.
// https://github.com/onedss/naza
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

// package assert 提供了单元测试时的断言功能，减少一些模板代码
package assert

import (
	"bytes"
	"reflect"
)

// 单元测试中的 *testing.T 和 *testing.B 都满足该接口
type TestingT interface {
	Errorf(format string, args ...interface{})
}

type tHelper interface {
	Helper()
}

func Equal(t TestingT, expected interface{}, actual interface{}, msg ...string) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if !IsEqual(expected, actual) {
		t.Errorf("%s expected=%+v, actual=%+v", msg, expected, actual)
	}
	return
}

// 比如有时我们需要对 error 类型不等于 nil 做断言，但是我们并不关心 error 的具体值是什么
func IsNotNil(t TestingT, actual interface{}, msg ...string) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if IsNil(actual) {
		t.Errorf("%s expected not nil, but actual=%+v", msg, actual)
	}
	return
}

func IsNil(actual interface{}) bool {
	if actual == nil {
		return true
	}
	v := reflect.ValueOf(actual)
	k := v.Kind()
	if k == reflect.Chan || k == reflect.Map || k == reflect.Ptr || k == reflect.Interface || k == reflect.Slice {
		return v.IsNil()
	}
	return false
}

// TODO chef: 考虑是否将EqualInteger放入Equal中，但是需考虑，会给Equal带来额外的性能开销
func IsEqual(expected, actual interface{}) bool {
	if expected == nil {
		return IsNil(actual)
	}

	exp, ok := expected.([]byte)
	if ok {
		act, ok := actual.([]byte)
		if !ok {
			return false
		}
		return bytes.Equal(exp, act)
	}

	return reflect.DeepEqual(expected, actual)
}
