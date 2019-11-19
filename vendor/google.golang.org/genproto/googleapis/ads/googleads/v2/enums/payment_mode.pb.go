// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v2/enums/payment_mode.proto

package enums

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

// Enum describing possible payment modes.
type PaymentModeEnum_PaymentMode int32

const (
	// Not specified.
	PaymentModeEnum_UNSPECIFIED PaymentModeEnum_PaymentMode = 0
	// Used for return value only. Represents value unknown in this version.
	PaymentModeEnum_UNKNOWN PaymentModeEnum_PaymentMode = 1
	// Pay per click.
	PaymentModeEnum_CLICKS PaymentModeEnum_PaymentMode = 4
	// Pay per conversion value. This mode is only supported by campaigns with
	// AdvertisingChannelType.HOTEL, BiddingStrategyType.COMMISSION, and
	// BudgetType.HOTEL_ADS_COMMISSION.
	PaymentModeEnum_CONVERSION_VALUE PaymentModeEnum_PaymentMode = 5
	// Pay per conversion. This mode is only supported by campaigns with
	// AdvertisingChannelType.DISPLAY (excluding
	// AdvertisingChannelSubType.DISPLAY_GMAIL), BiddingStrategyType.TARGET_CPA,
	// and BudgetType.FIXED_CPA. The customer must also be eligible for this
	// mode. See Customer.eligibility_failure_reasons for details.
	PaymentModeEnum_CONVERSIONS PaymentModeEnum_PaymentMode = 6
)

var PaymentModeEnum_PaymentMode_name = map[int32]string{
	0: "UNSPECIFIED",
	1: "UNKNOWN",
	4: "CLICKS",
	5: "CONVERSION_VALUE",
	6: "CONVERSIONS",
}

var PaymentModeEnum_PaymentMode_value = map[string]int32{
	"UNSPECIFIED":      0,
	"UNKNOWN":          1,
	"CLICKS":           4,
	"CONVERSION_VALUE": 5,
	"CONVERSIONS":      6,
}

func (x PaymentModeEnum_PaymentMode) String() string {
	return proto.EnumName(PaymentModeEnum_PaymentMode_name, int32(x))
}

func (PaymentModeEnum_PaymentMode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_3680001112a7164a, []int{0, 0}
}

// Container for enum describing possible payment modes.
type PaymentModeEnum struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PaymentModeEnum) Reset()         { *m = PaymentModeEnum{} }
func (m *PaymentModeEnum) String() string { return proto.CompactTextString(m) }
func (*PaymentModeEnum) ProtoMessage()    {}
func (*PaymentModeEnum) Descriptor() ([]byte, []int) {
	return fileDescriptor_3680001112a7164a, []int{0}
}

func (m *PaymentModeEnum) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PaymentModeEnum.Unmarshal(m, b)
}
func (m *PaymentModeEnum) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PaymentModeEnum.Marshal(b, m, deterministic)
}
func (m *PaymentModeEnum) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PaymentModeEnum.Merge(m, src)
}
func (m *PaymentModeEnum) XXX_Size() int {
	return xxx_messageInfo_PaymentModeEnum.Size(m)
}
func (m *PaymentModeEnum) XXX_DiscardUnknown() {
	xxx_messageInfo_PaymentModeEnum.DiscardUnknown(m)
}

var xxx_messageInfo_PaymentModeEnum proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("google.ads.googleads.v2.enums.PaymentModeEnum_PaymentMode", PaymentModeEnum_PaymentMode_name, PaymentModeEnum_PaymentMode_value)
	proto.RegisterType((*PaymentModeEnum)(nil), "google.ads.googleads.v2.enums.PaymentModeEnum")
}

func init() {
	proto.RegisterFile("google/ads/googleads/v2/enums/payment_mode.proto", fileDescriptor_3680001112a7164a)
}

