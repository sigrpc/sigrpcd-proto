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
	"errors"
	"io"
	"strings"
	"unsafe"

	"github.com/sigrpc/sigrpcd/pkg/domain/model/msg"
	msgcodec "github.com/sigrpc/sigrpcd/pkg/domain/repository/msg"
	"github.com/sigrpc/sigrpcd/pkg/grpc/x64"
)

type LoadLibCodec struct {
	msgcodec.RPCHeader
}

func NewLoadLibCodec(rpcHeaderCodec msgcodec.RPCHeader) msgcodec.LoadLib {
	return &LoadLibCodec{
		RPCHeader: rpcHeaderCodec,
	}
}

func (h *LoadLibCodec) Encode(loadlib *msg.LoadLibMsg) []byte {
	pidSize := 4
	header := msg.RPCHeader{
		X64: loadlib.X64.Header,
	}
	payloadSize := uintptr(len(loadlib.X64.LibraryName) + /* null byte */ 1)
	for _, addr2sym := range loadlib.X64.GetAddr2Sym() {
		payloadSize += unsafe.Sizeof(addr2sym.Address) +
			uintptr(len(addr2sym.Name)) +
			/* null byte */ uintptr(1)
	}
	size := unsafe.Sizeof(header.X64.MsgType) +
		unsafe.Sizeof(header.X64.Status) +
		uintptr(pidSize) +
		unsafe.Sizeof(header.X64.PayloadSize) +
		payloadSize

	header.X64.PayloadSize = uint64(payloadSize)
	byteHeader := h.RPCHeader.Encode(&header)

	if payloadSize == 0 {
		return byteHeader
	}
	byteLoadLib := make([]byte, size)
	copy(byteLoadLib, byteHeader)

	offset := unsafe.Sizeof(header.X64.MsgType) +
		unsafe.Sizeof(header.X64.Status) +
		unsafe.Sizeof(header.X64.PayloadSize)
	/* encode library name */
	copy(byteLoadLib[offset:], []byte(loadlib.X64.LibraryName))
	offset += uintptr(len(loadlib.X64.LibraryName)) + /* null byte */ 1

	/* encode byte addr2sym */
	for _, addr2sym := range loadlib.X64.Addr2Sym {
		binary.LittleEndian.PutUint64(byteLoadLib[offset:], addr2sym.Address)
		offset += unsafe.Sizeof(addr2sym.Address)
		/* encode symbol name */
		copy(byteLoadLib[offset:], []byte(addr2sym.Name))
		offset += uintptr(len(addr2sym.Name)) + /* null byte */ 1
	}
	return byteLoadLib
}

func (h *LoadLibCodec) Decode(reader io.Reader, header *msg.RPCHeader) (*msg.LoadLibMsg, error) {
	offset := 0
	loadlib := msg.LoadLibMsg{
		X64: &x64.LoadLibMsg{},
	}
	loadlib.X64.Header = header.X64
	if loadlib.X64.Header.PayloadSize == 0 {
		return &loadlib, nil
	}
	bytePayload := make([]byte, loadlib.X64.Header.PayloadSize)
	size, err := reader.Read(bytePayload)
	if err != nil || size != len(bytePayload) {
		return nil, err
	}
	nullIndex := strings.Index(string(bytePayload[offset:]), "\x00")
	if nullIndex < 0 {
		return nil, errors.New("non null terminated string")
	}
	loadlib.X64.LibraryName = string(bytePayload[offset:nullIndex])
	offset += len(loadlib.X64.LibraryName) + /* null byte */ 1
	loadlib.X64.Addr2Sym = make([]*x64.Addr2Sym, 0, 10)
	for offset < int(loadlib.X64.Header.PayloadSize) {
		addr := binary.LittleEndian.Uint64(bytePayload[offset:])
		offset += /* address size */ 8
		nullIndex = strings.Index(string(bytePayload[offset:]), "\x00")
		if nullIndex < 0 {
			return nil, errors.New("non null terminated string")
		}
		name := string(bytePayload[offset : offset+nullIndex])
		offset += len(name) + /* null byte */ 1
		addr2sym := x64.Addr2Sym{
			Address: addr,
			Name:    name,
		}
		loadlib.X64.Addr2Sym = append(loadlib.X64.Addr2Sym, &addr2sym)
	}
	return &loadlib, nil
}
