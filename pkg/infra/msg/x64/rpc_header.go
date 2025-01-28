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
	"encoding/binary"
	"net"
	"strconv"
	"strings"
	"unsafe"

	"github.com/sigrpc/sigrpcd/pkg/domain/model/msg"
	msgcodec "github.com/sigrpc/sigrpcd/pkg/domain/repository/msg"
	"github.com/sigrpc/sigrpcd/pkg/grpc/x64"
)

type RPCHeaderCodec struct {
	clientID string
}

func NewRPCHeaderCodec(clientID string) msgcodec.RPCHeader {
	return &RPCHeaderCodec{
		clientID: clientID,
	}
}

func (h *RPCHeaderCodec) Encode(header *msg.RPCHeader) []byte {
	pidStartPos := strings.LastIndex(header.X64.ClientId, "-")
	if pidStartPos == -1 || pidStartPos+1 >= len(header.X64.ClientId) {
		return nil
	}
	pidStr := header.X64.ClientId[pidStartPos+1:]
	pidU64, err := strconv.ParseUint(pidStr, 16, 32)
	if err != nil {
		return nil
	}
	pid := uint32(pidU64)
	byteHeader := make([]byte,
		unsafe.Sizeof(header.X64.MsgType)+
			unsafe.Sizeof(header.X64.Status)+
			unsafe.Sizeof(pid)+
			unsafe.Sizeof(header.X64.PayloadSize))
	offset := 0
	binary.LittleEndian.PutUint32(byteHeader[offset:], header.X64.MsgType)
	offset += int(unsafe.Sizeof(header.X64.MsgType))
	binary.LittleEndian.PutUint32(byteHeader[offset:], header.X64.Status)
	offset += int(unsafe.Sizeof(header.X64.Status))
	binary.LittleEndian.PutUint32(byteHeader[offset:], pid)
	offset += int(unsafe.Sizeof(pid))
	binary.LittleEndian.PutUint64(byteHeader[offset:], header.X64.PayloadSize)
	return byteHeader
}

func (h *RPCHeaderCodec) Decode(conn net.Conn) (*msg.RPCHeader, error) {
	header := msg.RPCHeader{
		X64: &x64.RPCHeader{},
	}
	var pid uint32
	buf := make(
		[]byte,
		unsafe.Sizeof(header.X64.MsgType)+
			unsafe.Sizeof(header.X64.Status)+
			unsafe.Sizeof(pid)+
			unsafe.Sizeof(header.X64.PayloadSize))
	readTotal := uint64(0)
	for readTotal < uint64(len(buf)) {
		size, err := conn.Read(buf[readTotal:])
		if err != nil && size == 0 {
			return nil, err
		}
		readTotal += uint64(size)
	}
	header.X64.MsgType = binary.LittleEndian.Uint32(buf)
	buf = buf[unsafe.Sizeof(header.X64.MsgType):]
	header.X64.Status = binary.LittleEndian.Uint32(buf)
	buf = buf[unsafe.Sizeof(header.X64.Status):]
	pid = binary.LittleEndian.Uint32(buf)
	header.X64.ClientId = h.clientID + "-" + strconv.FormatUint(uint64(pid), 16)
	buf = buf[unsafe.Sizeof(pid):]
	header.X64.PayloadSize = binary.LittleEndian.Uint64(buf)
	return &header, nil
}
