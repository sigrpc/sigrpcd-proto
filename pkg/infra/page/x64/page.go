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
	"io"
	"unsafe"

	"github.com/sigrpc/sigrpcd/pkg/domain/model/page"
	pagecodec "github.com/sigrpc/sigrpcd/pkg/domain/repository/page"
	"github.com/sigrpc/sigrpcd/pkg/grpc/x64"
)

type PageCodec struct{}

func NewPageCodec() pagecodec.Page {
	return &PageCodec{}
}

func (h *PageCodec) Encode(page *page.Page) []byte {
	propertySize := unsafe.Sizeof(page.X64.Address) +
		unsafe.Sizeof(page.X64.RuntimeRevision) +
		unsafe.Sizeof(page.X64.ClientRevision) +
		unsafe.Sizeof(page.X64.ContentSize)
	bytePage := make(
		[]byte,
		propertySize,
		propertySize+uintptr(page.X64.ContentSize))
	offset := 0
	binary.LittleEndian.PutUint64(bytePage[offset:], page.X64.Address)
	offset += int(unsafe.Sizeof(page.X64.Address))
	binary.LittleEndian.PutUint64(bytePage[offset:], page.X64.RuntimeRevision)
	offset += int(unsafe.Sizeof(page.X64.RuntimeRevision))
	binary.LittleEndian.PutUint64(bytePage[offset:], page.X64.ClientRevision)
	offset += int(unsafe.Sizeof(page.X64.ClientRevision))
	binary.LittleEndian.PutUint32(bytePage[offset:], page.X64.ContentSize)
	if len(page.X64.Content) > 0 {
		bytePage = append(bytePage, page.X64.Content...)
	}

	return bytePage
}

func (h *PageCodec) Decode(reader io.Reader) *page.Page {
	page := page.Page{
		X64: &x64.Page{},
	}
	err := binary.Read(reader, binary.LittleEndian, &page.X64.Address)
	if err != nil {
		return nil
	}
	err = binary.Read(reader, binary.LittleEndian, &page.X64.RuntimeRevision)
	if err != nil {
		return nil
	}
	err = binary.Read(reader, binary.LittleEndian, &page.X64.ClientRevision)
	if err != nil {
		return nil
	}
	err = binary.Read(reader, binary.LittleEndian, &page.X64.ContentSize)
	if err != nil {
		return nil
	}
	content := make([]byte, page.X64.ContentSize)
	err = binary.Read(reader, binary.LittleEndian, &content)
	if err == io.EOF {
		return &page
	}
	if err != nil {
		return nil
	}
	page.X64.Content = content
	return &page
}
