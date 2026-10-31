package sms

import "testing"

// statusReportPDU builds a minimal SMS-STATUS-REPORT PDU (TS 27.005 form:
// absent SMSC octet + TPDU).
func statusReportPDU(mr byte, recipient string, st byte) []byte {
	var pdu []byte
	pdu = append(pdu, 0x02) // MTI=10 (status report)
	pdu = append(pdu, mr)   // TP-MR
	pdu = appendAddress(pdu, recipient)
	pdu = append(pdu, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00) // TP-SCTS
	pdu = append(pdu, 0x00)                                     // TP-DT
	pdu = append(pdu, st)                                       // TP-ST
	return append([]byte{0}, pdu...)
}

func classifyStatus(st byte) (final, delivered bool) {
	switch {
	case st >= 0x20 && st <= 0x2f:
		return true, true
	case st >= 0x30 && st <= 0x7f:
		return true, false
	default:
		return false, false
	}
}

func TestStatusReportDecode(t *testing.T) {
	for _, tc := range []struct {
		name      string
		st        byte
		final     bool
		delivered bool
	}{
		{"delivered", 0x20, true, true},
		{"permanent_failure", 0x41, true, false},
		{"temporary_failure", 0x30, true, false},
		{"intermediate_accepted", 0x00, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := statusReportPDU(0x07, "8613750267346", tc.st)
			var part Part
			if err := part.UnmarshalBinary(raw); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !part.Message.DeliveryReport {
				t.Fatalf("expected DeliveryReport=true, got %v", part.Message.DeliveryReport)
			}
			if part.Message.MessageReference != 7 {
				t.Fatalf("TP-MR = %d, want 7", part.Message.MessageReference)
			}
			if part.Message.Number != "8613750267346" {
				t.Fatalf("TP-RA = %q, want 8613750267346", part.Message.Number)
			}
			if part.Message.DeliveryStatus != tc.st {
				t.Fatalf("TP-ST = %#x, want %#x", part.Message.DeliveryStatus, tc.st)
			}
			final, delivered := classifyStatus(tc.st)
			if final != tc.final || delivered != tc.delivered {
				t.Fatalf("classify(%#x) = (%v,%v), want (%v,%v)", tc.st, final, delivered, tc.final, tc.delivered)
			}
		})
	}
}
