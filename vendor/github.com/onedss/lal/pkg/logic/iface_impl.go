// Copyright 2021, Chef.  All rights reserved.
// https://github.com/onedss/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package logic

import (
	"github.com/onedss/lal/pkg/base"
	"github.com/onedss/lal/pkg/hls"
	"github.com/onedss/lal/pkg/httpflv"
	"github.com/onedss/lal/pkg/httpts"
	"github.com/onedss/lal/pkg/remux"
	"github.com/onedss/lal/pkg/rtmp"
	"github.com/onedss/lal/pkg/rtsp"
)

// TODO(chef): 整理所有Server类型Session的生命周期管理
//   -
//   - rtmp没有独立的Pub、Sub Session结构体类型，而是直接使用ServerSession
//   - write失败，需要反应到loop来
//   - rtsp是否也应该上层使用Command作为代理，避免生命周期管理混乱
//
// server.pub:  rtmp(), rtsp
// server.sub:  rtmp(), rtsp, flv, ts
//
// client.push: rtmp, rtsp
// client.pull: rtmp, rtsp, flv
//
// other:       rtmp.ClientSession, rtmp.ServerSession
//              rtsp.BaseInSession, rtsp.BaseOutSession, rtsp.ClientCommandSession, rtsp.ServerCommandSessionS
//              base.HttpSubSession

// IClientSession: 所有Client Session都满足
var (
	_ base.IClientSession = &rtmp.PushSession{}
	_ base.IClientSession = &rtmp.PullSession{}
	_ base.IClientSession = &rtsp.PushSession{}
	_ base.IClientSession = &rtsp.PullSession{}
	_ base.IClientSession = &httpflv.PullSession{}
)

// IServerSession
var (
	_ base.IServerSession = &rtmp.ServerSession{}
	_ base.IServerSession = &httpflv.SubSession{}
	_ base.IServerSession = &httpts.SubSession{}

	// 这两个比较特殊，它们没有RunLoop函数，RunLoop在rtsp.ServerCommandSession上
	//_ base.IServerSession = &rtsp.PubSession{}
	//_ base.IServerSession = &rtsp.SubSession{}
)

// IClientSessionLifecycle: 所有Client Session都满足
var (
	// client
	_ base.IClientSessionLifecycle = &rtmp.PushSession{}
	_ base.IClientSessionLifecycle = &rtmp.PullSession{}
	_ base.IClientSessionLifecycle = &rtsp.PushSession{}
	_ base.IClientSessionLifecycle = &rtsp.PullSession{}
	_ base.IClientSessionLifecycle = &httpflv.PullSession{}

	// other
	_ base.IClientSessionLifecycle = &rtmp.ClientSession{}
	_ base.IClientSessionLifecycle = &rtsp.ClientCommandSession{}
)

// IServerSessionLifecycle
var (
	// server session
	_ base.IServerSessionLifecycle = &rtmp.ServerSession{}
	_ base.IServerSessionLifecycle = &httpflv.SubSession{}
	_ base.IServerSessionLifecycle = &httpts.SubSession{}

	// 这两个比较特殊，它们没有RunLoop函数，RunLoop在rtsp.ServerCommandSession上
	//_ base.IServerSessionLifecycle = &rtsp.PubSession{}
	//_ base.IServerSessionLifecycle = &rtsp.SubSession{}
	// other
	_ base.IServerSessionLifecycle = &base.HttpSubSession{}
	_ base.IServerSessionLifecycle = &rtsp.ServerCommandSession{}
)

// ISessionStat: 所有Session(client/server)都满足
var (
	// client
	_ base.ISessionStat = &rtmp.PushSession{}
	_ base.ISessionStat = &rtsp.PushSession{}
	_ base.ISessionStat = &rtmp.PullSession{}
	_ base.ISessionStat = &rtsp.PullSession{}
	_ base.ISessionStat = &httpflv.PullSession{}
	// server session
	_ base.ISessionStat = &rtmp.ServerSession{}
	_ base.ISessionStat = &rtsp.PubSession{}
	_ base.ISessionStat = &rtsp.SubSession{}
	_ base.ISessionStat = &httpflv.SubSession{}
	_ base.ISessionStat = &httpts.SubSession{}
	// other
	_ base.ISessionStat = &base.HttpSubSession{}
	_ base.ISessionStat = &rtmp.ClientSession{}
	_ base.ISessionStat = &rtsp.BaseInSession{}
	_ base.ISessionStat = &rtsp.BaseOutSession{}
	_ base.ISessionStat = &rtsp.ServerCommandSession{}
)

