package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/interfaces"
	"github.com/hr3lxphr6j/bililive-go/src/listeners"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/events"
)

var liveStatusGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "bgo",
	Name:      "live_status",
	Help:      "live status",
}, []string{"live_id", "live_url", "live_host_name", "live_room_name"})

type collector struct{}

func NewCollector() interfaces.Module {
	return new(collector)
}

func (collector) callback(ctx context.Context, eventType events.EventType) events.EventHandler {
	return func(event *events.Event) {
		var value float64
		switch eventType {
		case listeners.ListenStart, listeners.LiveEnd:
			value = 0
		case listeners.ListenStop:
			value = -1
		case listeners.LiveStart:
			value = 1
		}
		l := event.Object.(live.Live)
		var info *live.Info
		obj, err := instance.GetInstance(ctx).Cache.Get(l)
		if err != nil {
			info, err = l.GetInfo()
			if err != nil {
				return
			}
		} else {
			info = obj.(*live.Info)
		}
		liveStatusGauge.WithLabelValues(
			string(l.GetLiveId()),
			l.GetRawUrl(),
			info.HostName,
			info.RoomName,
		).Set(value)
	}
}

func (c collector) registryListener(ctx context.Context, ed events.Dispatcher) {
	for _, evt := range []events.EventType{listeners.ListenStart, listeners.ListenStop, listeners.LiveStart, listeners.LiveEnd} {
		ed.AddEventListener(evt, events.NewEventListener(c.callback(ctx, evt)))
	}
}

func (c *collector) Start(ctx context.Context) error {
	ed := instance.GetInstance(ctx).EventDispatcher.(events.Dispatcher)
	c.registryListener(ctx, ed)
	return nil
}

func (c *collector) Close(_ context.Context) {}
