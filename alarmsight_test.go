package alarmsight_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/fujiwara/alarmsight"
)

func TestParsePayload(t *testing.T) {
	b, err := os.ReadFile("testdata/payload.json")
	if err != nil {
		t.Fatal(err)
	}
	var p alarmsight.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	alarmName, state, err := alarmsight.ParsePayload(p)
	if err != nil {
		t.Fatal(err)
	}
	if alarmName != "lambda-demo-metric-alarm" {
		t.Errorf("unexpected alarmName: %s", alarmName)
	}
	if state != "ALARM" {
		t.Errorf("unexpected state: %s", state)
	}
}
