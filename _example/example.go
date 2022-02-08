package main

import (
	"context"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/DisgoOrg/disgolink/lavalink"
	"github.com/DisgoOrg/snowflake"
)

var (
	urlPattern = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
)

func main() {
	bot := &Bot{Link: lavalink.New(lavalink.WithUserID("0123456789"))}
	bot.registerNodes()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

type Bot struct {
	Link lavalink.Lavalink
}

func (b *Bot) messageCreateHandler() {
	command := "!play channelID url"
	channelID := snowflake.Snowflake("")
	guildID := snowflake.Snowflake("")
	args := strings.Split(command, " ")
	if len(args) < 3 {
		// TODO: send error message
		return
	}
	query := strings.Join(args[2:], " ")
	if !urlPattern.MatchString(query) {
		query = "ytsearch:" + query
	}
	_ = b.Link.BestRestClient().LoadItemHandler(context.TODO(), query, lavalink.NewResultHandler(
		func(track lavalink.AudioTrack) {
			b.Play(guildID, snowflake.Snowflake(args[1]), channelID, track)
		},
		func(playlist lavalink.AudioPlaylist) {
			b.Play(guildID, snowflake.Snowflake(args[1]), channelID, playlist.Tracks[0])
		},
		func(tracks []lavalink.AudioTrack) {
			b.Play(guildID, snowflake.Snowflake(args[1]), channelID, tracks[0])
		},
		func() {
			// TODO: send error message
		},
		func(ex lavalink.FriendlyException) {
			// TODO: send error message
		},
	))
}

func (b *Bot) Play(guildID snowflake.Snowflake, voiceChannelID snowflake.Snowflake, channelID snowflake.Snowflake, track lavalink.AudioTrack) {
	// TODO: join voice channel

	if err := b.Link.Player(guildID).Play(track); err != nil {
		// TODO: send error message
		return
	}
	// TODO: send playing message
}

func (b *Bot) registerNodes() {
	secure, _ := strconv.ParseBool(os.Getenv("lavalink_secure"))
	node, _ := b.Link.AddNode(context.TODO(), lavalink.NodeConfig{
		Name:        "test",
		Host:        os.Getenv("lavalink_host"),
		Port:        os.Getenv("lavalink_port"),
		Password:    os.Getenv("lavalink_password"),
		Secure:      secure,
		ResumingKey: os.Getenv("lavalink_resuming_key"),
	})
	version, _ := node.RestClient().Version(context.TODO())
	println("Lavalink Server Version: ", version)
}
