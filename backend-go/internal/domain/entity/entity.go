package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONSlice represents a string array stored as JSON in the database.
type JSONSlice []string

func (j JSONSlice) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONSlice) Scan(value interface{}) error {
	if value == nil {
		*j = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, okStr := value.(string)
		if !okStr {
			return errors.New("type assertion failed")
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, j)
}

// JSONMap represents a map stored as JSON in the database.
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, okStr := value.(string)
		if !okStr {
			return errors.New("type assertion failed")
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, j)
}

// ─── DOMAIN ENTITIES ────────────────────────────────────────────────────────

type User struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Email            string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password         string         `gorm:"size:255;not null" json:"-"`
	IsActive         bool           `gorm:"default:true" json:"is_active"`
	IsStaff          bool           `gorm:"default:false" json:"is_staff"`
	IsSuperuser      bool           `gorm:"default:false" json:"is_superuser"`
	IsVerified       bool           `gorm:"default:false" json:"is_verified"`
	TwoFactorSecret  string         `gorm:"size:64" json:"-"`
	TwoFactorEnabled bool           `gorm:"default:false" json:"two_factor_enabled"`
	AdminRoleID      *uuid.UUID     `gorm:"type:uuid;index" json:"admin_role_id,omitempty"`
	OTPCode          string         `gorm:"size:6" json:"otp_code,omitempty"`
	OTPExpiry        *time.Time     `json:"otp_expiry,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	LastLogin        *time.Time     `json:"last_login,omitempty"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	Profile         *Profile          `gorm:"foreignKey:UserID" json:"profile,omitempty"`
	AdminRole       *AdminRole        `gorm:"foreignKey:AdminRoleID" json:"admin_role,omitempty"`
	Wallet          *Wallet           `gorm:"foreignKey:UserID" json:"wallet,omitempty"`
	BackgroundCheck *BackgroundCheck  `gorm:"foreignKey:ProviderID" json:"background_check,omitempty"`
}

func (User) TableName() string { return "accounts_user" }

type Profile struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID               uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	SubscriptionTier     string     `gorm:"size:10;default:'NONE'" json:"subscription_tier"`         // NONE, SILVER, GOLD, PLATINUM
	SubscriptionInterval string     `gorm:"size:10;default:'none'" json:"subscription_interval"`     // none, month, year
	PreferredPaymentMode string     `gorm:"size:10;default:'ON_SITE'" json:"preferred_payment_mode"` // IN_APP, ON_SITE
	FirstName            string     `gorm:"size:100" json:"first_name"`
	LastName             string     `gorm:"size:100" json:"last_name"`
	DateOfBirth          *time.Time `json:"date_of_birth,omitempty"`
	Service              string     `gorm:"size:255" json:"service"`
	Country              string     `gorm:"size:100" json:"country"`
	State                string     `gorm:"size:100" json:"state"`
	ZipCode              string     `gorm:"size:20" json:"zip_code"`
	Gender               string     `gorm:"size:10" json:"gender"`
	CountryCode          string     `gorm:"size:10" json:"country_code"`
	Phone                string     `gorm:"size:20" json:"phone"`
	City                 string     `gorm:"size:100" json:"city"`
	Address              string     `gorm:"type:text" json:"address"`
	ProfilePicture       string     `gorm:"size:500" json:"profile_picture"`
	UserType             string     `gorm:"size:10;default:'SEEKER'" json:"user_type"` // SEEKER, PROVIDER
	Longitude            float64    `gorm:"type:numeric(100,50);default:0" json:"longitude"`
	Latitude             float64    `gorm:"type:numeric(100,50);default:0" json:"latitude"`
	DeviceToken          string     `gorm:"size:255" json:"device_token"`
	AverageRating        float64    `gorm:"default:0.00" json:"average_rating"`
	TotalReviews         int        `gorm:"default:0" json:"total_reviews"`
	IsIdentityVerified   bool       `gorm:"default:false" json:"is_identity_verified"`
	Bio                  string     `gorm:"type:text" json:"bio"`
	IsOnline             bool       `gorm:"default:false" json:"is_online"`
	LastSeen             time.Time  `gorm:"autoUpdateTime" json:"last_seen"`
	StreakCount          int        `gorm:"default:0" json:"streak_count"`
	LastCheckIn          *time.Time `json:"last_check_in,omitempty"`
	XP                   int        `gorm:"default:0" json:"xp"`
	Level                int        `gorm:"default:1" json:"level"`
	NeighborScore        int        `gorm:"default:500" json:"neighbor_score"`
	MaxCatalogServices   int        `gorm:"default:1" json:"max_catalog_services"`
	CreatedAt            time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	User            *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CatalogServices []CatalogService `gorm:"many2many:accounts_profile_catalog_services;" json:"catalog_services,omitempty"`
	PortfolioItems  []PortfolioItem  `gorm:"foreignKey:ProfileID" json:"portfolio_items,omitempty"`
	ServicePackages []ServicePackage `gorm:"foreignKey:ProfileID" json:"service_packages,omitempty"`

	// Serializer & Frontend Parity Fields
	CatalogServiceName  string   `gorm:"-" json:"catalog_service_name,omitempty"`
	CatalogServiceNames []string `gorm:"-" json:"catalog_service_names,omitempty"`
	CatalogServiceIDs   []string `gorm:"-" json:"catalog_service_ids,omitempty"`
	ProfilePictureURL   string   `gorm:"-" json:"profile_picture_url,omitempty"`
	ReviewsReceived     []Review `gorm:"-" json:"reviews_received,omitempty"`
	MatchScore          float64  `gorm:"-" json:"match_score,omitempty"`
	MatchPercentage     int      `gorm:"-" json:"match_percentage,omitempty"`
	MatchReason         string   `gorm:"-" json:"match_reason,omitempty"`
}

func (Profile) TableName() string { return "accounts_profile" }

