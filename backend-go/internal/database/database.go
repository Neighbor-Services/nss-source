package database

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	logLevel := logger.Info
	if cfg.Env == "production" {
		logLevel = logger.Error
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	switch cfg.DBDriver {
	case "sqlite":
		dialector = sqlite.Open(cfg.DBDSN)
	case "postgres":
		fallthrough
	default:
		dialector = postgres.Open(cfg.DBDSN)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db
	log.Printf("✓ Database connected successfully (%s)", cfg.DBDriver)

	return db, nil
}

// AutoMigrate migrates all models matching Django and Go database table schemas.
func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		// Accounts & Profiles
		&entity.User{},
		&entity.Profile{},
		&entity.About{},
		&entity.PortfolioItem{},
		&entity.ServicePackage{},
		&entity.PerformanceBadge{},
		&entity.LegalDocument{},
		&entity.OTPVerification{},

		// Services & Proposals
		&entity.Category{},
		&entity.CatalogService{},
		&entity.ServiceRequest{},
		&entity.Proposal{},

		// Interactions & Appointments
		&entity.Favorite{},
		&entity.Review{},
		&entity.Appointment{},
		&entity.Dispute{},

		// Chat & Messaging
		&entity.Conversation{},
		&entity.Message{},
		&entity.ChatBlock{},

		// Notifications
		&entity.Notification{},
		&entity.DeviceToken{},

		// Payments, Wallets & Subscriptions
		&entity.Customer{},
		&entity.Wallet{},
		&entity.WalletTransaction{},
		&entity.PayoutRequest{},
		&entity.SubscriptionPlan{},
		&entity.Subscription{},

		// Moderation & Audit
		&entity.Report{},
		&entity.ProviderVerification{},
		&entity.BackgroundCheck{},
		&entity.ModerationSetting{},
		&entity.AuditLog{},

		// Core Admin, Enterprise & Security
		&entity.FeatureFlag{},
		&entity.AdminRole{},
		&entity.FraudRiskAlert{},
		&entity.NotificationTemplate{},
		&entity.BackupSnapshot{},

		// Public Landing Site CMS
		&entity.HeroSection{},
		&entity.Feature{},
		&entity.Testimonial{},
		&entity.FAQ{},
		&entity.AboutContent{},
		&entity.HowItWorksStep{},
		&entity.SiteStat{},
		&entity.SiteSetting{},
		&entity.ContactMessage{},
		&entity.ResolutionReport{},
		&entity.EmailCampaignLog{},
	)
	if err != nil {
		return err
	}

	createCompositeIndexes(db)
	seedDefaultSubscriptionPlans(db)
	seedDefaultCategoriesAndServices(db)
	seedDefaultLegalDocuments(db)
	return nil
}

