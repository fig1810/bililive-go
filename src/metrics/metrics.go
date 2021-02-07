package metrics

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/interfaces"
	"github.com/hr3lxphr6j/bililive-go/src/listeners"
	"github.com/hr3lxphr6j/bililive-go/src/live"
)

var liveStatus = prometheus.NewDesc(
	prometheus.BuildFQName("bgo", "", "live_status"),
	"live status",
	[]string{"live_id", "live_url", "live_host_name", "live_room_name", "live_listening"},
	nil,
)

type collector struct {
	inst *instance.Instance
}

func NewCollector(ctx context.Context) interfaces.Module {
	return &collector{
		inst: instance.GetInstance(ctx),
	}
}

func bool2float64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func (c collector) Collect(ch chan<- prometheus.Metric) {
	for id, l := range c.inst.Lives {
		var info *live.Info
		obj, err := c.inst.Cache.Get(l)
		if err != nil {
			info, err = l.GetInfo()
			if err != nil {
				return
			}
		} else {
			info = obj.(*live.Info)
		}
		listening := c.inst.ListenerManager.(listeners.Manager).HasListener(context.Background(), id)
		ch <- prometheus.MustNewConstMetric(
			liveStatus, prometheus.GaugeValue, bool2float64(info.Status),
			string(id), l.GetRawUrl(), info.HostName, info.RoomName, fmt.Sprintf("%v", listening),
		)
	}
}

func (collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- liveStatus
}

func (c *collector) Start(_ context.Context) error {
	return prometheus.Register(c)
}

func (c *collector) Close(_ context.Context) {}