func (p *Profile) RecordActivity() bool {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if p.LastCheckIn != nil {
		lastDate := time.Date(p.LastCheckIn.Year(), p.LastCheckIn.Month(), p.LastCheckIn.Day(), 0, 0, 0, 0, time.UTC)
		if lastDate.Equal(today) {
			return false
		}
		yesterday := today.AddDate(0, 0, -1)
		if lastDate.Equal(yesterday) {
			p.StreakCount++
		} else {
			p.StreakCount = 1
		}
	} else {
		p.StreakCount = 1
	}
	p.LastCheckIn = &today
	return true
}

func (p *Profile) AwardXP(amount int) {
	p.XP += amount
	newLevel := (p.XP / 1000) + 1
	if newLevel > p.Level {
		p.Level = newLevel
	}
}

func (p *Profile) PriorityScore() float64 {
	switch p.SubscriptionTier {
	case "PLATINUM":
		return 1.5
	case "GOLD":
		return 1.2
	case "SILVER":
		return 1.1
	default:
		return 1.0
	}
}

func (p *Profile) GetCommissionRate() float64 {
	switch p.SubscriptionTier {
	case "PLATINUM":
		return 0.05
	case "GOLD":
		return 0.10
	case "SILVER":
		return 0.15
	default:
		return 0.20
	}
}

func (p *Profile) GetMaxCatalogServices() int {
	switch p.SubscriptionTier {
	case "PLATINUM":
		return 10
	case "GOLD":
		return 5
	case "SILVER":
		return 3
	default:
		return 1
	}
}

func (p *Profile) EnrichCatalogServices() {
	if len(p.CatalogServices) > 0 {
		p.CatalogServiceIDs = make([]string, 0, len(p.CatalogServices))
		p.CatalogServiceNames = make([]string, 0, len(p.CatalogServices))
		for _, cs := range p.CatalogServices {
			p.CatalogServiceIDs = append(p.CatalogServiceIDs, cs.ID.String())
			if cs.Name != "" {
				p.CatalogServiceNames = append(p.CatalogServiceNames, cs.Name)
			}
		}
		if len(p.CatalogServiceNames) > 0 {
			p.CatalogServiceName = p.CatalogServiceNames[0]
			if p.Service == "" {
				p.Service = strings.Join(p.CatalogServiceNames, ", ")
			}
		}
	}
}

type About struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Name            string    `gorm:"size:255;not null" json:"name"`
	Address         string    `gorm:"type:text" json:"address"`
	CountryCode     string    `gorm:"size:5" json:"country_code"`
	Specification   string    `gorm:"size:255" json:"specification"`
	Description     string    `gorm:"type:text" json:"description"`
	ImageURLs       JSONSlice `gorm:"type:text" json:"image_urls"`
	ExperienceYears int       `gorm:"default:0" json:"experience_years"`
	Skills          JSONSlice `gorm:"type:text" json:"skills"`
	Education       string    `gorm:"type:text" json:"education"`
	Languages       JSONSlice `gorm:"type:text" json:"languages"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (About) TableName() string { return "accounts_about" }

type PortfolioItem struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProfileID   uuid.UUID `gorm:"type:uuid;index;not null" json:"profile_id"`
	Image       string    `gorm:"size:500;not null" json:"image"`
	Description string    `gorm:"size:255" json:"description"`
	Tags        JSONSlice `gorm:"type:text" json:"tags"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Serializer Parity
	ImageURL string `gorm:"-" json:"image_url,omitempty"`
}

func (PortfolioItem) TableName() string { return "accounts_portfolioitem" }

type ServicePackage struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProfileID    uuid.UUID `gorm:"type:uuid;index;not null" json:"profile_id"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	Price        float64   `gorm:"not null" json:"price"`
	Description  string    `gorm:"type:text" json:"description"`
	Revisions    int       `gorm:"default:0" json:"revisions"`
	DeliveryTime int       `gorm:"default:1" json:"delivery_time"`
	Features     JSONSlice `gorm:"type:text" json:"features"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ServicePackage) TableName() string { return "accounts_servicepackage" }

type PerformanceBadge struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProfileID   uuid.UUID `gorm:"type:uuid;index;not null" json:"profile_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	IconType    string    `gorm:"size:50;not null" json:"icon_type"`
	AwardedAt   time.Time `gorm:"autoCreateTime" json:"awarded_at"`
	Description string    `gorm:"type:text" json:"description"`
}

func (PerformanceBadge) TableName() string { return "accounts_performancebadge" }

type LegalDocument struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DocType   string    `gorm:"size:10;not null" json:"doc_type"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Version   string    `gorm:"size:20;default:'1.0'" json:"version"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (LegalDocument) TableName() string { return "accounts_legaldocument" }

type OTPVerification struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email     string    `gorm:"size:255;index;not null" json:"email"`
	OTPCode   string    `gorm:"size:10;not null" json:"otp_code"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsUsed    bool      `gorm:"default:false" json:"is_used"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (OTPVerification) TableName() string { return "accounts_otpverification" }

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Image       string    `gorm:"size:500" json:"image"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Services []CatalogService `gorm:"foreignKey:CategoryID" json:"services,omitempty"`

	// Serializer Parity
	ImageURL string `gorm:"-" json:"image_url,omitempty"`
}

func (Category) TableName() string { return "services_category" }

type CatalogService struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CategoryID  uuid.UUID `gorm:"type:uuid;index;not null" json:"category_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	BasePrice   *float64  `json:"base_price,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (CatalogService) TableName() string { return "services_catalogservice" }

