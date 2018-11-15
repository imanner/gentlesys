// Code generated by protoc-gen-go. DO NOT EDIT.
// source: comment.proto

package comment

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
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type CommentStory struct {
	Commentdata          []*CommentData `protobuf:"bytes,1,rep,name=commentdata" json:"commentdata,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *CommentStory) Reset()         { *m = CommentStory{} }
func (m *CommentStory) String() string { return proto.CompactTextString(m) }
func (*CommentStory) ProtoMessage()    {}
func (*CommentStory) Descriptor() ([]byte, []int) {
	return fileDescriptor_749aee09ea917828, []int{0}
}

func (m *CommentStory) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CommentStory.Unmarshal(m, b)
}
func (m *CommentStory) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CommentStory.Marshal(b, m, deterministic)
}
func (m *CommentStory) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommentStory.Merge(m, src)
}
func (m *CommentStory) XXX_Size() int {
	return xxx_messageInfo_CommentStory.Size(m)
}
func (m *CommentStory) XXX_DiscardUnknown() {
	xxx_messageInfo_CommentStory.DiscardUnknown(m)
}

var xxx_messageInfo_CommentStory proto.InternalMessageInfo

func (m *CommentStory) GetCommentdata() []*CommentData {
	if m != nil {
		return m.Commentdata
	}
	return nil
}

type CommentData struct {
	UserName             *string  `protobuf:"bytes,1,opt,name=userName" json:"userName,omitempty"`
	AnswerName           *string  `protobuf:"bytes,2,opt,name=answerName" json:"answerName,omitempty"`
	Time                 *string  `protobuf:"bytes,3,opt,name=time" json:"time,omitempty"`
	Content              *string  `protobuf:"bytes,4,opt,name=content" json:"content,omitempty"`
	Id                   *int32   `protobuf:"varint,5,opt,name=id" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CommentData) Reset()         { *m = CommentData{} }
func (m *CommentData) String() string { return proto.CompactTextString(m) }
func (*CommentData) ProtoMessage()    {}
func (*CommentData) Descriptor() ([]byte, []int) {
	return fileDescriptor_749aee09ea917828, []int{1}
}

func (m *CommentData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CommentData.Unmarshal(m, b)
}
func (m *CommentData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CommentData.Marshal(b, m, deterministic)
}
func (m *CommentData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommentData.Merge(m, src)
}
func (m *CommentData) XXX_Size() int {
	return xxx_messageInfo_CommentData.Size(m)
}
func (m *CommentData) XXX_DiscardUnknown() {
	xxx_messageInfo_CommentData.DiscardUnknown(m)
}

var xxx_messageInfo_CommentData proto.InternalMessageInfo

func (m *CommentData) GetUserName() string {
	if m != nil && m.UserName != nil {
		return *m.UserName
	}
	return ""
}

func (m *CommentData) GetAnswerName() string {
	if m != nil && m.AnswerName != nil {
		return *m.AnswerName
	}
	return ""
}

func (m *CommentData) GetTime() string {
	if m != nil && m.Time != nil {
		return *m.Time
	}
	return ""
}

func (m *CommentData) GetContent() string {
	if m != nil && m.Content != nil {
		return *m.Content
	}
	return ""
}

func (m *CommentData) GetId() int32 {
	if m != nil && m.Id != nil {
		return *m.Id
	}
	return 0
}

