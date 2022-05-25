// Copyright 2020, Chef.  All rights reserved.
// https://github.com/onedss/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package rtmp_test

import (
	"testing"

	"github.com/onedss/onedss/lal/base"

	"github.com/onedss/onedss/lal/assert"
	"github.com/onedss/onedss/lal/nazalog"
	"github.com/onedss/onedss/rtmp"
)

func TestMetadata(t *testing.T) {
	b, err := rtmp.BuildMetadata(1024, 768, 10, 7)
	assert.Equal(t, nil, err)

	opa, err := rtmp.ParseMetadata(b)
	assert.Equal(t, nil, err)
	nazalog.Debugf("%+v", opa)

	assert.Equal(t, 5, len(opa))
	v := opa.Find("width")
	assert.Equal(t, float64(1024), v.(float64))
	v = opa.Find("height")
	assert.Equal(t, float64(768), v.(float64))
	v = opa.Find("audiocodecid")
	assert.Equal(t, float64(10), v.(float64))
	v = opa.Find("videocodecid")
	assert.Equal(t, float64(7), v.(float64))
	v = opa.Find("version")
	assert.Equal(t, base.LalRtmpBuildMetadataEncoder, v.(string))
}