type ServiceRequest struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID               uuid.UUID  `gorm:"type:uuid;index;not null" json:"user"`
	TargetProviderID     *uuid.UUID `gorm:"type:uuid;index" json:"target_provider,omitempty"`
	CatalogServiceID     *uuid.UUID `gorm:"type:uuid;index" json:"catalog_service,omitempty"`
	Title                string     `gorm:"size:255;not null" json:"title"`
	Description          string     `gorm:"type:text;not null" json:"description"`
	Price                *float64   `json:"price,omitempty"`
	Status               string     `gorm:"size:20;default:'OPEN'" json:"status"`
	PreferredPaymentMode string     `gorm:"size:20;default:'IN_APP'" json:"preferred_payment_mode"`
	ServiceType          string     `gorm:"size:255" json:"service_type,omitempty"`
	WithImage            bool       `gorm:"default:false" json:"with_image"`
	Image                string     `gorm:"size:500" json:"image,omitempty"`
	Longitude            *float64   `gorm:"type:numeric(100,50)" json:"longitude,omitempty"`
	Latitude             *float64   `gorm:"type:numeric(100,50)" json:"latitude,omitempty"`
	ScheduledTime        *time.Time     `json:"scheduled_time,omitempty"`
	CreatedAt            time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`

	User           *User           `gorm:"foreignKey:UserID" json:"user_details,omitempty"`
	TargetProvider *User           `gorm:"foreignKey:TargetProviderID" json:"target_provider_details,omitempty"`
	CatalogService *CatalogService `gorm:"foreignKey:CatalogServiceID" json:"catalog_service_details,omitempty"`
	Proposals      []Proposal      `gorm:"foreignKey:RequestID" json:"proposals,omitempty"`

	// Serializer & Frontend Parity Fields
	UserEmail          string      `gorm:"-" json:"user_email,omitempty"`
	UserProfile        *Profile    `gorm:"-" json:"user_profile,omitempty"`
	CatalogServiceName string      `gorm:"-" json:"catalog_service_name,omitempty"`
	ProposalsCount     int         `gorm:"-" json:"proposals_count"`
	Approved           bool        `gorm:"-" json:"approved"`
	ApprovedUser       interface{} `gorm:"-" json:"approved_user,omitempty"`
	AppointmentID      *string     `gorm:"-" json:"appointment_id,omitempty"`
	IsFunded           bool        `gorm:"-" json:"is_funded"`
	ImageURL           string      `gorm:"-" json:"image_url,omitempty"`
	Distance           *float64    `gorm:"-" json:"distance,omitempty"`
}

func (ServiceRequest) TableName() string { return "services_servicerequest" }

type Proposal struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RequestID  uuid.UUID `gorm:"type:uuid;index;not null" json:"request"`
	ProviderID uuid.UUID `gorm:"type:uuid;index;not null" json:"provider"`
	IsApproved bool      `gorm:"default:false" json:"is_approved"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Request  *ServiceRequest `gorm:"foreignKey:RequestID" json:"request_details,omitempty"`
	Provider *User           `gorm:"foreignKey:ProviderID" json:"provider_details,omitempty"`

	// Serializer Parity
	ProviderProfile *Profile `gorm:"-" json:"provider_profile,omitempty"`
	ProviderEmail   string   `gorm:"-" json:"provider_email,omitempty"`
	SeekerProfile   *Profile `gorm:"-" json:"seeker_profile,omitempty"`
}

func (Proposal) TableName() string { return "services_proposal" }

type Favorite struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;index;not null" json:"user"`
	FavoriteUserID uuid.UUID `gorm:"type:uuid;index;not null" json:"favorite_user"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`

	User         *User `gorm:"foreignKey:UserID" json:"user_details,omitempty"`
	FavoriteUser *User `gorm:"foreignKey:FavoriteUserID" json:"favorite_user_details,omitempty"`
}

func (Favorite) TableName() string { return "interactions_favorite" }

type Review struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProviderID uuid.UUID `gorm:"type:uuid;index;not null" json:"provider"`
	ReviewerID uuid.UUID `gorm:"type:uuid;index;not null" json:"reviewer"`
	Rating     float64   `gorm:"type:numeric(3,1);not null" json:"rating"`
	Comment    string    `gorm:"type:text" json:"comment"`
	IsHidden   bool      `gorm:"default:false" json:"is_hidden"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Provider *User `gorm:"foreignKey:ProviderID" json:"provider_details,omitempty"`
	Reviewer *User `gorm:"foreignKey:ReviewerID" json:"reviewer_details,omitempty"`

	// Serializer Parity
	ReviewerProfile    *Profile `gorm:"-" json:"reviewer_profile,omitempty"`
	ProviderProfile    *Profile `gorm:"-" json:"provider_profile,omitempty"`
	CreatedAtFormatted string   `gorm:"-" json:"created_at_formatted,omitempty"`
}

func (Review) TableName() string { return "interactions_review" }

type Appointment struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SeekerID            uuid.UUID  `gorm:"type:uuid;index;not null" json:"seeker"`
	ProviderID          uuid.UUID  `gorm:"type:uuid;index;not null" json:"provider"`
	Title               string     `gorm:"size:255" json:"title"`
	Description         string     `gorm:"type:text" json:"description"`
	AppointmentDate     *time.Time `json:"appointment_date,omitempty"`
	Status              string     `gorm:"size:20;default:'SCHEDULED'" json:"status"`
	PaymentMode         string     `gorm:"size:20;default:'IN_APP'" json:"payment_mode"`
	IsConsultation      bool       `gorm:"default:false" json:"is_consultation"`
	ConsultationChannel string     `gorm:"size:255" json:"consultation_channel,omitempty"`
	IsFunded            bool       `gorm:"default:false" json:"is_funded"`
	PaymentIntentID     string     `gorm:"size:255" json:"payment_intent_id,omitempty"`
	TotalPrice          float64    `gorm:"default:0.00" json:"total_price"`
	ServiceRequestID    *uuid.UUID `gorm:"type:uuid;index" json:"service_request,omitempty"`
	ProposalID          *uuid.UUID `gorm:"type:uuid;index" json:"proposal,omitempty"`
	ReminderDaySent     bool           `gorm:"default:false" json:"reminder_day_sent"`
	ReminderHourSent    bool           `gorm:"default:false" json:"reminder_hour_sent"`
	ReviewReminderSent  bool           `gorm:"default:false" json:"review_reminder_sent"`
	NoShowProcessed     bool           `gorm:"default:false" json:"no_show_processed"`
	Version             int            `gorm:"default:1" json:"version"`
	SecretCode          string         `gorm:"size:6" json:"secret_code,omitempty"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	Seeker         *User           `gorm:"foreignKey:SeekerID" json:"seeker_details,omitempty"`
	Provider       *User           `gorm:"foreignKey:ProviderID" json:"provider_details,omitempty"`
	ServiceRequest *ServiceRequest `gorm:"foreignKey:ServiceRequestID" json:"service_request_details,omitempty"`
	Proposal       *Proposal       `gorm:"foreignKey:ProposalID" json:"proposal_details,omitempty"`

	// Serializer Parity
	SeekerProfile         *Profile               `gorm:"-" json:"seeker_profile,omitempty"`
	ProviderProfile       *Profile               `gorm:"-" json:"provider_profile,omitempty"`
	ServiceRequestPayload map[string]interface{} `gorm:"-" json:"service_request_details_payload,omitempty"`
}

