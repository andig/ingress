package actions

import (
	"testing"
	"time"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/data"
)

func timerData(sec int64, value float64) api.Data {
	tm := time.Unix(sec, 0)
	d := data.NewData("ev", value)
	d.SetTimestamp(tm.UnixNano() / 1e6)
	return d
}

func expect(t *testing.T, res api.Data, timestamp int64, val float64) {
	if res == nil {
		t.Fatalf("Missing output")
	}

	if res.GetTimestamp() != timestamp {
		t.Errorf("Unexpected output %v", res)
	}
	if res.GetValue() != val {
		t.Errorf("Unexpected output %v", res)
	}
}

func TestAggregateMax(t *testing.T) {
	var res api.Data
	a := NewAggregateAction("max", 60*time.Second)

	// 10 sec -> start of sequence
	if a.Process(timerData(10, 10)) != nil {
		t.Errorf("Unexpected output %v", res)
	}

	// 30 sec
	if a.Process(timerData(30, 30)) != nil {
		t.Errorf("Unexpected output %v", res)
	}

	// 60 sec
	if a.Process(timerData(60, 20)) != nil {
		t.Errorf("Unexpected output %v", res)
	}

	// 70 sec -> end of aggregation period
	expect(t, a.Process(timerData(70, 10)), 70000, 30)

	// 80 sec
	if a.Process(timerData(80, 80)) != nil {
		t.Errorf("Unexpected output %v", res)
	}

	// 130 sec
	expect(t, a.Process(timerData(130, 130)), 130000, 130)

	// 240 sec
	expect(t, a.Process(timerData(240, 240)), 240000, 240)
}
func TestAggregateSum(t *testing.T) {
	var res api.Data
	a := NewAggregateAction("sum", 60*time.Second)

	// 10 sec -> start of sequence
	if a.Process(timerData(10, 10)) != nil {
		t.Errorf("Unexpected output %v", res)
	}

	// 30 sec
	if a.Process(timerData(30, 30)) != nil {
		t.Errorf("Unexpected output %v", res)
	}

	// 60 sec
	if a.Process(timerData(60, 20)) != nil {
		t.Errorf("Unexpected output %v", res)
	}

	// 70 sec -> end of aggregation period
	expect(t, a.Process(timerData(70, 10)), 70000, 70)

	// 80 sec
	if a.Process(timerData(80, 80)) != nil {
		t.Errorf("Unexpected output %v", res)
	}

	// 130 sec
	expect(t, a.Process(timerData(130, 130)), 130000, 210)

	// 240 sec
	expect(t, a.Process(timerData(240, 240)), 240000, 240)
}

func TestAggregateAvg(t *testing.T) {
	var res api.Data
	a := NewAggregateAction("avg", 60*time.Second)

	// 10 sec -> start of sequence
	if a.Process(timerData(10, 10)) != nil {
		t.Errorf("Unexpected output %v", res)
	}

	// 10 sec -> start of sequence
	if a.Process(timerData(20, 20)) != nil {
		t.Errorf("Unexpected output %v", res)
	}

	// 70 sec -> end of aggregation period
	val := float64((20-10)*20+(70-20)*10) / (70 - 10)
	expect(t, a.Process(timerData(70, 10)), 70000, val)

	// 130 sec -> end of aggregation period
	expect(t, a.Process(timerData(130, 130)), 130000, 130)
}
