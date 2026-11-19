package flashalpha_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	flashalpha "github.com/FlashAlpha-lab/flashalpha-go"
)

// serveGammaFixture serves a real production response captured on 2026-11-19 (lists
// trimmed) byte for byte, and records the request.
func serveGammaFixture(t *testing.T, status int, name string) (*flashalpha.Client, *capturedRequest) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", "gamma_context", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	captured := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.Path = r.URL.Path
		captured.Query = r.URL.RawQuery
		captured.Header = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return flashalpha.NewClientWithURL("test-key", srv.URL), captured
}

func TestExpiryImpactOmitsUnsetOptions(t *testing.T) {
	client, cap := serveGammaFixture(t, 200, "expiry_impact_live.json")
	if _, err := client.ExpiryImpact(context.Background(), "SPY"); err != nil {
		t.Fatalf("ExpiryImpact: %v", err)
	}
	if cap.Path != "/v1/exposure/expiry-impact/SPY" || cap.Query != "" {
		t.Errorf("path = %q query = %q", cap.Path, cap.Query)
	}
}

func TestExpiryImpactSendsExpiryAndLimit(t *testing.T) {
	client, cap := serveGammaFixture(t, 200, "expiry_impact_live.json")
	_, err := client.ExpiryImpact(context.Background(), "QQQ",
		flashalpha.WithExpiryImpactExpiry("2026-11-20"),
		flashalpha.WithExpiryImpactLimit(3),
	)
	if err != nil {
		t.Fatalf("ExpiryImpact: %v", err)
	}
	if cap.Path != "/v1/exposure/expiry-impact/QQQ" || cap.Query != "expiry=2026-11-20&limit=3" {
		t.Errorf("path = %q query = %q", cap.Path, cap.Query)
	}
}