func (Appointment) TableName() string { return "interactions_appointment" }

type Dispute struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	RaisedByID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"raised_by"`
	DefendantID     *uuid.UUID `gorm:"type:uuid;index" json:"defendant,omitempty"`
	AppointmentID   *uuid.UUID `gorm:"type:uuid;index" json:"appointment,omitempty"`
	Reason          string     `gorm:"size:255;not null" json:"reason"`
	Description     string     `gorm:"type:text;not null" json:"description"`
	Status          string     `gorm:"size:20;default:'OPEN'" json:"status"`
	ResolutionNotes string     `gorm:"type:text" json:"resolution_notes,omitempty"`
	Evidence        string     `gorm:"size:500" json:"evidence,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	RaisedBy    *User        `gorm:"foreignKey:RaisedByID" json:"raised_by_user,omitempty"`
	Defendant   *User        `gorm:"foreignKey:DefendantID" json:"defendant_user,omitempty"`
	Appointment *Appointment `gorm:"foreignKey:AppointmentID" json:"appointment_details,omitempty"`

	// Serializer Parity
	RaisedByDetails  *Profile `gorm:"-" json:"raised_by_details,omitempty"`
	DefendantDetails *Profile `gorm:"-" json:"defendant_details,omitempty"`
	EvidenceURL      string   `gorm:"-" json:"evidence_url,omitempty"`
}

func (Dispute) TableName() string { return "interactions_dispute" }

type Conversation struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Participants []User    `gorm:"many2many:chat_conversation_participants;" json:"participants,omitempty"`
	Messages     []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

func (Conversation) TableName() string { return "chat_conversation" }

type Message struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ConversationID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"conversation"`
	SenderID          uuid.UUID  `gorm:"type:uuid;index;not null" json:"sender"`
	Content           string     `gorm:"type:text" json:"content"`
	Message           string     `gorm:"type:text" json:"message"`
	WithImage         bool       `gorm:"default:false" json:"with_image"`
	IsCalender        bool       `gorm:"default:false" json:"is_calender"`
	WithImageAndText  bool       `gorm:"default:false" json:"with_image_and_text"`
	CalenderDate      *time.Time `json:"calender_date,omitempty"`
	MediaURL          string     `gorm:"size:500" json:"media_url,omitempty"`
	Image             string     `gorm:"size:500" json:"image,omitempty"`
	FileName          string     `gorm:"size:255" json:"file_name,omitempty"`
	IsDelivered       bool       `gorm:"default:false" json:"is_delivered"`
	IsSeen            bool       `gorm:"default:false" json:"is_seen"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`

	Sender *User `gorm:"foreignKey:SenderID" json:"sender_details,omitempty"`

	// Serializer Parity
	SenderEmail string  `gorm:"-" json:"sender_email,omitempty"`
	ChatRoomID  string  `gorm:"-" json:"chat_room_id,omitempty"`
	Read        bool    `gorm:"-" json:"read"`
	Receiver    *string `gorm:"-" json:"receiver,omitempty"`
}

func (Message) TableName() string { return "chat_message" }

type ChatBlock struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	BlockerID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"blocker"`
	BlockedID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"blocked"`
	ConversationID *uuid.UUID `gorm:"type:uuid;index" json:"conversation,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (ChatBlock) TableName() string { return "chat_chatblock" }

type Notification struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID           uuid.UUID  `gorm:"type:uuid;index;not null" json:"user"`
	SenderID         *uuid.UUID `gorm:"type:uuid;index" json:"sender,omitempty"`
	NotificationType string     `gorm:"size:20;not null" json:"notification_type"`
	Title            string     `gorm:"size:255;not null" json:"title"`
	Message          string     `gorm:"type:text;not null" json:"message"`
	Data             JSONMap    `gorm:"type:text" json:"data"`
	IsRead           bool       `gorm:"default:false" json:"is_read"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`

	Sender *User `gorm:"foreignKey:SenderID" json:"sender_details,omitempty"`

	// Serializer Parity
	SenderProfile *Profile `gorm:"-" json:"sender_profile,omitempty"`
}

func (Notification) TableName() string { return "notifications_notification" }

type DeviceToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user"`
	Token     string    `gorm:"size:255;uniqueIndex;not null" json:"token"`
	Platform  string    `gorm:"size:10;not null" json:"platform"`
	DeviceID  string    `gorm:"size:255" json:"device_id,omitempty"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (DeviceToken) TableName() string { return "notifications_devicetoken" }

type Customer struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID               uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user"`
	StripeCustomerID     string    `gorm:"size:255" json:"stripe_customer_id"`
	StripeAccountID      string    `gorm:"size:255" json:"stripe_account_id"`
	DefaultPaymentMethod string    `gorm:"size:255" json:"default_payment_method"`
	EphemeralSecret      string    `gorm:"size:255" json:"ephemeral_secret"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Customer) TableName() string { return "payments_customer" }

type Wallet struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user"`
	Balance         float64   `gorm:"default:0.00" json:"balance"`
	Currency        string    `gorm:"size:3;default:'USD'" json:"currency"`
	StripeConnectID string    `gorm:"size:100" json:"stripe_connect_id"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user_details,omitempty"`
}

func (Wallet) TableName() string { return "payments_wallet" }

type WalletTransaction struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	WalletID        uuid.UUID `gorm:"type:uuid;index;not null" json:"wallet"`
	Amount          float64   `gorm:"not null" json:"amount"`
	TransactionType string    `gorm:"size:10;not null" json:"transaction_type"`
	Description     string    `gorm:"size:255;not null" json:"description"`
	Status          string    `gorm:"size:20;default:'COMPLETED'" json:"status"`
	ReferenceID     string    `gorm:"size:100" json:"reference_id"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`

	Wallet *Wallet `gorm:"foreignKey:WalletID" json:"wallet_details,omitempty"`
}

func (WalletTransaction) TableName() string { return "payments_wallettransaction" }

type PayoutRequest struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	WalletID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"wallet"`
	Amount      float64    `gorm:"not null" json:"amount"`
	Status      string     `gorm:"size:20;default:'PENDING'" json:"status"`
	AdminNotes  string     `gorm:"type:text" json:"admin_notes,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`

	Wallet *Wallet `gorm:"foreignKey:WalletID" json:"wallet_details,omitempty"`
}

func (PayoutRequest) TableName() string { return "payments_payoutrequest" }

type SubscriptionPlan struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name               string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Tier               string    `gorm:"size:10;default:'SILVER'" json:"tier"`
	Interval           string    `gorm:"size:10;default:'month'" json:"interval"`
	Description        string    `gorm:"type:text" json:"description"`
	Price              float64   `gorm:"not null" json:"price"`
	Currency           string    `gorm:"size:3;default:'USD'" json:"currency"`
	Features           JSONSlice `gorm:"type:text" json:"features"`
	AppleProductID     string    `gorm:"size:255" json:"apple_product_id"`
	GoogleProductID    string    `gorm:"size:255" json:"google_product_id"`
	MaxCatalogServices int       `gorm:"default:1" json:"max_catalog_services"`
	IsActive           bool      `gorm:"default:true" json:"is_active"`
	DisplayOrder       int       `gorm:"default:0" json:"display_order"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Serializer Parity
	FormattedPrice string `gorm:"-" json:"formatted_price,omitempty"`
}

func (SubscriptionPlan) TableName() string { return "payments_subscriptionplan" }

type Subscription struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID             uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"user"`
	PlanID             *uuid.UUID `gorm:"type:uuid;index" json:"plan,omitempty"`
	StoreTransactionID string     `gorm:"size:512" json:"store_transaction_id"`
	NextPayment        *time.Time `json:"next_payment,omitempty"`
	IsActive           bool       `gorm:"default:false" json:"is_active"`
	IsSandbox          bool       `gorm:"default:false" json:"is_sandbox"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	User *User             `gorm:"foreignKey:UserID" json:"user_details,omitempty"`
	Plan *SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan_details,omitempty"`

	// Serializer Parity
	PlanPrice string `gorm:"-" json:"plan_price,omitempty"`
}

func (Subscription) TableName() string { return "payments_subscription" }

// Aliases for Clean Architecture naming consistency
type Portfolio = PortfolioItem
type UserSubscription = Subscription
type PaymentTransaction = WalletTransaction

type Report struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ReporterID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"reporter"`
	ReportedUserID *uuid.UUID `gorm:"type:uuid;index" json:"reported_user,omitempty"`
	ResourceType   string     `gorm:"size:100" json:"resource_type"`
	ResourceID     string     `gorm:"size:100" json:"resource_id"`
	Reason         string     `gorm:"type:text;not null" json:"reason"`
	Status         string     `gorm:"size:20;default:'PENDING'" json:"status"`
	AdminNote      string     `gorm:"type:text" json:"admin_note,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Reporter     *User `gorm:"foreignKey:ReporterID" json:"reporter_details,omitempty"`
	ReportedUser *User `gorm:"foreignKey:ReportedUserID" json:"reported_user_details,omitempty"`

	// Serializer Parity
	ReporterEmail string `gorm:"-" json:"reporter_email,omitempty"`
}

func (Report) TableName() string { return "moderation_report" }

