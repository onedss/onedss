// Copyright 2021, Chef.  All rights reserved.
// https://github.com/onedss/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package hls

var (
	PathStrategy IPathStrategy = &DefaultPathStrategy{}
)

var (
	calcFragmentHeaderQueueSize = 16
)
