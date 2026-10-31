package smsforward

import "testing"

func TestRetryPolicyNormalize(t *testing.T) {
	tests := []struct {
		name     string
		in       RetryPolicy
		wantTo   int
		wantMax  int
		wantIntv int
	}{
		{"all zero -> defaults", RetryPolicy{}, 30, 1, 5},
		{"explicit values kept", RetryPolicy{TimeoutMinutes: 10, MaxRetries: 3, IntervalMinutes: 2}, 10, 3, 2},
		{"zero max defaults to one", RetryPolicy{TimeoutMinutes: 15, MaxRetries: 0, IntervalMinutes: 7}, 15, 1, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.Normalize()
			if got.TimeoutMinutes != tt.wantTo || got.MaxRetries != tt.wantMax || got.IntervalMinutes != tt.wantIntv {
				t.Fatalf("Normalize() = %+v, want timeout=%d max=%d interval=%d", got, tt.wantTo, tt.wantMax, tt.wantIntv)
			}
		})
	}
}
