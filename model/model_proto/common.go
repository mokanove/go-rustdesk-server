package model_proto

import (
	"go-rustdesk-server/cmd"
	"google.golang.org/protobuf/reflect/protoreflect"
	"reflect"
	"strings"
)

func NewRendezvousMessage(msg protoreflect.ProtoMessage) *RendezvousMessage {
	typeR := reflect.TypeOf(msg)
	ts := strings.ReplaceAll(strings.ReplaceAll(typeR.String(), "*", ""), "model_proto.", "")
	typeR2 := getTypeByName("RendezvousMessage_" + ts)
	if typeR2 == nil {
		cmd.Fatal("not find type ", "RendezvousMessage_"+ts)
		return nil
	}
	newMsg := reflect.New(typeR2)
	f := newMsg.Elem().FieldByName(ts)
	if !f.CanSet() {
		return nil
	}
	f.Set(reflect.ValueOf(msg))
	message := &RendezvousMessage{}
	name := reflect.ValueOf(message).Elem().FieldByName("Union")
	name.Set(newMsg)
	return message
}