var fileDescriptor_3680001112a7164a = []byte{
	// 312 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x50, 0xcf, 0x4a, 0xc3, 0x30,
	0x1c, 0x76, 0x53, 0x27, 0x64, 0x87, 0x85, 0xe2, 0x49, 0xdc, 0x61, 0x7b, 0x80, 0x54, 0xea, 0x2d,
	0x9e, 0xb2, 0x5a, 0x47, 0xd9, 0xcc, 0x8a, 0x65, 0x15, 0xa4, 0x38, 0xa2, 0x09, 0x61, 0xb0, 0x26,
	0x75, 0xe9, 0x06, 0xbe, 0x8e, 0x47, 0x1f, 0xc5, 0xf7, 0xf0, 0xe2, 0x53, 0x48, 0x13, 0x57, 0x77,
	0xd1, 0x4b, 0xf8, 0xf8, 0x7d, 0x7f, 0xf8, 0xf2, 0x81, 0x0b, 0xa9, 0xb5, 0x5c, 0x09, 0x9f, 0x71,
	0xe3, 0x3b, 0x58, 0xa3, 0x6d, 0xe0, 0x0b, 0xb5, 0x29, 0x8c, 0x5f, 0xb2, 0xd7, 0x42, 0xa8, 0x6a,
	0x51, 0x68, 0x2e, 0x50, 0xb9, 0xd6, 0x95, 0xf6, 0xfa, 0x4e, 0x86, 0x18, 0x37, 0xa8, 0x71, 0xa0,
	0x6d, 0x80, 0xac, 0xe3, 0xec, 0x7c, 0x17, 0x58, 0x2e, 0x7d, 0xa6, 0x94, 0xae, 0x58, 0xb5, 0xd4,
	0xca, 0x38, 0xf3, 0xf0, 0x05, 0xf4, 0x12, 0x17, 0x79, 0xab, 0xb9, 0x88, 0xd4, 0xa6, 0x18, 0x3e,
	0x82, 0xee, 0xde, 0xc9, 0xeb, 0x81, 0xee, 0x9c, 0xa6, 0x49, 0x14, 0xc6, 0x37, 0x71, 0x74, 0x0d,
	0x0f, 0xbc, 0x2e, 0x38, 0x99, 0xd3, 0x09, 0x9d, 0xdd, 0x53, 0xd8, 0xf2, 0x00, 0xe8, 0x84, 0xd3,
	0x38, 0x9c, 0xa4, 0xf0, 0xc8, 0x3b, 0x05, 0x30, 0x9c, 0xd1, 0x2c, 0xba, 0x4b, 0xe3, 0x19, 0x5d,
	0x64, 0x64, 0x3a, 0x8f, 0xe0, 0x71, 0xed, 0xff, 0xbd, 0xa6, 0xb0, 0x33, 0xfa, 0x6c, 0x81, 0xc1,
	0xb3, 0x2e, 0xd0, 0xbf, 0xb5, 0x47, 0x70, 0xaf, 0x43, 0x52, 0x57, 0x4d, 0x5a, 0x0f, 0xa3, 0x1f,
	0x8b, 0xd4, 0x2b, 0xa6, 0x24, 0xd2, 0x6b, 0xe9, 0x4b, 0xa1, 0xec, 0x47, 0x76, 0x5b, 0x95, 0x4b,
	0xf3, 0xc7, 0x74, 0x57, 0xf6, 0x7d, 0x6b, 0x1f, 0x8e, 0x09, 0x79, 0x6f, 0xf7, 0xc7, 0x2e, 0x8a,
	0x70, 0x83, 0x1c, 0xac, 0x51, 0x16, 0xa0, 0x7a, 0x02, 0xf3, 0xb1, 0xe3, 0x73, 0xc2, 0x4d, 0xde,
	0xf0, 0x79, 0x16, 0xe4, 0x96, 0xff, 0x6a, 0x0f, 0xdc, 0x11, 0x63, 0xc2, 0x0d, 0xc6, 0x8d, 0x02,
	0xe3, 0x2c, 0xc0, 0xd8, 0x6a, 0x9e, 0x3a, 0xb6, 0xd8, 0xe5, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff,
	0xe2, 0x41, 0xc7, 0x1f, 0xd2, 0x01, 0x00, 0x00,
}