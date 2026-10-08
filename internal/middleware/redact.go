package middleware

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	redactedValue = "***"
)

var secretFieldNames = map[protoreflect.Name]struct{}{
	"access_token":       {},
	"password":           {},
	"secret_access_key":  {},
	"web_ui_password":    {},
	"secret":             {},
	"token":              {},
	"registration_token": {},
}

func redactSecrets(message any) any {
	protoMessage, ok := message.(proto.Message)
	if !ok || protoMessage == nil {
		return message
	}

	reflected := protoMessage.ProtoReflect()
	if !reflected.IsValid() {
		return message
	}

	clone := proto.Clone(protoMessage)
	redactMessage(clone.ProtoReflect())

	return clone
}

func redactMessage(message protoreflect.Message) {
	message.Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		switch {
		case field.IsMap():
			redactMap(field, value.Map())
		case field.IsList():
			redactList(field, value.List())
		case field.Kind() == protoreflect.MessageKind || field.Kind() == protoreflect.GroupKind:
			redactMessage(value.Message())
		case field.Kind() == protoreflect.StringKind:
			redactString(message, field, value.String())
		}

		return true
	})
}

func redactString(message protoreflect.Message, field protoreflect.FieldDescriptor, value string) {
	if !isSecretName(field.Name()) || value == "" {
		return
	}

	message.Set(field, protoreflect.ValueOfString(redactedValue))
}

func redactList(field protoreflect.FieldDescriptor, list protoreflect.List) {
	isMessage := field.Kind() == protoreflect.MessageKind || field.Kind() == protoreflect.GroupKind
	isSecretString := field.Kind() == protoreflect.StringKind && isSecretName(field.Name())

	for i := range list.Len() {
		switch {
		case isMessage:
			redactMessage(list.Get(i).Message())
		case isSecretString && list.Get(i).String() != "":
			list.Set(i, protoreflect.ValueOfString(redactedValue))
		}
	}
}

func redactMap(field protoreflect.FieldDescriptor, entries protoreflect.Map) {
	valueField := field.MapValue()
	isMessage := valueField.Kind() == protoreflect.MessageKind || valueField.Kind() == protoreflect.GroupKind
	isSecretString := valueField.Kind() == protoreflect.StringKind && isSecretName(field.Name())

	entries.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
		switch {
		case isMessage:
			redactMessage(value.Message())
		case isSecretString && value.String() != "":
			entries.Set(key, protoreflect.ValueOfString(redactedValue))
		}

		return true
	})
}

func isSecretName(name protoreflect.Name) bool {
	_, isSecret := secretFieldNames[name]

	return isSecret
}
