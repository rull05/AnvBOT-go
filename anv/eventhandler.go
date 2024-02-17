package anv

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type EvtHandler struct {
	Client *whatsmeow.Client
	Log    waLog.Logger
}

type logLevels string

const (
	DEBUG logLevels = "DEBUG"
	WARN  logLevels = "WARN"
	INFO  logLevels = "INFO"
	ERROR logLevels = "ERROR"
)

func NewEvtHandler(client *whatsmeow.Client, logLevel logLevels) *EvtHandler {
	return &EvtHandler{
		Client: client,
		Log:    waLog.Stdout("EventHandler", string(logLevel), true),
	}
}

func (h *EvtHandler) HandleEvent(rawEvt interface{}) {

	switch evt := rawEvt.(type) {
	case *events.Connected, *events.PushNameSetting:
		if len(h.Client.Store.PushName) == 0 {
			return
		}
		// Send presence available when connecting and when the pushname is changed.
		// This makes sure that outgoing messages always have the right pushname.
		err := h.Client.SendPresence(types.PresenceAvailable)
		if err != nil {
			h.Log.Warnf("Failed to send available presence: %v", err)
		} else {
			h.Log.Infof("Marked self as available")
		}
		h.handleConected(evt.(*events.Connected))
	case *events.Message:
		h.handleMessage(evt)

	}
}

func (h *EvtHandler) handleConected(evt *events.Connected) {
	h.Log.Infof("Connected to WhatsApp")
}

// handle message
func (h *EvtHandler) handleMessage(evt *events.Message) {
	h.Log.Infof(evt.Info.SourceString())
	h.Log.Infof("Received message %s from %s: %+v", evt.Info.ID, evt.Info.SourceString(), evt.Message)
}
