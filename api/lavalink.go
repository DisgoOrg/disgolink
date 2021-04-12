package api


type Lavalink interface {
	AddNode(options NodeOptions)
	RemoveNode(name string)
	Link(guildID Snowflake) Link
	ExistingLink(guildID Snowflake) Link
	Links() map[Snowflake]Link
	UserID() Snowflake
	ClientName() string
	Shutdown()
	OnVoiceServerUpdate(voiceServerUpdateEvent *VoiceServer)
	OnVoiceStateUpdate(voiceStateUpdateEvent *VoiceState)
}
