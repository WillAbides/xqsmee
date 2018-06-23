// Code generated by protoc-gen-go. DO NOT EDIT.
// source: queue.proto

/*
Package queue is a generated protocol buffer package.

It is generated from these files:
	queue.proto

It has these top-level messages:
	Header
	WebRequest
	PopRequest
	PopResponse
	PeekRequest
	PeekResponse
*/
package queue

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Header struct {
	Name  string   `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Value []string `protobuf:"bytes,2,rep,name=value" json:"value,omitempty"`
}

func (m *Header) Reset()                    { *m = Header{} }
func (m *Header) String() string            { return proto.CompactTextString(m) }
func (*Header) ProtoMessage()               {}
func (*Header) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Header) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Header) GetValue() []string {
	if m != nil {
		return m.Value
	}
	return nil
}

type WebRequest struct {
	ReceivedAt int64     `protobuf:"varint,1,opt,name=ReceivedAt" json:"ReceivedAt,omitempty"`
	Header     []*Header `protobuf:"bytes,2,rep,name=Header" json:"Header,omitempty"`
	Host       string    `protobuf:"bytes,3,opt,name=Host" json:"Host,omitempty"`
	Body       string    `protobuf:"bytes,4,opt,name=Body" json:"Body,omitempty"`
}

func (m *WebRequest) Reset()                    { *m = WebRequest{} }
func (m *WebRequest) String() string            { return proto.CompactTextString(m) }
func (*WebRequest) ProtoMessage()               {}
func (*WebRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *WebRequest) GetReceivedAt() int64 {
	if m != nil {
		return m.ReceivedAt
	}
	return 0
}

func (m *WebRequest) GetHeader() []*Header {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *WebRequest) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

func (m *WebRequest) GetBody() string {
	if m != nil {
		return m.Body
	}
	return ""
}

type PopRequest struct {
	QueueName string `protobuf:"bytes,1,opt,name=QueueName" json:"QueueName,omitempty"`
	Timeout   int64  `protobuf:"varint,2,opt,name=Timeout" json:"Timeout,omitempty"`
}

func (m *PopRequest) Reset()                    { *m = PopRequest{} }
func (m *PopRequest) String() string            { return proto.CompactTextString(m) }
func (*PopRequest) ProtoMessage()               {}
func (*PopRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PopRequest) GetQueueName() string {
	if m != nil {
		return m.QueueName
	}
	return ""
}

func (m *PopRequest) GetTimeout() int64 {
	if m != nil {
		return m.Timeout
	}
	return 0
}

type PopResponse struct {
	WebRequest *WebRequest `protobuf:"bytes,1,opt,name=WebRequest" json:"WebRequest,omitempty"`
}

func (m *PopResponse) Reset()                    { *m = PopResponse{} }
func (m *PopResponse) String() string            { return proto.CompactTextString(m) }
func (*PopResponse) ProtoMessage()               {}
func (*PopResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *PopResponse) GetWebRequest() *WebRequest {
	if m != nil {
		return m.WebRequest
	}
	return nil
}

type PeekRequest struct {
	QueueName string `protobuf:"bytes,1,opt,name=QueueName" json:"QueueName,omitempty"`
	Count     int64  `protobuf:"varint,2,opt,name=Count" json:"Count,omitempty"`
}

func (m *PeekRequest) Reset()                    { *m = PeekRequest{} }
func (m *PeekRequest) String() string            { return proto.CompactTextString(m) }
func (*PeekRequest) ProtoMessage()               {}
func (*PeekRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *PeekRequest) GetQueueName() string {
	if m != nil {
		return m.QueueName
	}
	return ""
}

func (m *PeekRequest) GetCount() int64 {
	if m != nil {
		return m.Count
	}
	return 0
}

type PeekResponse struct {
	WebRequest []*WebRequest `protobuf:"bytes,1,rep,name=WebRequest" json:"WebRequest,omitempty"`
}

func (m *PeekResponse) Reset()                    { *m = PeekResponse{} }
func (m *PeekResponse) String() string            { return proto.CompactTextString(m) }
func (*PeekResponse) ProtoMessage()               {}
func (*PeekResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *PeekResponse) GetWebRequest() []*WebRequest {
	if m != nil {
		return m.WebRequest
	}
	return nil
}

func init() {
	proto.RegisterType((*Header)(nil), "Header")
	proto.RegisterType((*WebRequest)(nil), "WebRequest")
	proto.RegisterType((*PopRequest)(nil), "PopRequest")
	proto.RegisterType((*PopResponse)(nil), "PopResponse")
	proto.RegisterType((*PeekRequest)(nil), "PeekRequest")
	proto.RegisterType((*PeekResponse)(nil), "PeekResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Queue service

type QueueClient interface {
	Pop(ctx context.Context, in *PopRequest, opts ...grpc.CallOption) (*PopResponse, error)
	Peek(ctx context.Context, in *PeekRequest, opts ...grpc.CallOption) (*PeekResponse, error)
}

type queueClient struct {
	cc *grpc.ClientConn
}

func NewQueueClient(cc *grpc.ClientConn) QueueClient {
	return &queueClient{cc}
}

func (c *queueClient) Pop(ctx context.Context, in *PopRequest, opts ...grpc.CallOption) (*PopResponse, error) {
	out := new(PopResponse)
	err := grpc.Invoke(ctx, "/Queue/Pop", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueClient) Peek(ctx context.Context, in *PeekRequest, opts ...grpc.CallOption) (*PeekResponse, error) {
	out := new(PeekResponse)
	err := grpc.Invoke(ctx, "/Queue/Peek", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Queue service

type QueueServer interface {
	Pop(context.Context, *PopRequest) (*PopResponse, error)
	Peek(context.Context, *PeekRequest) (*PeekResponse, error)
}

func RegisterQueueServer(s *grpc.Server, srv QueueServer) {
	s.RegisterService(&_Queue_serviceDesc, srv)
}

func _Queue_Pop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PopRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueServer).Pop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Queue/Pop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueServer).Pop(ctx, req.(*PopRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Queue_Peek_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PeekRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueServer).Peek(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Queue/Peek",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueServer).Peek(ctx, req.(*PeekRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Queue_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Queue",
	HandlerType: (*QueueServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Pop",
			Handler:    _Queue_Pop_Handler,
		},
		{
			MethodName: "Peek",
			Handler:    _Queue_Peek_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "queue.proto",
}

func init() { proto.RegisterFile("queue.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 279 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x31, 0x6b, 0xc3, 0x30,
	0x10, 0x85, 0x49, 0x6c, 0x27, 0xf8, 0xe4, 0x2e, 0x47, 0x06, 0x11, 0x4a, 0x6b, 0xd4, 0x25, 0x50,
	0xd0, 0xe0, 0x6e, 0xed, 0x94, 0xb6, 0x43, 0xa6, 0xe0, 0x8a, 0x42, 0x67, 0xa7, 0xbe, 0x21, 0xb4,
	0xb1, 0x9c, 0xd8, 0x0a, 0xf4, 0xdf, 0x17, 0x49, 0x18, 0x6b, 0x69, 0xc9, 0x76, 0xef, 0x09, 0xbd,
	0xef, 0x9d, 0x10, 0xb0, 0xa3, 0x21, 0x43, 0xb2, 0x3d, 0xe9, 0x5e, 0x8b, 0x02, 0x66, 0x1b, 0xaa,
	0x6a, 0x3a, 0x21, 0x42, 0xdc, 0x54, 0x07, 0xe2, 0x93, 0x7c, 0xb2, 0x4a, 0x95, 0x9b, 0x71, 0x01,
	0xc9, 0xb9, 0xfa, 0x36, 0xc4, 0xa7, 0x79, 0xb4, 0x4a, 0x95, 0x17, 0xc2, 0x00, 0x7c, 0xd0, 0x4e,
	0xd1, 0xd1, 0x50, 0xd7, 0xe3, 0x0d, 0x80, 0xa2, 0x4f, 0xda, 0x9f, 0xa9, 0x5e, 0xf7, 0xee, 0x76,
	0xa4, 0x02, 0x07, 0x6f, 0x07, 0x82, 0x0b, 0x61, 0xc5, 0x5c, 0x7a, 0xa9, 0x02, 0xf0, 0x46, 0x77,
	0x3d, 0x8f, 0x3c, 0xd8, 0xce, 0xd6, 0x7b, 0xd6, 0xf5, 0x0f, 0x8f, 0xbd, 0x67, 0x67, 0xf1, 0x0a,
	0x50, 0xea, 0x76, 0xc0, 0x5e, 0x43, 0xfa, 0x66, 0xf7, 0xd8, 0x8e, 0x9d, 0x47, 0x03, 0x39, 0xcc,
	0xdf, 0xf7, 0x07, 0xd2, 0xa6, 0xe7, 0x53, 0xd7, 0x68, 0x90, 0xe2, 0x11, 0x98, 0x4b, 0xe9, 0x5a,
	0xdd, 0x74, 0x84, 0xf7, 0xe1, 0x2e, 0x2e, 0x87, 0x15, 0x4c, 0x8e, 0x96, 0x0a, 0x8e, 0xc5, 0x1a,
	0x58, 0x49, 0xf4, 0x75, 0x59, 0x85, 0x05, 0x24, 0x2f, 0xda, 0x34, 0x43, 0x01, 0x2f, 0xc4, 0x13,
	0x64, 0x3e, 0xe2, 0x0f, 0x7e, 0xf4, 0x0f, 0xbf, 0xd8, 0x42, 0xe2, 0xf2, 0x31, 0x87, 0xa8, 0xd4,
	0x2d, 0x32, 0x39, 0x3e, 0xc8, 0x32, 0x93, 0xe1, 0x5e, 0x77, 0x10, 0x5b, 0x0e, 0x66, 0x32, 0x68,
	0xbc, 0xbc, 0x92, 0x21, 0x7c, 0x37, 0x73, 0x7f, 0xe0, 0xe1, 0x37, 0x00, 0x00, 0xff, 0xff, 0x4e,
	0xc9, 0xf1, 0x92, 0x12, 0x02, 0x00, 0x00,
}