type ProviderVerification struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProviderID    uuid.UUID `gorm:"type:uuid;index;not null" json:"provider"`
	DocumentFront string    `gorm:"size:500;not null" json:"document_front"`
	DocumentBack  string    `gorm:"size:500" json:"document_back,omitempty"`
	Status        string    `gorm:"size:20;default:'PENDING'" json:"status"`
	ReviewerNotes string    `gorm:"type:text" json:"reviewer_notes,omitempty"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Provider *User `gorm:"foreignKey:ProviderID" json:"provider_details,omitempty"`

	// Serializer Parity
	DocumentFrontURL string `gorm:"-" json:"document_front_url,omitempty"`
	DocumentBackURL  string `gorm:"-" json:"document_back_url,omitempty"`
}

func (ProviderVerification) TableName() string { return "moderation_providerverification" }

type BackgroundCheck struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ProviderID         uuid.UUID  `gorm:"type:uuid;index;not null" json:"provider"`
	CheckrCandidateID  string     `gorm:"size:255" json:"checkr_candidate_id,omitempty"`
	CheckrReportID     string     `gorm:"size:255;index" json:"checkr_report_id,omitempty"`
	CheckrInvitationID string     `gorm:"size:255" json:"checkr_invitation_id,omitempty"`
	InvitationURL      string     `gorm:"size:1024" json:"invitation_url,omitempty"`
	Package            string     `gorm:"size:100;default:'tasker_standard'" json:"package"`
	Status             string     `gorm:"size:30;default:'pending';index" json:"status"`
	Adjudication       string     `gorm:"size:30" json:"adjudication,omitempty"`
	Result             JSONMap    `gorm:"type:text" json:"result,omitempty"`
	LastSyncedAt       *time.Time `json:"last_synced_at,omitempty"`
	SyncAttemptCount   int        `gorm:"default:0" json:"sync_attempt_count"`
	Notes              string     `gorm:"type:text" json:"notes,omitempty"`
	PaymentIntentID    string     `gorm:"size:255;uniqueIndex" json:"payment_intent_id,omitempty"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Provider *User `gorm:"foreignKey:ProviderID" json:"provider_details,omitempty"`

	// Serializer Parity
	ProviderEmail string `gorm:"-" json:"provider_email,omitempty"`
	IsClear       bool   `gorm:"-" json:"is_clear"`
	IsTerminal    bool   `gorm:"-" json:"is_terminal"`
}

func (BackgroundCheck) TableName() string { return "moderation_backgroundcheck" }

type ModerationSetting struct {
	ID                         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	BackgroundCheckPaymentMode string    `gorm:"size:20;default:'IN_APP_STRIPE'" json:"background_check_payment_mode"` // IN_APP_STRIPE, INVOICE
	BackgroundCheckFee         float64   `gorm:"default:29.99" json:"background_check_fee"`
	BroadcastRadiusKm          float64   `gorm:"default:25.0" json:"broadcast_radius_km"`
	MatchRadiusKm              float64   `gorm:"default:25.0" json:"match_radius_km"`
}

func (ModerationSetting) TableName() string { return "moderation_moderationsetting" }

type AuditLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action       string     `gorm:"size:255;not null" json:"action"`
	ResourceType string     `gorm:"size:100;not null" json:"resource_type"`
	ResourceID   string     `gorm:"size:100" json:"resource_id,omitempty"`
	Details      JSONMap    `gorm:"type:text" json:"details"`
	IPAddress    string     `gorm:"size:45" json:"ip_address,omitempty"`
	CreatedAt    time.Time  `gorm:"autoCreateTime;index" json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (AuditLog) TableName() string { return "audit_auditlog" }

// Admin DTOs & Models
type FeatureFlag struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Key         string    `gorm:"size:100;uniqueIndex;not null" json:"key"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	IsEnabled   bool      `gorm:"default:true" json:"is_enabled"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (FeatureFlag) TableName() string { return "core_featureflag" }

type FinancialDayBreakdown struct {
	Day              string  `json:"day"`
	GMV              float64 `json:"gmv"`
	PlatformFee      float64 `json:"platform_fee"`
	PayoutsProcessed float64 `json:"payouts_processed"`
}

type FinancialReportSummary struct {
	StartDate                  string                  `json:"start_date"`
	EndDate                    string                  `json:"end_date"`
	TotalGMV                   float64                 `json:"total_gmv"`
	TotalPlatformFeeRevenue    float64                 `json:"total_platform_fee_revenue"`
	TotalPayoutsProcessed      float64                 `json:"total_payouts_processed"`
	PendingPayoutsAmount       float64                 `json:"pending_payouts_amount"`
	ActiveSubscriptionsCount   int64                   `json:"active_subscriptions_count"`
	SubscriptionRevenueEstimate float64                 `json:"subscription_revenue_estimate"`
	DailyBreakdown             []FinancialDayBreakdown `json:"daily_breakdown"`
}

type SystemHealthStatus struct {
	Status              string    `json:"status"`
	DatabaseStatus       string    `json:"database_status"`
	ActiveDBConnections int       `json:"active_db_connections"`
	IdleDBConnections   int       `json:"idle_db_connections"`
	OpenDBConnections   int       `json:"open_db_connections"`
	TotalUsersCount     int64     `json:"total_users_count"`
	UptimeSeconds       int64     `json:"uptime_seconds"`
	GoRoutinesCount     int       `json:"goroutines_count"`
	MemoryAllocMB       float64   `json:"memory_alloc_mb"`
	Timestamp           time.Time `json:"timestamp"`
}

type GDPRUserData struct {
	User          User                `json:"user"`
	Profile       *Profile            `json:"profile,omitempty"`
	Wallet        *Wallet             `json:"wallet,omitempty"`
	Transactions  []WalletTransaction `json:"transactions,omitempty"`
	Appointments  []Appointment       `json:"appointments,omitempty"`
	Reviews       []Review            `json:"reviews,omitempty"`
	Disputes      []Dispute           `json:"disputes,omitempty"`
	MessagesCount int64               `json:"messages_count"`
	ExportedAt    time.Time           `json:"exported_at"`
}

type AdminDashboardStats struct {
	TotalUsers            int64                  `json:"total_users"`
	TotalSeekers          int64                  `json:"total_seekers"`
	TotalProviders        int64                  `json:"total_providers"`
	TotalStaff            int64                  `json:"total_staff"`
	TotalVerified         int64                  `json:"total_verified"`
	TotalSuspended        int64                  `json:"total_suspended"`
	ActiveSubscriptions   int64                  `json:"active_subscriptions"`
	TotalWalletBalance    float64                `json:"total_wallet_balance"`
	PendingPayoutsCount   int64                  `json:"pending_payouts_count"`
	PendingPayoutsAmount  float64                `json:"pending_payouts_amount"`
	AppointmentsCount     int64                  `json:"appointments_count"`
	AppointmentsCompleted int64                  `json:"appointments_completed"`
	AppointmentsPending   int64                  `json:"appointments_pending"`
	OpenDisputesCount     int64                  `json:"open_disputes_count"`
	PendingVerifications  int64                  `json:"pending_verifications"`
	PendingBackground     int64                  `json:"pending_background_checks"`
	RecentAuditLogs       []AuditLog             `json:"recent_audit_logs"`
	SignupsLast30Days     map[string]int64       `json:"signups_last_30_days,omitempty"`
}

type AdminRole struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Slug        string    `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	Permissions JSONSlice `gorm:"type:text" json:"permissions"` // e.g. ["users:read", "finance:write", ...]
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AdminRole) TableName() string { return "admin_role" }

