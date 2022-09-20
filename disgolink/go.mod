module github.com/disgoorg/disgolink/disgolink

go 1.18

replace (
	github.com/disgoorg/disgolink/lavalink => ../lavalink
)

require (
	github.com/disgoorg/disgo v0.13.20
	github.com/disgoorg/disgolink/lavalink v1.7.1
)

require (
	github.com/disgoorg/log v1.2.0 // indirect
	github.com/disgoorg/snowflake/v2 v2.0.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/sasha-s/go-csync v0.0.0-20210812194225-61421b77c44b // indirect
	golang.org/x/exp v0.0.0-20220325121720-054d8573a5d8 // indirect
)
