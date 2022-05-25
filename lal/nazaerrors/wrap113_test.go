// Copyright 2021, Chef.  All rights reserved.
// https://github.com/onedss/naza
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

// +build go1.13

package nazaerrors

import (
	"errors"
	"io"
	"testing"

	"github.com/onedss/onedss/lal/assert"

	"github.com/onedss/onedss/lal/nazalog"
)

func TestWrap(t *testing.T) {
	err := Wrap(io.EOF)
	nazalog.Debugf("%+v", err)
	assert.Equal(t, true, errors.Is(err, io.EOF))
	err = Wrap(err)
	nazalog.Debugf("%+v", err)
	assert.Equal(t, true, errors.Is(err, io.EOF))
}
