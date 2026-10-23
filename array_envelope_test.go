package flashalpha

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// A bare JSON array is the one body shape that cannot carry an envelope, so the API
// moves it to headers. These tests pin both halves: the rows survive decoding, and the
// envelope is still reachable.
func arrayServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Endpoint-Version", "2026.08.25")
		w.Header().Set("X-Data-As-Of", `{"node":"fa2","equity_feed":"2026-08-25T18:48:58.204Z",`+
			`"equity_options_feed":"2026-08-25T18:48:57.900Z","index_feed":null,`+
			`"index_options_feed":null,"futures_feed":null,"futures_options_feed":null,`+
			`"flow_feed":null,"oi_feed":"2026-08-24T20:00:00.000Z","macro_feed":null}`)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"strike":500,"type":"call"},{"strike":505,"type":"put"}]`))
	}))
}

// Regression: decoding an array through map[string]interface{} dropped it silently and
// produced a single zero-valued quote. Anything less than the real rows means that path
// has come back.
func TestArrayBodyKeepsItsRows(t *testing.T) {
	srv := arrayServer(t)
	defer srv.Close()

	quotes, _, err := NewClientWithURL("k", srv.URL).OptionQuoteWithMetadata(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if len(quotes) != 2 {
		t.Fatalf("got %d quotes, want 2 (1 means the array was silently discarded)", len(quotes))
	}
}

func TestArrayBodyExposesTheEnvelopeFromHeaders(t *testing.T) {
	srv := arrayServer(t)
	defer srv.Close()

	_, meta, err := NewClientWithURL("k", srv.URL).OptionQuoteWithMetadata(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if meta.EndpointVersion != "2026.08.25" {
		t.Errorf("endpoint_version = %q", meta.EndpointVersion)
	}
	if meta.DataAsOf == nil {
		t.Fatal("data_as_of nil; the header was not parsed")
	}
	if meta.DataAsOf.Node != "fa2" {
		t.Errorf("node = %q", meta.DataAsOf.Node)
	}
	if got := deref(meta.DataAsOf.OiFeed); got != "2026-08-24T20:00:00.000Z" {
		t.Errorf("oi_feed = %q", got)
	}
	if meta.DataAsOf.IndexFeed != nil {
		t.Error("unseen feed should stay nil")
	}
}

// Provenance is diagnostic: a malformed header must not turn a good response into an
// error, or a header change upstream would break every array call.
func TestMalformedEnvelopeHeaderDoesNotFailTheCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Data-As-Of", "not json")
		_, _ = w.Write([]byte(`[{"strike":500}]`))
	}))
	defer srv.Close()

	quotes, meta, err := NewClientWithURL("k", srv.URL).OptionQuoteWithMetadata(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("a bad header should not fail the call: %v", err)
	}
	if len(quotes) != 1 {
		t.Errorf("got %d quotes, want 1", len(quotes))
	}
	if meta.DataAsOf != nil {
		t.Error("unparseable header should leave DataAsOf nil")
	}
}

// The filtered call returns a single object rather than an array; both normalise to a list.
func TestSingleObjectShapeStillDecodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"strike":500,"type":"call"}`))
	}))
	defer srv.Close()

	quotes, _, err := NewClientWithURL("k", srv.URL).OptionQuoteWithMetadata(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if len(quotes) != 1 {
		t.Fatalf("got %d quotes, want 1", len(quotes))
	}
}
