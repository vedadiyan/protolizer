package protolizer

import (
	"encoding/hex"
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

func TestSerializeStruct_User(t *testing.T) {
	// prepare values
	id := int64(123)
	emergency := "911"
	profile := "avatar.png"

	u := &User{
		Id:               &id,
		FirstName:        "John",
		LastName:         "Doe",
		Email:            "john.doe@example.com",
		Phone:            "555-1234",
		EmergencyContact: &emergency,
		DateOfBirth:      "2000-01-01",
		ProfilePicture:   &profile,
		Password:         "secret",
	}

	// register type explicitly
	RegisterType(reflect.TypeOf(new(User)))

	// serialize
	b, err := SerializeStruct(u)
	if err != nil {
		t.Fatalf("SerializeStruct failed: %v", err)
	}

	if len(b) == 0 {
		t.Fatal("expected non-empty serialization")
	}

	// log encoded bytes for inspection
	t.Logf("Encoded User protobuf: %s", hex.EncodeToString(b))
}

func TestSerializeDeserialize_User(t *testing.T) {
	// prepare values
	id := int64(123)
	emergency := "911"
	profile := "avatar.png"

	u := &User{
		Id:               &id,
		FirstName:        "John",
		LastName:         "Doe",
		Email:            "john.doe@example.com",
		Phone:            "555-1234",
		EmergencyContact: &emergency,
		DateOfBirth:      "2000-01-01",
		ProfilePicture:   &profile,
		Password:         "secret",
	}

	// register type explicitly
	RegisterType(reflect.TypeOf(new(User)))

	// serialize
	b, err := SerializeStruct(u)
	if err != nil {
		t.Fatalf("SerializeStruct failed: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty serialization")
	}
	t.Logf("Encoded User protobuf: %s", hex.EncodeToString(b))

	// deserialize into a new User
	var u2 User
	if err := DeserializeStruct(b, &u2); err != nil {
		t.Fatalf("DeserializeStruct failed: %v", err)
	}

	// compare round trip
	if u2.FirstName != u.FirstName ||
		u2.LastName != u.LastName ||
		u2.Email != u.Email ||
		u2.Phone != u.Phone ||
		u2.DateOfBirth != u.DateOfBirth ||
		u2.Password != u.Password {
		t.Errorf("mismatch after round trip: got %+v, want %+v", u2, u)
	}
	if u2.Id == nil || *u2.Id != *u.Id {
		t.Errorf("Id mismatch: got %v, want %v", u2.Id, u.Id)
	}
	if u2.EmergencyContact == nil || *u2.EmergencyContact != *u.EmergencyContact {
		t.Errorf("EmergencyContact mismatch: got %v, want %v", u2.EmergencyContact, u.EmergencyContact)
	}
	if u2.ProfilePicture == nil || *u2.ProfilePicture != *u.ProfilePicture {
		t.Errorf("ProfilePicture mismatch: got %v, want %v", u2.ProfilePicture, u.ProfilePicture)
	}
}
