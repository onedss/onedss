// Copyright 2020, Chef.  All rights reserved.
// https://github.com/onedss/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package hls

import (
	"net/http"

	"github.com/onedss/lal/pkg/base"

	"github.com/onedss/naza/pkg/nazalog"
)

type ServerHandler struct {
	outPath string
	//addr    string
	//ln      net.Listener
	//httpSrv *http.Server
}

func NewServerHandler(outPath string) *ServerHandler {
	return &ServerHandler{
		outPath: outPath,
	}
}

//
//func (s *Server) Listen() (err error) {
//	if s.ln, err = net.Listen("tcp", s.addr); err != nil {
//		return
//	}
//	s.httpSrv = &http.Server{Addr: s.addr, Handler: s}
//	nazalog.Infof("start hls server listen. addr=%s", s.addr)
//	return
//}
//
//func (s *Server) RunLoop() error {
//	return s.httpSrv.Serve(s.ln)
//}
//
//func (s *Server) Dispose() {
//	if err := s.httpSrv.Close(); err != nil {
//		nazalog.Error(err)
//	}
//}

func (s *ServerHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	//nazalog.Debugf("%+v", req)

	// TODO chef:
	// - check appname in URI path

	ri := PathStrategy.GetRequestInfo(req.RequestURI, s.outPath)
	//nazalog.Debugf("%+v", ri)

	if ri.FileName == "" || ri.StreamName == "" || ri.FileNameWithPath == "" || (ri.FileType != "m3u8" && ri.FileType != "ts") {
		nazalog.Warnf("invalid hls request. uri=%s, request=%+v", req.RequestURI, ri)
		resp.WriteHeader(404)
		return
	}

	content, err := ReadFile(ri.FileNameWithPath)
	if err != nil {
		nazalog.Warnf("read hls file failed. request=%+v, err=%+v", ri, err)
		resp.WriteHeader(404)
		return
	}

	switch ri.FileType {
	case "m3u8":
		resp.Header().Add("Content-Type", "application/x-mpegurl")
		resp.Header().Add("Server", base.LalHlsM3u8Server)
	case "ts":
		resp.Header().Add("Content-Type", "video/mp2t")
		resp.Header().Add("Server", base.LalHlsTsServer)
	}
	resp.Header().Add("Cache-Control", "no-cache")
	resp.Header().Add("Access-Control-Allow-Origin", "*")

	_, _ = resp.Write(content)
	return
}

// m3u8文件用这个也行
//resp.Header().Add("Content-Type", "application/vnd.apple.mpegurl")

//resp.Header().Add("Access-Control-Allow-Origin", "*")
//resp.Header().Add("Access-Control-Allow-Credentials", "true")
//resp.Header().Add("Access-Control-Allow-Methods", "*")
//resp.Header().Add("Access-Control-Allow-Headers", "Content-Type,Access-Token")
//resp.Header().Add("Access-Control-Allow-Expose-Headers", "*")