// ISessionUrlContext: 所有Session(client/server)都满足
var (
	// client
	_ base.ISessionUrlContext = &rtmp.PushSession{}
	_ base.ISessionUrlContext = &rtsp.PushSession{}
	_ base.ISessionUrlContext = &rtmp.PullSession{}
	_ base.ISessionUrlContext = &rtsp.PullSession{}
	_ base.ISessionUrlContext = &httpflv.PullSession{}
	// server session
	_ base.ISessionUrlContext = &rtmp.ServerSession{}
	_ base.ISessionUrlContext = &rtsp.PubSession{}
	_ base.ISessionUrlContext = &rtsp.SubSession{}
	_ base.ISessionUrlContext = &httpflv.SubSession{}
	_ base.ISessionUrlContext = &httpts.SubSession{}
	// other
	_ base.ISessionUrlContext = &base.HttpSubSession{}
	_ base.ISessionUrlContext = &rtmp.ClientSession{}
	_ base.ISessionUrlContext = &rtsp.ClientCommandSession{}
)

// IObject: 所有Session(client/server)都满足
var (
	//// client
	_ base.IObject = &rtmp.PushSession{}
	_ base.IObject = &rtsp.PushSession{}
	_ base.IObject = &rtmp.PullSession{}
	_ base.IObject = &rtsp.PullSession{}
	_ base.IObject = &httpflv.PullSession{}
	// server session
	_ base.IObject = &rtmp.ServerSession{}
	_ base.IObject = &rtsp.PubSession{}
	_ base.IObject = &rtsp.SubSession{}
	_ base.IObject = &httpflv.SubSession{}
	_ base.IObject = &httpts.SubSession{}
	//// other
	_ base.IObject = &base.HttpSubSession{}
	_ base.IObject = &rtmp.ClientSession{}
	_ base.IObject = &rtsp.BaseInSession{}
	_ base.IObject = &rtsp.BaseOutSession{}
	_ base.IObject = &rtsp.ClientCommandSession{}
	_ base.IObject = &rtsp.ServerCommandSession{}
)

// ---------------------------------------------------------------------------------------------------------------------

var _ rtmp.ServerObserver = &ServerManager{}
var _ rtsp.ServerObserver = &ServerManager{}
var _ HttpServerHandlerObserver = &ServerManager{}

var _ HttpApiServerObserver = &ServerManager{}

var _ rtmp.PubSessionObserver = &Group{} //
var _ rtsp.PullSessionObserver = &Group{}
var _ rtsp.PullSessionObserver = &remux.AvPacket2RtmpRemuxer{}
var _ rtsp.PubSessionObserver = &Group{}
var _ rtsp.PubSessionObserver = &remux.AvPacket2RtmpRemuxer{}
var _ hls.MuxerObserver = &Group{}
var _ rtsp.BaseInSessionObserver = &Group{} //
var _ rtsp.BaseInSessionObserver = &remux.AvPacket2RtmpRemuxer{}

var _ rtmp.ServerSessionObserver = &rtmp.Server{}
var _ rtmp.IHandshakeClient = &rtmp.HandshakeClientSimple{}
var _ rtmp.IHandshakeClient = &rtmp.HandshakeClientComplex{}

var _ rtsp.ServerCommandSessionObserver = &rtsp.Server{}
var _ rtsp.ClientCommandSessionObserver = &rtsp.PushSession{}
var _ rtsp.ClientCommandSessionObserver = &rtsp.PullSession{}
var _ rtsp.IInterleavedPacketWriter = &rtsp.PushSession{}
var _ rtsp.IInterleavedPacketWriter = &rtsp.PullSession{}
var _ rtsp.IInterleavedPacketWriter = &rtsp.PubSession{}
var _ rtsp.IInterleavedPacketWriter = &rtsp.SubSession{}
var _ rtsp.IInterleavedPacketWriter = &rtsp.ClientCommandSession{}
var _ rtsp.IInterleavedPacketWriter = &rtsp.ServerCommandSession{}

var _ hls.StreamerObserver = &hls.Muxer{}
