package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// ─── OTP & AUTHENTICATION EMAILS ─────────────────────────────────────────────

func SendOTPEmail(cfg *Config, toEmail, otp string) error {
	subject := "Verify your email address - Neighbor Service"
	plainText := fmt.Sprintf("Welcome to Neighbor Service\n\nThank you for registering. Please use the verification code below to verify your account:\n\n%s\n\nThis code will expire in 10 minutes. If you did not request this verification, please disregard this email.\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", otp)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Email Verification</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .header { text-align: center; margin-bottom: 24px; }
        .header h2 { color: #0f172a; margin: 0; font-size: 22px; font-weight: 700; }
        .otp-code { font-size: 36px; font-weight: 800; letter-spacing: 8px; color: #2563eb; text-align: center; padding: 18px; background-color: #eff6ff; border-radius: 8px; margin: 24px 0; border: 1px dashed #93c5fd; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Welcome to Neighbor Service</h2>
        </div>
        <p>Hello,</p>
        <p>Thank you for registering. Please use the 4-digit verification code below to verify your account:</p>
        <div class="otp-code">%s</div>
        <p style="color: #64748b; font-size: 14px;">This code will expire in 10 minutes. If you did not request this verification, please disregard this email.</p>
        <div class="footer">
            <p>You received this transactional email because of activity on your Neighbor Service account.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, otp)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

func SendPasswordResetEmail(cfg *Config, toEmail, otp string) error {
	subject := "Reset your password - Neighbor Service"
	plainText := fmt.Sprintf("Password Reset Request\n\nWe received a request to reset your password. Use the verification code below to set a new password:\n\n%s\n\nThis code will expire in 10 minutes. If you did not request a password reset, your account is safe and you can safely ignore this email.\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", otp)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .header { text-align: center; margin-bottom: 24px; }
        .header h2 { color: #0f172a; margin: 0; font-size: 22px; font-weight: 700; }
        .otp-code { font-size: 36px; font-weight: 800; letter-spacing: 8px; color: #dc2626; text-align: center; padding: 18px; background-color: #fef2f2; border-radius: 8px; margin: 24px 0; border: 1px dashed #fca5a5; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Password Reset Request</h2>
        </div>
        <p>Hello,</p>
        <p>We received a request to reset your password. Use the verification code below to set a new password:</p>
        <div class="otp-code">%s</div>
        <p style="color: #64748b; font-size: 14px;">This code will expire in 10 minutes. If you did not request a password reset, your account is safe and you can safely ignore this email.</p>
        <div class="footer">
            <p>You received this transactional email because of activity on your Neighbor Service account.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, otp)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── PROPOSALS & REQUESTS EMAILS ─────────────────────────────────────────────

func SendNewProposalEmail(cfg *Config, toEmail, seekerName, providerName, requestTitle string, price float64, description string) error {
	subject := fmt.Sprintf("New Proposal Received for: %s", requestTitle)
	plainText := fmt.Sprintf("Hello %s,\n\n%s has submitted a proposal for your request \"%s\".\n\nProposed Amount: $%.2f\nCover Letter: %s\n\nOpen your Neighbor Service app to view the provider's profile, chat, and accept the proposal.\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", seekerName, providerName, requestTitle, price, description)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Proposal Received</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .badge { display: inline-block; padding: 4px 10px; background: #e0f2fe; color: #0284c7; border-radius: 4px; font-weight: 600; font-size: 13px; }
        .card { background: #f8fafc; border-radius: 6px; padding: 16px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .price { font-size: 24px; font-weight: 700; color: #16a34a; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">New Proposal</span>
        <h2 style="margin-top: 12px; color: #0f172a;">Hello %s,</h2>
        <p><strong>%s</strong> has submitted a proposal for your request <strong>"%s"</strong>.</p>
        <div class="card">
            <p style="margin:0 0 8px 0; font-size:14px; color:#64748b;">Proposed Amount:</p>
            <div class="price">$%.2f</div>
            <p style="margin:12px 0 0 0; font-size:14px;"><strong>Cover Letter:</strong> %s</p>
        </div>
        <p>Open your Neighbor Service app to view the provider's profile, chat, and accept the proposal.</p>
        <div class="footer">
            <p>You received this transactional email because of activity on your Neighbor Service account.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, seekerName, providerName, requestTitle, price, description)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

func SendProposalApprovedEmail(cfg *Config, toEmail, providerName, seekerName, seekerEmail, requestTitle string, price float64, scheduledTime string) error {
	subject := fmt.Sprintf("Proposal Accepted: %s", requestTitle)
	plainText := fmt.Sprintf("Congratulations %s!\n\nYour proposal for \"%s\" was accepted by %s.\n\nAgreed Price: $%.2f\nClient Contact: %s\nScheduled Time: %s\n\nAn appointment has been scheduled. Please arrive on time and ask the client for their arrival code when you arrive.\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", providerName, requestTitle, seekerName, price, seekerEmail, scheduledTime)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Proposal Accepted</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .badge { display: inline-block; padding: 4px 10px; background: #dcfce7; color: #16a34a; border-radius: 4px; font-weight: 600; font-size: 13px; }
        .card { background: #f8fafc; border-radius: 6px; padding: 16px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">Proposal Accepted</span>
        <h2 style="margin-top: 12px; color: #0f172a;">Congratulations %s!</h2>
        <p>Your proposal for <strong>"%s"</strong> was accepted by <strong>%s</strong>.</p>
        <div class="card">
            <p style="margin:0 0 6px 0;"><strong>Agreed Price:</strong> $%.2f</p>
            <p style="margin:0 0 6px 0;"><strong>Client Contact:</strong> %s</p>
            <p style="margin:0;"><strong>Scheduled Time:</strong> %s</p>
        </div>
        <p>An appointment has been scheduled. Please arrive on time and ask the client for their 4-digit arrival code when you arrive.</p>
        <div class="footer">
            <p>You received this transactional email because of activity on your Neighbor Service account.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, providerName, requestTitle, seekerName, price, seekerEmail, scheduledTime)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

func SendDirectRequestEmail(cfg *Config, toEmail, providerName, seekerName, requestTitle string, price float64, scheduledTime, description string) error {
	subject := fmt.Sprintf("Direct Service Request: %s", requestTitle)
	plainText := fmt.Sprintf("Hello %s,\n\n%s has sent a direct request exclusively to you for \"%s\".\n\nBudget: $%.2f\nScheduled For: %s\nDetails: %s\n\nOpen Neighbor Service to accept or submit a customized proposal.\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", providerName, seekerName, requestTitle, price, scheduledTime, description)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Direct Service Request</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .badge { display: inline-block; padding: 4px 10px; background: #fef3c7; color: #d97706; border-radius: 4px; font-weight: 600; font-size: 13px; }
        .card { background: #f8fafc; border-radius: 6px; padding: 16px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">Direct Request</span>
        <h2 style="margin-top: 12px; color: #0f172a;">Hello %s,</h2>
        <p><strong>%s</strong> has sent a direct request exclusively to you for <strong>"%s"</strong>.</p>
        <div class="card">
            <p style="margin:0 0 6px 0;"><strong>Budget:</strong> $%.2f</p>
            <p style="margin:0 0 6px 0;"><strong>Scheduled For:</strong> %s</p>
            <p style="margin:0;"><strong>Details:</strong> %s</p>
        </div>
        <p>Open Neighbor Service to accept or submit a customized proposal.</p>
        <div class="footer">
            <p>You received this transactional email because of activity on your Neighbor Service account.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, providerName, seekerName, requestTitle, price, scheduledTime, description)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

func SendBroadcastRequestEmail(cfg *Config, toEmail, providerName, seekerName, requestTitle string, distance float64, price float64, scheduledTime, description string) error {
	subject := fmt.Sprintf("New Service Request Nearby: %s", requestTitle)
	plainText := fmt.Sprintf("Hello %s,\n\nA new service request for \"%s\" was posted %.1f km away from your location.\n\nCustomer: %s\nBudget: $%.2f\nScheduled For: %s\nDetails: %s\n\nOpen Neighbor Service to submit your proposal before another provider is selected!\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", providerName, requestTitle, distance, seekerName, price, scheduledTime, description)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nearby Service Request</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .badge { display: inline-block; padding: 4px 10px; background: #e0e7ff; color: #4338ca; border-radius: 4px; font-weight: 600; font-size: 13px; }
        .card { background: #f8fafc; border-radius: 6px; padding: 16px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">Nearby Request</span>
        <h2 style="margin-top: 12px; color: #0f172a;">Hello %s,</h2>
        <p>A new service request for <strong>"%s"</strong> was posted <strong>%.1f km</strong> away from your location.</p>
        <div class="card">
            <p style="margin:0 0 6px 0;"><strong>Customer:</strong> %s</p>
            <p style="margin:0 0 6px 0;"><strong>Budget:</strong> $%.2f</p>
            <p style="margin:0 0 6px 0;"><strong>Scheduled For:</strong> %s</p>
            <p style="margin:0;"><strong>Details:</strong> %s</p>
        </div>
        <p>Open Neighbor Service to submit your proposal before another provider is selected!</p>
        <div class="footer">
            <p>You received this transactional email because of activity on your Neighbor Service account.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, providerName, requestTitle, distance, seekerName, price, scheduledTime, description)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── APPOINTMENT LIFECYCLE EMAILS ───────────────────────────────────────────

func SendAppointmentBookedEmail(cfg *Config, toEmail, recipientName, otherPartyName, serviceTitle, scheduledTime, address string) error {
	subject := fmt.Sprintf("Appointment Scheduled: %s", serviceTitle)
	locationLine := ""
	if address != "" {
		locationLine = fmt.Sprintf("Location: %s\n", address)
	}
	plainText := fmt.Sprintf("Appointment Confirmed\n\nHello %s,\n\nYour appointment for \"%s\" with %s is confirmed.\n\nTime: %s\n%s\nYou can track the appointment status and communicate directly in the app.\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", recipientName, serviceTitle, otherPartyName, scheduledTime, locationLine)
	
	locHTML := ""
	if address != "" {
		locHTML = fmt.Sprintf("<p style=\"margin:0;\"><strong>Location:</strong> %s</p>", address)
	}
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Appointment Confirmed</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .card { background: #f8fafc; border-radius: 6px; padding: 16px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <h2 style="margin: 0 0 16px 0; color: #0f172a;">Appointment Confirmed</h2>
        <p>Hello %s,</p>
        <p>Your appointment for <strong>"%s"</strong> with <strong>%s</strong> is confirmed.</p>
        <div class="card">
            <p style="margin:0 0 6px 0;"><strong>Time:</strong> %s</p>
            %s
        </div>
        <p>You can track the appointment status and communicate directly in the app.</p>
        <div class="footer">
            <p>You received this transactional email because of activity on your Neighbor Service account.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, recipientName, serviceTitle, otherPartyName, scheduledTime, locHTML)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

func SendAppointmentCompletedEmail(cfg *Config, toEmail, recipientName, otherPartyName, serviceTitle string, totalAmount float64) error {
	subject := fmt.Sprintf("Service Completed: %s", serviceTitle)
	plainText := fmt.Sprintf("Service Completed\n\nHello %s,\n\nThe appointment for \"%s\" has been marked completed.\n\nOther Party: %s\nTotal Amount: $%.2f\n\nFunds held in escrow have been released to the provider. Don't forget to leave a review!\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", recipientName, serviceTitle, otherPartyName, totalAmount)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Service Completed</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .card { background: #f0fdf4; border-radius: 6px; padding: 16px; margin: 20px 0; border: 1px solid #bbf7d0; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <h2 style="margin: 0 0 16px 0; color: #15803d;">Service Completed</h2>
        <p>Hello %s,</p>
        <p>The appointment for <strong>"%s"</strong> has been marked completed.</p>
        <div class="card">
            <p style="margin:0 0 6px 0;"><strong>Other Party:</strong> %s</p>
            <p style="margin:0;"><strong>Total Amount:</strong> $%.2f</p>
        </div>
        <p>Funds held in escrow have been released to the provider. Don't forget to leave a review!</p>
        <div class="footer">
            <p>You received this transactional email because of activity on your Neighbor Service account.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, recipientName, serviceTitle, otherPartyName, totalAmount)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

func SendAppointmentCancelledEmail(cfg *Config, toEmail, recipientName, otherPartyName, serviceTitle, reason string) error {
	subject := fmt.Sprintf("Appointment Cancelled: %s", serviceTitle)
	plainText := fmt.Sprintf("Appointment Cancelled\n\nHello %s,\n\nYour appointment for \"%s\" with %s has been cancelled.\n\nReason: %s\n\nIf funds were held in escrow, any applicable refund will be processed automatically.\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", recipientName, serviceTitle, otherPartyName, reason)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Appointment Cancelled</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .card { background: #fef2f2; border-radius: 6px; padding: 16px; margin: 20px 0; border: 1px solid #fecaca; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <h2 style="margin: 0 0 16px 0; color: #b91c1c;">Appointment Cancelled</h2>
        <p>Hello %s,</p>
        <p>Your appointment for <strong>"%s"</strong> with <strong>%s</strong> has been cancelled.</p>
        <div class="card">
            <p style="margin:0;"><strong>Reason:</strong> %s</p>
        </div>
        <p>If funds were held in escrow, any applicable refund will be processed automatically.</p>
        <div class="footer">
            <p>You received this transactional email because of activity on your Neighbor Service account.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, recipientName, serviceTitle, otherPartyName, reason)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── PAYOUTS & DISPUTES EMAILS ───────────────────────────────────────────────

func SendPayoutStatusEmail(cfg *Config, toEmail, providerName string, amount float64, status, reason string) error {
	subject := fmt.Sprintf("Payout Request %s - Neighbor Service", status)
	reasonText := ""
	if reason != "" {
		reasonText = fmt.Sprintf("Notes: %s\n", reason)
	}
	plainText := fmt.Sprintf("Payout Status Update\n\nHello %s,\n\nYour payout request for $%.2f has been %s.\n%s\n---\nNeighbor Service Solutions LLC. All rights reserved.", providerName, amount, status, reasonText)
	
	reasonHTML := ""
	if reason != "" {
		reasonHTML = fmt.Sprintf(`<div class="card"><p style="margin:0;"><strong>Notes:</strong> %s</p></div>`, reason)
	}
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Payout Status Update</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .card { background: #f8fafc; border-radius: 6px; padding: 16px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <h2 style="margin: 0 0 16px 0; color: #0f172a;">Payout Status Update</h2>
        <p>Hello %s,</p>
        <p>Your payout request for <strong>$%.2f</strong> has been <strong>%s</strong>.</p>
        %s
        <div class="footer">
            <p>You received this transactional email because of activity on your Neighbor Service account.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, providerName, amount, status, reasonHTML)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

func SendDisputeFiledEmail(cfg *Config, toEmail, recipientName, appointmentID, reason string) error {
	subject := "Dispute Opened on Appointment - Neighbor Service"
	plainText := fmt.Sprintf("Dispute Opened\n\nHello %s,\n\nA dispute has been opened for Appointment ID %s.\n\nReason: %s\n\nOur moderation team is reviewing this case. You can upload additional evidence or notes in the app.\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", recipientName, appointmentID, reason)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dispute Opened</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .card { background: #fef2f2; border-radius: 6px; padding: 16px; margin: 20px 0; border: 1px solid #fecaca; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <h2 style="margin: 0 0 16px 0; color: #b91c1c;">Dispute Opened</h2>
        <p>Hello %s,</p>
        <p>A dispute has been opened for Appointment ID <code>%s</code>.</p>
        <div class="card">
            <p style="margin:0;"><strong>Reason:</strong> %s</p>
        </div>
        <p>Our moderation team is reviewing this case. You can upload additional evidence or notes in the app.</p>
        <div class="footer">
            <p>You received this transactional email because of activity on your Neighbor Service account.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, recipientName, appointmentID, reason)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

func SendDisputeResolvedEmail(cfg *Config, toEmail, recipientName, appointmentID, notes, status string) error {
	subject := fmt.Sprintf("Dispute %s: Appointment %s", status, appointmentID)
	plainText := fmt.Sprintf("Dispute Decision\n\nHello %s,\n\nThe dispute on Appointment %s has been marked %s.\n\nResolution Notes: %s\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", recipientName, appointmentID, status, notes)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dispute Decision</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .card { background: #f8fafc; border-radius: 6px; padding: 16px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <h2 style="margin: 0 0 16px 0; color: #0f172a;">Dispute Decision</h2>
        <p>Hello %s,</p>
        <p>The dispute on Appointment <code>%s</code> has been marked <strong>%s</strong>.</p>
        <div class="card">
            <p style="margin:0;"><strong>Resolution Notes:</strong> %s</p>
        </div>
        <div class="footer">
            <p>You received this transactional email because of activity on your Neighbor Service account.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, recipientName, appointmentID, status, notes)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── AUTOMATED LIFECYCLE & RETENTION EMAIL CAMPAIGNS ────────────────────────

// SendFavoriteProOpenSlotsEmail re-engages seekers when their previously booked pro has open calendar slots
func SendFavoriteProOpenSlotsEmail(cfg *Config, toEmail, seekerName, providerName, serviceName string) error {
	subject := fmt.Sprintf("%s has open slots this week in your area!", providerName)
	plainText := fmt.Sprintf("Hello %s,\n\nYour favorite neighborhood pro, %s (%s), has open availability this week!\n\nBook ahead to secure your preferred time slot before their calendar fills up.\n\nOpen Neighbor Service to book instantly: https://neighborservice.com\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", seekerName, providerName, serviceName)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Book Your Favorite Pro</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 12px; border: 1px solid #e2e8f0; }
        .badge { display: inline-block; padding: 6px 12px; background: #eff6ff; color: #2563eb; border-radius: 20px; font-weight: 700; font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px; }
        .card { background: #f8fafc; border-radius: 8px; padding: 20px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .btn { display: inline-block; padding: 14px 28px; background: #2563eb; color: #ffffff !important; border-radius: 8px; font-weight: 700; text-decoration: none; margin-top: 16px; text-align: center; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">Open Availability</span>
        <h2 style="margin-top: 14px; color: #0f172a;">Hello %s,</h2>
        <p>Great news! Your trusted provider <strong>%s</strong> has open calendar slots this week for <strong>%s</strong>.</p>
        <div class="card">
            <h3 style="margin:0 0 8px 0; font-size:16px; color:#0f172a;">%s</h3>
            <p style="margin:0; font-size:14px; color:#64748b;">Specialist: %s &bull; Verified Neighbor Pro</p>
        </div>
        <p>Book your upcoming appointment now to guarantee your preferred day and time.</p>
        <div style="text-align: center;">
            <a href="https://neighborservice.com" class="btn">Book Available Slot</a>
        </div>
        <div class="footer">
            <p>You received this recommendation based on your past positive experience with %s on Neighbor Service.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, seekerName, providerName, serviceName, serviceName, providerName, providerName)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// SendSeasonalCareReminderEmail sends seasonal home maintenance reminders
func SendSeasonalCareReminderEmail(cfg *Config, toEmail, seekerName, seasonName, tipsSummary string) error {
	subject := fmt.Sprintf("Seasonal Home Care: %s checklist for your home 🏡", seasonName)
	plainText := fmt.Sprintf("Hello %s,\n\n%s is here! Protect your home with our seasonal maintenance checklist:\n\n%s\n\nFind top-rated local pros on Neighbor Service to take care of these tasks hassle-free.\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", seekerName, seasonName, tipsSummary)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Seasonal Home Care</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 12px; border: 1px solid #e2e8f0; }
        .badge { display: inline-block; padding: 6px 12px; background: #fef3c7; color: #d97706; border-radius: 20px; font-weight: 700; font-size: 12px; text-transform: uppercase; }
        .card { background: #f8fafc; border-radius: 8px; padding: 20px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .btn { display: inline-block; padding: 14px 28px; background: #16a34a; color: #ffffff !important; border-radius: 8px; font-weight: 700; text-decoration: none; margin-top: 16px; text-align: center; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">%s Home Checklist</span>
        <h2 style="margin-top: 14px; color: #0f172a;">Hello %s,</h2>
        <p>As the season changes, keeping your home in top shape prevents costly repairs. Here is your quick seasonal maintenance guide:</p>
        <div class="card">
            <p style="margin:0; font-size:14px; line-height:1.6;">%s</p>
        </div>
        <p>Don't want to do it all yourself? Connect with verified local neighborhood pros in minutes.</p>
        <div style="text-align: center;">
            <a href="https://neighborservice.com" class="btn">Find Local Pros</a>
        </div>
        <div class="footer">
            <p>You received this maintenance guide as part of your Neighbor Service membership.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, seasonName, seekerName, tipsSummary)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// SendPostJobReviewAndReferralEmail triggers 24h after completion for reviews and neighbor referrals
func SendPostJobReviewAndReferralEmail(cfg *Config, toEmail, seekerName, providerName, serviceName, referralCode string) error {
	subject := fmt.Sprintf("How was your service with %s? Refer a neighbor & earn $10!", providerName)
	plainText := fmt.Sprintf("Hello %s,\n\nWe hope your recent %s with %s went smoothly!\n\n1. Leave a quick review in the app to support your local pro.\n2. Share your referral code [%s] with neighbors. When they complete their first booking, you both receive a $10 credit!\n\nOpen Neighbor Service: https://neighborservice.com\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", seekerName, serviceName, providerName, referralCode)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Review & Referral</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 12px; border: 1px solid #e2e8f0; }
        .badge { display: inline-block; padding: 6px 12px; background: #dcfce7; color: #16a34a; border-radius: 20px; font-weight: 700; font-size: 12px; }
        .card { background: #f8fafc; border-radius: 8px; padding: 20px; margin: 20px 0; border: 1px solid #e2e8f0; text-align: center; }
        .ref-code { font-size: 26px; font-weight: 800; letter-spacing: 4px; color: #2563eb; background: #eff6ff; padding: 12px; border-radius: 6px; margin: 12px 0; border: 1px dashed #93c5fd; }
        .btn { display: inline-block; padding: 14px 28px; background: #2563eb; color: #ffffff !important; border-radius: 8px; font-weight: 700; text-decoration: none; margin-top: 14px; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">Feedback & Referral</span>
        <h2 style="margin-top: 14px; color: #0f172a;">Hello %s,</h2>
        <p>How was your experience with <strong>%s</strong> for <strong>"%s"</strong>? Your feedback helps keep our community trusted and reliable.</p>
        <div class="card">
            <h3 style="margin:0 0 6px 0; color:#0f172a;">Give $10, Get $10</h3>
            <p style="margin:0; font-size:13px; color:#64748b;">Share your unique referral code with neighbors:</p>
            <div class="ref-code">%s</div>
            <p style="margin:0; font-size:12px; color:#64748b;">When they complete their first booking, you both get $10 credit!</p>
        </div>
        <div style="text-align: center;">
            <a href="https://neighborservice.com" class="btn">Leave a Star Review</a>
        </div>
        <div class="footer">
            <p>Thank you for supporting local neighborhood service professionals.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, seekerName, providerName, serviceName, referralCode)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// SendDormantUserWinbackEmail re-engages users inactive for 45+ days
func SendDormantUserWinbackEmail(cfg *Config, toEmail, userName, recentProsCount string) error {
	subject := "We miss you! Verified local pros are ready in your neighborhood 🏡"
	plainText := fmt.Sprintf("Hello %s,\n\nWe haven't seen you in a while! %s verified service professionals are active in your neighborhood right now.\n\nFrom home repairs and lawn care to pet sitting and cleaning, get your tasks completed with trusted neighbors.\n\nOpen the app today: https://neighborservice.com\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", userName, recentProsCount)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome Back to Neighbor Service</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 12px; border: 1px solid #e2e8f0; }
        .badge { display: inline-block; padding: 6px 12px; background: #f3e8ff; color: #9333ea; border-radius: 20px; font-weight: 700; font-size: 12px; }
        .card { background: #f8fafc; border-radius: 8px; padding: 20px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .btn { display: inline-block; padding: 14px 28px; background: #7c3aed; color: #ffffff !important; border-radius: 8px; font-weight: 700; text-decoration: none; margin-top: 14px; text-align: center; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">Welcome Back</span>
        <h2 style="margin-top: 14px; color: #0f172a;">Hello %s,</h2>
        <p>We miss having you in the neighborhood! Over <strong>%s verified pros</strong> are ready to help you with your upcoming home, garden, and personal projects.</p>
        <div class="card">
            <h4 style="margin:0 0 8px 0; color:#0f172a;">Popular in your area this week:</h4>
            <p style="margin:0; font-size:13px; color:#475569;">&bull; Lawn & Landscaping &bull; Handyman Repairs &bull; House Cleaning &bull; Plumbing & Electrical</p>
        </div>
        <div style="text-align: center;">
            <a href="https://neighborservice.com" class="btn">Explore Neighborhood Services</a>
        </div>
        <div class="footer">
            <p>You received this update because you are a registered member of Neighbor Service.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, userName, recentProsCount)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// SendWeeklyProviderDigestEmail sends Monday morning performance summaries to active providers
func SendWeeklyProviderDigestEmail(cfg *Config, toEmail, providerName string, weeklyEarnings float64, profileViews int, openJobsCount int) error {
	subject := fmt.Sprintf("Your Weekly Pro Digest: $%.2f earned & %d open jobs near you 📈", weeklyEarnings, openJobsCount)
	plainText := fmt.Sprintf("Hello %s,\n\nHere is your Neighbor Service weekly performance summary:\n\n- Weekly Earnings: $%.2f\n- Profile Views: %d\n- Open Local Requests: %d waiting for bids\n\nOpen your app to review open client requests and submit proposals.\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", providerName, weeklyEarnings, profileViews, openJobsCount)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Weekly Pro Digest</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 12px; border: 1px solid #e2e8f0; }
        .badge { display: inline-block; padding: 6px 12px; background: #e0f2fe; color: #0369a1; border-radius: 20px; font-weight: 700; font-size: 12px; }
        .stat-grid { display: table; width: 100%%; margin: 20px 0; }
        .stat-col { display: table-cell; width: 33.33%%; text-align: center; padding: 14px 8px; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; }
        .stat-val { font-size: 22px; font-weight: 800; color: #0f172a; margin-bottom: 4px; }
        .stat-label { font-size: 11px; font-weight: 600; color: #64748b; text-transform: uppercase; }
        .btn { display: inline-block; padding: 14px 28px; background: #0284c7; color: #ffffff !important; border-radius: 8px; font-weight: 700; text-decoration: none; margin-top: 16px; text-align: center; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">Weekly Performance</span>
        <h2 style="margin-top: 14px; color: #0f172a;">Hello %s,</h2>
        <p>Here is your weekly summary of earnings, client engagement, and local service demand:</p>
        <div class="stat-grid">
            <div class="stat-col">
                <div class="stat-val" style="color: #16a34a;">$%.2f</div>
                <div class="stat-label">Past 7 Days</div>
            </div>
            <div class="stat-col" style="border-left:none; border-right:none;">
                <div class="stat-val">%d</div>
                <div class="stat-label">Profile Views</div>
            </div>
            <div class="stat-col">
                <div class="stat-val" style="color: #2563eb;">%d</div>
                <div class="stat-label">Open Leads</div>
            </div>
        </div>
        <p>There are <strong>%d active client requests</strong> nearby looking for verified pros like you.</p>
        <div style="text-align: center;">
            <a href="https://neighborservice.com" class="btn">Review & Bid on Leads</a>
        </div>
        <div class="footer">
            <p>You received this digest as an active service provider on Neighbor Service.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, providerName, weeklyEarnings, profileViews, openJobsCount, openJobsCount)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// SendAppointmentUpcomingReminderEmail sends 24h and 2h appointment reminders to prevent no-shows
func SendAppointmentUpcomingReminderEmail(cfg *Config, toEmail, recipientName, otherPartyName, title, scheduledTime, role string) error {
	subject := fmt.Sprintf("Reminder: Upcoming Appointment with %s (%s)", otherPartyName, scheduledTime)
	arrivalNote := ""
	if role == "PROVIDER" {
		arrivalNote = "Please remember to arrive on time and collect the client's 4-digit arrival code upon arrival."
	} else {
		arrivalNote = "Your 4-digit arrival code is available inside the app. Provide it to your pro upon arrival to start the service."
	}
	plainText := fmt.Sprintf("Hello %s,\n\nThis is a friendly reminder for your upcoming appointment \"%s\" with %s.\n\nScheduled Time: %s\n\n%s\n\nOpen Neighbor Service: https://neighborservice.com\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", recipientName, title, otherPartyName, scheduledTime, arrivalNote)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Appointment Reminder</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 12px; border: 1px solid #e2e8f0; }
        .badge { display: inline-block; padding: 6px 12px; background: #fef3c7; color: #b45309; border-radius: 20px; font-weight: 700; font-size: 12px; }
        .card { background: #f8fafc; border-radius: 8px; padding: 20px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .btn { display: inline-block; padding: 14px 28px; background: #2563eb; color: #ffffff !important; border-radius: 8px; font-weight: 700; text-decoration: none; margin-top: 14px; text-align: center; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">Appointment Reminder</span>
        <h2 style="margin-top: 14px; color: #0f172a;">Hello %s,</h2>
        <p>You have an upcoming appointment scheduled for <strong>"%s"</strong> with <strong>%s</strong>.</p>
        <div class="card">
            <p style="margin:0 0 6px 0;"><strong>Scheduled Date & Time:</strong> %s</p>
            <p style="margin:0; font-size:13px; color:#475569;">%s</p>
        </div>
        <div style="text-align: center;">
            <a href="https://neighborservice.com" class="btn">View Appointment in App</a>
        </div>
        <div class="footer">
            <p>You received this automated reminder for your confirmed booking on Neighbor Service.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, recipientName, title, otherPartyName, scheduledTime, arrivalNote)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// SendProMilestonePerksEmail congratulates providers upon achieving milestones
func SendProMilestonePerksEmail(cfg *Config, toEmail, providerName, milestoneText string) error {
	subject := fmt.Sprintf("Congratulations %s! You reached a new Pro Milestone 🏆", providerName)
	plainText := fmt.Sprintf("Congratulations %s!\n\nYou just achieved: %s.\n\nYour dedication and quality service make our neighborhood community thrive. Keep up the amazing work!\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", providerName, milestoneText)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pro Milestone</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 12px; border: 1px solid #e2e8f0; text-align: center; }
        .badge { display: inline-block; padding: 6px 14px; background: #fef3c7; color: #d97706; border-radius: 20px; font-weight: 800; font-size: 13px; }
        .milestone-box { background: #fffbeb; border: 2px solid #fde68a; border-radius: 12px; padding: 24px; margin: 20px 0; }
        .btn { display: inline-block; padding: 14px 28px; background: #d97706; color: #ffffff !important; border-radius: 8px; font-weight: 700; text-decoration: none; margin-top: 14px; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">Community Milestone</span>
        <h2 style="margin-top: 14px; color: #0f172a;">Bravo, %s!</h2>
        <div class="milestone-box">
            <h3 style="margin:0; font-size:20px; color:#92400e;">%s</h3>
        </div>
        <p style="color:#475569;">Your dedication and high-quality service make Neighbor Service the most trusted community platform. Thank you for being a standout pro!</p>
        <div style="text-align: center;">
            <a href="https://neighborservice.com" class="btn">View Your Pro Profile</a>
        </div>
        <div class="footer">
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, providerName, milestoneText)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── CORE SMTP TRANSPORT ENGINE (ANTI-SPAM COMPLIANT) ─────────────────────────

// SendHTML sends an email formatted with multipart MIME fallback
func SendHTML(cfg *Config, toEmail, subject, htmlBody string) error {
	plainText := htmlToPlainText(htmlBody)
	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// SendMultipartEmail sends a production-grade multipart/alternative MIME email with
// strict RFC 5322 compliance, Date, Message-ID, and anti-spam optimization headers.
func SendMultipartEmail(cfg *Config, toEmail, subject, plainText, htmlBody string) error {
	if cfg == nil || cfg.Host == "" {
		return nil
	}

	fromRaw := cfg.From
	if fromRaw == "" {
		fromRaw = cfg.User
	}
	cleanFromEmail := extractEmail(fromRaw)

	// Format From with a trusted Display Name if not already present
	fromFormatted := fromRaw
	if !strings.Contains(fromRaw, "<") {
		fromFormatted = fmt.Sprintf("\"Neighbor Service\" <%s>", cleanFromEmail)
	}

	// Extract domain for Message-ID and headers
	domain := "neighborservice.com"
	if idx := strings.LastIndex(cleanFromEmail, "@"); idx != -1 {
		domain = strings.Trim(cleanFromEmail[idx+1:], " >")
	}

	msgID := fmt.Sprintf("<%s@%s>", uuid.New().String(), domain)
	dateHeader := time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700")
	boundary := fmt.Sprintf("=_boundary_%s", uuid.New().String())

	// Build RFC 5322 & anti-spam compliant headers
	var msgBuilder strings.Builder
	msgBuilder.WriteString(fmt.Sprintf("From: %s\r\n", fromFormatted))
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	msgBuilder.WriteString(fmt.Sprintf("Reply-To: %s\r\n", fromFormatted))
	msgBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msgBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", dateHeader))
	msgBuilder.WriteString(fmt.Sprintf("Message-ID: %s\r\n", msgID))
	msgBuilder.WriteString("MIME-Version: 1.0\r\n")
	msgBuilder.WriteString("X-Priority: 3\r\n")
	msgBuilder.WriteString("Importance: Normal\r\n")
	msgBuilder.WriteString("Auto-Submitted: auto-generated\r\n")
	msgBuilder.WriteString("X-Auto-Response-Suppress: All\r\n")
	msgBuilder.WriteString(fmt.Sprintf("List-Unsubscribe: <mailto:support@%s?subject=Unsubscribe>\r\n", domain))
	msgBuilder.WriteString("List-Unsubscribe-Post: List-Unsubscribe=One-Click\r\n")
	msgBuilder.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	msgBuilder.WriteString("\r\n")

	// 1. Plain text part (Critical for spam filter scoring)
	msgBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msgBuilder.WriteString("Content-Type: text/plain; charset=UTF-8; format=flowed\r\n")
	msgBuilder.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msgBuilder.WriteString(plainText)
	msgBuilder.WriteString("\r\n\r\n")

	// 2. HTML part
	msgBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msgBuilder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msgBuilder.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	msgBuilder.WriteString(htmlBody)
	msgBuilder.WriteString("\r\n\r\n")

	// End boundary
	msgBuilder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	message := msgBuilder.String()
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)

	// Direct SSL/TLS (Port 465)
	if cfg.Port == 465 {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         cfg.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			return err
		}
		defer client.Quit()

		if err = client.Auth(auth); err != nil {
			return err
		}
		if err = client.Mail(cleanFromEmail); err != nil {
			return err
		}
		if err = client.Rcpt(toEmail); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(message))
		if err != nil {
			return err
		}
		return w.Close()
	}

	// Standard STARTTLS (Port 587 / 25)
	return smtp.SendMail(addr, auth, cleanFromEmail, []string{toEmail}, []byte(message))
}

// extractEmail parses plain email from formats like "Neighbor Service <noreply@domain.com>"
func extractEmail(addr string) string {
	if start := strings.Index(addr, "<"); start != -1 {
		if end := strings.Index(addr, ">"); end != -1 && end > start {
			return strings.TrimSpace(addr[start+1 : end])
		}
	}
	return strings.TrimSpace(addr)
}

// htmlToPlainText strips HTML tags to produce a readable plain text fallback
func htmlToPlainText(html string) string {
	// Replace break tags and paragraph tags with newlines
	reBr := regexp.MustCompile(`(?i)<br\s*/?>|</p>|</div>|</h1>|</h2>|</h3>`)
	text := reBr.ReplaceAllString(html, "\n")

	// Strip remaining HTML tags
	reTags := regexp.MustCompile(`<[^>]*>`)
	text = reTags.ReplaceAllString(text, "")

	// Clean excess whitespace
	reSpaces := regexp.MustCompile(`[ \t]+`)
	text = reSpaces.ReplaceAllString(text, " ")

	reNewlines := regexp.MustCompile(`\n{3,}`)
	text = reNewlines.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	name := parts[0]
	domain := parts[1]
	if len(name) <= 2 {
		return name + "@" + domain
	}
	return string(name[0]) + strings.Repeat("*", len(name)-2) + string(name[len(name)-1]) + "@" + domain
}

func SendNearbyJobAlertEmail(cfg *Config, toEmail, recipientName, requestTitle, serviceType string, distanceKm float64, price float64) error {
	subject := fmt.Sprintf("New Job Opportunity Nearby: %s", requestTitle)
	priceStr := "Negotiable"
	if price > 0 {
		priceStr = fmt.Sprintf("$%.2f", price)
	}

	plainText := fmt.Sprintf("Hello %s,\n\nA new job matching your location is now available:\n\nTitle: %s\nService: %s\nDistance: approx %.1f km away\nBudget: %s\n\nOpen the Neighbor Service app to view full details and send your proposal!\n\n---\nNeighbor Service Solutions LLC", recipientName, requestTitle, serviceType, distanceKm, priceStr)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Job Opportunity Nearby</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.5; margin: 0; }
        .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; padding: 36px; border-radius: 8px; border: 1px solid #e2e8f0; }
        .header { text-align: center; margin-bottom: 24px; }
        .header h2 { color: #0f172a; margin: 0; font-size: 22px; font-weight: 700; }
        .badge { display: inline-block; background-color: #dbeafe; color: #1e40af; padding: 4px 12px; border-radius: 9999px; font-size: 13px; font-weight: 600; margin-bottom: 16px; }
        .card { background-color: #f1f5f9; border-radius: 8px; padding: 20px; margin: 20px 0; border: 1px solid #cbd5e1; }
        .card-row { display: flex; justify-content: space-between; margin-bottom: 8px; font-size: 14px; }
        .btn { display: inline-block; background-color: #2563eb; color: #ffffff; text-decoration: none; padding: 12px 28px; border-radius: 6px; font-weight: 600; font-size: 15px; margin-top: 16px; text-align: center; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <span class="badge">📍 Nearby Job Alert</span>
            <h2>New Job Opportunity Near You!</h2>
        </div>
        <p>Hello <strong>%s</strong>,</p>
        <p>A customer in your area just posted a service request that matches your profile:</p>
        <div class="card">
            <h3 style="margin-top: 0; color: #0f172a;">%s</h3>
            <p style="margin: 4px 0; color: #475569; font-size: 14px;"><strong>Service:</strong> %s</p>
            <p style="margin: 4px 0; color: #475569; font-size: 14px;"><strong>Distance:</strong> ~%.1f km away</p>
            <p style="margin: 4px 0; color: #475569; font-size: 14px;"><strong>Budget:</strong> %s</p>
        </div>
        <p style="text-align: center;">
            <a href="https://neighborservice.com" class="btn">View Request in App</a>
        </p>
        <div class="footer">
            <p>You received this email because you are registered as a service provider on Neighbor Service.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, recipientName, requestTitle, serviceType, distanceKm, priceStr)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}


// ─── STALE REQUEST ────────────────────────────────────────────────────────────

func SendStaleRequestExpiredEmail(cfg *Config, toEmail, seekerName, requestTitle string) error {
	subject := "Your service request has expired - Neighbor Service"
	plainText := fmt.Sprintf("Hi %s,\n\nYour service request '%s' has been automatically closed after 30 days with no activity.\n\nFeel free to post a new request anytime.\n\n— Neighbor Service Team", seekerName, requestTitle)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:'Segoe UI',Arial,sans-serif;background:#f1f5f9;margin:0;padding:0;}
.container{max-width:600px;margin:40px auto;background:#fff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08);}
.header{background:linear-gradient(135deg,#64748b,#94a3b8);padding:40px 32px;text-align:center;color:#fff;}
.header h1{margin:0;font-size:24px;}.body{padding:32px;}
.card{background:#f8fafc;border:1px solid #e2e8f0;border-radius:12px;padding:20px;margin:20px 0;}
.btn{display:inline-block;background:linear-gradient(135deg,#6366f1,#8b5cf6);color:#fff;text-decoration:none;padding:14px 32px;border-radius:10px;font-weight:600;margin-top:16px;}
.footer{background:#f8fafc;padding:24px 32px;text-align:center;color:#94a3b8;font-size:12px;}
</style></head><body><div class="container">
<div class="header"><h1>📋 Request Expired</h1></div>
<div class="body">
<h2 style="color:#0f172a;margin-top:0;">Hi %s,</h2>
<p style="color:#475569;">Your service request has been automatically closed after <strong>30 days</strong> with no provider activity.</p>
<div class="card"><p style="margin:0;font-size:15px;font-weight:600;color:#0f172a;">"%s"</p><p style="margin:8px 0 0;color:#64748b;font-size:13px;">Status: Expired</p></div>
<p style="color:#475569;">No worries — you can post a new request anytime and reach hundreds of providers near you.</p>
<p style="text-align:center;"><a href="https://neighborservice.com" class="btn">Post a New Request</a></p>
</div>
<div class="footer"><p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p></div>
</div></body></html>`, seekerName, requestTitle)
	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── PAYOUT PROCESSING ────────────────────────────────────────────────────────

func SendPayoutProcessingEmail(cfg *Config, toEmail, providerName string, amount float64, status, notes string) error {
	var emoji, headline, detail string
	switch status {
	case "PROCESSING":
		emoji, headline = "💸", "Your payout is being processed!"
		detail = fmt.Sprintf("Your withdrawal of <strong>$%.2f</strong> is now in progress. Funds typically arrive within 2–5 business days.", amount)
	case "FAILED":
		emoji, headline = "⚠️", "Payout could not be processed"
		detail = fmt.Sprintf("We were unable to process your withdrawal of <strong>$%.2f</strong>. Reason: %s", amount, notes)
	default:
		emoji, headline = "📤", "Payout status update"
		detail = fmt.Sprintf("Your payout of <strong>$%.2f</strong> — status: %s.", amount, status)
	}
	subject := fmt.Sprintf("%s Payout %s - Neighbor Service", emoji, status)
	plainText := fmt.Sprintf("Hi %s,\n\n%s\nAmount: $%.2f\n%s\n\n— Neighbor Service Team", providerName, headline, amount, notes)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:'Segoe UI',Arial,sans-serif;background:#f1f5f9;margin:0;padding:0;}
.container{max-width:600px;margin:40px auto;background:#fff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08);}
.header{background:linear-gradient(135deg,#059669,#10b981);padding:40px 32px;text-align:center;color:#fff;}
.header h1{margin:0;font-size:24px;}.body{padding:32px;}
.amount{font-size:36px;font-weight:700;color:#059669;text-align:center;margin:16px 0;}
.card{background:#f0fdf4;border:1px solid #bbf7d0;border-radius:12px;padding:20px;margin:20px 0;text-align:center;}
.btn{display:inline-block;background:linear-gradient(135deg,#059669,#10b981);color:#fff;text-decoration:none;padding:14px 32px;border-radius:10px;font-weight:600;margin-top:16px;}
.footer{background:#f8fafc;padding:24px 32px;text-align:center;color:#94a3b8;font-size:12px;}
</style></head><body><div class="container">
<div class="header"><h1>%s %s</h1></div>
<div class="body">
<h2 style="color:#0f172a;margin-top:0;">Hi %s,</h2>
<p style="color:#475569;">%s</p>
<div class="card"><div class="amount">$%.2f</div><p style="color:#64748b;margin:0;font-size:13px;">Withdrawal Amount</p></div>
<p style="text-align:center;"><a href="https://neighborservice.com" class="btn">View Wallet</a></p>
</div>
<div class="footer"><p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p></div>
</div></body></html>`, emoji, headline, providerName, detail, amount)
	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── SUBSCRIPTION EXPIRY ──────────────────────────────────────────────────────

func SendSubscriptionExpiryReminderEmail(cfg *Config, toEmail, userName, planName, expiryDate string) error {
	subject := "Your subscription is expiring soon ⚠️ - Neighbor Service"
	plainText := fmt.Sprintf("Hi %s,\n\nYour '%s' plan expires on %s. Renew now to keep your premium benefits.\n\n— Neighbor Service Team", userName, planName, expiryDate)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:'Segoe UI',Arial,sans-serif;background:#f1f5f9;margin:0;padding:0;}
.container{max-width:600px;margin:40px auto;background:#fff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08);}
.header{background:linear-gradient(135deg,#f59e0b,#fbbf24);padding:40px 32px;text-align:center;color:#fff;}
.header h1{margin:0;font-size:24px;}.body{padding:32px;}
.card{background:#fffbeb;border:1px solid #fde68a;border-radius:12px;padding:20px;margin:20px 0;}
.btn{display:inline-block;background:linear-gradient(135deg,#f59e0b,#f97316);color:#fff;text-decoration:none;padding:14px 32px;border-radius:10px;font-weight:600;margin-top:16px;}
.footer{background:#f8fafc;padding:24px 32px;text-align:center;color:#94a3b8;font-size:12px;}
</style></head><body><div class="container">
<div class="header"><h1>⚠️ Subscription Expiring Soon</h1></div>
<div class="body">
<h2 style="color:#0f172a;margin-top:0;">Hi %s,</h2>
<p style="color:#475569;">Your subscription is about to expire. Don't lose access to your premium features!</p>
<div class="card"><p style="margin:0;font-size:15px;font-weight:600;color:#92400e;">Plan: %s</p><p style="margin:8px 0 0;color:#b45309;font-size:14px;">Expires: %s</p></div>
<p style="color:#475569;">Renew now to keep enjoying priority listings, unlimited messages, and more.</p>
<p style="text-align:center;"><a href="https://neighborservice.com" class="btn">Renew Subscription</a></p>
</div>
<div class="footer"><p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p></div>
</div></body></html>`, userName, planName, expiryDate)
	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

func SendSubscriptionExpiredEmail(cfg *Config, toEmail, userName, planName string) error {
	subject := "Your subscription has expired - Neighbor Service"
	plainText := fmt.Sprintf("Hi %s,\n\nYour '%s' plan has expired. Resubscribe to restore your premium benefits.\n\n— Neighbor Service Team", userName, planName)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:'Segoe UI',Arial,sans-serif;background:#f1f5f9;margin:0;padding:0;}
.container{max-width:600px;margin:40px auto;background:#fff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08);}
.header{background:linear-gradient(135deg,#dc2626,#ef4444);padding:40px 32px;text-align:center;color:#fff;}
.header h1{margin:0;font-size:24px;}.body{padding:32px;}
.card{background:#fef2f2;border:1px solid #fecaca;border-radius:12px;padding:20px;margin:20px 0;}
.btn{display:inline-block;background:linear-gradient(135deg,#6366f1,#8b5cf6);color:#fff;text-decoration:none;padding:14px 32px;border-radius:10px;font-weight:600;margin-top:16px;}
.footer{background:#f8fafc;padding:24px 32px;text-align:center;color:#94a3b8;font-size:12px;}
</style></head><body><div class="container">
<div class="header"><h1>❌ Subscription Expired</h1></div>
<div class="body">
<h2 style="color:#0f172a;margin-top:0;">Hi %s,</h2>
<p style="color:#475569;">Your premium subscription has expired and your account has been moved to the free tier.</p>
<div class="card"><p style="margin:0;font-size:15px;font-weight:600;color:#dc2626;">Plan: %s — Expired</p></div>
<p style="color:#475569;">Resubscribe today to restore priority listings, unlimited messages, and more premium perks.</p>
<p style="text-align:center;"><a href="https://neighborservice.com" class="btn">Resubscribe Now</a></p>
</div>
<div class="footer"><p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p></div>
</div></body></html>`, userName, planName)
	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── DISPUTE ESCALATION ───────────────────────────────────────────────────────

func SendDisputeEscalatedEmail(cfg *Config, toEmail, recipientName, disputeID, reason string) error {
	subject := "Your dispute has been escalated to our support team - Neighbor Service"
	plainText := fmt.Sprintf("Hi %s,\n\nDispute #%s regarding '%s' has been escalated after 48 hours without resolution. Our team will respond within 24 hours.\n\n— Neighbor Service Team", recipientName, disputeID, reason)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:'Segoe UI',Arial,sans-serif;background:#f1f5f9;margin:0;padding:0;}
.container{max-width:600px;margin:40px auto;background:#fff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08);}
.header{background:linear-gradient(135deg,#7c3aed,#a855f7);padding:40px 32px;text-align:center;color:#fff;}
.header h1{margin:0;font-size:24px;}.body{padding:32px;}
.card{background:#faf5ff;border:1px solid #e9d5ff;border-radius:12px;padding:20px;margin:16px 0;}
.step{padding:10px 0;border-bottom:1px solid #ede9fe;color:#475569;font-size:14px;}
.step:last-child{border-bottom:none;}
.btn{display:inline-block;background:linear-gradient(135deg,#7c3aed,#a855f7);color:#fff;text-decoration:none;padding:14px 32px;border-radius:10px;font-weight:600;margin-top:16px;}
.footer{background:#f8fafc;padding:24px 32px;text-align:center;color:#94a3b8;font-size:12px;}
</style></head><body><div class="container">
<div class="header"><h1>⚖️ Dispute Escalated</h1></div>
<div class="body">
<h2 style="color:#0f172a;margin-top:0;">Hi %s,</h2>
<p style="color:#475569;">Your dispute has been escalated to our support team after 48 hours of inactivity.</p>
<div class="card"><p style="margin:0;font-size:13px;color:#7c3aed;font-weight:600;">DISPUTE #%s</p><p style="margin:8px 0 0;font-size:15px;font-weight:600;color:#0f172a;">%s</p></div>
<div class="card">
<div class="step">✅ Dispute filed — awaiting resolution</div>
<div class="step">⏱️ 48h elapsed — auto-escalated to support</div>
<div class="step">🔍 Under review — response within 24 hours</div>
</div>
<p style="text-align:center;"><a href="https://neighborservice.com" class="btn">View Dispute Details</a></p>
</div>
<div class="footer"><p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p></div>
</div></body></html>`, recipientName, disputeID, reason)
	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── INACTIVE USER RE-ENGAGEMENT ─────────────────────────────────────────────

func SendInactiveUserReengagementEmail(cfg *Config, toEmail, userName, campaignType string) error {
	var headline, sub, cta string
	switch campaignType {
	case "INACTIVE_14D":
		headline = "We miss you, " + userName + "! 👋"
		sub = "It's been 2 weeks since your last visit. New providers have joined in your area and exciting jobs are waiting."
		cta = "See What's New"
	default:
		headline = "Still there, " + userName + "? 🤔"
		sub = "It's been over a month. Dozens of trusted providers are near you — don't miss out!"
		cta = "Come Back & Explore"
	}
	subject := headline + " — Neighbor Service"
	plainText := fmt.Sprintf("%s\n\n%s\n\nhttps://neighborservice.com\n\n— Neighbor Service Team", headline, sub)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:'Segoe UI',Arial,sans-serif;background:#f1f5f9;margin:0;padding:0;}
.container{max-width:600px;margin:40px auto;background:#fff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08);}
.header{background:linear-gradient(135deg,#0ea5e9,#6366f1);padding:48px 32px;text-align:center;color:#fff;}
.header h1{margin:0;font-size:26px;font-weight:700;}.body{padding:40px 32px;}
.features{margin:24px 0;background:#f8fafc;border-radius:12px;padding:20px;}
.feature{padding:10px 0;border-bottom:1px solid #e2e8f0;color:#475569;font-size:14px;}
.feature:last-child{border-bottom:none;}
.btn{display:inline-block;background:linear-gradient(135deg,#0ea5e9,#6366f1);color:#fff;text-decoration:none;padding:16px 40px;border-radius:12px;font-weight:700;font-size:16px;margin-top:8px;}
.footer{background:#f8fafc;padding:24px 32px;text-align:center;color:#94a3b8;font-size:12px;}
</style></head><body><div class="container">
<div class="header"><h1>%s</h1></div>
<div class="body">
<p style="color:#475569;font-size:16px;text-align:center;">%s</p>
<div class="features">
<div class="feature">🔍 Find nearby verified providers instantly</div>
<div class="feature">⭐ Top-rated pros with real reviews</div>
<div class="feature">💬 Chat, book, and pay — all in one place</div>
</div>
<p style="text-align:center;"><a href="https://neighborservice.com" class="btn">%s</a></p>
</div>
<div class="footer"><p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p></div>
</div></body></html>`, headline, sub, cta)
	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── PERFORMANCE BADGE ────────────────────────────────────────────────────────

func SendBadgeAwardedEmail(cfg *Config, toEmail, providerName, badgeName, badgeDescription string) error {
	subject := fmt.Sprintf("🏅 You've earned a new badge: %s - Neighbor Service", badgeName)
	plainText := fmt.Sprintf("Hi %s,\n\nCongratulations! You've earned the '%s' badge.\n\n%s\n\nKeep up the great work!\n\n— Neighbor Service Team", providerName, badgeName, badgeDescription)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:'Segoe UI',Arial,sans-serif;background:#f1f5f9;margin:0;padding:0;}
.container{max-width:600px;margin:40px auto;background:#fff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08);}
.header{background:linear-gradient(135deg,#f59e0b,#eab308);padding:48px 32px;text-align:center;color:#fff;}
.medal{font-size:64px;display:block;margin-bottom:12px;}
.header h1{margin:0;font-size:24px;}.body{padding:40px 32px;text-align:center;}
.badge-card{background:linear-gradient(135deg,#fef3c7,#fde68a);border:2px solid #f59e0b;border-radius:16px;padding:28px;margin:24px auto;max-width:360px;}
.badge-name{font-size:22px;font-weight:700;color:#92400e;margin:0 0 8px;}
.badge-desc{font-size:14px;color:#b45309;margin:0;}
.btn{display:inline-block;background:linear-gradient(135deg,#f59e0b,#f97316);color:#fff;text-decoration:none;padding:14px 32px;border-radius:10px;font-weight:600;margin-top:16px;}
.footer{background:#f8fafc;padding:24px 32px;text-align:center;color:#94a3b8;font-size:12px;}
</style></head><body><div class="container">
<div class="header"><span class="medal">🏅</span><h1>Badge Unlocked!</h1></div>
<div class="body">
<h2 style="color:#0f172a;margin-top:0;">Congratulations, %s!</h2>
<p style="color:#475569;">Your hard work has paid off. You've earned a new achievement!</p>
<div class="badge-card"><p class="badge-name">%s</p><p class="badge-desc">%s</p></div>
<p style="color:#475569;">This badge is now visible on your public profile, helping you attract more clients.</p>
<a href="https://neighborservice.com" class="btn">View Your Profile</a>
</div>
<div class="footer"><p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p></div>
</div></body></html>`, providerName, badgeName, badgeDescription)
	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── WALLET RECONCILIATION ────────────────────────────────────────────────────

func SendWalletDiscrepancyAlertEmail(cfg *Config, toEmail, adminName, userID, walletID string, bookBalance, txSum, diff float64) error {
	subject := fmt.Sprintf("🚨 Wallet Discrepancy Detected — User %s - Neighbor Service Admin", userID[:8])
	plainText := fmt.Sprintf("Admin Alert\n\nWallet discrepancy detected.\n\nUser ID: %s\nWallet ID: %s\nBook Balance: $%.2f\nTransaction Sum: $%.2f\nDifference: $%.2f\n\nPlease review immediately.\n\n— Neighbor Service System", userID, walletID, bookBalance, txSum, diff)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:'Segoe UI',Arial,sans-serif;background:#f1f5f9;margin:0;padding:0;}
.container{max-width:600px;margin:40px auto;background:#fff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08);}
.header{background:linear-gradient(135deg,#dc2626,#b91c1c);padding:40px 32px;text-align:center;color:#fff;}
.header h1{margin:0;font-size:24px;}.body{padding:32px;}
.alert-box{background:#fef2f2;border:1px solid #fca5a5;border-radius:12px;padding:20px;margin:20px 0;}
.row{display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #fee2e2;font-size:14px;}
.row:last-child{border-bottom:none;}
.label{color:#64748b;}.val{font-weight:600;color:#0f172a;}
.diff{color:#dc2626;font-size:22px;font-weight:700;text-align:center;margin:16px 0;}
.btn{display:inline-block;background:linear-gradient(135deg,#dc2626,#b91c1c);color:#fff;text-decoration:none;padding:14px 32px;border-radius:10px;font-weight:600;margin-top:16px;}
.footer{background:#f8fafc;padding:24px 32px;text-align:center;color:#94a3b8;font-size:12px;}
</style></head><body><div class="container">
<div class="header"><h1>🚨 Wallet Discrepancy Alert</h1></div>
<div class="body">
<h2 style="color:#0f172a;margin-top:0;">Hi %s,</h2>
<p style="color:#dc2626;font-weight:600;">A wallet balance discrepancy has been detected and requires immediate review.</p>
<div class="alert-box">
<div class="row"><span class="label">User ID</span><span class="val">%s</span></div>
<div class="row"><span class="label">Wallet ID</span><span class="val">%s</span></div>
<div class="row"><span class="label">Book Balance</span><span class="val">$%.2f</span></div>
<div class="row"><span class="label">Transaction Sum</span><span class="val">$%.2f</span></div>
</div>
<div class="diff">Discrepancy: $%.2f</div>
<p style="text-align:center;"><a href="https://neighborservice.com/admin" class="btn">Review in Admin Panel</a></p>
</div>
<div class="footer"><p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p></div>
</div></body></html>`, adminName, userID, walletID, bookBalance, txSum, diff)
	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── APPOINTMENT NO-SHOW & REVIEW REMINDERS ──────────────────────────────────

func SendAppointmentNoShowEmail(cfg *Config, toEmail, userName, role, aptTitle string, aptTime time.Time) error {
	subject := fmt.Sprintf("Appointment Follow-up: %s - Neighbor Service", aptTitle)
	timeFormatted := aptTime.Format("Jan 02, 2006 at 3:04 PM")
	plainText := fmt.Sprintf("Hi %s,\n\nYour scheduled appointment '%s' on %s was not marked as completed.\nIf you experienced any issues or need to reschedule or report a dispute, please visit your Neighbor Service dashboard.\n\n— Neighbor Service Team", userName, aptTitle, timeFormatted)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:'Segoe UI',Arial,sans-serif;background:#f8fafc;margin:0;padding:0;}
.container{max-width:580px;margin:30px auto;background:#fff;border-radius:14px;border:1px solid #e2e8f0;overflow:hidden;}
.header{background:linear-gradient(135deg,#f59e0b,#d97706);padding:30px 24px;text-align:center;color:#fff;}
.header h2{margin:0;font-size:22px;}.body{padding:28px 24px;color:#1e293b;line-height:1.6;}
.details{background:#fef3c7;border:1px solid #fde68a;border-radius:10px;padding:16px;margin:18px 0;}
.btn{display:inline-block;background:#d97706;color:#fff;text-decoration:none;padding:12px 24px;border-radius:8px;font-weight:600;}
.footer{background:#f8fafc;padding:18px;text-align:center;color:#94a3b8;font-size:12px;border-top:1px solid #f1f5f9;}
</style></head><body><div class="container">
<div class="header"><h2>Appointment Follow-up</h2></div>
<div class="body">
<p>Hi <strong>%s</strong>,</p>
<p>We noticed your scheduled appointment did not take place as planned:</p>
<div class="details">
<strong>Service:</strong> %s<br>
<strong>Scheduled Time:</strong> %s
</div>
<p>If you need to reschedule, submit a completion request, or contact support, please access your appointments.</p>
<p style="text-align:center;margin-top:24px;"><a href="https://neighborservice.com/appointments" class="btn">View Appointment</a></p>
</div>
<div class="footer"><p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p></div>
</div></body></html>`, userName, aptTitle, timeFormatted)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

func SendReviewReminderEmail(cfg *Config, toEmail, seekerName, providerName, aptTitle string) error {
	subject := fmt.Sprintf("How was your experience with %s? - Neighbor Service", providerName)
	plainText := fmt.Sprintf("Hi %s,\n\nYour service '%s' with %s was recently completed. We would love to hear your feedback! Leaving a review helps trusted neighbors find great local providers.\n\nLeave a review: https://neighborservice.com/appointments\n\n— Neighbor Service Team", seekerName, aptTitle, providerName)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:'Segoe UI',Arial,sans-serif;background:#f8fafc;margin:0;padding:0;}
.container{max-width:580px;margin:30px auto;background:#fff;border-radius:14px;border:1px solid #e2e8f0;overflow:hidden;}
.header{background:linear-gradient(135deg,#6366f1,#4f46e5);padding:32px 24px;text-align:center;color:#fff;}
.header h2{margin:0;font-size:22px;}.stars{font-size:26px;letter-spacing:4px;margin-top:8px;}
.body{padding:28px 24px;color:#1e293b;line-height:1.6;}
.btn{display:inline-block;background:linear-gradient(135deg,#6366f1,#4f46e5);color:#fff;text-decoration:none;padding:14px 28px;border-radius:8px;font-weight:600;}
.footer{background:#f8fafc;padding:18px;text-align:center;color:#94a3b8;font-size:12px;border-top:1px solid #f1f5f9;}
</style></head><body><div class="container">
<div class="header">
<h2>Rate Your Recent Experience</h2>
<div class="stars">⭐⭐⭐⭐⭐</div>
</div>
<div class="body">
<p>Hi <strong>%s</strong>,</p>
<p>Your recent appointment for <strong>%s</strong> with <strong>%s</strong> has concluded. How did it go?</p>
<p>Leaving a quick review takes less than a minute and helps your community discover top-rated neighbors.</p>
<p style="text-align:center;margin:28px 0;"><a href="https://neighborservice.com/appointments" class="btn">Leave a Review</a></p>
</div>
<div class="footer"><p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p></div>
</div></body></html>`, seekerName, aptTitle, providerName)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

func SendProviderRankingEmail(cfg *Config, toEmail, providerName, rankTier string, score int) error {
	subject := fmt.Sprintf("🌟 Provider Performance Update: You're at %s Tier! - Neighbor Service", rankTier)
	plainText := fmt.Sprintf("Hi %s,\n\nYour neighbor score is %d, placing you in the %s provider tier. Keep up the amazing work to maintain high search visibility and client trust!\n\n— Neighbor Service Team", providerName, score, rankTier)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:'Segoe UI',Arial,sans-serif;background:#f8fafc;margin:0;padding:0;}
.container{max-width:580px;margin:30px auto;background:#fff;border-radius:14px;border:1px solid #e2e8f0;overflow:hidden;}
.header{background:linear-gradient(135deg,#059669,#10b981);padding:32px 24px;text-align:center;color:#fff;}
.header h2{margin:0;font-size:22px;}.body{padding:28px 24px;color:#1e293b;line-height:1.6;}
.score-box{background:#ecfdf5;border:1px solid #a7f3d0;border-radius:10px;padding:18px;text-align:center;margin:20px 0;}
.score-val{font-size:32px;font-weight:800;color:#059669;}
.btn{display:inline-block;background:#059669;color:#fff;text-decoration:none;padding:12px 24px;border-radius:8px;font-weight:600;}
.footer{background:#f8fafc;padding:18px;text-align:center;color:#94a3b8;font-size:12px;border-top:1px solid #f1f5f9;}
</style></head><body><div class="container">
<div class="header"><h2>🌟 Provider Ranking Update</h2></div>
<div class="body">
<p>Hi <strong>%s</strong>,</p>
<p>Our weekly performance calculation just completed. Here is your current standing:</p>
<div class="score-box">
<div style="font-size:14px;color:#047857;text-transform:uppercase;font-weight:700;">Neighbor Score</div>
<div class="score-val">%d</div>
<div style="color:#065f46;font-weight:600;margin-top:4px;">Tier: %s</div>
</div>
<p>Higher scores give your profile priority placement in search results and client discovery.</p>
<p style="text-align:center;margin-top:24px;"><a href="https://neighborservice.com/provider/stats" class="btn">View Performance Analytics</a></p>
</div>
<div class="footer"><p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p></div>
</div></body></html>`, providerName, score, rankTier)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

// ─── PUBLIC SITE CONTACT & RESOLUTION EMAILS ────────────────────────────────

func SendContactConfirmationEmail(cfg *Config, toEmail, senderName, inquiryType, messageSnippet string) error {
	subject := "We've received your message - Neighbor Service Support"
	plainText := fmt.Sprintf("Hello %s,\n\nThank you for reaching out to Neighbor Service. We have received your message regarding \"%s\".\n\nOur customer care team is reviewing your request and will get back to you shortly (typically within 2 hours during support hours).\n\nYour message summary:\n\"%s\"\n\nIf you need immediate assistance, you can also browse our Help Center at https://neighborservice.com/support.\n\nWarm regards,\nNeighbor Service Support Team\n---\nNeighbor Service Solutions LLC. All rights reserved.", senderName, inquiryType, messageSnippet)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Message Received</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.6; margin: 0; }
        .container { max-width: 580px; margin: 20px auto; background-color: #ffffff; padding: 36px; border-radius: 12px; border: 1px solid #e2e8f0; }
        .badge { display: inline-block; padding: 6px 14px; background: #f0eafd; color: #5b16ee; border-radius: 20px; font-weight: 700; font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px; }
        .card { background: #f8fafc; border-radius: 8px; padding: 20px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .btn { display: inline-block; padding: 12px 24px; background: #5b16ee; color: #ffffff !important; border-radius: 6px; font-weight: 700; text-decoration: none; margin-top: 14px; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">Support Inquiry Received</span>
        <h2 style="margin-top: 14px; color: #0f172a;">Hello %s,</h2>
        <p>Thank you for contacting Neighbor Service. We have successfully received your inquiry regarding <strong>%s</strong>.</p>
        <div class="card">
            <h4 style="margin: 0 0 8px 0; color: #0f172a; font-size: 14px; text-transform: uppercase;">Message Summary</h4>
            <p style="margin: 0; font-size: 14px; color: #475569; font-style: italic;">"%s"</p>
        </div>
        <p>Our dedicated support team is reviewing your message and will respond to you shortly. Our average response time is under <strong>2 hours</strong> during business hours.</p>
        <div style="text-align: center;">
            <a href="https://neighborservice.com/support" class="btn">Visit Help &amp; Support</a>
        </div>
        <div class="footer">
            <p>You received this email because you submitted a contact inquiry on Neighbor Service.</p>
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, senderName, inquiryType, messageSnippet)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

func SendResolutionConfirmationEmail(cfg *Config, toEmail, submitterRole, issueType, bookingRef string) error {
	subject := fmt.Sprintf("Resolution Case Filed: %s - Neighbor Service", issueType)
	plainText := fmt.Sprintf("Hello,\n\nWe have received your dispute resolution submission for \"%s\" (Role: %s, Ref: %s).\n\nOur Trust & Safety moderation team will review the details submitted and coordinate resolution between parties within 24-48 business hours.\n\n---\nNeighbor Service Solutions LLC. All rights reserved.", issueType, submitterRole, bookingRef)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Resolution Case Received</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; line-height: 1.6; margin: 0; }
        .container { max-width: 580px; margin: 20px auto; background-color: #ffffff; padding: 36px; border-radius: 12px; border: 1px solid #e2e8f0; }
        .badge { display: inline-block; padding: 6px 14px; background: #fef3c7; color: #d97706; border-radius: 20px; font-weight: 700; font-size: 12px; text-transform: uppercase; }
        .card { background: #f8fafc; border-radius: 8px; padding: 20px; margin: 20px 0; border: 1px solid #e2e8f0; }
        .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">Resolution Case Received</span>
        <h2 style="margin-top: 14px; color: #0f172a;">Dispute Report Registered</h2>
        <p>Thank you for submitting your case to the Neighbor Service Resolution Center.</p>
        <div class="card">
            <p style="margin: 0 0 6px 0;"><strong>Issue Category:</strong> %s</p>
            <p style="margin: 0 0 6px 0;"><strong>Role:</strong> %s</p>
            <p style="margin: 0;"><strong>Booking/User Reference:</strong> %s</p>
        </div>
        <p>Our Trust &amp; Safety moderation team is reviewing the submission and will follow up with next steps within 24 to 48 hours.</p>
        <div class="footer">
            <p>&copy; Neighbor Service Solutions LLC. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, issueType, submitterRole, bookingRef)

	return SendMultipartEmail(cfg, toEmail, subject, plainText, htmlBody)
}

