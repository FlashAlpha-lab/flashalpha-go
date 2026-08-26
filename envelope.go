package flashalpha

import (
	"encoding/json"
	"net/http"
)

// DataAsOf reports when each upstream feed last delivered to the node that served
// the response.
//
// It is present on every successful response. The shape is fixed: every field
// exists on every endpoint, and a field is nil when that node has not received
// anything on that feed since it started.
//
// Spot and options are reported separately because they arrive over different
// pipes and fail independently - an index chain can be current while the index
// level behind it is not, and one timestamp cannot express that.
//
// Read each feed against its OWN cadence rather than against AsOf. OiFeed dated to
// the previous session's close is correct, because settled open interest is
// published once per session: on a Monday the newest figure that exists is
// Friday's. EquityOptionsFeed an hour behind during the regular session is not.
//
// A timestamp evidences that the feed delivered recently. It does not assert that
// every contract in a chain is equally current: an illiquid strike may not have
// quoted for hours while its feed is healthy.
type DataAsOf struct {
	// Node identifies which node answered. Nodes hydrate independently, so their
	// feeds can differ.
	Node string `json:"node"`
	// EquityFeed covers equity and ETF spot quotes. Ticks in seconds during market hours.
	EquityFeed *string `json:"equity_feed"`
	// EquityOptionsFeed covers equity and ETF option quotes. Ticks in seconds during market hours.
	EquityOptionsFeed *string `json:"equity_options_feed"`
	// IndexFeed covers index spot - SPX, RUT, VIX and the other index roots. Ticks in
	// seconds during market hours.
	IndexFeed *string `json:"index_feed"`
	// IndexOptionsFeed covers index option quotes. Ticks in seconds during market hours.
	IndexOptionsFeed *string `json:"index_options_feed"`
	// FuturesFeed covers futures prices. Ticks in seconds during the futures session.
	FuturesFeed *string `json:"futures_feed"`
	// FuturesOptionsFeed covers futures option quotes. Ticks in seconds during the futures session.
	FuturesOptionsFeed *string `json:"futures_options_feed"`
	// FlowFeed covers the classified options and stock trade tape. Ticks in seconds during market hours.
	FlowFeed *string `json:"flow_feed"`
	// OiFeed covers settled open interest, dated to the prior 16:00 ET close.
	// Published once per session, so trailing AsOf by a day - or by three across a
	// weekend - is correct rather than stale.
	OiFeed *string `json:"oi_feed"`
	// MacroFeed covers VIX, VVIX, SKEW, MOVE, SPX and Fear & Greed. Unlike the other
	// feeds this reports its OLDEST component, because these are independent series
	// rather than one pipe.
	MacroFeed *string `json:"macro_feed"`
}

// ResponseEnvelope is embedded in every response type and carries the envelope the
// API returns on all successful responses. Because it is embedded anonymously,
// encoding/json flattens it: the wire shape is unchanged and the fields are
// promoted, so they read as gex.DataAsOf and gex.EndpointVersion.
//
// Endpoints that return a bare JSON array cannot carry an envelope in the body and
// send the same information in the X-Data-As-Of and X-Endpoint-Version response
// headers instead.
type ResponseEnvelope struct {
	// EndpointVersion identifies the deployment that produced this response.
	EndpointVersion string `json:"endpoint_version,omitempty"`
	// DataAsOf is the per-feed freshness of the data behind this response.
	DataAsOf *DataAsOf `json:"data_as_of,omitempty"`
}

// ResponseMeta is the envelope for endpoints that return a bare JSON array.
//
// Those responses have nowhere to put an envelope in the body, so the API sends it in
// the X-Data-As-Of and X-Endpoint-Version headers instead. The *WithMetadata accessors
// return this alongside the decoded body so provenance is reachable there too.
type ResponseMeta struct {
	// EndpointVersion identifies the deployment that produced the response.
	EndpointVersion string
	// DataAsOf is the per-feed freshness, or nil if the header was absent.
	DataAsOf *DataAsOf
}

// parseMeta reads the envelope headers. A malformed or absent X-Data-As-Of leaves
// DataAsOf nil rather than failing the call: provenance is diagnostic, and losing it
// should never turn a good response into an error.
func parseMeta(h http.Header) ResponseMeta {
	m := ResponseMeta{EndpointVersion: h.Get("X-Endpoint-Version")}
	if raw := h.Get("X-Data-As-Of"); raw != "" {
		var d DataAsOf
		if err := json.Unmarshal([]byte(raw), &d); err == nil {
			m.DataAsOf = &d
		}
	}
	return m
}
