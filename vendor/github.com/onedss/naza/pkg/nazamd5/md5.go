// Copyright 2019, Chef.  All rights reserved.
// https://github.com/onedss/naza
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package nazamd5

import (
	"crypto/md5"
	"encoding/hex"
)

// 返回32字节小写字符串
func Md5(b []byte) string {
	h := md5.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
