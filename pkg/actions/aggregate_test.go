package actions

import (
	"testing"
	"time"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/data"
)

func timerData(sec int64, value float64) api.Data {
	ts := time.Unix(sec, 0)
	d := data.New("ev", value, ts)
	return d
}

func expectNil(t *testing.T, res api.Data) {
	if res != nil {
		t.Errorf("Unexpected output %v", res)
	}
}

func expectData(t *testing.T, res api.Data, timestamp int64, val float64) {
	if res == nil {
		t.Fatalf("Missing output")
	}

	ts := time.Unix(0, timestamp*1e6)
	if res.Timestamp() != ts {
		t.Errorf("Unexpected output %v", res)
	}
	if res.Value() != val {
		t.Errorf("Unexpected output %v", res)
	}
}

func TestAggregateMax(t *testing.T) {
	a := NewAggregateAction("max", 60*time.Second)

	// 10 sec -> start of sequence
	expectNil(t, a.Process(timerData(10, 10)))

	// 30 sec
	expectNil(t, a.Process(timerData(30, 30)))

	// 60 sec
	expectNil(t, a.Process(timerData(60, 20)))

	// 70 sec -> end of aggregation period
	expectData(t, a.Process(timerData(70, 10)), 70000, 30)

	// 80 sec
	expectNil(t, a.Process(timerData(80, 80)))

	// 130 sec
	expectData(t, a.Process(timerData(130, 130)), 130000, 130)

	// 240 sec
	expectData(t, a.Process(timerData(240, 240)), 240000, 240)
}
func TestAggregateSum(t *testing.T) {
	a := NewAggregateAction("sum", 60*time.Second)

	// 10 sec -> start of sequence
	expectNil(t, a.Process(timerData(10, 10)))

	// 30 sec
	expectNil(t, a.Process(timerData(30, 30)))

	// 60 sec
	expectNil(t, a.Process(timerData(60, 20)))

	// 70 sec -> end of aggregation period
	expectData(t, a.Process(timerData(70, 10)), 70000, 70)

	// 80 sec
	expectNil(t, a.Process(timerData(80, 80)))

	// 130 sec
	expectData(t, a.Process(timerData(130, 130)), 130000, 210)

	// 240 sec
	expectData(t, a.Process(timerData(240, 240)), 240000, 240)
}

func TestAggregateAvg(t *testing.T) {
	a := NewAggregateAction("avg", 60*time.Second)

	// 10 sec -> start of sequence
	expectNil(t, a.Process(timerData(10, 10)))

	// 10 sec -> start of sequence
	expectNil(t, a.Process(timerData(20, 20)))

	// 70 sec -> end of aggregation period
	val := float64((20-10)*20+(70-20)*10) / (70 - 10)
	expectData(t, a.Process(timerData(70, 10)), 70000, val)

	// 130 sec -> end of aggregation period
	expectData(t, a.Process(timerData(130, 130)), 130000, 130)

	// 240 sec -> after aggregation period
	expectData(t, a.Process(timerData(240, 240)), 240000, 240)
}
