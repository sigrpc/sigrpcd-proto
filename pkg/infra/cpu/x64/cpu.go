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

	"github.com/sigrpc/sigrpcd/pkg/domain/model/cpu"
	codec "github.com/sigrpc/sigrpcd/pkg/domain/repository/cpu"
	"github.com/sigrpc/sigrpcd/pkg/grpc/x64"
)

type CPUCodec struct{}

func NewCodec() codec.CPU {
	return &CPUCodec{}
}

func (h *CPUCodec) Encode(cpuState *cpu.CPU) []byte {
	gregMax := cpu.CR2 + 1
	gregsSize := 8 * gregMax
	fpstateSize := 512
	byteData := make([]byte, gregsSize+fpstateSize)
	head := byteData
	// x64 state
	// gregs
	for i := 0; i < gregMax; i++ {
		binary.LittleEndian.PutUint64(head, cpuState.X64.Gregs[i])
		head = head[unsafe.Sizeof(cpuState.X64.Gregs[i]):]
	}
	// fpregs
	binary.LittleEndian.PutUint16(head, uint16(cpuState.X64.Fpregs.Cwd))
	head = head[unsafe.Sizeof(uint16(cpuState.X64.Fpregs.Cwd)):]
	binary.LittleEndian.PutUint16(head, uint16(cpuState.X64.Fpregs.Swd))
	head = head[unsafe.Sizeof(uint16(cpuState.X64.Fpregs.Swd)):]
	binary.LittleEndian.PutUint16(head, uint16(cpuState.X64.Fpregs.Ftw))
	head = head[unsafe.Sizeof(uint16(cpuState.X64.Fpregs.Ftw)):]
	binary.LittleEndian.PutUint16(head, uint16(cpuState.X64.Fpregs.Fop))
	head = head[unsafe.Sizeof(uint16(cpuState.X64.Fpregs.Fop)):]
	binary.LittleEndian.PutUint64(head, cpuState.X64.Fpregs.Rip)
	head = head[unsafe.Sizeof(cpuState.X64.Fpregs.Rip):]
	binary.LittleEndian.PutUint64(head, cpuState.X64.Fpregs.Rdp)
	head = head[unsafe.Sizeof(cpuState.X64.Fpregs.Rdp):]
	binary.LittleEndian.PutUint32(head, cpuState.X64.Fpregs.Mxcsr)
	head = head[unsafe.Sizeof(cpuState.X64.Fpregs.Mxcsr):]
	binary.LittleEndian.PutUint32(head, cpuState.X64.Fpregs.MxcrMask)
	head = head[unsafe.Sizeof(cpuState.X64.Fpregs.MxcrMask):]
	for i := 0; i < len(cpuState.X64.Fpregs.St); i++ {
		for j := 0; j < len(cpuState.X64.Fpregs.St[i].Significand); j++ {
			binary.LittleEndian.PutUint16(head, uint16(cpuState.X64.Fpregs.St[i].Significand[j]))
			head = head[unsafe.Sizeof(uint16(cpuState.X64.Fpregs.St[i].Significand[j])):]
		}
		binary.LittleEndian.PutUint16(head, uint16(cpuState.X64.Fpregs.St[i].Exponent))
		head = head[unsafe.Sizeof(uint16(cpuState.X64.Fpregs.St[i].Exponent)):]
		for j := 0; j < len(cpuState.X64.Fpregs.St[i].Reserved); j++ {
			binary.LittleEndian.PutUint16(head, uint16(cpuState.X64.Fpregs.St[i].Reserved[j]))
			head = head[unsafe.Sizeof(uint16(cpuState.X64.Fpregs.St[i].Reserved[j])):]
		}
	}
	for i := 0; i < len(cpuState.X64.Fpregs.Xmm); i++ {
		for j := 0; j < len(cpuState.X64.Fpregs.Xmm[i].Element); j++ {
			binary.LittleEndian.PutUint32(head, cpuState.X64.Fpregs.Xmm[i].Element[j])
			head = head[unsafe.Sizeof(cpuState.X64.Fpregs.Xmm[i].Element[j]):]
		}
	}
	for i := 0; i < len(cpuState.X64.Fpregs.Reserved); i++ {
		binary.LittleEndian.PutUint32(head, cpuState.X64.Fpregs.Reserved[i])
		head = head[unsafe.Sizeof(cpuState.X64.Fpregs.Reserved[i]):]
	}
	return byteData
}

func (h *CPUCodec) Decode(reader io.Reader) *cpu.CPU {
	st := make([]*x64.X64FPXReg, 8)
	for i := 0; i < len(st); i++ {
		st[i] = &x64.X64FPXReg{
			Significand: make([]uint32, 4),
			Reserved:    make([]uint32, 3),
		}
	}
	xmm := make([]*x64.X64XMMReg, 16)
	for i := 0; i < len(xmm); i++ {
		xmm[i] = &x64.X64XMMReg{
			Element: make([]uint32, 4),
		}
	}
	cpu := cpu.CPU{
		X64: &x64.CPUState{
			Gregs: make([]uint64, cpu.CR2+1),
			Fpregs: &x64.X64FPRegs{
				St:       st,
				Xmm:      xmm,
				Reserved: make([]uint32, 24),
			},
		},
	}
	binary.Read(reader, binary.LittleEndian, &cpu.X64.Gregs)
	var shortReg uint16
	binary.Read(reader, binary.LittleEndian, &shortReg)
	cpu.X64.Fpregs.Cwd = uint32(shortReg)
	binary.Read(reader, binary.LittleEndian, &shortReg)
	cpu.X64.Fpregs.Swd = uint32(shortReg)
	binary.Read(reader, binary.LittleEndian, &shortReg)
	cpu.X64.Fpregs.Ftw = uint32(shortReg)
	binary.Read(reader, binary.LittleEndian, &shortReg)
	cpu.X64.Fpregs.Fop = uint32(shortReg)
	binary.Read(reader, binary.LittleEndian, &cpu.X64.Fpregs.Rip)
	binary.Read(reader, binary.LittleEndian, &cpu.X64.Fpregs.Rdp)
	binary.Read(reader, binary.LittleEndian, &cpu.X64.Fpregs.Mxcsr)
	binary.Read(reader, binary.LittleEndian, &cpu.X64.Fpregs.MxcrMask)
	for st := 0; st < len(cpu.X64.Fpregs.St); st++ {
		for sig := 0; sig < len(cpu.X64.Fpregs.St[st].Significand); sig++ {
			binary.Read(reader, binary.LittleEndian, &shortReg)
			cpu.X64.Fpregs.St[st].Significand[sig] = uint32(shortReg)
		}
		binary.Read(reader, binary.LittleEndian, &shortReg)
		cpu.X64.Fpregs.St[st].Exponent = uint32(shortReg)
		for res := 0; res < len(cpu.X64.Fpregs.St[st].Reserved); res++ {
			binary.Read(reader, binary.LittleEndian, &shortReg)
			cpu.X64.Fpregs.St[st].Reserved[res] = uint32(shortReg)
		}
	}
	for xmm := 0; xmm < len(cpu.X64.Fpregs.Xmm); xmm++ {
		for element := 0; element < len(cpu.X64.Fpregs.Xmm[xmm].Element); element++ {
			binary.Read(reader, binary.LittleEndian, &cpu.X64.Fpregs.Xmm[xmm].Element[element])
		}
	}
	for reserved := 0; reserved < len(cpu.X64.Fpregs.Reserved); reserved++ {
		binary.Read(reader, binary.LittleEndian, &cpu.X64.Fpregs.Reserved[reserved])
	}
	return &cpu
}
