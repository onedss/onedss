// Copyright 2019, Chef.  All rights reserved.
// https://github.com/onedss/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package httpflv

import (
	"net"

	"github.com/onedss/lal/pkg/base"

	"github.com/onedss/naza/pkg/connection"

	"github.com/onedss/naza/pkg/nazalog"
)

var flvHttpResponseHeader []byte

type SubSession struct {
	core                    *base.HttpSubSession
	IsFresh                 bool
	ShouldWaitVideoKeyFrame bool
}

func NewSubSession(conn net.Conn, urlCtx base.UrlContext, isWebSocket bool, websocketKey string) *SubSession {
	uk := base.GenUkFlvSubSession()
	s := &SubSession{
		core: base.NewHttpSubSession(base.HttpSubSessionOption{
			Conn: conn,
			ConnModOption: func(option *connection.Option) {
				option.WriteChanSize = SubSessionWriteChanSize
				option.WriteTimeoutMs = SubSessionWriteTimeoutMs
			},
			Uk:           uk,
			Protocol:     base.ProtocolHttpflv,
			UrlCtx:       urlCtx,
			IsWebSocket:  isWebSocket,
			WebSocketKey: websocketKey,
		}),
		IsFresh:                 true,
		ShouldWaitVideoKeyFrame: true,
	}
	nazalog.Infof("[%s] lifecycle new httpflv SubSession. session=%p, remote addr=%s", uk, s, conn.RemoteAddr().String())
	return s
}

// ---------------------------------------------------------------------------------------------------------------------
// IServerSessionLifecycle interface
// ---------------------------------------------------------------------------------------------------------------------

func (session *SubSession) RunLoop() error {
	return session.core.RunLoop()
}

func (session *SubSession) Dispose() error {
	nazalog.Infof("[%s] lifecycle dispose httpflv SubSession.", session.core.UniqueKey())
	return session.core.Dispose()
}

// ---------------------------------------------------------------------------------------------------------------------

func (session *SubSession) WriteHttpResponseHeader() {
	nazalog.Debugf("[%s] > W http response header.", session.core.UniqueKey())
	session.core.WriteHttpResponseHeader(flvHttpResponseHeader)
}

func (session *SubSession) WriteFlvHeader() {
	nazalog.Debugf("[%s] > W http flv header.", session.core.UniqueKey())
	session.core.Write(FlvHeader)
}

func (session *SubSession) WriteTag(tag *Tag) {
	session.core.Write(tag.Raw)
}

func (session *SubSession) Write(b []byte) {
	session.core.Write(b)
}

// ---------------------------------------------------------------------------------------------------------------------
// IObject interface
// ---------------------------------------------------------------------------------------------------------------------

func (session *SubSession) UniqueKey() string {
	return session.core.UniqueKey()
}

// ---------------------------------------------------------------------------------------------------------------------
// ISessionUrlContext interface
// ---------------------------------------------------------------------------------------------------------------------

func (session *SubSession) Url() string {
	return session.core.Url()
}

func (session *SubSession) AppName() string {
	return session.core.AppName()
}

func (session *SubSession) StreamName() string {
	return session.core.StreamName()
}

func (session *SubSession) RawQuery() string {
	return session.core.RawQuery()
}

// ---------------------------------------------------------------------------------------------------------------------
// ISessionStat interface
// ---------------------------------------------------------------------------------------------------------------------

func (session *SubSession) UpdateStat(intervalSec uint32) {
	session.core.UpdateStat(intervalSec)
}

func (session *SubSession) GetStat() base.StatSession {
	return session.core.GetStat()
}

func (session *SubSession) IsAlive() (readAlive, writeAlive bool) {
	return session.core.IsAlive()
}

// ---------------------------------------------------------------------------------------------------------------------

func init() {
	flvHttpResponseHeaderStr := "HTTP/1.1 200 OK\r\n" +
		"Server: " + base.LalHttpflvSubSessionServer + "\r\n" +
		"Cache-Control: no-cache\r\n" +
		"Content-Type: video/x-flv\r\n" +
		"Connection: close\r\n" +
		"Expires: -1\r\n" +
		"Pragma: no-cache\r\n" +
		"Access-Control-Allow-Credentials: true\r\n" +
		"Access-Control-Allow-Origin: *\r\n" +
		"\r\n"

	flvHttpResponseHeader = []byte(flvHttpResponseHeaderStr)
}