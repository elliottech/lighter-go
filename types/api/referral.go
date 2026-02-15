package api

// ReferralPoints represents referral program points
type ReferralPoints struct {
	BaseResponse
	AccountIndex       int64  `json:"account_index"`
	ReferralCode       string `json:"referral_code"`
	ReferredBy         int64  `json:"referred_by,omitempty"`
	TotalPoints        string `json:"total_points"`
	AvailablePoints    string `json:"available_points"`
	ClaimedPoints      string `json:"claimed_points"`
	DirectReferrals    int64  `json:"direct_referrals"`
	IndirectReferrals  int64  `json:"indirect_referrals"`
	TotalVolume        string `json:"total_volume"`
	KickbackPercentage string `json:"kickback_percentage"`
	TierLevel          int    `json:"tier_level"`
}

// ReferralStats represents referral statistics
type ReferralStats struct {
	TotalReferrals      int64  `json:"total_referrals"`
	ActiveReferrals     int64  `json:"active_referrals"`
	TotalEarnings       string `json:"total_earnings"`
	PendingEarnings     string `json:"pending_earnings"`
	TotalReferralVolume string `json:"total_referral_volume"`
}

// Referral represents a single referral
type Referral struct {
	ReferredAccount int64  `json:"referred_account"`
	ReferredAt      int64  `json:"referred_at"`
	TotalVolume     string `json:"total_volume"`
	TotalEarnings   string `json:"total_earnings"`
	IsActive        bool   `json:"is_active"`
}

// ReferralList is the response for referral list query
type ReferralList struct {
	BaseResponse
	Referrals []Referral `json:"referrals"`
	Cursor    Cursor     `json:"cursor,omitempty"`
}

// RespUpdateReferralCode is the response for updating referral code
type RespUpdateReferralCode struct {
	BaseResponse
	AccountIndex    int64  `json:"account_index"`
	NewReferralCode string `json:"new_referral_code"`
}

// RespUpdateKickback is the response for updating kickback percentage
type RespUpdateKickback struct {
	BaseResponse
	AccountIndex       int64  `json:"account_index"`
	KickbackPercentage string `json:"kickback_percentage"`
}

// ReferralTier represents a referral tier
type ReferralTier struct {
	Level              int    `json:"level"`
	Name               string `json:"name"`
	MinVolume          string `json:"min_volume"`
	CommissionRate     string `json:"commission_rate"`
	MaxKickbackRate    string `json:"max_kickback_rate"`
}

// ReferralTiers is the response for referral tiers query
type ReferralTiers struct {
	BaseResponse
	Tiers []ReferralTier `json:"tiers"`
}
