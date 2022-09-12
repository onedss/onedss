// Copyright 2020, Chef.  All rights reserved.
// https://github.com/onedss/naza
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package defertaskthread

var thread DeferTaskThread

func Go(deferMs int, task TaskFn, param ...interface{}) {
	thread.Go(deferMs, task, param...)
}

func init() {
	thread = NewDeferTaskThread()
}
