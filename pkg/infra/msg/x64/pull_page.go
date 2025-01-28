// Copyright 2025 Keita HAGIWARA. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package x64

import (
	"io"

	"github.com/sigrpc/sigrpcd/pkg/domain/model/msg"
	"github.com/sigrpc/sigrpcd/pkg/domain/model/page"
	msgcodec "github.com/sigrpc/sigrpcd/pkg/domain/repository/msg"
	pagecodec "github.com/sigrpc/sigrpcd/pkg/domain/repository/page"
	"github.com/sigrpc/sigrpcd/pkg/grpc/x64"
)

type PullPageCodec struct {
	pagecodec.Page
	msgcodec.RPCHeader
}

func NewPullPageCodec(pageCodec pagecodec.Page, rpcHeaderCodec msgcodec.RPCHeader) msgcodec.PullPage {
	return &PullPageCodec{
		Page:      pageCodec,
		RPCHeader: rpcHeaderCodec,
	}
}

func (h *PullPageCodec) Encode(pullpage *msg.PullPageMsg) []byte {
	var bytePayload []byte

	for _, x64page := range pullpage.X64.Page {
		page := page.Page{
			X64: x64page,
		}
		bytepage := h.Page.Encode(&page)
		bytePayload = append(bytePayload, bytepage...)
	}
	pullpage.X64.Header.PayloadSize = uint64(len(bytePayload))
	header := msg.RPCHeader{
		X64: pullpage.X64.Header,
	}
	byteHeader := h.RPCHeader.Encode(&header)
	bytePullPage := byteHeader
	bytePullPage = append(bytePullPage, bytePayload...)
	return bytePullPage
}

func (h *PullPageCodec) Decode(reader io.Reader, header *msg.RPCHeader) (*msg.PullPageMsg, error) {
	pullPageMsg := msg.PullPageMsg{
		X64: &x64.PullPageMsg{
			Page: make([]*x64.Page, 0),
		},
	}
	pullPageMsg.X64.Header = header.X64
	if pullPageMsg.X64.Header.PayloadSize == 0 {
		return &pullPageMsg, nil
	}
	for {
		p := h.Page.Decode(reader)
		if p == nil {
			break
		}
		pullPageMsg.X64.Page = append(pullPageMsg.X64.Page, p.X64)
	}
	return &pullPageMsg, nil
}
