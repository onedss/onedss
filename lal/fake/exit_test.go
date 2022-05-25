// Copyright 2019, Chef.  All rights reserved.
// https://github.com/onedss/naza
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package fake_test

import (
	"testing"

	"github.com/onedss/onedss/lal/assert"
	"github.com/onedss/onedss/lal/fake"
)

func TestWithFakeExit(t *testing.T) {
	var er fake.ExitResult
	er = fake.WithFakeOsExit(func() {
		fake.Os_Exit(1)
	})
	assert.Equal(t, true, er.HasExit)
	assert.Equal(t, 1, er.ExitCode)

	er = fake.WithFakeOsExit(func() {
	})
	assert.Equal(t, false, er.HasExit)

	er = fake.WithFakeOsExit(func() {
		fake.Os_Exit(2)
	})
	assert.Equal(t, true, er.HasExit)
	assert.Equal(t, 2, er.ExitCode)
}
