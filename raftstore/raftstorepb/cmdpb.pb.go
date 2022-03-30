// Code generated by protoc-gen-go. DO NOT EDIT.
// source: cmdpb.proto

package raftstorepb

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

type CmdType int32

const (
	CmdType_Get    CmdType = 0
	CmdType_Put    CmdType = 1
	CmdType_Delete CmdType = 2
)

var CmdType_name = map[int32]string{
	0: "Get",
	1: "Put",
	2: "Delete",
}
var CmdType_value = map[string]int32{
	"Get":    0,
	"Put":    1,
	"Delete": 2,
}

func (x CmdType) String() string {
	return proto.EnumName(CmdType_name, int32(x))
}
func (CmdType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{0}
}

type AdminCmdType int32

const (
	AdminCmdType_CompactLog AdminCmdType = 0
)

var AdminCmdType_name = map[int32]string{
	0: "CompactLog",
}
var AdminCmdType_value = map[string]int32{
	"CompactLog": 0,
}

func (x AdminCmdType) String() string {
	return proto.EnumName(AdminCmdType_name, int32(x))
}
func (AdminCmdType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{1}
}

type GetRequest struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetRequest) Reset()         { *m = GetRequest{} }
func (m *GetRequest) String() string { return proto.CompactTextString(m) }
func (*GetRequest) ProtoMessage()    {}
func (*GetRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{0}
}
func (m *GetRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetRequest.Unmarshal(m, b)
}
func (m *GetRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetRequest.Marshal(b, m, deterministic)
}
func (dst *GetRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetRequest.Merge(dst, src)
}
func (m *GetRequest) XXX_Size() int {
	return xxx_messageInfo_GetRequest.Size(m)
}
func (m *GetRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetRequest proto.InternalMessageInfo

func (m *GetRequest) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

type GetResponse struct {
	Value                []byte   `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetResponse) Reset()         { *m = GetResponse{} }
func (m *GetResponse) String() string { return proto.CompactTextString(m) }
func (*GetResponse) ProtoMessage()    {}
func (*GetResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{1}
}
func (m *GetResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetResponse.Unmarshal(m, b)
}
func (m *GetResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetResponse.Marshal(b, m, deterministic)
}
func (dst *GetResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetResponse.Merge(dst, src)
}
func (m *GetResponse) XXX_Size() int {
	return xxx_messageInfo_GetResponse.Size(m)
}
func (m *GetResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetResponse proto.InternalMessageInfo

func (m *GetResponse) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type PutRequest struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value                []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PutRequest) Reset()         { *m = PutRequest{} }
func (m *PutRequest) String() string { return proto.CompactTextString(m) }
func (*PutRequest) ProtoMessage()    {}
func (*PutRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{2}
}
func (m *PutRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PutRequest.Unmarshal(m, b)
}
func (m *PutRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PutRequest.Marshal(b, m, deterministic)
}
func (dst *PutRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PutRequest.Merge(dst, src)
}
func (m *PutRequest) XXX_Size() int {
	return xxx_messageInfo_PutRequest.Size(m)
}
func (m *PutRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PutRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PutRequest proto.InternalMessageInfo

func (m *PutRequest) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *PutRequest) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type PutResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PutResponse) Reset()         { *m = PutResponse{} }
func (m *PutResponse) String() string { return proto.CompactTextString(m) }
func (*PutResponse) ProtoMessage()    {}
func (*PutResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{3}
}
func (m *PutResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PutResponse.Unmarshal(m, b)
}
func (m *PutResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PutResponse.Marshal(b, m, deterministic)
}
func (dst *PutResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PutResponse.Merge(dst, src)
}
func (m *PutResponse) XXX_Size() int {
	return xxx_messageInfo_PutResponse.Size(m)
}
func (m *PutResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PutResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PutResponse proto.InternalMessageInfo

type DeleteRequest struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteRequest) Reset()         { *m = DeleteRequest{} }
func (m *DeleteRequest) String() string { return proto.CompactTextString(m) }
func (*DeleteRequest) ProtoMessage()    {}
func (*DeleteRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{4}
}
func (m *DeleteRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteRequest.Unmarshal(m, b)
}
func (m *DeleteRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteRequest.Marshal(b, m, deterministic)
}
func (dst *DeleteRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteRequest.Merge(dst, src)
}
func (m *DeleteRequest) XXX_Size() int {
	return xxx_messageInfo_DeleteRequest.Size(m)
}
func (m *DeleteRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteRequest proto.InternalMessageInfo

func (m *DeleteRequest) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

type DeleteResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteResponse) Reset()         { *m = DeleteResponse{} }
func (m *DeleteResponse) String() string { return proto.CompactTextString(m) }
func (*DeleteResponse) ProtoMessage()    {}
func (*DeleteResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{5}
}
func (m *DeleteResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteResponse.Unmarshal(m, b)
}
func (m *DeleteResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteResponse.Marshal(b, m, deterministic)
}
func (dst *DeleteResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteResponse.Merge(dst, src)
}
func (m *DeleteResponse) XXX_Size() int {
	return xxx_messageInfo_DeleteResponse.Size(m)
}
func (m *DeleteResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteResponse proto.InternalMessageInfo

type CompactLogRequest struct {
	CompactIndex         uint64   `protobuf:"varint,1,opt,name=compact_index,json=compactIndex" json:"compact_index,omitempty"`
	CompactTerm          uint64   `protobuf:"varint,2,opt,name=compact_term,json=compactTerm" json:"compact_term,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CompactLogRequest) Reset()         { *m = CompactLogRequest{} }
func (m *CompactLogRequest) String() string { return proto.CompactTextString(m) }
func (*CompactLogRequest) ProtoMessage()    {}
func (*CompactLogRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{6}
}
func (m *CompactLogRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CompactLogRequest.Unmarshal(m, b)
}
func (m *CompactLogRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CompactLogRequest.Marshal(b, m, deterministic)
}
func (dst *CompactLogRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CompactLogRequest.Merge(dst, src)
}
func (m *CompactLogRequest) XXX_Size() int {
	return xxx_messageInfo_CompactLogRequest.Size(m)
}
func (m *CompactLogRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CompactLogRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CompactLogRequest proto.InternalMessageInfo

func (m *CompactLogRequest) GetCompactIndex() uint64 {
	if m != nil {
		return m.CompactIndex
	}
	return 0
}

func (m *CompactLogRequest) GetCompactTerm() uint64 {
	if m != nil {
		return m.CompactTerm
	}
	return 0
}

type CompactLogResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CompactLogResponse) Reset()         { *m = CompactLogResponse{} }
func (m *CompactLogResponse) String() string { return proto.CompactTextString(m) }
func (*CompactLogResponse) ProtoMessage()    {}
func (*CompactLogResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{7}
}
func (m *CompactLogResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CompactLogResponse.Unmarshal(m, b)
}
func (m *CompactLogResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CompactLogResponse.Marshal(b, m, deterministic)
}
func (dst *CompactLogResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CompactLogResponse.Merge(dst, src)
}
func (m *CompactLogResponse) XXX_Size() int {
	return xxx_messageInfo_CompactLogResponse.Size(m)
}
func (m *CompactLogResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_CompactLogResponse.DiscardUnknown(m)
}

var xxx_messageInfo_CompactLogResponse proto.InternalMessageInfo

type RaftRequestHeader struct {
	Term                 uint64   `protobuf:"varint,1,opt,name=term" json:"term,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RaftRequestHeader) Reset()         { *m = RaftRequestHeader{} }
func (m *RaftRequestHeader) String() string { return proto.CompactTextString(m) }
func (*RaftRequestHeader) ProtoMessage()    {}
func (*RaftRequestHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{8}
}
func (m *RaftRequestHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RaftRequestHeader.Unmarshal(m, b)
}
func (m *RaftRequestHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RaftRequestHeader.Marshal(b, m, deterministic)
}
func (dst *RaftRequestHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RaftRequestHeader.Merge(dst, src)
}
func (m *RaftRequestHeader) XXX_Size() int {
	return xxx_messageInfo_RaftRequestHeader.Size(m)
}
func (m *RaftRequestHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_RaftRequestHeader.DiscardUnknown(m)
}

var xxx_messageInfo_RaftRequestHeader proto.InternalMessageInfo

func (m *RaftRequestHeader) GetTerm() uint64 {
	if m != nil {
		return m.Term
	}
	return 0
}

type RaftResponseHeader struct {
	CurrentTerm          uint64   `protobuf:"varint,1,opt,name=current_term,json=currentTerm" json:"current_term,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RaftResponseHeader) Reset()         { *m = RaftResponseHeader{} }
func (m *RaftResponseHeader) String() string { return proto.CompactTextString(m) }
func (*RaftResponseHeader) ProtoMessage()    {}
func (*RaftResponseHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{9}
}
func (m *RaftResponseHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RaftResponseHeader.Unmarshal(m, b)
}
func (m *RaftResponseHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RaftResponseHeader.Marshal(b, m, deterministic)
}
func (dst *RaftResponseHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RaftResponseHeader.Merge(dst, src)
}
func (m *RaftResponseHeader) XXX_Size() int {
	return xxx_messageInfo_RaftResponseHeader.Size(m)
}
func (m *RaftResponseHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_RaftResponseHeader.DiscardUnknown(m)
}

var xxx_messageInfo_RaftResponseHeader proto.InternalMessageInfo

func (m *RaftResponseHeader) GetCurrentTerm() uint64 {
	if m != nil {
		return m.CurrentTerm
	}
	return 0
}

type Request struct {
	CmdType              CmdType        `protobuf:"varint,1,opt,name=cmd_type,json=cmdType,enum=raftstorepb.CmdType" json:"cmd_type,omitempty"`
	Get                  *GetRequest    `protobuf:"bytes,2,opt,name=get" json:"get,omitempty"`
	Put                  *PutRequest    `protobuf:"bytes,3,opt,name=put" json:"put,omitempty"`
	Delete               *DeleteRequest `protobuf:"bytes,4,opt,name=delete" json:"delete,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{10}
}
func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (dst *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(dst, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetCmdType() CmdType {
	if m != nil {
		return m.CmdType
	}
	return CmdType_Get
}

func (m *Request) GetGet() *GetRequest {
	if m != nil {
		return m.Get
	}
	return nil
}

func (m *Request) GetPut() *PutRequest {
	if m != nil {
		return m.Put
	}
	return nil
}

func (m *Request) GetDelete() *DeleteRequest {
	if m != nil {
		return m.Delete
	}
	return nil
}

type Response struct {
	CmdType              CmdType         `protobuf:"varint,1,opt,name=cmd_type,json=cmdType,enum=raftstorepb.CmdType" json:"cmd_type,omitempty"`
	Get                  *GetResponse    `protobuf:"bytes,2,opt,name=get" json:"get,omitempty"`
	Put                  *PutResponse    `protobuf:"bytes,3,opt,name=put" json:"put,omitempty"`
	Delete               *DeleteResponse `protobuf:"bytes,4,opt,name=delete" json:"delete,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{11}
}
func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Response.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Response.Marshal(b, m, deterministic)
}
func (dst *Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Response.Merge(dst, src)
}
func (m *Response) XXX_Size() int {
	return xxx_messageInfo_Response.Size(m)
}
func (m *Response) XXX_DiscardUnknown() {
	xxx_messageInfo_Response.DiscardUnknown(m)
}

var xxx_messageInfo_Response proto.InternalMessageInfo

func (m *Response) GetCmdType() CmdType {
	if m != nil {
		return m.CmdType
	}
	return CmdType_Get
}

func (m *Response) GetGet() *GetResponse {
	if m != nil {
		return m.Get
	}
	return nil
}

func (m *Response) GetPut() *PutResponse {
	if m != nil {
		return m.Put
	}
	return nil
}

func (m *Response) GetDelete() *DeleteResponse {
	if m != nil {
		return m.Delete
	}
	return nil
}

type AdminRequest struct {
	CmdType              AdminCmdType       `protobuf:"varint,1,opt,name=cmd_type,json=cmdType,enum=raftstorepb.AdminCmdType" json:"cmd_type,omitempty"`
	CompactLog           *CompactLogRequest `protobuf:"bytes,2,opt,name=compact_log,json=compactLog" json:"compact_log,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *AdminRequest) Reset()         { *m = AdminRequest{} }
func (m *AdminRequest) String() string { return proto.CompactTextString(m) }
func (*AdminRequest) ProtoMessage()    {}
func (*AdminRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{12}
}
func (m *AdminRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AdminRequest.Unmarshal(m, b)
}
func (m *AdminRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AdminRequest.Marshal(b, m, deterministic)
}
func (dst *AdminRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AdminRequest.Merge(dst, src)
}
func (m *AdminRequest) XXX_Size() int {
	return xxx_messageInfo_AdminRequest.Size(m)
}
func (m *AdminRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_AdminRequest.DiscardUnknown(m)
}

var xxx_messageInfo_AdminRequest proto.InternalMessageInfo

func (m *AdminRequest) GetCmdType() AdminCmdType {
	if m != nil {
		return m.CmdType
	}
	return AdminCmdType_CompactLog
}

func (m *AdminRequest) GetCompactLog() *CompactLogRequest {
	if m != nil {
		return m.CompactLog
	}
	return nil
}

type AdminResponse struct {
	CmdType              AdminCmdType        `protobuf:"varint,1,opt,name=cmd_type,json=cmdType,enum=raftstorepb.AdminCmdType" json:"cmd_type,omitempty"`
	CompactLog           *CompactLogResponse `protobuf:"bytes,2,opt,name=compact_log,json=compactLog" json:"compact_log,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *AdminResponse) Reset()         { *m = AdminResponse{} }
func (m *AdminResponse) String() string { return proto.CompactTextString(m) }
func (*AdminResponse) ProtoMessage()    {}
func (*AdminResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{13}
}
func (m *AdminResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AdminResponse.Unmarshal(m, b)
}
func (m *AdminResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AdminResponse.Marshal(b, m, deterministic)
}
func (dst *AdminResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AdminResponse.Merge(dst, src)
}
func (m *AdminResponse) XXX_Size() int {
	return xxx_messageInfo_AdminResponse.Size(m)
}
func (m *AdminResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_AdminResponse.DiscardUnknown(m)
}

var xxx_messageInfo_AdminResponse proto.InternalMessageInfo

func (m *AdminResponse) GetCmdType() AdminCmdType {
	if m != nil {
		return m.CmdType
	}
	return AdminCmdType_CompactLog
}

func (m *AdminResponse) GetCompactLog() *CompactLogResponse {
	if m != nil {
		return m.CompactLog
	}
	return nil
}

type RaftCmdRequest struct {
	Header *RaftRequestHeader `protobuf:"bytes,1,opt,name=header" json:"header,omitempty"`
	// We can't enclose normal requests and administrator request
	// at same time.
	Request              *Request      `protobuf:"bytes,2,opt,name=request" json:"request,omitempty"`
	AdminRequest         *AdminRequest `protobuf:"bytes,3,opt,name=admin_request,json=adminRequest" json:"admin_request,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *RaftCmdRequest) Reset()         { *m = RaftCmdRequest{} }
func (m *RaftCmdRequest) String() string { return proto.CompactTextString(m) }
func (*RaftCmdRequest) ProtoMessage()    {}
func (*RaftCmdRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{14}
}
func (m *RaftCmdRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RaftCmdRequest.Unmarshal(m, b)
}
func (m *RaftCmdRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RaftCmdRequest.Marshal(b, m, deterministic)
}
func (dst *RaftCmdRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RaftCmdRequest.Merge(dst, src)
}
func (m *RaftCmdRequest) XXX_Size() int {
	return xxx_messageInfo_RaftCmdRequest.Size(m)
}
func (m *RaftCmdRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RaftCmdRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RaftCmdRequest proto.InternalMessageInfo

func (m *RaftCmdRequest) GetHeader() *RaftRequestHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *RaftCmdRequest) GetRequest() *Request {
	if m != nil {
		return m.Request
	}
	return nil
}

func (m *RaftCmdRequest) GetAdminRequest() *AdminRequest {
	if m != nil {
		return m.AdminRequest
	}
	return nil
}

type RaftCmdResponse struct {
	Header               *RaftResponseHeader `protobuf:"bytes,1,opt,name=header" json:"header,omitempty"`
	Response             *Response           `protobuf:"bytes,2,opt,name=response" json:"response,omitempty"`
	AdminResponse        *AdminResponse      `protobuf:"bytes,3,opt,name=admin_response,json=adminResponse" json:"admin_response,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *RaftCmdResponse) Reset()         { *m = RaftCmdResponse{} }
func (m *RaftCmdResponse) String() string { return proto.CompactTextString(m) }
func (*RaftCmdResponse) ProtoMessage()    {}
func (*RaftCmdResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_cmdpb_5faf1474e63d99ed, []int{15}
}
func (m *RaftCmdResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RaftCmdResponse.Unmarshal(m, b)
}
func (m *RaftCmdResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RaftCmdResponse.Marshal(b, m, deterministic)
}
func (dst *RaftCmdResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RaftCmdResponse.Merge(dst, src)
}
func (m *RaftCmdResponse) XXX_Size() int {
	return xxx_messageInfo_RaftCmdResponse.Size(m)
}
func (m *RaftCmdResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_RaftCmdResponse.DiscardUnknown(m)
}

var xxx_messageInfo_RaftCmdResponse proto.InternalMessageInfo

func (m *RaftCmdResponse) GetHeader() *RaftResponseHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *RaftCmdResponse) GetResponse() *Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (m *RaftCmdResponse) GetAdminResponse() *AdminResponse {
	if m != nil {
		return m.AdminResponse
	}
	return nil
}

func init() {
	proto.RegisterType((*GetRequest)(nil), "raftstorepb.GetRequest")
	proto.RegisterType((*GetResponse)(nil), "raftstorepb.GetResponse")
	proto.RegisterType((*PutRequest)(nil), "raftstorepb.PutRequest")
	proto.RegisterType((*PutResponse)(nil), "raftstorepb.PutResponse")
	proto.RegisterType((*DeleteRequest)(nil), "raftstorepb.DeleteRequest")
	proto.RegisterType((*DeleteResponse)(nil), "raftstorepb.DeleteResponse")
	proto.RegisterType((*CompactLogRequest)(nil), "raftstorepb.CompactLogRequest")
	proto.RegisterType((*CompactLogResponse)(nil), "raftstorepb.CompactLogResponse")
	proto.RegisterType((*RaftRequestHeader)(nil), "raftstorepb.RaftRequestHeader")
	proto.RegisterType((*RaftResponseHeader)(nil), "raftstorepb.RaftResponseHeader")
	proto.RegisterType((*Request)(nil), "raftstorepb.Request")
	proto.RegisterType((*Response)(nil), "raftstorepb.Response")
	proto.RegisterType((*AdminRequest)(nil), "raftstorepb.AdminRequest")
	proto.RegisterType((*AdminResponse)(nil), "raftstorepb.AdminResponse")
	proto.RegisterType((*RaftCmdRequest)(nil), "raftstorepb.RaftCmdRequest")
	proto.RegisterType((*RaftCmdResponse)(nil), "raftstorepb.RaftCmdResponse")
	proto.RegisterEnum("raftstorepb.CmdType", CmdType_name, CmdType_value)
	proto.RegisterEnum("raftstorepb.AdminCmdType", AdminCmdType_name, AdminCmdType_value)
}

func init() { proto.RegisterFile("cmdpb.proto", fileDescriptor_cmdpb_5faf1474e63d99ed) }

var fileDescriptor_cmdpb_5faf1474e63d99ed = []byte{
	// 589 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x55, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0xed, 0x36, 0x21, 0x89, 0xc6, 0x71, 0x70, 0x57, 0x41, 0x04, 0x90, 0x52, 0xea, 0x1e, 0x0a,
	0x39, 0x04, 0x91, 0x56, 0xf4, 0x06, 0x54, 0x41, 0x2a, 0x48, 0x1c, 0xaa, 0x55, 0x6f, 0x1c, 0x22,
	0xc7, 0xde, 0x86, 0x8a, 0x38, 0x36, 0x9b, 0x35, 0x22, 0x1f, 0x80, 0xf8, 0x25, 0x8e, 0x1c, 0x38,
	0xf0, 0x59, 0x68, 0x77, 0x67, 0x13, 0x3b, 0x76, 0x0e, 0xf4, 0xb6, 0x3b, 0x7e, 0x6f, 0xe6, 0xbd,
	0x99, 0x59, 0x19, 0x9c, 0x30, 0x8e, 0xd2, 0xe9, 0x30, 0x15, 0x89, 0x4c, 0xa8, 0x23, 0x82, 0x1b,
	0xb9, 0x94, 0x89, 0xe0, 0xe9, 0xd4, 0xef, 0x03, 0x5c, 0x72, 0xc9, 0xf8, 0xd7, 0x8c, 0x2f, 0x25,
	0xf5, 0xa0, 0xf6, 0x85, 0xaf, 0x7a, 0xe4, 0x29, 0x79, 0xd6, 0x66, 0xea, 0xe8, 0x1f, 0x83, 0xa3,
	0xbf, 0x2f, 0xd3, 0x64, 0xb1, 0xe4, 0xb4, 0x0b, 0xf7, 0xbe, 0x05, 0xf3, 0x8c, 0x23, 0xc4, 0x5c,
	0xfc, 0x33, 0x80, 0xab, 0x6c, 0x77, 0x92, 0x0d, 0x6b, 0x3f, 0xcf, 0x72, 0xc1, 0xd1, 0x2c, 0x93,
	0xda, 0x3f, 0x02, 0xf7, 0x1d, 0x9f, 0x73, 0xc9, 0x77, 0x8b, 0xf1, 0xa0, 0x63, 0x21, 0x48, 0xfa,
	0x04, 0x07, 0xe3, 0x24, 0x4e, 0x83, 0x50, 0x7e, 0x4c, 0x66, 0x96, 0x78, 0x0c, 0x6e, 0x68, 0x82,
	0x93, 0xdb, 0x45, 0xc4, 0xbf, 0xeb, 0x14, 0x75, 0xd6, 0xc6, 0xe0, 0x07, 0x15, 0xa3, 0x47, 0x60,
	0xef, 0x13, 0xc9, 0x45, 0xac, 0xa5, 0xd5, 0x99, 0x83, 0xb1, 0x6b, 0x2e, 0x62, 0xbf, 0x0b, 0x34,
	0x9f, 0x1c, 0x4b, 0x9e, 0xc0, 0x01, 0x0b, 0x6e, 0xac, 0xdb, 0xf7, 0x3c, 0x88, 0xb8, 0xa0, 0x14,
	0xea, 0x3a, 0x8b, 0xa9, 0xa4, 0xcf, 0xfe, 0x39, 0x50, 0x03, 0x34, 0x44, 0x44, 0xaa, 0xba, 0x99,
	0x10, 0x7c, 0x81, 0x75, 0x09, 0xd6, 0x35, 0x31, 0x5d, 0xf7, 0x0f, 0x81, 0xa6, 0xf5, 0xf2, 0x02,
	0x5a, 0x61, 0x1c, 0x4d, 0xe4, 0x2a, 0x35, 0x3d, 0xef, 0x8c, 0xba, 0xc3, 0xdc, 0xfc, 0x86, 0xe3,
	0x38, 0xba, 0x5e, 0xa5, 0x9c, 0x35, 0x43, 0x73, 0xa0, 0xcf, 0xa1, 0x36, 0xe3, 0x52, 0xdb, 0x71,
	0x46, 0x0f, 0x0b, 0xd8, 0xcd, 0xa0, 0x99, 0xc2, 0x28, 0x68, 0x9a, 0xc9, 0x5e, 0xad, 0x02, 0xba,
	0x19, 0x27, 0x53, 0x18, 0x3a, 0x82, 0x46, 0xa4, 0x3b, 0xdf, 0xab, 0x6b, 0xf4, 0xe3, 0x02, 0xba,
	0x30, 0x37, 0x86, 0x48, 0xff, 0x2f, 0x81, 0xd6, 0x7a, 0x71, 0xfe, 0xdb, 0xc7, 0x20, 0xef, 0xa3,
	0x57, 0xf6, 0x61, 0xf2, 0x1a, 0x23, 0x83, 0xbc, 0x91, 0x5e, 0xd9, 0x88, 0xc5, 0x2a, 0x27, 0xa7,
	0x5b, 0x4e, 0x9e, 0x54, 0x3a, 0x41, 0x86, 0xb5, 0xf2, 0x83, 0x40, 0xfb, 0x22, 0x8a, 0x6f, 0x17,
	0x76, 0x2c, 0x67, 0x25, 0x3b, 0x8f, 0x0a, 0x79, 0x34, 0xb8, 0xe4, 0xe9, 0x0d, 0xd8, 0xfd, 0x9a,
	0xcc, 0x93, 0x19, 0x7a, 0xeb, 0x17, 0xfb, 0xb0, 0xbd, 0xcd, 0x0c, 0xc2, 0x75, 0xc8, 0xff, 0x49,
	0xc0, 0x45, 0x1d, 0xd8, 0xd7, 0xbb, 0x09, 0x79, 0x5b, 0x25, 0xe4, 0x70, 0xa7, 0x10, 0xec, 0x46,
	0x5e, 0xc9, 0x2f, 0x02, 0x1d, 0xb5, 0xdd, 0xe3, 0x38, 0xb2, 0x3d, 0x79, 0x05, 0x8d, 0xcf, 0x7a,
	0xc7, 0xb5, 0x90, 0x6d, 0x63, 0xa5, 0x37, 0xc3, 0x10, 0x4d, 0x87, 0xd0, 0x14, 0xe6, 0x03, 0x0a,
	0x29, 0x6e, 0x86, 0xed, 0x83, 0x05, 0xd1, 0xd7, 0xe0, 0x06, 0xca, 0xd5, 0xc4, 0xb2, 0xcc, 0xdc,
	0x2b, 0x7c, 0x5b, 0x6a, 0x3b, 0xc8, 0xdd, 0xfc, 0xdf, 0x04, 0xee, 0xaf, 0xa5, 0x63, 0x1b, 0xcf,
	0xb7, 0xb4, 0x1f, 0x56, 0x68, 0xcf, 0x3f, 0xe3, 0xb5, 0xf8, 0x97, 0xd0, 0x12, 0xf8, 0x05, 0xd5,
	0x3f, 0xd8, 0x52, 0x8f, 0xcd, 0x5b, 0xc3, 0xe8, 0x05, 0x74, 0xac, 0x7e, 0x24, 0xd6, 0x2a, 0xde,
	0x54, 0x61, 0xcc, 0xcc, 0x0d, 0xf2, 0xd7, 0xc1, 0x09, 0x34, 0x71, 0xa6, 0xb4, 0x09, 0xb5, 0x4b,
	0x2e, 0xbd, 0x3d, 0x75, 0xb8, 0xca, 0xa4, 0x47, 0x28, 0x40, 0xc3, 0xac, 0xb1, 0xb7, 0x3f, 0xe8,
	0xe3, 0xde, 0x5a, 0x74, 0x07, 0x60, 0x33, 0x58, 0x6f, 0x6f, 0xda, 0xd0, 0xbf, 0x84, 0xd3, 0x7f,
	0x01, 0x00, 0x00, 0xff, 0xff, 0x99, 0xee, 0xeb, 0x37, 0x21, 0x06, 0x00, 0x00,
}