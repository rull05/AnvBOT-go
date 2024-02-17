package anv

import (
	"context"
	"strings"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (anv *Anv) Reply(ctx context.Context, to string, message string, quoted *events.Message) (resp whatsmeow.SendResponse, err error) {
	parts := strings.Split(to, "@")
	user := parts[0]
	server := parts[1]

	jid := types.JID{
		User:       user,
		Server:     server,
		Integrator: 0,
	}
	var msg *waProto.Message
	if quoted != nil {
		msg = &waProto.Message{
			ExtendedTextMessage: &waProto.ExtendedTextMessage{
				Text: &message,
				ContextInfo: &waProto.ContextInfo{
					QuotedMessage: quoted.Message,
				},
			},
		}
	} else {
		msg = &waProto.Message{
			Conversation: &message,
		}
	}

	resp, err = anv.Client.SendMessage(ctx, jid, msg)
	if err != nil {
		anv.Client.Log.Errorf("Error sending message: %v", err)
	}

	return resp, err
}
