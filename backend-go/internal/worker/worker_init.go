package worker

func init() {
	GlobalRegistry.RegisterWorker("Fraud Detection Engine", "Security & Risk", "10m", "Analyzes suspicious high-velocity payouts, chargebacks, and rapid account logins.")
	GlobalRegistry.RegisterWorker("Payout Rails Processor", "Finance & Escrow", "24h", "Dispatches scheduled Stripe Connect direct payouts to verified provider bank accounts.")
	GlobalRegistry.RegisterWorker("Wallet Reconciliation Worker", "Finance & Escrow", "24h", "Double-entry ledger reconciliation between local ledger entries and Stripe escrow balances.")
	GlobalRegistry.RegisterWorker("Dispute Escalation Worker", "Moderation", "30m", "Monitors unanswered dispute tickets and auto-escalates to senior staff based on SLA.")
	GlobalRegistry.RegisterWorker("Appointment No-Show Worker", "Operations", "15m", "Detects appointments unattended past grace period, triggers penalties & auto-refunds.")
	GlobalRegistry.RegisterWorker("Review Reminder Worker", "Engagement", "2h", "Sends push & in-app feedback notifications to seekers 2 hours after service completion.")
	GlobalRegistry.RegisterWorker("Provider Ranking Worker", "Algorithm & Search", "6h", "Computes dynamic composite ranking scores based on reviews, response speed, and SLA history.")
	GlobalRegistry.RegisterWorker("Nearby Request Dispatch Worker", "Realtime Push", "5m", "Calculates geospatial haversine match to broadcast new requests to active radius providers.")
	GlobalRegistry.RegisterWorker("Stale Request Worker", "Operations", "1h", "Auto-expires unaccepted booking requests older than 24 hours and releases funds.")
	GlobalRegistry.RegisterWorker("Performance Badge Worker", "Engagement", "6h", "Awards 'Top Rated', 'Fast Responder', and 'Neighborhood Hero' badges to qualified providers.")
	GlobalRegistry.RegisterWorker("Subscription Expiry Worker", "Subscriptions", "12h", "Enforces tier access limits, grace periods, and alerts providers before renewal.")
	GlobalRegistry.RegisterWorker("Automated Email Marketing Worker", "Marketing", "24h", "Dispatches personalized weekly provider digest and neighborhood promotional campaigns.")
	GlobalRegistry.RegisterWorker("Inactive User Worker", "User Lifecycle", "24h", "Sends re-engagement reminders to accounts with no activity in 45+ days.")
	GlobalRegistry.RegisterWorker("Expired OTP Cleanup Worker", "Security", "1h", "Flushes expired SMS OTPs, password reset nonces, and unverified registration codes.")
	GlobalRegistry.RegisterWorker("Database Backup Worker", "Infrastructure", "24h", "Generates daily automated PostgreSQL database snapshots with 14-day retention rotation.")
}
