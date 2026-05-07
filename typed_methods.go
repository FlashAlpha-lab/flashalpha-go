package flashalpha

// Strongly-typed wrapper methods for the existing map-returning client
// endpoints. Each `*Typed` method delegates to the canonical untyped method
// (preserving identical request semantics) and decodes the response into
// the appropriate typed struct from this package.
//
// The original untyped methods continue to return map[string]interface{}
// unchanged — adding these wrappers is purely additive.

import (
	"context"
	"encoding/json"
	"fmt"
)

// decodeTyped re-encodes a map[string]interface{} response and decodes it
// into the supplied typed-struct pointer. Used by the *Typed wrappers below
// to keep them concise and uniform.
func decodeTyped(label string, raw map[string]interface{}, out interface{}) error {
	buf, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("flashalpha: re-encode %s: %w", label, err)
	}
	if err := json.Unmarshal(buf, out); err != nil {
		return fmt.Errorf("flashalpha: decode %s: %w", label, err)
	}
	return nil
}

// StockSummaryTyped is the strongly-typed variant of StockSummary. The
// original StockSummary continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *StockSummaryResponse for the given symbol.
func (c *Client) StockSummaryTyped(ctx context.Context, symbol string) (*StockSummaryResponse, error) {
	raw, err := c.StockSummary(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &StockSummaryResponse{}
	if err := decodeTyped("stock summary", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// NarrativeTyped is the strongly-typed variant of Narrative. The original
// Narrative continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *NarrativeResponse for the given symbol.
func (c *Client) NarrativeTyped(ctx context.Context, symbol string) (*NarrativeResponse, error) {
	raw, err := c.Narrative(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &NarrativeResponse{}
	if err := decodeTyped("narrative", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// LevelsTyped is the strongly-typed variant of ExposureLevels. The original
// ExposureLevels continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *LevelsResponse for the given symbol.
func (c *Client) LevelsTyped(ctx context.Context, symbol string) (*LevelsResponse, error) {
	raw, err := c.ExposureLevels(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &LevelsResponse{}
	if err := decodeTyped("levels", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GreeksTyped is the strongly-typed variant of Greeks. The original Greeks
// continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *PricingGreeksResponse computed from the supplied
// BSM inputs.
func (c *Client) GreeksTyped(ctx context.Context, p GreeksParams) (*PricingGreeksResponse, error) {
	raw, err := c.Greeks(ctx, p)
	if err != nil {
		return nil, err
	}
	out := &PricingGreeksResponse{}
	if err := decodeTyped("greeks", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// MaxPainTyped is the strongly-typed variant of MaxPain. The original MaxPain
// continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *MaxPainResponse for the given symbol. Pass
// WithMaxPainExpiration to scope the response to a single expiry.
func (c *Client) MaxPainTyped(ctx context.Context, symbol string, opts ...MaxPainOption) (*MaxPainResponse, error) {
	raw, err := c.MaxPain(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &MaxPainResponse{}
	if err := decodeTyped("max pain", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// VrpTyped is the strongly-typed variant of Vrp. Vrp already returns
// *VrpResponse — VrpTyped is provided for API symmetry alongside the other
// *Typed wrappers.
//
// Returns a fully-populated *VrpResponse for the given symbol.
func (c *Client) VrpTyped(ctx context.Context, symbol string) (*VrpResponse, error) {
	return c.Vrp(ctx, symbol)
}

// ExposureSummaryTyped is the strongly-typed variant of ExposureSummary. The
// original ExposureSummary continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *ExposureSummaryResponse for the given symbol.
func (c *Client) ExposureSummaryTyped(ctx context.Context, symbol string) (*ExposureSummaryResponse, error) {
	raw, err := c.ExposureSummary(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &ExposureSummaryResponse{}
	if err := decodeTyped("exposure summary", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// VolatilityTyped is the strongly-typed variant of Volatility. The original
// Volatility continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *VolatilityResponse for the given symbol.
func (c *Client) VolatilityTyped(ctx context.Context, symbol string) (*VolatilityResponse, error) {
	raw, err := c.Volatility(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &VolatilityResponse{}
	if err := decodeTyped("volatility", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// AdvVolatilityTyped is the strongly-typed variant of AdvVolatility. The
// original AdvVolatility continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *AdvVolatilityResponse for the given symbol.
func (c *Client) AdvVolatilityTyped(ctx context.Context, symbol string) (*AdvVolatilityResponse, error) {
	raw, err := c.AdvVolatility(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &AdvVolatilityResponse{}
	if err := decodeTyped("adv volatility", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SurfaceTyped is the strongly-typed variant of Surface. The original
// Surface continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *SurfaceResponse for the given symbol.
func (c *Client) SurfaceTyped(ctx context.Context, symbol string) (*SurfaceResponse, error) {
	raw, err := c.Surface(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &SurfaceResponse{}
	if err := decodeTyped("surface", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GexTyped is the strongly-typed variant of Gex. The original Gex continues
// to return map[string]interface{} unchanged.
//
// Returns a fully-populated *GexResponse for the given symbol.
func (c *Client) GexTyped(ctx context.Context, symbol string, opts ...GexOption) (*GexResponse, error) {
	raw, err := c.Gex(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &GexResponse{}
	if err := decodeTyped("gex", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DexTyped is the strongly-typed variant of Dex. The original Dex continues
// to return map[string]interface{} unchanged.
//
// Returns a fully-populated *DexResponse for the given symbol.
func (c *Client) DexTyped(ctx context.Context, symbol string, opts ...DexOption) (*DexResponse, error) {
	raw, err := c.Dex(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &DexResponse{}
	if err := decodeTyped("dex", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// VexTyped is the strongly-typed variant of Vex. The original Vex continues
// to return map[string]interface{} unchanged.
//
// Returns a fully-populated *VexResponse for the given symbol.
func (c *Client) VexTyped(ctx context.Context, symbol string, opts ...VexOption) (*VexResponse, error) {
	raw, err := c.Vex(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &VexResponse{}
	if err := decodeTyped("vex", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ChexTyped is the strongly-typed variant of Chex. The original Chex
// continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *ChexResponse for the given symbol.
func (c *Client) ChexTyped(ctx context.Context, symbol string, opts ...ChexOption) (*ChexResponse, error) {
	raw, err := c.Chex(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &ChexResponse{}
	if err := decodeTyped("chex", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// OptionQuoteTyped is the strongly-typed variant of OptionQuote. The
// original OptionQuote continues to return map[string]interface{} unchanged.
//
// The endpoint returns either a single object (when expiry, strike, and
// type are all supplied) or an array of quotes. To preserve a single typed
// shape, this wrapper always returns a slice of OptionQuote — single-object
// responses are wrapped into a 1-element slice. Inspect the underlying
// shape via OptionQuote(...) when you need to distinguish.
func (c *Client) OptionQuoteTyped(ctx context.Context, ticker string, opts ...OptionQuoteOption) ([]OptionQuote, error) {
	raw, err := c.OptionQuote(ctx, ticker, opts...)
	if err != nil {
		return nil, err
	}
	// The /optionquote endpoint may return either a list or a single
	// object. The map[string]interface{} return type from the unwrapped
	// method already discards the outer-array shape, so re-marshal the
	// raw payload and try both shapes.
	if quotes, ok := raw["quotes"]; ok {
		// Shape: { "quotes": [ ... ] } — some servers wrap.
		buf, mErr := json.Marshal(quotes)
		if mErr != nil {
			return nil, fmt.Errorf("flashalpha: re-encode option quote: %w", mErr)
		}
		var out []OptionQuote
		if err := json.Unmarshal(buf, &out); err != nil {
			return nil, fmt.Errorf("flashalpha: decode option quote: %w", err)
		}
		return out, nil
	}
	// Single-object shape — decode as one OptionQuote and wrap.
	single := OptionQuote{}
	if err := decodeTyped("option quote", raw, &single); err != nil {
		return nil, err
	}
	return []OptionQuote{single}, nil
}

// StockQuoteTyped is the strongly-typed variant of StockQuote. The original
// StockQuote continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *StockQuoteResponse for the given ticker.
func (c *Client) StockQuoteTyped(ctx context.Context, ticker string) (*StockQuoteResponse, error) {
	raw, err := c.StockQuote(ctx, ticker)
	if err != nil {
		return nil, err
	}
	out := &StockQuoteResponse{}
	if err := decodeTyped("stock quote", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// IvTyped is the strongly-typed variant of IV. The original IV continues to
// return map[string]interface{} unchanged.
//
// Returns a fully-populated *PricingIvResponse computed from the supplied
// inputs.
func (c *Client) IvTyped(ctx context.Context, p IVParams) (*PricingIvResponse, error) {
	raw, err := c.IV(ctx, p)
	if err != nil {
		return nil, err
	}
	out := &PricingIvResponse{}
	if err := decodeTyped("iv", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// KellyTyped is the strongly-typed variant of Kelly. The original Kelly
// continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *PricingKellyResponse computed from the supplied
// inputs. On negative-EV setups the Sizing fields are all zero and
// Recommendation explains the no-position outcome.
func (c *Client) KellyTyped(ctx context.Context, p KellyParams) (*PricingKellyResponse, error) {
	raw, err := c.Kelly(ctx, p)
	if err != nil {
		return nil, err
	}
	out := &PricingKellyResponse{}
	if err := decodeTyped("kelly", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// AccountTyped is the strongly-typed variant of Account. The original
// Account continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *AccountResponse for the authenticated user.
func (c *Client) AccountTyped(ctx context.Context) (*AccountResponse, error) {
	raw, err := c.Account(ctx)
	if err != nil {
		return nil, err
	}
	out := &AccountResponse{}
	if err := decodeTyped("account", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// TickersTyped is the strongly-typed variant of Tickers. The original
// Tickers continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *TickersResponse with the alphabetised ticker
// list and total count.
func (c *Client) TickersTyped(ctx context.Context) (*TickersResponse, error) {
	raw, err := c.Tickers(ctx)
	if err != nil {
		return nil, err
	}
	out := &TickersResponse{}
	if err := decodeTyped("tickers", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SymbolsTyped is the strongly-typed variant of Symbols. The original
// Symbols continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *SymbolsResponse listing the live in-store
// symbols plus a refresh timestamp.
func (c *Client) SymbolsTyped(ctx context.Context) (*SymbolsResponse, error) {
	raw, err := c.Symbols(ctx)
	if err != nil {
		return nil, err
	}
	out := &SymbolsResponse{}
	if err := decodeTyped("symbols", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// OptionsTyped is the strongly-typed variant of Options. The original
// Options continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *OptionsResponse with the listed expirations
// and per-expiry strike grids for the given ticker.
func (c *Client) OptionsTyped(ctx context.Context, ticker string) (*OptionsResponse, error) {
	raw, err := c.Options(ctx, ticker)
	if err != nil {
		return nil, err
	}
	out := &OptionsResponse{}
	if err := decodeTyped("options", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// HealthTyped is the strongly-typed variant of Health. The original Health
// continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *HealthResponse — Status is the literal "ok" on
// healthy responses.
func (c *Client) HealthTyped(ctx context.Context) (*HealthResponse, error) {
	raw, err := c.Health(ctx)
	if err != nil {
		return nil, err
	}
	out := &HealthResponse{}
	if err := decodeTyped("health", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ScreenerTyped is the strongly-typed variant of Screener. The original
// Screener continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *ScreenerResponse — Meta has the pagination /
// universe / tier / as-of metadata; Data is the result rows typed as
// map[string]any (row columns depend on the request's `select` argument and
// any defined `formulas`).
func (c *Client) ScreenerTyped(ctx context.Context, req ScreenerRequest) (*ScreenerResponse, error) {
	raw, err := c.Screener(ctx, req)
	if err != nil {
		return nil, err
	}
	out := &ScreenerResponse{}
	if err := decodeTyped("screener", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}
