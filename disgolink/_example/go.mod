module github.com/DisgoOrg/disgolink/disgolink/_example

go 1.17

replace (
	github.com/DisgoOrg/disgolink/disgolink => ../
	github.com/DisgoOrg/disgolink/lavalink => ../../lavalink
)

require (
	github.com/DisgoOrg/disgo v0.7.0
	github.com/DisgoOrg/disgolink/disgolink v0.2.5-0.20220112205449-450f387bf713
	github.com/DisgoOrg/disgolink/lavalink v1.1.2
	github.com/DisgoOrg/log v1.1.2
	github.com/DisgoOrg/snowflake v1.0.3
)

require (
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sasha-s/go-csync v0.0.0-20210812194225-61421b77c44b // indirect
)
