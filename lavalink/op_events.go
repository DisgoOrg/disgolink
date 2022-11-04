package lavalink

import (
	"github.com/disgoorg/json"

	"github.com/disgoorg/snowflake/v2"
)

type UnmarshalEvent struct {
	Event
}

func (e *UnmarshalEvent) UnmarshalJSON(data []byte) error {
	var eType struct {
		Type EventType `json:"type"`
	}
	if err := json.Unmarshal(data, &eType); err != nil {
		return err
	}

	var err error

	switch eType.Type {
	case EventTypeTrackStart:
		var v TrackStartEvent
		err = json.Unmarshal(data, &v)
		e.Event = v

	case EventTypeTrackEnd:
		var v TrackEndEvent
		err = json.Unmarshal(data, &v)
		e.Event = v

	case EventTypeTrackException:
		var v TrackExceptionEvent
		err = json.Unmarshal(data, &v)
		e.Event = v

	case EventTypeTrackStuck:
		var v TrackStuckEvent
		err = json.Unmarshal(data, &v)
		e.Event = v

	case EventTypeWebSocketClosed:
		var v WebsocketClosedEvent
		err = json.Unmarshal(data, &v)
		e.Event = v

	default:
		var v UnknownEvent
		err = json.Unmarshal(data, &v)
		e.Event = v
	}

	return err
}

var (
	_ Event = (*TrackStartEvent)(nil)
	_ Event = (*TrackEndEvent)(nil)
	_ Event = (*TrackExceptionEvent)(nil)
	_ Event = (*TrackStuckEvent)(nil)
	_ Event = (*WebsocketClosedEvent)(nil)

	_ TrackEvent = (*TrackStartEvent)(nil)
	_ TrackEvent = (*TrackEndEvent)(nil)
	_ TrackEvent = (*TrackExceptionEvent)(nil)
	_ TrackEvent = (*TrackStuckEvent)(nil)
)

type TrackEvent interface {
	Track() string
}

type TrackStartEvent struct {
	GID          snowflake.ID `json:"guildId"`
	EncodedTrack string       `json:"track"`
}

func (TrackStartEvent) Event() EventType        { return EventTypeTrackStart }
func (TrackStartEvent) Op() OpType              { return OpTypeEvent }
func (e TrackStartEvent) GuildID() snowflake.ID { return e.GID }
func (e TrackStartEvent) Track() string         { return e.EncodedTrack }
func (TrackStartEvent) OpEvent()                {}

type TrackEndEvent struct {
	GID          snowflake.ID        `json:"guildId"`
	EncodedTrack string              `json:"track"`
	Reason       AudioTrackEndReason `json:"reason"`
}

func (TrackEndEvent) Event() EventType        { return EventTypeTrackEnd }
func (TrackEndEvent) Op() OpType              { return OpTypeEvent }
func (e TrackEndEvent) GuildID() snowflake.ID { return e.GID }
func (e TrackEndEvent) Track() string         { return e.EncodedTrack }
func (TrackEndEvent) OpEvent()                {}

type TrackExceptionEvent struct {
	GID          snowflake.ID      `json:"guildId"`
	EncodedTrack string            `json:"track"`
	Exception    FriendlyException `json:"exception"`
}

func (TrackExceptionEvent) Event() EventType        { return EventTypeTrackException }
func (TrackExceptionEvent) Op() OpType              { return OpTypeEvent }
func (e TrackExceptionEvent) GuildID() snowflake.ID { return e.GID }
func (e TrackExceptionEvent) Track() string         { return e.EncodedTrack }
func (TrackExceptionEvent) OpEvent()                {}

type TrackStuckEvent struct {
	GID          snowflake.ID `json:"guildId"`
	EncodedTrack string       `json:"track"`
	ThresholdMs  Duration     `json:"threasholdMs"`
}

func (TrackStuckEvent) Event() EventType        { return EventTypeTrackStuck }
func (TrackStuckEvent) Op() OpType              { return OpTypeEvent }
func (e TrackStuckEvent) GuildID() snowflake.ID { return e.GID }
func (e TrackStuckEvent) Track() string         { return e.EncodedTrack }
func (TrackStuckEvent) OpEvent()                {}

type WebsocketClosedEvent struct {
	GID      snowflake.ID `json:"guildId"`
	Code     int          `json:"code"`
	Reason   string       `json:"reason"`
	ByRemote bool         `json:"byRemote"`
}

func (WebsocketClosedEvent) Event() EventType        { return EventTypeWebSocketClosed }
func (WebsocketClosedEvent) Op() OpType              { return OpTypeEvent }
func (e WebsocketClosedEvent) GuildID() snowflake.ID { return e.GID }
func (WebsocketClosedEvent) OpEvent()                {}

type UnknownEvent struct {
	event   EventType
	guildID snowflake.ID
	Data    []byte
}

func (e *UnknownEvent) UnmarshalJSON(data []byte) error {
	var v struct {
		Event   EventType    `json:"type"`
		GuildID snowflake.ID `json:"guildId"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	e.event = v.Event
	e.guildID = v.GuildID
	e.Data = data
	return nil
}

func (e UnknownEvent) MarshalJSON() ([]byte, error) {
	return e.Data, nil
}

func (e UnknownEvent) Event() EventType      { return e.event }
func (UnknownEvent) Op() OpType              { return OpTypeEvent }
func (e UnknownEvent) GuildID() snowflake.ID { return e.guildID }
func (UnknownEvent) OpEvent()                {}
