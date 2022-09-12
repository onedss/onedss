// Copyright 2020, Chef.  All rights reserved.
// https://github.com/onedss/naza
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package nazahttp

import "errors"

var (
	ErrHttpHeader   = errors.New("nazahttp: fxxk")
	ErrParamMissing = errors.New("nazahttp: param missing")
)

const (
	HeaderFieldContentLength = "Content-Length"
	HeaderFieldContentType   = "application/json"
)
