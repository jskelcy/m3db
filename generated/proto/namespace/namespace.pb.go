// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Code generated by protoc-gen-go.
// source: namespace.proto
// DO NOT EDIT!

/*
Package namespace is a generated protocol buffer package.

It is generated from these files:
	namespace.proto

It has these top-level messages:
	RetentionOptions
	NamespaceOptions
	Registry
*/
package namespace

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type RetentionOptions struct {
	RetentionPeriodNanos                     int64 `protobuf:"varint,1,opt,name=retentionPeriodNanos" json:"retentionPeriodNanos,omitempty"`
	BlockSizeNanos                           int64 `protobuf:"varint,2,opt,name=blockSizeNanos" json:"blockSizeNanos,omitempty"`
	BufferFutureNanos                        int64 `protobuf:"varint,3,opt,name=bufferFutureNanos" json:"bufferFutureNanos,omitempty"`
	BufferPastNanos                          int64 `protobuf:"varint,4,opt,name=bufferPastNanos" json:"bufferPastNanos,omitempty"`
	BlockDataExpiry                          bool  `protobuf:"varint,5,opt,name=blockDataExpiry" json:"blockDataExpiry,omitempty"`
	BlockDataExpiryAfterNotAccessPeriodNanos int64 `protobuf:"varint,6,opt,name=blockDataExpiryAfterNotAccessPeriodNanos" json:"blockDataExpiryAfterNotAccessPeriodNanos,omitempty"`
}

func (m *RetentionOptions) Reset()                    { *m = RetentionOptions{} }
func (m *RetentionOptions) String() string            { return proto.CompactTextString(m) }
func (*RetentionOptions) ProtoMessage()               {}
func (*RetentionOptions) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type NamespaceOptions struct {
	NeedsBootstrap      bool              `protobuf:"varint,1,opt,name=needsBootstrap" json:"needsBootstrap,omitempty"`
	NeedsFlush          bool              `protobuf:"varint,2,opt,name=needsFlush" json:"needsFlush,omitempty"`
	WritesToCommitLog   bool              `protobuf:"varint,3,opt,name=writesToCommitLog" json:"writesToCommitLog,omitempty"`
	NeedsFilesetCleanup bool              `protobuf:"varint,4,opt,name=needsFilesetCleanup" json:"needsFilesetCleanup,omitempty"`
	NeedsRepair         bool              `protobuf:"varint,5,opt,name=needsRepair" json:"needsRepair,omitempty"`
	RetentionOptions    *RetentionOptions `protobuf:"bytes,6,opt,name=retentionOptions" json:"retentionOptions,omitempty"`
}

func (m *NamespaceOptions) Reset()                    { *m = NamespaceOptions{} }
func (m *NamespaceOptions) String() string            { return proto.CompactTextString(m) }
func (*NamespaceOptions) ProtoMessage()               {}
func (*NamespaceOptions) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *NamespaceOptions) GetRetentionOptions() *RetentionOptions {
	if m != nil {
		return m.RetentionOptions
	}
	return nil
}