type FraudRiskAlert struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	RiskScore  int       `gorm:"default:0" json:"risk_score"` // 0 - 100
	RiskLevel  string    `gorm:"size:20;default:'LOW'" json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
	Flags      JSONSlice `gorm:"type:text" json:"flags"` // reasons flagged
	Details    JSONMap   `gorm:"type:text" json:"details"`
	Status     string    `gorm:"size:20;default:'OPEN'" json:"status"` // OPEN, REVIEWED, DISMISSED, ACTIONED
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (FraudRiskAlert) TableName() string { return "admin_fraudriskalert" }

type NotificationTemplate struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Key       string    `gorm:"size:100;uniqueIndex;not null" json:"key"` // e.g. "otp_verification", "payout_processed"
	Name      string    `gorm:"size:255;not null" json:"name"`
	Channel   string    `gorm:"size:20;default:'EMAIL'" json:"channel"` // EMAIL, PUSH, SMS
	Subject   string    `gorm:"size:255" json:"subject"`
	BodyHTML  string    `gorm:"type:text" json:"body_html"`
	BodyText  string    `gorm:"type:text" json:"body_text"`
	Variables JSONSlice `gorm:"type:text" json:"variables"` // dynamic placeholders e.g. ["name", "code"]
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (NotificationTemplate) TableName() string { return "admin_notificationtemplate" }

type BackupSnapshot struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Filename    string     `gorm:"size:255;not null" json:"filename"`
	FileSize    int64      `gorm:"default:0" json:"file_size"`
	Status      string     `gorm:"size:20;default:'COMPLETED'" json:"status"` // PENDING, COMPLETED, FAILED
	StoragePath string     `gorm:"size:500" json:"storage_path"`
	Checksum    string     `gorm:"size:128" json:"checksum"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

func (BackupSnapshot) TableName() string { return "admin_backupsnapshot" }

type TOTPSetupResponse struct {
	Secret     string `json:"secret"`
	OTPAuthURL string `json:"otp_auth_url"`
	QRCodeURL  string `json:"qr_code_url"`
}

type ImpersonationResult struct {
	AccessToken  string    `json:"access_token"`
	ExpiresIn    int       `json:"expires_in"` // seconds
	TargetUserID uuid.UUID `json:"target_user_id"`
	TargetEmail  string    `json:"target_email"`
}

// ─── PUBLIC SITE CMS MODELS ──────────────────────────────────────────────────

type HeroSection struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Headline    string    `gorm:"size:255;not null" json:"headline"`
	Subheadline string    `gorm:"type:text;not null" json:"subheadline"`
	CtaText     string    `gorm:"size:50;default:'Get Started'" json:"cta_text"`
	CtaLink     string    `gorm:"size:255;default:'#'" json:"cta_link"`
	Image       string    `gorm:"size:500" json:"image,omitempty"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (HeroSection) TableName() string { return "public_site_herosection" }

type Feature struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string    `gorm:"size:100;not null" json:"title"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Icon        string    `gorm:"size:50" json:"icon"`
	Order       int       `gorm:"default:0" json:"order"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
}

func (Feature) TableName() string { return "public_site_feature" }

type Testimonial struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Role      string    `gorm:"size:100" json:"role,omitempty"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Image     string    `gorm:"size:500" json:"image,omitempty"`
	Rating    int       `gorm:"default:5" json:"rating"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Testimonial) TableName() string { return "public_site_testimonial" }

type FAQ struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Question string    `gorm:"size:255;not null" json:"question"`
	Answer   string    `gorm:"type:text;not null" json:"answer"`
	Order    int       `gorm:"default:0" json:"order"`
	Category string    `gorm:"size:50;default:'other'" json:"category"`
	IsActive bool      `gorm:"default:true" json:"is_active"`
}

func (FAQ) TableName() string { return "public_site_faq" }

type AboutContent struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Title         string    `gorm:"size:255;default:'Our Mission'" json:"title"`
	StoryHeadline string    `gorm:"size:255;default:'Our Story'" json:"story_headline"`
	StoryText1    string    `gorm:"type:text;not null" json:"story_text_1"`
	StoryText2    string    `gorm:"type:text" json:"story_text_2,omitempty"`
	MissionText   string    `gorm:"type:text;not null" json:"mission_text"`
	VisionText    string    `gorm:"type:text;not null" json:"vision_text"`
	YearFounded   string    `gorm:"size:10;default:'2025'" json:"year_founded"`
	CitiesCovered string    `gorm:"size:50;default:'50+'" json:"cities_covered"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AboutContent) TableName() string { return "public_site_aboutcontent" }

type HowItWorksStep struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string    `gorm:"size:100;not null" json:"title"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Order       int       `gorm:"default:1" json:"order"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
}

func (HowItWorksStep) TableName() string { return "public_site_howitworksstep" }

type SiteStat struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Label    string    `gorm:"size:100;not null" json:"label"`
	Value    string    `gorm:"size:50;not null" json:"value"`
	Order    int       `gorm:"default:0" json:"order"`
	IsActive bool      `gorm:"default:true" json:"is_active"`
}

func (SiteStat) TableName() string { return "public_site_sitestat" }