func TestExpiryImpactBindsEveryLevel(t *testing.T) {
	client, _ := serveGammaFixture(t, 200, "expiry_impact_live.json")
	r, err := client.ExpiryImpact(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("ExpiryImpact: %v", err)
	}

	// Shared metadata, promoted from GammaContextMetadata.
	if r.Symbol != "SPY" || r.UnderlyingPrice != 735.885 || r.AnalysisType != "expiry_removal_sensitivity" {
		t.Errorf("metadata = %q %v %q", r.Symbol, r.UnderlyingPrice, r.AnalysisType)
	}
	if r.DataStatus != "stale" || r.PricingBasis != "captured_cached_gamma_current_spot" || r.EligibleContractCount != 9891 {
		t.Errorf("status = %q %q %d", r.DataStatus, r.PricingBasis, r.EligibleContractCount)
	}
	if r.QuoteAt == nil || *r.QuoteAt != "2026-11-19T01:00:00.26Z" || r.OiAsOf != nil {
		t.Errorf("quote_at = %v oi_as_of = %v", r.QuoteAt, r.OiAsOf)
	}
	if r.ObservedMinStrike == nil || *r.ObservedMinStrike != 50 || r.ObservedMaxStrike == nil || *r.ObservedMaxStrike != 1480 {
		t.Errorf("observed strikes = %v %v", r.ObservedMinStrike, r.ObservedMaxStrike)
	}
	if len(r.Reasons) != 4 || r.Reasons[0].ContractCount == nil || *r.Reasons[0].ContractCount != 600 || r.Reasons[2].ContractCount != nil {
		t.Errorf("reasons = %+v", r.Reasons)
	}

	// Envelope.
	if r.EndpointVersion != "2026.09.30" || r.DataAsOf == nil || r.DataAsOf.Node != "fa4" {
		t.Errorf("envelope = %q %+v", r.EndpointVersion, r.DataAsOf)
	}

	// Selection and totals.
	if r.SelectedExpiry != "2026-11-19" || r.SelectionSource != "nearest_active" || len(r.AvailableExpiries) != 3 {
		t.Errorf("selection = %q %q %v", r.SelectedExpiry, r.SelectionSource, r.AvailableExpiries)
	}
	if r.Current.ContractCount != 9891 || r.Current.GrossGexUsd != 33170572651.203655 || r.Current.Status != nil {
		t.Errorf("current = %+v", r.Current)
	}
	if r.ExcludedExpiry.PutGexUsd != -573935208.9391723 || r.ExcludedExpiry.ContractCount != 153 {
		t.Errorf("excluded_expiry = %+v", r.ExcludedExpiry)
	}
	if r.ExcludingExpiry.Status == nil || *r.ExcludingExpiry.Status != "available" || r.ExcludingExpiry.ContractCount != 9738 {
		t.Errorf("excluding_expiry = %+v", r.ExcludingExpiry)
	}

	// Impact and its nested strike rows.
	im := r.Impact
	if im.RemovedGrossSharePct == nil || *im.RemovedGrossSharePct != 4.374061987278384 || im.NetGexChangeUsd != -303030991.4205142 {
		t.Errorf("impact = %+v", im)
	}
	if im.NetSignBefore != "negative" || im.NetSignAfter != "negative" || im.NetSignChanged == nil || *im.NetSignChanged {
		t.Errorf("impact signs = %q %q %v", im.NetSignBefore, im.NetSignAfter, im.NetSignChanged)
	}
	if len(im.LargestGrossReductions) != 2 || im.LargestGrossReductions[0].Strike != 740 || im.LargestGrossReductions[0].ExcludedGrossGexUsd != 124993987.37306339 {
		t.Errorf("largest_gross_reductions = %+v", im.LargestGrossReductions)
	}
	if s := im.StrongestRemainingConcentrations; len(s) != 2 || s[1].Strike != 730 || s[1].ExcludedShareAtStrikePct == nil || *s[1].ExcludedShareAtStrikePct != 1.126324536749193 {
		t.Errorf("strongest_remaining_concentrations = %+v", s)
	}

	// Integer JSON values bind into float64 fields.
	if len(r.Strikes) != 3 || r.Strikes[0].Strike != 50 || r.Strikes[0].ExcludedGrossGexUsd != 0 || r.Strikes[0].DistanceFromSpotPct != -93.20546009226986 {
		t.Errorf("strikes = %+v", r.Strikes)
	}

	// Facts carry a free-form supporting_values object.
	if len(r.Facts) != 2 || r.Facts[0].Code != "expiry_gross_share" {
		t.Fatalf("facts = %+v", r.Facts)
	}
	sv := r.Facts[0].SupportingValues
	if sv["expiry"] != "2026-11-19" || sv["gross_share_pct"] != 4.374061987278384 {
		t.Errorf("facts[0].supporting_values = %v", sv)
	}
	if r.Facts[1].SupportingValues["remaining_contract_count"] != float64(9738) {
		t.Errorf("facts[1].supporting_values = %v", r.Facts[1].SupportingValues)
	}
}

func TestStrikeDependencySendsExactStrike(t *testing.T) {
	cases := map[float64]string{
		672.5:  "strike=672.5",
		670:    "strike=670",
		1480:   "strike=1480",
		52.25:  "strike=52.25",
		0.0005: "strike=0.0005",
	}
	for strike, want := range cases {
		client, cap := serveGammaFixture(t, 200, "strike_dependency_live.json")
		if _, err := client.StrikeDependency(context.Background(), "SPY", strike); err != nil {
			t.Fatalf("StrikeDependency(%v): %v", strike, err)
		}
		if cap.Path != "/v1/exposure/strike-dependency/SPY" || cap.Query != want {
			t.Errorf("strike %v: path = %q query = %q, want %q", strike, cap.Path, cap.Query, want)
		}
	}
}

