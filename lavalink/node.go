package lavalink

import (
	"context"
	"fmt"
	"github.com/DisgoOrg/disgo/json"
	"github.com/DisgoOrg/disgolink/info"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type Node interface {
	Lavalink() Lavalink
	Send(cmd OpCommand) error

	Open(ctx context.Context) error
	ReOpen(ctx context.Context) error
	Close(ctx context.Context)

	Name() string
	RestClient() RestClient
	RestURL() string
	Config() NodeConfig
	Stats() Stats
}

type NodeConfig struct {
	Name     string
	Host     string
	Port     string
	Password string
	Secure   bool
}

type nodeImpl struct {
	config     NodeConfig
	lavalink   Lavalink
	conn       *websocket.Conn
	quit       chan interface{}
	status     NodeStatus
	stats      Stats
	available  bool
	restClient RestClient
}

func (n *nodeImpl) RestURL() string {
	scheme := "http"
	if n.config.Secure {
		scheme += "s"
	}

	return fmt.Sprintf("%s://%s:%s", scheme, n.config.Host, n.config.Port)
}

func (n *nodeImpl) Lavalink() Lavalink {
	return n.lavalink
}

func (n *nodeImpl) Config() NodeConfig {
	return n.config
}

func (n *nodeImpl) RestClient() RestClient {
	return n.restClient
}

func (n *nodeImpl) Name() string {
	return n.config.Name
}

func (n *nodeImpl) Send(cmd OpCommand) error {
	err := n.conn.WriteJSON(cmd)
	if err != nil {
		return errors.Wrap(err, "error while sending to lavalink websocket")
	}
	return nil
}

func (n *nodeImpl) Status() NodeStatus {
	return n.status
}

func (n *nodeImpl) Stats() Stats {
	return n.stats
}

func (n *nodeImpl) reconnect(ctx context.Context, delay time.Duration) {
	go func() {
		time.Sleep(delay)

		if n.Status() == Connecting || n.Status() == Reconnecting {
			n.lavalink.Logger().Error("tried to reconnect gateway while connecting/reconnecting")
			return
		}
		n.lavalink.Logger().Info("reconnecting gateway...")
		if err := n.Open(ctx); err != nil {
			n.lavalink.Logger().Errorf("failed to reconnect gateway: %s", err)
			n.status = Disconnected
			n.reconnect(ctx, delay*2)
		}
	}()
}

func (n *nodeImpl) listen() {
	defer func() {
		n.lavalink.Logger().Info("shut down listen goroutine")
	}()
	for {
		select {
		case <-n.quit:
			n.lavalink.Logger().Infof("existed listen routine")
			return
		default:
			if n.conn == nil {
				return
			}
			_, data, err := n.conn.ReadMessage()
			if err != nil {
				n.lavalink.Logger().Errorf("error while reading from ws. error: %s", err)
				n.Close(context.TODO())
				n.reconnect(context.TODO(), 1*time.Second)
				return
			}

			var v UnmarshalOp
			if err = json.Unmarshal(data, &v); err != nil {
				n.lavalink.Logger().Errorf("error while unmarshalling op. error: %s", err)
				continue
			}
			switch op := v.Op.(type) {
			case PlayerUpdateOp:
				n.onPlayerUpdate(op)

			case OpEvent:
				n.onEvent(op)

			case StatsOp:
				n.onStatsEvent(op)

			default:
				n.lavalink.Logger().Warnf("unexpected op received: %T, data: ", op, string(data))
			}
		}
	}
}

func (n *nodeImpl) onPlayerUpdate(playerUpdate PlayerUpdateOp) {
	if player := n.lavalink.ExistingPlayer(playerUpdate.GuildID); player != nil {
		player.PlayerUpdate(playerUpdate.State)
		player.EmitEvent(func(listener PlayerEventListener) {
			listener.OnPlayerUpdate(player, playerUpdate.State)
		})
	}
}

func (n *nodeImpl) onEvent(event OpEvent) {
	p := n.lavalink.ExistingPlayer(event.GuildID())
	if p == nil {
		return
	}

	switch e := event.(type) {
	case TrackStartEvent:
		track := NewTrack(e.Track)
		p.SetTrack(track)
		p.EmitEvent(func(listener PlayerEventListener) {
			listener.OnTrackStart(p, track)
		})

	case TrackEndEvent:
		p.EmitEvent(func(listener PlayerEventListener) {
			listener.OnTrackEnd(p, NewTrack(e.Track), e.Reason)
		})

	case TrackExceptionEvent:
		p.EmitEvent(func(listener PlayerEventListener) {
			listener.OnTrackException(p, NewTrack(e.Track), e.Exception)
		})

	case TrackStuckEvent:
		p.EmitEvent(func(listener PlayerEventListener) {
			listener.OnTrackStuck(p, NewTrack(e.Track), e.ThresholdMs)
		})

	case WebsocketClosedEvent:
		p.EmitEvent(func(listener PlayerEventListener) {
			listener.OnWebSocketClosed(p, e.Code, e.Reason, e.ByRemote)
		})

	default:
		n.lavalink.Logger().Warnf("unexpected event received: %T", event)
		return
	}
}

func (n *nodeImpl) onStatsEvent(stats StatsOp) {
	n.stats = stats.Stats
}

func (n *nodeImpl) Open(ctx context.Context) error {
	scheme := "ws"
	if n.config.Secure {
		scheme += "s"
	}
	header := http.Header{}
	header.Add("Authorization", n.config.Password)
	header.Add("User-Id", n.lavalink.UserID().String())
	header.Add("Client-Name", fmt.Sprintf("%s/%s", info.Name, info.Version))

	var err error
	n.conn, _, err = websocket.DefaultDialer.Dial(fmt.Sprintf("%s://%s:%s", scheme, n.config.Host, n.config.Port), header)

	go n.listen()

	return err
}

func (n *nodeImpl) ReOpen(ctx context.Context) error {
	return nil
}

func (n *nodeImpl) Close(ctx context.Context) {
	n.status = Disconnected
	if n.quit != nil {
		n.lavalink.Logger().Info("closing ws goroutines...")
		close(n.quit)
		n.quit = nil
		n.lavalink.Logger().Info("closed ws goroutines")
	}
	if n.conn != nil {
		if err := n.conn.Close(); err != nil {
			n.lavalink.Logger().Errorf("error while closing wsconn: %s", err)
		}
	}
}
