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

package usecase

import (
	"io"

	"github.com/sigrpc/sigrpcd/pkg/domain/model/cpu"
	cpucodec "github.com/sigrpc/sigrpcd/pkg/domain/repository/cpu"
)

type CPUCodec struct {
	cpucodec.CPU
}

func NewCPUCodec(codec cpucodec.CPU) CPUCodec {
	return CPUCodec{codec}
}

func (h *CPUCodec) Encode(cpu *cpu.CPU) []byte {
	return h.CPU.Encode(cpu)
}

func (h *CPUCodec) Decode(reader io.Reader) *cpu.CPU {
	return h.CPU.Decode(reader)
}
