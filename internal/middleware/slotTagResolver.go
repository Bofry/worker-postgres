package middleware

import (
	"github.com/Bofry/structproto"
)

var _ structproto.TagResolver = SlotTagResolver

func SlotTagResolver(fieldname, token string) (*structproto.Tag, error) {
	var tag *structproto.Tag
	if token != "" && token != "-" {
		tag = &structproto.Tag{
			Name: token,
		}
	}
	return tag, nil
}
