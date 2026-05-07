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
