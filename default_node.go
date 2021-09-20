package disgolink

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type defaultNode struct {
	options    *NodeOptions
	lavalink   Lavalink
	conn       *websocket.Conn
	quit       chan interface{}
	status     NodeStatus
	stats      *Stats
	available  bool
	restClient RestClient
}

func (n defaultNode) RestURL() string {
	scheme := "http"
	if n.options.Secure {
		scheme += "s"
	}

	return fmt.Sprintf("%s://%s:%s", scheme, n.options.Host, n.options.Port)
}

func (n *defaultNode) Lavalink() Lavalink {
	return n.lavalink
}

func (n *defaultNode) Options() *NodeOptions {
	return n.options
}

func (n *defaultNode) RestClient() RestClient {
	return n.restClient
}

func (n *defaultNode) Name() string {
	return n.options.Name
}

func (n *defaultNode) Send(d interface{}) error {
	err := n.conn.WriteJSON(d)
	if err != nil {
		return errors.Wrap(err, "error while sending to lavalink websocket")
	}
	return nil
}

func (n *defaultNode) Status() NodeStatus {
	return n.status
}

func (n *defaultNode) Stats() *Stats {
	return n.stats
}

func (n *defaultNode) reconnect(delay time.Duration) {
	go func() {
		time.Sleep(delay)

		if n.Status() == Connecting || n.Status() == Reconnecting {
			n.lavalink.Logger().Error("tried to reconnect gateway while connecting/reconnecting")
			return
		}
		n.lavalink.Logger().Info("reconnecting gateway...")
		if err := n.Open(); err != nil {
			n.lavalink.Logger().Errorf("failed to reconnect gateway: %s", err)
			n.status = Disconnected
			n.reconnect(delay * 2)
		}
	}()
}

func (n *defaultNode) Close() {
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

func (n *defaultNode) listen() {
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
			mt, data, err := n.conn.ReadMessage()
			if err != nil {
				n.lavalink.Logger().Errorf("error while reading from ws. error: %s", err)
				n.Close()
				n.reconnect(1 * time.Second)
				return
			}
			op, err := n.getOp(mt, data)
			if err != nil {
				n.lavalink.Logger().Errorf("error getting op from websocket message: %s", err)
				continue
			}
			switch *op {
			case OpPlayerUpdate:
				n.onPlayerUpdate(data)
			case OpEvent:
				n.onTrackEvent(data)
			case OpStats:
				n.onStatsEvent(data)
			default:
				n.lavalink.Logger().Warnf("unexpected op received: %s", op)
			}
		}
	}
}

func (n *defaultNode) getOp(mt int, data []byte) (*Op, error) {
	if mt != websocket.TextMessage {
		return nil, fmt.Errorf("recieved unexpected mt type: %d", mt)
	}

	var op GenericOp
	if err := json.Unmarshal(data, &op); err != nil {
		return nil, err
	}
	return &op.Op, nil
}

func (n *defaultNode) onPlayerUpdate(data []byte) {
	var playerUpdate PlayerUpdateEvent
	err := json.Unmarshal(data, &playerUpdate)
	if err != nil {
		n.lavalink.Logger().Errorf("error unmarshalling PlayerUpdateEvent: %s", err)
		return
	}
	if player := n.lavalink.ExistingPlayer(playerUpdate.GuildID); player != nil {
		player.PlayerUpdate(playerUpdate.State)
		player.EmitEvent(func(listener PlayerEventListener) {
			listener.OnPlayerUpdate(player, playerUpdate.State)
		})
	}
}

func (n *defaultNode) onTrackEvent(data []byte) {
	var event GenericPlayerEvent
	err := json.Unmarshal(data, &event)
	if err != nil {
		n.lavalink.Logger().Errorf("error unmarshalling GenericPlayerEvent: %s", err)
		return
	}
	p := n.lavalink.ExistingPlayer(event.GuildID)
	if p == nil {
		return
	}

	switch event.Type {
	case WebsocketEventTrackStart:
		var trackStartEvent TrackStartEvent
		if err = json.Unmarshal(data, &trackStartEvent); err != nil {
			n.lavalink.Logger().Errorf("error unmarshalling TrackStartEvent: %s", err)
			return
		}
		track := trackStartEvent.Track()
		p.SetTrack(track)
		p.EmitEvent(func(listener PlayerEventListener) {
			listener.OnTrackStart(p, track)
		})

	case WebsocketEventTrackEnd:
		var trackEndEvent TrackEndEvent
		if err = json.Unmarshal(data, &trackEndEvent); err != nil {
			n.lavalink.Logger().Errorf("error unmarshalling TrackEndEvent: %s", err)
			return
		}
		p.EmitEvent(func(listener PlayerEventListener) {
			listener.OnTrackEnd(p, trackEndEvent.Track(), trackEndEvent.EndReason)
		})

	case WebsocketEventTrackException:
		var trackExceptionEvent TrackExceptionEvent
		if err = json.Unmarshal(data, &trackExceptionEvent); err != nil {
			n.lavalink.Logger().Errorf("error unmarshalling TrackExceptionEvent: %s", err)
			return
		}
		p.EmitEvent(func(listener PlayerEventListener) {
			listener.OnTrackException(p, trackExceptionEvent.Track(), trackExceptionEvent.Exception)
		})

	case WebsocketEventTrackStuck:
		var trackStuckEvent TrackStuckEvent
		if err = json.Unmarshal(data, &trackStuckEvent); err != nil {
			n.lavalink.Logger().Errorf("error unmarshalling TrackStuckEvent: %s", err)
			return
		}
		p.EmitEvent(func(listener PlayerEventListener) {
			listener.OnTrackStuck(p, trackStuckEvent.Track(), trackStuckEvent.ThresholdMs)
		})

	case WebSocketEventClosed:
		var websocketClosedEvent WebSocketClosedEvent
		if err = json.Unmarshal(data, &websocketClosedEvent); err != nil {
			n.lavalink.Logger().Errorf("error unmarshalling WebSocketClosedEvent: %s", err)
			return
		}
		p.EmitEvent(func(listener PlayerEventListener) {
			listener.OnWebSocketClosed(p, websocketClosedEvent.Code, websocketClosedEvent.Reason, websocketClosedEvent.ByRemote)
		})

	default:
		n.lavalink.Logger().Warnf("unexpected event received: %s", string(data))
		return
	}
}

func (n *defaultNode) onStatsEvent(data []byte) {
	var event StatsEvent
	err := json.Unmarshal(data, &event)
	if err != nil {
		n.lavalink.Logger().Errorf("error unmarshalling StatsEvent: %s", err)
		return
	}
	n.stats = event.Stats
}

func (n *defaultNode) Open() error {
	scheme := "ws"
	if n.options.Secure {
		scheme += "s"
	}
	header := http.Header{}
	header.Add("Authorization", n.options.Password)
	header.Add("User-Id", n.lavalink.UserID().String())
	header.Add("Client-Name", n.lavalink.ClientName())
	u := url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%v:%v", n.options.Host, n.options.Port),
	}

	var err error
	n.conn, _, err = websocket.DefaultDialer.Dial(u.String(), header)

	go n.listen()

	return err
}
