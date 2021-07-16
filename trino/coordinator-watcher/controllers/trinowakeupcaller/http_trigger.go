package trinowakeupcaller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
)

type Httptrigger struct {
	Log logr.Logger
}

func (ht *Httptrigger) Trigger(sink string) error {
	_, err := http.Get(sink)
	if err != nil {
		ht.Log.Error(err, "sink is unreachable")
		return err
	}
	now := time.Now()
	secs := now.Unix()
	ht.Log.Info(fmt.Sprintf("send wakeup message to Trino %s %d", sink, secs))
	return nil

}