func TestStrikeDependencyBindsEveryLevel(t *testing.T) {
	client, _ := serveGammaFixture(t, 200, "strike_dependency_live.json")
	r, err := client.StrikeDependency(context.Background(), "SPY", 670)
	if err != nil {
		t.Fatalf("StrikeDependency: %v", err)
	}

	if r.Strike != 670 || r.DistanceFromSpotPct != -8.953165236416016 || r.ContractCount != 48 {
		t.Errorf("top level = %v %v %d", r.Strike, r.DistanceFromSpotPct, r.ContractCount)
	}
	if r.GrossGexUsd != 187959991.78647596 || r.PutGexUsd != -181746678.9477364 {
		t.Errorf("totals = %v %v", r.GrossGexUsd, r.PutGexUsd)
	}
	if r.DominantExpiry == nil || *r.DominantExpiry != "2026-12-18" || r.DominantExpirySharePct == nil || *r.DominantExpirySharePct != 22.973602773735948 {
		t.Errorf("dominant = %v %v", r.DominantExpiry, r.DominantExpirySharePct)
	}
	if r.AnalysisType != "strike_attribution" || r.SnapshotID != "5eea47dec15e0c187d02cb26c72169df7e20d2d767b306db67ee9fd56323de86" {
		t.Errorf("metadata = %q %q", r.AnalysisType, r.SnapshotID)
	}
	if r.EndpointVersion != "2026.09.30" || r.DataAsOf == nil || r.DataAsOf.Node != "fa4" {
		t.Errorf("envelope = %q %+v", r.EndpointVersion, r.DataAsOf)
	}

	if len(r.ByExpiry) != 3 {
		t.Fatalf("by_expiry = %+v", r.ByExpiry)
	}
	row := r.ByExpiry[0]
	if row.Expiry != "2026-11-19" || row.ValuationCutoffAt != "2026-11-19T21:00:00Z" || row.LastTradingAt != "2026-11-19T21:15:00Z" {
		t.Errorf("by_expiry[0] clocks = %+v", row)
	}
	if row.LastTradingStatus != "scheduled_regular_session" || row.SettlementAt != nil || row.SettlementStatus != "unknown" || row.CalendarDaysToExpiry != 0 {
		t.Errorf("by_expiry[0] status = %+v", row)
	}
	if row.GrossSharePct == nil || *row.GrossSharePct != 0.10318767058863045 || row.ShareOfStrikeGrossPct == nil || *row.ShareOfStrikeGrossPct != 0.10318767058863045 {
		t.Errorf("by_expiry[0] shares = %v %v", row.GrossSharePct, row.ShareOfStrikeGrossPct)
	}
	// GammaTotals fields are promoted onto the row.
	if row.GrossGexUsd != 193951.53716304564 || row.CallGexUsd != 0 || row.ContractCount != 1 || row.Status != nil {
		t.Errorf("by_expiry[0] totals = %+v", row.GammaTotals)
	}
	if r.ByExpiry[2].CalendarDaysToExpiry != 4 {
		t.Errorf("by_expiry[2] = %+v", r.ByExpiry[2])
	}
}

func TestGammaContextBelowGrowthIsTierRestricted(t *testing.T) {
	_, client, _ := newCapturingServer(t, 403, map[string]interface{}{
		"error": "tier_restricted", "message": "Requires Growth plan.", "current_plan": "basic", "required_plan": "growth",
	})
	_, err := client.ExpiryImpact(context.Background(), "SPY")
	var te *flashalpha.TierRestrictedError
	if !errors.As(err, &te) || te.RequiredPlan != "growth" {
		t.Errorf("ExpiryImpact err = %T %v", err, err)
	}
	_, err = client.StrikeDependency(context.Background(), "SPY", 670)
	if !errors.As(err, &te) {
		t.Errorf("StrikeDependency err = %T %v", err, err)
	}
}

func TestGammaContextUnsupportedSymbolSurfaces422(t *testing.T) {
	client, _ := serveGammaFixture(t, 422, "error_unsupported_symbol.json")
	for name, call := range map[string]func() error{
		"ExpiryImpact": func() error { _, err := client.ExpiryImpact(context.Background(), "AAPL"); return err },
		"StrikeDependency": func() error {
			_, err := client.StrikeDependency(context.Background(), "AAPL", 200)
			return err
		},
	} {
		err := call()
		var apiErr *flashalpha.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != 422 {
			t.Fatalf("%s err = %T %v", name, err, err)
		}
		if apiErr.Response["error"] != "unsupported_symbol" || apiErr.Message != "This operation currently supports standard SPY and QQQ options." {
			t.Errorf("%s 422 body = %v", name, apiErr.Response)
		}
	}
}

func TestStrikeDependencyStrikeNotFoundKeepsBody(t *testing.T) {
	client, _ := serveGammaFixture(t, 404, "error_strike_not_found.json")
	_, err := client.StrikeDependency(context.Background(), "SPY", 671.3)
	var nf *flashalpha.NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("err = %T %v", err, err)
	}
	if nf.Response["error"] != "strike_not_found" || nf.Response["requested_strike"] != 671.3 {
		t.Errorf("404 body = %v", nf.Response)
	}
}
