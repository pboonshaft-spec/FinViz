package models

import "time"

type AssetType struct {
	ID                int       `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	DefaultReturn     float64   `json:"defaultReturn" db:"default_return"`
	DefaultVolatility float64   `json:"defaultVolatility" db:"default_volatility"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
}

type Asset struct {
	ID               int        `json:"id" db:"id"`
	UserID           int        `json:"userId" db:"user_id"`
	Name             string     `json:"name" db:"name"`
	TypeID           int        `json:"typeId" db:"type_id"`
	CurrentValue     float64    `json:"currentValue" db:"current_value"`
	CustomReturn     *float64   `json:"customReturn,omitempty" db:"custom_return"`
	CustomVolatility *float64   `json:"customVolatility,omitempty" db:"custom_volatility"`
	PlaidAccountID   *string    `json:"plaidAccountId,omitempty" db:"plaid_account_id"`
	CreatedAt        time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time  `json:"updatedAt" db:"updated_at"`
	AssetType        *AssetType `json:"assetType,omitempty" db:"-"`
}

type CreateAssetRequest struct {
	Name             string   `json:"name"`
	TypeID           int      `json:"typeId"`
	CurrentValue     float64  `json:"currentValue"`
	CustomReturn     *float64 `json:"customReturn,omitempty"`
	CustomVolatility *float64 `json:"customVolatility,omitempty"`
}

type UpdateAssetRequest struct {
	Name             *string  `json:"name,omitempty"`
	TypeID           *int     `json:"typeId,omitempty"`
	CurrentValue     *float64 `json:"currentValue,omitempty"`
	CustomReturn     *float64 `json:"customReturn,omitempty"`
	CustomVolatility *float64 `json:"customVolatility,omitempty"`
}

// GetReturn returns the effective return rate for this asset
func (a *Asset) GetReturn() float64 {
	if a.CustomReturn != nil {
		return *a.CustomReturn
	}
	if a.AssetType != nil {
		return a.AssetType.DefaultReturn
	}
	return 0
}

// GetVolatility returns the effective volatility for this asset
func (a *Asset) GetVolatility() float64 {
	if a.CustomVolatility != nil {
		return *a.CustomVolatility
	}
	if a.AssetType != nil {
		return a.AssetType.DefaultVolatility
	}
	return 0
}