type Registry struct {
	Namespaces map[string]*NamespaceOptions `protobuf:"bytes,1,rep,name=namespaces" json:"namespaces,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Registry) Reset()                    { *m = Registry{} }
func (m *Registry) String() string            { return proto.CompactTextString(m) }
func (*Registry) ProtoMessage()               {}
func (*Registry) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Registry) GetNamespaces() map[string]*NamespaceOptions {
	if m != nil {
		return m.Namespaces
	}
	return nil
}

func init() {
	proto.RegisterType((*RetentionOptions)(nil), "namespace.RetentionOptions")
	proto.RegisterType((*NamespaceOptions)(nil), "namespace.NamespaceOptions")
	proto.RegisterType((*Registry)(nil), "namespace.Registry")
}

func init() { proto.RegisterFile("namespace.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 403 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x93, 0x41, 0x6f, 0xda, 0x30,
	0x14, 0xc7, 0x95, 0x64, 0xa0, 0xf0, 0x90, 0x20, 0xf3, 0x76, 0x40, 0x9b, 0x34, 0x21, 0x26, 0xa1,
	0x1c, 0x26, 0xb4, 0xb1, 0xcb, 0xb4, 0x1b, 0x63, 0xb0, 0xcb, 0xc4, 0x90, 0xd7, 0x13, 0x37, 0x13,
	0x1e, 0xd4, 0x22, 0xc4, 0x91, 0xed, 0xb4, 0x4d, 0xbf, 0x4f, 0x4f, 0xfd, 0x2a, 0xfd, 0x50, 0x55,
	0x1c, 0xa0, 0xc1, 0x70, 0xe8, 0x05, 0xa1, 0xdf, 0xff, 0xff, 0xe2, 0xe4, 0x97, 0x17, 0x68, 0x27,
	0x6c, 0x87, 0x2a, 0x65, 0x11, 0x0e, 0x52, 0x29, 0xb4, 0x20, 0x8d, 0x23, 0xe8, 0x3d, 0xb9, 0x10,
	0x50, 0xd4, 0x98, 0x68, 0x2e, 0x92, 0x7f, 0x69, 0xf1, 0xab, 0xc8, 0x10, 0xde, 0xcb, 0x03, 0x9b,
	0xa3, 0xe4, 0x62, 0x35, 0x63, 0x89, 0x50, 0x1d, 0xa7, 0xeb, 0x84, 0x1e, 0xbd, 0x98, 0x91, 0x3e,
	0xb4, 0x96, 0xb1, 0x88, 0xb6, 0xff, 0xf9, 0x3d, 0x96, 0x6d, 0xd7, 0xb4, 0x2d, 0x4a, 0xbe, 0xc0,
	0xdb, 0x65, 0xb6, 0x5e, 0xa3, 0x9c, 0x66, 0x3a, 0x93, 0xfb, 0xaa, 0x67, 0xaa, 0xe7, 0x01, 0x09,
	0xa1, 0x5d, 0xc2, 0x39, 0x53, 0xba, 0xec, 0xbe, 0x31, 0x5d, 0x1b, 0x9b, 0x66, 0x71, 0xd2, 0x6f,
	0xa6, 0xd9, 0xe4, 0x2e, 0xe5, 0x32, 0xef, 0xd4, 0xba, 0x4e, 0xe8, 0x53, 0x1b, 0x93, 0x05, 0x84,
	0x16, 0x1a, 0xad, 0x35, 0xca, 0x99, 0xd0, 0xa3, 0x28, 0x42, 0xa5, 0xaa, 0x4f, 0x5c, 0x37, 0x87,
	0xbd, 0xba, 0xdf, 0x7b, 0x70, 0x21, 0x98, 0x1d, 0xe4, 0x1e, 0x74, 0xf6, 0xa1, 0x95, 0x20, 0xae,
	0xd4, 0x2f, 0x21, 0xb4, 0xd2, 0x92, 0xa5, 0x46, 0xa4, 0x4f, 0x2d, 0x4a, 0x3e, 0x01, 0x18, 0x32,
	0x8d, 0x33, 0x75, 0x6d, 0xf4, 0xf9, 0xb4, 0x42, 0x0a, 0x75, 0xb7, 0x92, 0x6b, 0x54, 0x57, 0x62,
	0x2c, 0x76, 0x3b, 0xae, 0xff, 0x8a, 0x8d, 0x51, 0xe7, 0xd3, 0xf3, 0x80, 0x7c, 0x85, 0x77, 0xe5,
	0x2c, 0x8f, 0x51, 0xa1, 0x1e, 0xc7, 0xc8, 0x92, 0x2c, 0x35, 0xfa, 0x7c, 0x7a, 0x29, 0x22, 0x5d,
	0x68, 0x1a, 0x4c, 0x31, 0x65, 0x5c, 0xee, 0xf5, 0x55, 0x11, 0xf9, 0x03, 0x81, 0xb4, 0x96, 0xc5,
	0x28, 0x6a, 0x0e, 0x3f, 0x0e, 0x5e, 0x96, 0xcc, 0xde, 0x27, 0x7a, 0x36, 0xd4, 0x7b, 0x74, 0xc0,
	0xa7, 0xb8, 0xe1, 0x4a, 0xcb, 0x9c, 0x8c, 0x01, 0x8e, 0xc3, 0xc5, 0x92, 0x79, 0x61, 0x73, 0xf8,
	0xf9, 0xe4, 0x7a, 0x65, 0x71, 0x70, 0x34, 0xab, 0x26, 0x89, 0x96, 0x39, 0xad, 0x8c, 0x7d, 0x58,
	0x40, 0xdb, 0x8a, 0x49, 0x00, 0xde, 0x16, 0x73, 0x23, 0xbb, 0x41, 0x8b, 0xbf, 0xe4, 0x1b, 0xd4,
	0x6e, 0x58, 0x9c, 0xa1, 0x91, 0x7b, 0x7a, 0xd3, 0xf6, 0x5b, 0xa3, 0x65, 0xf3, 0xa7, 0xfb, 0xc3,
	0x59, 0xd6, 0xcd, 0x67, 0xf3, 0xfd, 0x39, 0x00, 0x00, 0xff, 0xff, 0xac, 0x0c, 0x58, 0x12, 0x49,
	0x03, 0x00, 0x00,
}
