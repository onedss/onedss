// Copyright 2020, Chef.  All rights reserved.
// https://github.com/onedss/naza
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package nazahttp

import (
	"net/http"
	"strings"
)

type LineReader interface {
	ReadLine() (line []byte, isPrefix bool, err error)
}

// @return firstLine: request的request line或response的status line
// @return headers: request header fileds的键值对
func ReadHttpHeader(r LineReader) (firstLine string, headers http.Header, err error) {
	headers = make(http.Header)

	var line []byte
	var isPrefix bool
	line, isPrefix, err = r.ReadLine()
	if err != nil {
		return
	}
	if len(line) == 0 || isPrefix {
		err = ErrHttpHeader
		return
	}
	firstLine = string(line)

	for {
		line, isPrefix, err = r.ReadLine()
		if len(line) == 0 { // 读到一个空的 \r\n 表示http头全部读取完毕了
			break
		}
		if isPrefix {
			err = ErrHttpHeader
			return
		}
		if err != nil {
			return
		}
		l := string(line)
		pos := strings.Index(l, ":")
		if pos == -1 {
			err = ErrHttpHeader
			return
		}
		headers.Add(strings.Trim(l[0:pos], " "), strings.Trim(l[pos+1:], " "))
	}
	return
}

// Request-Line = Method SP URI SP Version CRLF
func ParseHttpRequestLine(line string) (method string, uri string, version string, err error) {
	return parseFirstLine(line)
}

// Status-Line = Version SP Status-Code SP Reason CRLF
func ParseHttpStatusLine(line string) (version string, statusCode string, reason string, err error) {
	return parseFirstLine(line)
}

func parseFirstLine(line string) (item1, item2, item3 string, err error) {
	f := strings.Index(line, " ")
	if f == -1 {
		err = ErrHttpHeader
		return
	}
	s := strings.Index(line[f+1:], " ")
	if s == -1 || f+1+s+1 == len(line) {
		err = ErrHttpHeader
		return
	}

	return line[0:f], line[f+1 : f+1+s], line[f+1+s+1:], nil
}
