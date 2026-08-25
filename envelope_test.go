package flashalpha

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

const envelopeBody = `{
  "symbol": "SPY",
  "net_gex": 1234.5,
  "endpoint_version": "2026.08.25",
  "data_as_of": {
    "node": "fa2",
    "equity_feed": "2026-08-25T18:48:58.204Z",
    "equity_options_feed": "2026-08-25T18:48:57.900Z",
    "index_feed": null,
    "index_options_feed": null,
    "futures_feed": null,
    "futures_options_feed": null,
    "flow_feed": "2026-08-25T18:48:55.100Z",
    "oi_feed": "2026-08-22T20:00:00.000Z",
    "macro_feed": "2026-08-25T18:45:00.000Z"
  }
}`

// The envelope is embedded rather than repeated on each of the 83 response types.
// That only works if encoding/json flattens the anonymous field, so this exercises
// the unmarshal rather than the declaration.
func TestEnvelopeFlattensThroughEmbedding(t *testing.T) {
	var gex GexResponse
	if err := json.Unmarshal([]byte(envelopeBody), &gex); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if gex.Symbol != "SPY" {
		t.Errorf("symbol = %q, want SPY", gex.Symbol)
	}
	if gex.EndpointVersion != "2026.08.25" {
		t.Errorf("endpoint_version = %q, want 2026.08.25", gex.EndpointVersion)
	}
	if gex.DataAsOf == nil {
		t.Fatal("data_as_of is nil; embedding did not flatten")
	}
	if gex.DataAsOf.Node != "fa2" {
		t.Errorf("node = %q, want fa2", gex.DataAsOf.Node)
	}
	if got := deref(gex.DataAsOf.EquityOptionsFeed); got != "2026-08-25T18:48:57.900Z" {
		t.Errorf("equity_options_feed = %q", got)
	}
}

// A feed the node has never seen is nil, which is not the same as the feed being
// unhealthy. Distinguishing the two is the point of the field, so nil must survive
// as nil rather than collapsing to "".
func TestUnseenFeedsStayNil(t *testing.T) {
	var gex GexResponse
	if err := json.Unmarshal([]byte(envelopeBody), &gex); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for name, got := range map[string]*string{
		"index_feed":           gex.DataAsOf.IndexFeed,
		"futures_feed":         gex.DataAsOf.FuturesFeed,
		"futures_options_feed": gex.DataAsOf.FuturesOptionsFeed,
	} {
		if got != nil {
			t.Errorf("%s = %q, want nil", name, *got)
		}
	}
}

// Settled open interest is published once per session, so it trails the response by
// design. Assert it passes through untouched rather than being normalised toward
// response time, because that trailing gap is the signal a caller checks.
func TestSettledOpenInterestTrailsUnmodified(t *testing.T) {
	var gex GexResponse
	if err := json.Unmarshal([]byte(envelopeBody), &gex); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got := deref(gex.DataAsOf.OiFeed); got != "2026-08-22T20:00:00.000Z" {
		t.Errorf("oi_feed = %q, want the prior session close passed through", got)
	}
}

// Responses predating the envelope must still unmarshal.
func TestPreEnvelopeResponsesStillUnmarshal(t *testing.T) {
	var gex GexResponse
	if err := json.Unmarshal([]byte(`{"symbol":"SPY","net_gex":1.0}`), &gex); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if gex.Symbol != "SPY" {
		t.Errorf("symbol = %q, want SPY", gex.Symbol)
	}
	if gex.DataAsOf != nil {
		t.Error("data_as_of should be nil when absent")
	}
	if gex.EndpointVersion != "" {
		t.Errorf("endpoint_version = %q, want empty", gex.EndpointVersion)
	}
}

// Marshalling must not nest the envelope under a "ResponseEnvelope" key. Embedding
// only preserves the wire shape while the field stays anonymous; naming it would
// silently change every serialized payload.
func TestEnvelopeDoesNotNestOnMarshal(t *testing.T) {
	var gex GexResponse
	if err := json.Unmarshal([]byte(envelopeBody), &gex); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	out, err := json.Marshal(gex)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out), "ResponseEnvelope") {
		t.Errorf("envelope nested instead of flattened: %s", out)
	}
	if !strings.Contains(string(out), `"data_as_of"`) {
		t.Errorf("data_as_of missing from output: %s", out)
	}
}

// Guard the sweep itself: every exported *Response struct must embed the envelope.
// Go cannot enumerate a package's types at runtime, so this parses the source - which
// is the stronger check anyway, since it reads the declarations directly. Trusting
// that one regex touched all 83 files is exactly the assumption worth testing, and a
// type added later would otherwise slip through silently.
func TestEveryResponseTypeEmbedsTheEnvelope(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parse package: %v", err)
	}

	checked := 0
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				spec, ok := n.(*ast.TypeSpec)
				if !ok || !strings.HasSuffix(spec.Name.Name, "Response") {
					return true
				}
				st, ok := spec.Type.(*ast.StructType)
				if !ok {
					return true
				}
				checked++
				for _, f := range st.Fields.List {
					if len(f.Names) != 0 {
						continue // named field, not an embed
					}
					if id, ok := f.Type.(*ast.Ident); ok && id.Name == "ResponseEnvelope" {
						return true
					}
				}
				t.Errorf("%s does not embed ResponseEnvelope", spec.Name.Name)
				return true
			})
		}
	}

	if checked == 0 {
		t.Fatal("parsed no *Response structs; the guard is not actually checking anything")
	}
	t.Logf("checked %d response structs", checked)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