type SiteSetting struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SiteName       string    `gorm:"size:100;default:'Neighbor Service'" json:"site_name"`
	ContactEmail   string    `gorm:"size:100;default:'support@neighborservice.com'" json:"contact_email"`
	ContactPhone   string    `gorm:"size:50;default:'+1 (555) 000-0000'" json:"contact_phone"`
	ContactAddress string    `gorm:"type:text;default:'123 Community City, CC 12345'" json:"contact_address"`
	FacebookURL    string    `gorm:"size:255" json:"facebook_url,omitempty"`
	TwitterURL     string    `gorm:"size:255" json:"twitter_url,omitempty"`
	InstagramURL   string    `gorm:"size:255" json:"instagram_url,omitempty"`
	LinkedinURL    string    `gorm:"size:255" json:"linkedin_url,omitempty"`
}

func (SiteSetting) TableName() string { return "public_site_sitesetting" }

type ContactMessage struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	FirstName   string    `gorm:"size:100;not null" json:"first_name"`
	LastName    string    `gorm:"size:100;not null" json:"last_name"`
	Email       string    `gorm:"size:100;not null" json:"email"`
	InquiryType string    `gorm:"size:100;not null" json:"inquiry_type"`
	Message     string    `gorm:"type:text;not null" json:"message"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	IsResolved  bool      `gorm:"default:false" json:"is_resolved"`
}

func (ContactMessage) TableName() string { return "public_site_contactmessage" }

type ResolutionReport struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Role            string     `gorm:"size:50;not null" json:"role"`
	IssueType       string     `gorm:"size:100;not null" json:"issue_type"`
	BookingRef      string     `gorm:"size:100" json:"booking_ref,omitempty"`
	OtherNeighbor   string     `gorm:"size:100" json:"other_neighbor,omitempty"`
	DateOfService   *time.Time `json:"date_of_service,omitempty"`
	Description     string     `gorm:"type:text;not null" json:"description"`
	ExpectedOutcome string     `gorm:"type:text;not null" json:"expected_outcome"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	IsReviewed      bool       `gorm:"default:false" json:"is_reviewed"`
}

func (ResolutionReport) TableName() string { return "public_site_resolutionreport" }

// ─── ADMIN STAFF INTERNAL NOTES ─────────────────────────────────────────────
type StaffNote struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	AuthorAdminID uuid.UUID `gorm:"type:uuid;not null" json:"author_admin_id"`
	AuthorEmail   string    `gorm:"size:255;not null" json:"author_email"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	Category      string    `gorm:"size:50;default:'GENERAL'" json:"category"` // GENERAL, FRAUD, COMPLIANCE, BILLING
	IsPinned      bool      `gorm:"default:false" json:"is_pinned"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (StaffNote) TableName() string { return "admin_staff_notes" }

// ─── ADMIN NOTIFICATION & LIVE EVENT FEED ────────────────────────────────────
type AdminNotification struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`       // DISPUTE, CHECKR_FLAG, PAYOUT_HOLD, FRAUD_VELOCITY, SYSTEM_ALERT
	Severity   string    `json:"severity"`   // HIGH, MEDIUM, INFO
	Title      string    `json:"title"`
	Message    string    `json:"message"`
	ResourceID string    `json:"resource_id"`
	Route      string    `json:"route"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

// ─── PROVIDER ONBOARDING FUNNEL ─────────────────────────────────────────────
type ProviderOnboardingFunnel struct {
	TotalSignedUp     int64   `json:"total_signed_up"`
	IDUploaded        int64   `json:"id_uploaded"`
	CheckrCompleted   int64   `json:"checkr_completed"`
	StripeConnected   int64   `json:"stripe_connected"`
	FirstBookingDone  int64   `json:"first_booking_done"`
	ConversionRatePct float64 `json:"conversion_rate_pct"`
}

// ─── BATCH PAYOUT DISBURSEMENTS ─────────────────────────────────────────────
type BatchPayoutRequest struct {
	PayoutIDs []uuid.UUID `json:"payout_ids" binding:"required"`
	Reason    string      `json:"reason,omitempty"`
}

type BatchPayoutResponse struct {
	ProcessedCount int64   `json:"processed_count"`
	TotalAmount    float64 `json:"total_amount"`
	FailedCount    int64   `json:"failed_count"`
	Message        string  `json:"message"`
}

// ─── PROMO CODES & MARKETING CAMPAIGNS ───────────────────────────────────────
type PromoCode struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Code          string     `gorm:"size:50;uniqueIndex;not null" json:"code"`
	DiscountType  string     `gorm:"size:20;default:'PERCENTAGE'" json:"discount_type"` // PERCENTAGE or FIXED
	DiscountValue float64    `gorm:"not null" json:"discount_value"`
	MinSpend      float64    `gorm:"default:0.00" json:"min_spend"`
	MaxDiscount   float64    `gorm:"default:0.00" json:"max_discount"`
	MaxUses       int        `gorm:"default:0" json:"max_uses"` // 0 = unlimited
	UsesCount     int        `gorm:"default:0" json:"uses_count"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PromoCode) TableName() string { return "billing_promocode" }

// ─── EMAIL MARKETING & CAMPAIGN LOGS ─────────────────────────────────────────
type EmailCampaignLog struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	TargetEmail  string    `gorm:"size:255;not null" json:"target_email"`
	CampaignType string    `gorm:"size:100;index;not null" json:"campaign_type"` // FAVORITE_PRO_SLOTS, SEASONAL_REMINDER, POST_JOB_REVIEW, DORMANT_WINBACK, WEEKLY_DIGEST, UPCOMING_REMINDER, PRO_MILESTONE
	ReferenceID  *string   `gorm:"size:100" json:"reference_id,omitempty"`       // e.g. appointment_id, provider_id
	SentAt       time.Time `gorm:"autoCreateTime" json:"sent_at"`
}

func (EmailCampaignLog) TableName() string { return "email_campaign_logs" }