type UserTopics struct {
	Usertopicdata        []*UserTopicData `protobuf:"bytes,1,rep,name=usertopicdata" json:"usertopicdata,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *UserTopics) Reset()         { *m = UserTopics{} }
func (m *UserTopics) String() string { return proto.CompactTextString(m) }
func (*UserTopics) ProtoMessage()    {}
func (*UserTopics) Descriptor() ([]byte, []int) {
	return fileDescriptor_749aee09ea917828, []int{2}
}

func (m *UserTopics) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UserTopics.Unmarshal(m, b)
}
func (m *UserTopics) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UserTopics.Marshal(b, m, deterministic)
}
func (m *UserTopics) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UserTopics.Merge(m, src)
}
func (m *UserTopics) XXX_Size() int {
	return xxx_messageInfo_UserTopics.Size(m)
}
func (m *UserTopics) XXX_DiscardUnknown() {
	xxx_messageInfo_UserTopics.DiscardUnknown(m)
}

var xxx_messageInfo_UserTopics proto.InternalMessageInfo

func (m *UserTopics) GetUsertopicdata() []*UserTopicData {
	if m != nil {
		return m.Usertopicdata
	}
	return nil
}

type UserTopicData struct {
	Sid                  *int32   `protobuf:"varint,1,opt,name=sid" json:"sid,omitempty"`
	Aid                  *int32   `protobuf:"varint,2,opt,name=aid" json:"aid,omitempty"`
	Time                 *string  `protobuf:"bytes,3,opt,name=time" json:"time,omitempty"`
	Title                *string  `protobuf:"bytes,4,opt,name=title" json:"title,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UserTopicData) Reset()         { *m = UserTopicData{} }
func (m *UserTopicData) String() string { return proto.CompactTextString(m) }
func (*UserTopicData) ProtoMessage()    {}
func (*UserTopicData) Descriptor() ([]byte, []int) {
	return fileDescriptor_749aee09ea917828, []int{3}
}

func (m *UserTopicData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UserTopicData.Unmarshal(m, b)
}
func (m *UserTopicData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UserTopicData.Marshal(b, m, deterministic)
}
func (m *UserTopicData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UserTopicData.Merge(m, src)
}
func (m *UserTopicData) XXX_Size() int {
	return xxx_messageInfo_UserTopicData.Size(m)
}
func (m *UserTopicData) XXX_DiscardUnknown() {
	xxx_messageInfo_UserTopicData.DiscardUnknown(m)
}

var xxx_messageInfo_UserTopicData proto.InternalMessageInfo

func (m *UserTopicData) GetSid() int32 {
	if m != nil && m.Sid != nil {
		return *m.Sid
	}
	return 0
}

func (m *UserTopicData) GetAid() int32 {
	if m != nil && m.Aid != nil {
		return *m.Aid
	}
	return 0
}

func (m *UserTopicData) GetTime() string {
	if m != nil && m.Time != nil {
		return *m.Time
	}
	return ""
}

func (m *UserTopicData) GetTitle() string {
	if m != nil && m.Title != nil {
		return *m.Title
	}
	return ""
}

func init() {
	proto.RegisterType((*CommentStory)(nil), "comment.CommentStory")
	proto.RegisterType((*CommentData)(nil), "comment.CommentData")
	proto.RegisterType((*UserTopics)(nil), "comment.UserTopics")
	proto.RegisterType((*UserTopicData)(nil), "comment.UserTopicData")
}

func init() { proto.RegisterFile("comment.proto", fileDescriptor_749aee09ea917828) }

var fileDescriptor_749aee09ea917828 = []byte{
	// 244 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x8f, 0x31, 0x4b, 0x04, 0x31,
	0x14, 0x84, 0xc9, 0xee, 0x85, 0xd3, 0xb7, 0xae, 0xc8, 0xe3, 0x90, 0x60, 0x21, 0xcb, 0x56, 0x5b,
	0x5d, 0x61, 0x61, 0x65, 0xa7, 0x58, 0x58, 0x58, 0x44, 0x2d, 0x2d, 0xc2, 0x26, 0x45, 0xc0, 0xdd,
	0x1c, 0xc9, 0x13, 0xf1, 0x17, 0xf8, 0xb7, 0x25, 0xb9, 0xdc, 0x92, 0x83, 0xeb, 0x66, 0xe6, 0x9b,
	0x90, 0x37, 0xd0, 0x8e, 0x6e, 0x9a, 0xcc, 0x4c, 0xdb, 0x9d, 0x77, 0xe4, 0x70, 0x9d, 0x6d, 0xff,
	0x0c, 0x17, 0x8f, 0x7b, 0xf9, 0x46, 0xce, 0xff, 0xe2, 0x3d, 0x34, 0x19, 0x69, 0x45, 0x4a, 0xb0,
	0xae, 0x1e, 0x9a, 0xbb, 0xcd, 0xf6, 0xf0, 0x3a, 0x77, 0x9f, 0x14, 0x29, 0x59, 0x16, 0xfb, 0x3f,
	0x06, 0x4d, 0x01, 0xf1, 0x06, 0xce, 0xbe, 0x83, 0xf1, 0xaf, 0x6a, 0x32, 0x82, 0x75, 0x6c, 0x38,
	0x97, 0x8b, 0xc7, 0x5b, 0x00, 0x35, 0x87, 0x9f, 0x4c, 0xab, 0x44, 0x8b, 0x04, 0x11, 0x56, 0x64,
	0x27, 0x23, 0xea, 0x44, 0x92, 0x46, 0x01, 0xeb, 0xd1, 0xcd, 0x64, 0x66, 0x12, 0xab, 0x14, 0x1f,
	0x2c, 0x5e, 0x42, 0x65, 0xb5, 0xe0, 0x1d, 0x1b, 0xb8, 0xac, 0xac, 0xee, 0x5f, 0x00, 0x3e, 0x82,
	0xf1, 0xef, 0x6e, 0x67, 0xc7, 0x80, 0x0f, 0xd0, 0xc6, 0x7f, 0x29, 0xba, 0x62, 0xd1, 0xf5, 0xb2,
	0x68, 0xe9, 0xa6, 0x4d, 0xc7, 0xe5, 0xfe, 0x13, 0xda, 0x23, 0x8e, 0x57, 0x50, 0x07, 0xab, 0xd3,
	0x22, 0x2e, 0xa3, 0x8c, 0x89, 0xb2, 0x3a, 0xad, 0xe0, 0x32, 0xca, 0x93, 0xe7, 0x6f, 0x80, 0x93,
	0xa5, 0x2f, 0x93, 0x8f, 0xdf, 0x9b, 0xff, 0x00, 0x00, 0x00, 0xff, 0xff, 0x7e, 0x3e, 0xf5, 0xe7,
	0x95, 0x01, 0x00, 0x00,
}