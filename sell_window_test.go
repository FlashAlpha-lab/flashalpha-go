package flashalpha_test

import (
	"context"
	"encoding/json"
	"testing"

	flashalpha "github.com/FlashAlpha-lab/flashalpha-go"
)

const sellWindowBody = `{"symbol":"SPY","session":"2026-10-16","spot":772.255,"quoted_at":"2026-10-16T13:35:00Z","legs_indicative":false,"windows":[{"expiry":"2026-10-16","days_to_expiry":0,"action":"sell","candidates":[{"structure":"short_put","exit":"expiry","legs":[{"side":"sell","quantity":1,"type":"put","strike":769,"delta":-0.19,"bid":0.39,"ask":0.4}],"mid_credit":0.395}]},{"expiry":"2026-10-23","days_to_expiry":7,"action":"skip"},{"expiry":"2026-10-30","days_to_expiry":14,"action":"skip"},{"action":"unavailable"}],"endpoint_version":"2026.09.30","data_as_of":{"node":"fa5"}}`

func TestSellWindowBindsEveryField(t *testing.T) {
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(sellWindowBody), &body); err != nil {
		t.Fatal(err)
	}
	_, client, cap := newCapturingServer(t, 200, body)

	r, err := client.SellWindow(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("SellWindow: %v", err)
	}
	if cap.Path != "/v1/strategies/sell-window/SPY" || cap.Query != "" {
		t.Errorf("path = %q query = %q", cap.Path, cap.Query)
	}
	if r.Session != "2026-10-16" || r.Spot == nil || *r.Spot != 772.255 || r.LegsIndicative == nil || *r.LegsIndicative {
		t.Errorf("top level = %+v", r)
	}
	if len(r.Windows) != 4 || r.Windows[0].Action != "sell" {
		t.Fatalf("windows = %+v", r.Windows)
	}
	trade := r.Windows[0].Candidates[0]
	if trade.Structure != "short_put" || trade.Exit != "expiry" || trade.MidCredit == nil || *trade.MidCredit != 0.395 {
		t.Errorf("trade = %+v", trade)
	}
	if leg := trade.Legs[0]; leg.Strike != 769 || leg.Delta != -0.19 || leg.Side != "sell" || leg.Type != "put" || leg.Quantity != 1 {
		t.Errorf("leg = %+v", leg)
	}
	if r.Windows[3].Action != "unavailable" || r.Windows[3].Expiry != nil || r.Windows[3].DaysToExpiry != nil {
		t.Errorf("unavailable window = %+v", r.Windows[3])
	}
	if r.EndpointVersion != "2026.09.30" || r.DataAsOf == nil || r.DataAsOf.Node != "fa5" {
		t.Errorf("envelope = %q %+v", r.EndpointVersion, r.DataAsOf)
	}
}

func TestSellWindowSkipDayLeavesOptionalFieldsNil(t *testing.T) {
	_, client, _ := newCapturingServer(t, 200, map[string]interface{}{
		"symbol": "SPY", "session": "2026-11-05",
		"windows": []interface{}{map[string]interface{}{"expiry": "2026-11-05", "days_to_expiry": 0, "action": "skip"}},
	})
	r, err := client.SellWindow(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("SellWindow: %v", err)
	}
	if r.Spot != nil || r.QuotedAt != nil || r.LegsIndicative != nil || r.Windows[0].Candidates != nil {
		t.Errorf("skip day should leave optionals nil: %+v", r)
	}
}

func TestSellWindowBelowRequiredPlan(t *testing.T) {
	_, client, _ := newCapturingServer(t, 403, map[string]interface{}{"error": "tier_restricted", "message": "Requires Basic plan."})
	_, err := client.SellWindow(context.Background(), "SPY")
	if _, ok := err.(*flashalpha.TierRestrictedError); !ok {
		t.Errorf("err = %T %v", err, err)
	}
}
