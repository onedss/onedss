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
	"log"
	"testing"

	"github.com/onedss/onedss/lal/assert"
)

func TestWrap(t *testing.T) {
	err := Wrap(io.EOF)
	log.Printf("%+v", err)
	assert.Equal(t, true, errors.Is(err, io.EOF))
	err = Wrap(err)
	log.Printf("%+v", err)
	assert.Equal(t, true, errors.Is(err, io.EOF))
}
