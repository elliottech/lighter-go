package txtypes

import "testing"

func TestUpdateMarginValidateUSDCAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount int64
		want   error
	}{
		{"valid", OneUSDC, nil},
		{"max", MaxTransferAmount, nil},
		{"zero", 0, ErrTransferAmountTooLow},
		{"negative", -1, ErrTransferAmountTooLow},
		{"min int64", -1 << 63, ErrTransferAmountTooLow},
		{"too high", MaxTransferAmount + 1, ErrTransferAmountTooHigh},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txInfo := &L2UpdateMarginTxInfo{
				AccountIndex: 1,
				MarketIndex:  0,
				USDCAmount:   tt.amount,
				Direction:    AddToIsolatedMargin,
				ExpiredAt:    1,
			}
			if err := txInfo.Validate(); err != tt.want {
				t.Errorf("Validate() = %v, want %v", err, tt.want)
			}
		})
	}
}
