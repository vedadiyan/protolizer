package protolizer

import (
	"reflect"
	"testing"
)

type User struct {
	Id               *int64  `protobuf:"varint,1,opt,name=id,proto3,oneof" json:"id,omitempty"`
	FirstName        string  `protobuf:"bytes,2,opt,name=first_name,proto3" json:"first_name,omitempty"`
	LastName         string  `protobuf:"bytes,3,opt,name=last_name,proto3" json:"last_name,omitempty"`
	Email            string  `protobuf:"bytes,4,opt,name=email,proto3" json:"email,omitempty"`
	Phone            string  `protobuf:"bytes,5,opt,name=phone,proto3" json:"phone,omitempty"`
	EmergencyContact *string `protobuf:"bytes,6,opt,name=emergency_contact,proto3,oneof" json:"emergency_contact,omitempty"`
	DateOfBirth      string  `protobuf:"bytes,7,opt,name=date_of_birth,proto3" json:"date_of_birth,omitempty"`
	ProfilePicture   *string `protobuf:"bytes,8,opt,name=profile_picture,proto3,oneof" json:"profile_picture,omitempty"`
	Password         string  `protobuf:"bytes,9,opt,name=password,proto3" json:"password,omitempty"`
}

func TestRegisterType(t *testing.T) {
	RegisterType(reflect.TypeOf(new(User)))
	RegisterType(reflect.TypeOf(new(User)))
}
