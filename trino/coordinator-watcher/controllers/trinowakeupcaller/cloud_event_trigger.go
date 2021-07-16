package trinowakeupcaller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/go-logr/logr"
)

type CloudEvent struct {
	eventType     string
	source        string
	extensionsmap map[string]interface{}
	log           logr.Logger
}

type Heartbeat struct {
	Enventime string `json:"enventime"`
	Label     string `json:"label"`
}

func (ce *CloudEvent) Trigger(sink string) error {

	p, err := cloudevents.NewHTTP(cloudevents.WithTarget(sink))

	if err != nil {
		ce.log.Error(err, "failed to create http", "protocol: ", "sink")
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithUUIDs(), cloudevents.WithTimeNow())
	if err != nil {
		ce.log.Error(err, "failed to create client")
	}

	event := cloudevents.NewEvent("1.0")
	event.SetType(ce.eventType)
	event.SetSource(ce.source) //todo

	for k, v := range ce.extensionsmap {
		event.SetExtension(k, v)

	}

	now := time.Now()
	secs := now.Unix()

	hb := &Heartbeat{
		Enventime: fmt.Sprintf("%d", secs),
		Label:     "Wakeup Trino",
	}

	hb.Enventime = strconv.FormatInt(time.Now().UnixNano(), 10)

	if err := event.SetData(cloudevents.ApplicationJSON, hb); err != nil {
		ce.log.Error(err, "failed to set cloudevents ")
	}

	if res := c.Send(context.Background(), event); !cloudevents.IsACK(res) {
		ce.log.Error(nil, "failed to send", "cloudevent:", res)
	}

	return nil

}
