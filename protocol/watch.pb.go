// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protocol/watch.proto

package protocol

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type WatchRequest struct {
	Names                []string `protobuf:"bytes,1,rep,name=names,proto3" json:"names,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WatchRequest) Reset()         { *m = WatchRequest{} }
func (m *WatchRequest) String() string { return proto.CompactTextString(m) }
func (*WatchRequest) ProtoMessage()    {}
func (*WatchRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_e860bb3920cca293, []int{0}
}

func (m *WatchRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WatchRequest.Unmarshal(m, b)
}
func (m *WatchRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WatchRequest.Marshal(b, m, deterministic)
}
func (m *WatchRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WatchRequest.Merge(m, src)
}
func (m *WatchRequest) XXX_Size() int {
	return xxx_messageInfo_WatchRequest.Size(m)
}
func (m *WatchRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_WatchRequest.DiscardUnknown(m)
}

var xxx_messageInfo_WatchRequest proto.InternalMessageInfo

func (m *WatchRequest) GetNames() []string {
	if m != nil {
		return m.Names
	}
	return nil
}

type WatchFile struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	LastStatus           string   `protobuf:"bytes,2,opt,name=lastStatus,proto3" json:"lastStatus,omitempty"`
	LastSyncAt           string   `protobuf:"bytes,3,opt,name=lastSyncAt,proto3" json:"lastSyncAt,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WatchFile) Reset()         { *m = WatchFile{} }
func (m *WatchFile) String() string { return proto.CompactTextString(m) }
func (*WatchFile) ProtoMessage()    {}
func (*WatchFile) Descriptor() ([]byte, []int) {
	return fileDescriptor_e860bb3920cca293, []int{1}
}

func (m *WatchFile) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WatchFile.Unmarshal(m, b)
}
func (m *WatchFile) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WatchFile.Marshal(b, m, deterministic)
}
func (m *WatchFile) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WatchFile.Merge(m, src)
}
func (m *WatchFile) XXX_Size() int {
	return xxx_messageInfo_WatchFile.Size(m)
}
func (m *WatchFile) XXX_DiscardUnknown() {
	xxx_messageInfo_WatchFile.DiscardUnknown(m)
}

var xxx_messageInfo_WatchFile proto.InternalMessageInfo

func (m *WatchFile) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *WatchFile) GetLastStatus() string {
	if m != nil {
		return m.LastStatus
	}
	return ""
}

func (m *WatchFile) GetLastSyncAt() string {
	if m != nil {
		return m.LastSyncAt
	}
	return ""
}

type WatchResponse struct {
	Files                []*WatchFile `protobuf:"bytes,1,rep,name=files,proto3" json:"files,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *WatchResponse) Reset()         { *m = WatchResponse{} }
func (m *WatchResponse) String() string { return proto.CompactTextString(m) }
func (*WatchResponse) ProtoMessage()    {}
func (*WatchResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_e860bb3920cca293, []int{2}
}

func (m *WatchResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WatchResponse.Unmarshal(m, b)
}
func (m *WatchResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WatchResponse.Marshal(b, m, deterministic)
}
func (m *WatchResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WatchResponse.Merge(m, src)
}
func (m *WatchResponse) XXX_Size() int {
	return xxx_messageInfo_WatchResponse.Size(m)
}
func (m *WatchResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_WatchResponse.DiscardUnknown(m)
}

var xxx_messageInfo_WatchResponse proto.InternalMessageInfo

func (m *WatchResponse) GetFiles() []*WatchFile {
	if m != nil {
		return m.Files
	}
	return nil
}

func init() {
	proto.RegisterType((*WatchRequest)(nil), "protocol.WatchRequest")
	proto.RegisterType((*WatchFile)(nil), "protocol.WatchFile")
	proto.RegisterType((*WatchResponse)(nil), "protocol.WatchResponse")
}

func init() { proto.RegisterFile("protocol/watch.proto", fileDescriptor_e860bb3920cca293) }

var fileDescriptor_e860bb3920cca293 = []byte{
	// 174 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x29, 0x28, 0xca, 0x2f,
	0xc9, 0x4f, 0xce, 0xcf, 0xd1, 0x2f, 0x4f, 0x2c, 0x49, 0xce, 0xd0, 0x03, 0x73, 0x85, 0x38, 0x60,
	0xa2, 0x4a, 0x2a, 0x5c, 0x3c, 0xe1, 0x20, 0x89, 0xa0, 0xd4, 0xc2, 0xd2, 0xd4, 0xe2, 0x12, 0x21,
	0x11, 0x2e, 0xd6, 0xbc, 0xc4, 0xdc, 0xd4, 0x62, 0x09, 0x46, 0x05, 0x66, 0x0d, 0xce, 0x20, 0x08,
	0x47, 0x29, 0x9e, 0x8b, 0x13, 0xac, 0xca, 0x2d, 0x33, 0x27, 0x55, 0x48, 0x88, 0x8b, 0x05, 0x24,
	0x2a, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1, 0x19, 0x04, 0x66, 0x0b, 0xc9, 0x71, 0x71, 0xe5, 0x24, 0x16,
	0x97, 0x04, 0x97, 0x24, 0x96, 0x94, 0x16, 0x4b, 0x30, 0x81, 0x65, 0x90, 0x44, 0xe0, 0xf2, 0x95,
	0x79, 0xc9, 0x8e, 0x25, 0x12, 0xcc, 0x48, 0xf2, 0x60, 0x11, 0x25, 0x2b, 0x2e, 0x5e, 0xa8, 0x33,
	0x8a, 0x0b, 0xf2, 0xf3, 0x8a, 0x53, 0x85, 0x34, 0xb9, 0x58, 0xd3, 0x32, 0x73, 0xa0, 0xee, 0xe0,
	0x36, 0x12, 0xd6, 0x83, 0xb9, 0x58, 0x0f, 0xee, 0x90, 0x20, 0x88, 0x8a, 0x24, 0x36, 0xb0, 0x94,
	0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0xaa, 0xab, 0x64, 0xc1, 0xeb, 0x00, 0x00, 0x00,
}