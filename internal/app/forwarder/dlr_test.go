package forwarder

import "testing"

func TestClassifyDeliveryStatus(t *testing.T) {
	cases := []struct {
		st        uint8
		final     bool
		delivered bool
	}{
		{0x00, false, false}, // intermediate
		{0x1f, false, false}, // intermediate
		{0x20, true, true},   // delivered to MS
		{0x2f, true, true},   // delivered
		{0x30, true, false},  // temporary failure
		{0x41, true, false},  // permanent failure
		{0x7f, true, false},  // error
	}
	for _, c := range cases {
		final, delivered := classifyDeliveryStatus(c.st)
		if final != c.final || delivered != c.delivered {
			t.Errorf("classifyDeliveryStatus(%#x) = (%v,%v), want (%v,%v)", c.st, final, delivered, c.final, c.delivered)
		}
	}
}

func TestCorrelationNumber(t *testing.T) {
	cases := map[string]string{
		"+8613750267346": "8613750267346",
		"8613750267346":  "8613750267346",
		"13812345678":    "13812345678",
	}
	for in, want := range cases {
		if got := correlationNumber(in); got != want {
			t.Errorf("correlationNumber(%q) = %q, want %q", in, got, want)
		}
	}
}
