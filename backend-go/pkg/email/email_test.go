package email_test

import (
	"testing"

	"backend-go/pkg/email"
)

func TestEmailTemplates(t *testing.T) {
	// A dummy config with no host - SendHTML safely no-ops without panic
	cfg := &email.Config{
		Host: "",
	}

	if err := email.SendOTPEmail(cfg, "test@example.com", "1234"); err != nil {
		t.Fatalf("SendOTPEmail returned error: %v", err)
	}

	if err := email.SendPasswordResetEmail(cfg, "test@example.com", "5678"); err != nil {
		t.Fatalf("SendPasswordResetEmail returned error: %v", err)
	}

	if err := email.SendNewProposalEmail(cfg, "test@example.com", "John", "Alice", "Plumbing", 150.0, "Details"); err != nil {
		t.Fatalf("SendNewProposalEmail returned error: %v", err)
	}

	if err := email.SendProposalApprovedEmail(cfg, "test@example.com", "Alice", "John", "john@example.com", "Plumbing", 150.0, "Tomorrow"); err != nil {
		t.Fatalf("SendProposalApprovedEmail returned error: %v", err)
	}

	if err := email.SendDirectRequestEmail(cfg, "test@example.com", "Alice", "John", "Plumbing", 150.0, "Tomorrow", "Details"); err != nil {
		t.Fatalf("SendDirectRequestEmail returned error: %v", err)
	}

	if err := email.SendBroadcastRequestEmail(cfg, "test@example.com", "Alice", "John", "Plumbing", 5.2, 150.0, "Tomorrow", "Details"); err != nil {
		t.Fatalf("SendBroadcastRequestEmail returned error: %v", err)
	}

	if err := email.SendAppointmentBookedEmail(cfg, "test@example.com", "John", "Alice", "Plumbing", "Tomorrow", "123 Main St"); err != nil {
		t.Fatalf("SendAppointmentBookedEmail returned error: %v", err)
	}

	if err := email.SendAppointmentCompletedEmail(cfg, "test@example.com", "John", "Alice", "Plumbing", 150.0); err != nil {
		t.Fatalf("SendAppointmentCompletedEmail returned error: %v", err)
	}

	if err := email.SendAppointmentCancelledEmail(cfg, "test@example.com", "John", "Alice", "Plumbing", "Rescheduled"); err != nil {
		t.Fatalf("SendAppointmentCancelledEmail returned error: %v", err)
	}

	if err := email.SendPayoutStatusEmail(cfg, "test@example.com", "Alice", 250.0, "COMPLETED", "Bank processed"); err != nil {
		t.Fatalf("SendPayoutStatusEmail returned error: %v", err)
	}

	if err := email.SendDisputeFiledEmail(cfg, "test@example.com", "John", "apt-123", "Not satisfied"); err != nil {
		t.Fatalf("SendDisputeFiledEmail returned error: %v", err)
	}

	if err := email.SendDisputeResolvedEmail(cfg, "test@example.com", "John", "apt-123", "Refund granted", "REFUNDED"); err != nil {
		t.Fatalf("SendDisputeResolvedEmail returned error: %v", err)
	}
}

func TestMaskEmail(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"test@example.com", "t**t@example.com"},
		{"a@b.com", "a@b.com"},
		{"invalid", "invalid"},
	}

	for _, c := range cases {
		masked := email.MaskEmail(c.input)
		if masked != c.expected {
			t.Errorf("MaskEmail(%s) = %s, expected %s", c.input, masked, c.expected)
		}
	}
}