func createCompositeIndexes(db *gorm.DB) {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_servicerequest_status_created ON services_servicerequest (status, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_servicerequest_user_status ON services_servicerequest (user_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_payoutrequest_status_created ON payments_payoutrequest (status, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_dispute_status_created ON interactions_dispute (status, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_wallettx_wallet_created ON payments_wallettransaction (wallet_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_appointment_status_date ON interactions_appointment (status, appointment_date)",
		"CREATE INDEX IF NOT EXISTS idx_otp_email_created ON accounts_otpverification (email, created_at DESC)",
	}
	for _, idx := range indexes {
		_ = db.Exec(idx).Error
	}
}

func seedDefaultCategoriesAndServices(db *gorm.DB) {
	var catCount int64
	db.Model(&entity.Category{}).Count(&catCount)
	if catCount > 0 {
		return
	}

	categories := []struct {
		Name        string
		Description string
		Image       string
		Services    []string
	}{
		{
			Name:        "Home Cleaning",
			Description: "Professional residential and commercial cleaning services",
			Image:       "spray-can",
			Services:    []string{"Deep House Cleaning", "Standard Cleaning", "Move-In / Move-Out Cleaning", "Carpet Cleaning", "Window Cleaning"},
		},
		{
			Name:        "Plumbing",
			Description: "Licensed plumbing repair, installation, and maintenance",
			Image:       "wrench",
			Services:    []string{"Pipe Leak Repair", "Drain Unclogging", "Water Heater Installation", "Faucet & Toilet Repair", "Garbage Disposal Fix"},
		},
		{
			Name:        "Electrical",
			Description: "Certified electrical diagnostics, wiring, and fixtures",
			Image:       "bolt",
			Services:    []string{"Lighting & Fixture Installation", "Outlet & Switch Repair", "Circuit Breaker Troubleshooting", "Ceiling Fan Installation", "EV Charger Setup"},
		},
		{
			Name:        "Carpentry & Handyman",
			Description: "General home repair, custom carpentry, and assembly",
			Image:       "hammer",
			Services:    []string{"Furniture Assembly", "Drywall Repair & Patching", "Door & Lock Installation", "Custom Shelving", "Deck & Fence Repair"},
		},
		{
			Name:        "Lawn & Garden",
			Description: "Landscaping, lawn mowing, and garden care",
			Image:       "leaf",
			Services:    []string{"Lawn Mowing & Edging", "Hedge Trimming", "Garden Weeding & Planting", "Tree Pruning", "Irrigation Sprinkler Repair"},
		},
		{
			Name:        "Painting",
			Description: "Interior and exterior painting, staining, and finishing",
			Image:       "paintbrush",
			Services:    []string{"Interior Room Painting", "Exterior House Painting", "Cabinet Refinishing", "Deck Staining", "Wallpaper Removal"},
		},
	}

	for _, c := range categories {
		cat := entity.Category{
			ID:          uuid.New(),
			Name:        c.Name,
			Description: c.Description,
			Image:       c.Image,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := db.Create(&cat).Error; err == nil {
			for _, sName := range c.Services {
				svc := entity.CatalogService{
					ID:          uuid.New(),
					CategoryID:  cat.ID,
					Name:        sName,
					Description: sName,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				_ = db.Create(&svc)
			}
		}
	}
}

func seedDefaultSubscriptionPlans(db *gorm.DB) {
	var count int64
	db.Model(&entity.SubscriptionPlan{}).Count(&count)
	if count > 0 {
		return
	}

	defaultPlans := []entity.SubscriptionPlan{
		{
			ID:                 uuid.New(),
			Name:               "Silver Monthly",
			Tier:               "SILVER",
			Interval:           "month",
			Description:        "Basic provider tier with essential listing and direct client messaging.",
			Price:              19.99,
			Currency:           "USD",
			Features:           entity.JSONSlice{"Standard Profile Placement", "Direct Client Messaging", "Up to 1 Catalog Service"},
			AppleProductID:     "silver_monthly",
			GoogleProductID:    "silver_monthly",
			MaxCatalogServices: 1,
			IsActive:           true,
			DisplayOrder:       1,
		},
		{
			ID:                 uuid.New(),
			Name:               "Gold Monthly",
			Tier:               "GOLD",
			Interval:           "month",
			Description:        "Featured badge, top search placement, and multiple service catalog packages.",
			Price:              38.89,
			Currency:           "USD",
			Features:           entity.JSONSlice{"Featured Provider Badge", "Priority Local Search Placement", "Up to 3 Catalog Services", "Instant Booking Acceptance"},
			AppleProductID:     "gold_monthly",
			GoogleProductID:    "gold_monthly",
			MaxCatalogServices: 3,
			IsActive:           true,
			DisplayOrder:       2,
		},
		{
			ID:                 uuid.New(),
			Name:               "Platinum Monthly",
			Tier:               "PLATINUM",
			Interval:           "month",
			Description:        "Unlimited catalog services, dedicated concierge support, and maximum visibility.",
			Price:              49.99,
			Currency:           "USD",
			Features:           entity.JSONSlice{"Elite Verified Badge", "Top-Tier Homepage Spotlight", "Up to 5 Catalog Services", "Dedicated Concierge Support", "Advanced Analytics"},
			AppleProductID:     "platinum_monthly",
			GoogleProductID:    "platinum_monthly",
			MaxCatalogServices: 5,
			IsActive:           true,
			DisplayOrder:       3,
		},
		{
			ID:                 uuid.New(),
			Name:               "Silver Yearly",
			Tier:               "SILVER",
			Interval:           "year",
			Description:        "Annual basic provider plan with two months free.",
			Price:              199.99,
			Currency:           "USD",
			Features:           entity.JSONSlice{"Standard Profile Placement", "Direct Client Messaging", "Up to 1 Catalog Service", "Save 16% Annually"},
			AppleProductID:     "silver_yearly",
			GoogleProductID:    "silver_yearly",
			MaxCatalogServices: 1,
			IsActive:           true,
			DisplayOrder:       4,
		},
		{
			ID:                 uuid.New(),
			Name:               "Gold Yearly",
			Tier:               "GOLD",
			Interval:           "year",
			Description:        "Annual Gold membership with priority local search and featured badge.",
			Price:              389.99,
			Currency:           "USD",
			Features:           entity.JSONSlice{"Featured Provider Badge", "Priority Local Search Placement", "Up to 3 Catalog Services", "Instant Booking Acceptance", "Save 16% Annually"},
			AppleProductID:     "gold_yearly",
			GoogleProductID:    "gold_yearly",
			MaxCatalogServices: 3,
			IsActive:           true,
			DisplayOrder:       5,
		},
		{
			ID:                 uuid.New(),
			Name:               "Platinum Yearly",
			Tier:               "PLATINUM",
			Interval:           "year",
			Description:        "Annual Platinum access with top spotlight and dedicated support.",
			Price:              499.99,
			Currency:           "USD",
			Features:           entity.JSONSlice{"Elite Verified Badge", "Top-Tier Homepage Spotlight", "Up to 5 Catalog Services", "Dedicated Concierge Support", "Save 16% Annually"},
			AppleProductID:     "platinum_yearly",
			GoogleProductID:    "platinum_yearly",
			MaxCatalogServices: 5,
			IsActive:           true,
			DisplayOrder:       6,
		},
	}

	for _, p := range defaultPlans {
		_ = db.Create(&p)
	}
}

func seedDefaultLegalDocuments(db *gorm.DB) {
	var termsCount int64
	db.Model(&entity.LegalDocument{}).Where("doc_type = ?", "TERMS").Count(&termsCount)
	if termsCount == 0 {
		termsDoc := entity.LegalDocument{
			ID:        uuid.New(),
			DocType:   "TERMS",
			Title:     "Terms of Service",
			Version:   "1.0",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Content: `Welcome to Neighbor Service App ("NSA", "we", "our", or "us"). These Terms of Service ("Terms") govern your access to and use of the Neighbor Service App platform, including our mobile applications, website, and related services (collectively, the "Platform"). By accessing or using the Platform, you agree to be bound by these Terms. If you do not agree to these Terms, you may not use the Platform.

1. PLATFORM OVERVIEW
Neighbor Service App is a technology platform that connects individuals seeking home and personal services ("Seekers") with independent, background-verified service professionals ("Providers"). NSA is a marketplace facilitator and does not directly provide services, employ service providers, or supervise their day-to-day work. All agreements for services are made directly between Seekers and Providers.

2. ELIGIBILITY & REGISTRATION
To use the Platform, you must be at least 18 years of age, legally capable of entering into binding contracts, and provide accurate and complete registration information. You agree to maintain the confidentiality of your account credentials and immediately notify NSA of any unauthorized use or security breaches.

3. IDENTITY VERIFICATION & BACKGROUND SCREENING
Provider safety and trust are core to our community. Providers must complete our identity verification and background screening process (powered by Checkr and accredited screening partners). While NSA conducts screenings, we encourage Seekers to review provider reviews, ratings, and profile qualifications prior to booking.

4. SERVICE REQUESTS, PROPOSALS & BOOKINGS
Seekers may post service requests detailing job scope, location, and requirements. Providers may submit proposals with transparent pricing. A booking is formed when a Seeker accepts a provider proposal or books an appointment. Both parties agree to communicate respectfully and confirm scheduling details through the Platform.

5. PAYMENTS & ESCROW PROTECTION
All transactions must be processed securely through the Platform's integrated payment system (powered by Stripe). Payments are held securely in escrow until service completion and satisfaction verification. Direct cash transactions off-platform are strictly prohibited and forfeit user protection, guarantees, and dispute resolution.

6. RATINGS, REVIEWS & REPUTATION
Seekers and Providers may leave honest, constructive ratings and reviews following completed appointments. Reviews must reflect genuine firsthand experiences and comply with our community guidelines. Fraudulent, abusive, defamatory, or retaliatory reviews are subject to immediate removal.

7. SUBSCRIPTIONS & PROVIDER TIERS
Provider subscriptions (Silver, Gold, Platinum) offer enhanced profile visibility, direct client messaging, and expanded service catalog listings. Subscriptions renew automatically unless cancelled prior to the end of the billing period through your account settings or app store provider.

8. CANCELLATION, RESCHEDULING & REFUNDS
Cancellations made more than 24 hours prior to a scheduled appointment are eligible for a full refund. Cancellations within 24 hours may incur a standard cancellation fee to compensate the provider for reserved time. In cases of provider non-arrival or unsatisfactory service, funds in escrow may be refunded following dispute review.

9. USER CONDUCT & PROHIBITED ACTIVITIES
You agree not to engage in unlawful conduct, harassment, discrimination, property damage, fraud, or circumventing platform payments. Violations will result in immediate suspension, account termination, and potential legal action.

10. LIMITATION OF LIABILITY
NSA provides the Platform on an "as is" and "as available" basis. To the maximum extent permitted by applicable law, NSA shall not be liable for indirect, incidental, special, consequential, or punitive damages arising out of your use of the Platform or interactions between users.

11. DISPUTE RESOLUTION & ARBITRATION
Any dispute arising out of or relating to these Terms or the Platform shall be resolved through good-faith negotiation, followed if necessary by binding individual arbitration under the rules of the American Arbitration Association (AAA), rather than in court.

12. CONTACT INFORMATION
For questions regarding these Terms, please contact our Legal & Compliance Team at support@neighborservice.com or visit https://neighborservice.com.`,
		}
		_ = db.Create(&termsDoc)
	}

	var privacyCount int64
	db.Model(&entity.LegalDocument{}).Where("doc_type = ?", "PRIVACY").Count(&privacyCount)
	if privacyCount == 0 {
		privacyDoc := entity.LegalDocument{
			ID:        uuid.New(),
			DocType:   "PRIVACY",
			Title:     "Privacy Policy",
			Version:   "1.0",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Content: `Neighbor Service App ("NSA", "we", "our", or "us") is dedicated to protecting your privacy and personal data. This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you access or use our mobile applications, website, and related services.

1. INFORMATION WE COLLECT
We collect information you provide directly to us when creating an account, updating your profile, submitting service requests, completing background checks, or contacting support. This includes:
- Personal Identifiers: Name, email address, phone number, physical address, and profile photo.
- Professional Credentials: Government ID, SSN/tax ID (for provider verification and Checkr background checks), licenses, certifications, and portfolio media.
- Transactional & Financial Data: Payment card details and bank account payout info (processed securely via Stripe; NSA does not store complete card numbers).
- Location Data: Precise or approximate GPS location (with your permission) to match local seekers and providers and facilitate navigation.
- Communications & Chat: In-app chat messages, audio/video consultation session metadata, and customer service inquiries.

2. HOW WE USE YOUR INFORMATION
We use collected data to:
- Facilitate matching between Seekers and nearby qualified Providers.
- Process secure payments, escrow holds, and provider payouts.
- Conduct identity verification and background screening to maintain platform trust.
- Provide real-time location tracking for active service appointments.
- Send transactional notifications, booking updates, and security alerts.
- Detect fraud, abuse, and safety violations.
- Improve and personalize our platform features and user experience.

3. INFORMATION SHARING & DISCLOSURE
We respect your privacy and do not sell your personal data. We share information only in limited circumstances:
- Between Seekers and Providers: Contact info, job location, and reviews necessary to coordinate and complete booked services.
- Service Providers & Partners: Trusted third-party vendors including Stripe (payments), Checkr (background checks), Agora (in-app consultations), and Firebase (push notifications).
- Legal & Safety Compliance: When required by law, subpoena, or to protect the vital safety of our users and the public.

4. DATA SECURITY & RETENTION
We employ industry-standard administrative, technical, and physical safeguards (including AES-256 encryption in transit and at rest, tokenized auth, and regular security audits) to protect your personal information. We retain personal data only as long as necessary to fulfill the purposes outlined in this policy or comply with legal obligations.

5. YOUR PRIVACY RIGHTS & CHOICES
Depending on your jurisdiction, you have the right to:
- Access and download a copy of your personal data (GDPR data export).
- Correct or update inaccurate account information.
- Delete your account and associated personal data.
- Opt out of promotional communications and manage device notification permissions.
- Control location sharing permissions through your mobile operating system settings.

6. CHILDREN'S PRIVACY
Our Platform is not directed to children under 18 years of age. We do not knowingly collect personal data from minors. If we discover that a minor has created an account, we will promptly delete it.

7. CHANGES TO THIS PRIVACY POLICY
We may update this Privacy Policy from time to time to reflect changes in our practices or legal obligations. We will notify you of material changes through app notices or email prior to the effective date.

8. CONTACT US
If you have questions or concerns regarding this Privacy Policy or our data handling practices, please reach out to our Data Protection Officer at privacy@neighborservice.com.`,
		}
		_ = db.Create(&privacyDoc)
	}
}
