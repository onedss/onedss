// Copyright 2020, Chef.  All rights reserved.
// https://github.com/onedss/naza
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package fake_test

import (
	"testing"

	"github.com/onedss/onedss/lal/fake"
)

func TestWithRecover(t *testing.T) {
	fake.WithRecover(func() {
		// noop
	})
	fake.WithRecover(func() {
		panic(0)
	})
}
