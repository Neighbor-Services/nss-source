package aimatcher

import (
	"testing"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

func TestVectorMatchPlumbing(t *testing.T) {
	candidates := []entity.Profile{
		{
			ID:                 uuid.New(),
			FirstName:          "John",
			LastName:           "Plumber",
			Service:            "Licensed Master Plumbing & Pipe Repair",
			Bio:                "Specializing in drain cleaning, leaking pipes, faucet repair and water heaters.",
			AverageRating:      4.9,
			TotalReviews:       34,
			IsIdentityVerified: true,
			CatalogServices: []entity.CatalogService{
				{Name: "Emergency Pipe Leak Repair"},
				{Name: "Sink & Faucet Installation"},
			},
		},
		{
			ID:            uuid.New(),
			FirstName:     "Alice",
			LastName:      "Painter",
			Service:       "Interior & Exterior Wall Painting",
			Bio:           "Professional home painting, drywall repair, and staining.",
			AverageRating: 4.8,
			TotalReviews:  20,
		},
	}

	query := "My bathroom pipe is leaking under the sink and needs immediate repair"
	results := RankProviders(query, candidates)

	for i, res := range results {
		t.Logf("Result #%d: %s %s - Score: %f (Pct: %d%%), Reason: %s",
			i, res.Profile.FirstName, res.Profile.LastName, res.Score, res.MatchPercentage, res.MatchReason)
	}

	if len(results) == 0 {
		t.Fatalf("Expected matched providers, got none")
	}

	if results[0].Profile.FirstName != "John" {
		t.Errorf("Expected top match to be John (Plumber), got %s", results[0].Profile.FirstName)
	}
}

func TestVectorMatchMobileBarber(t *testing.T) {
	candidates := []entity.Profile{
		{
			ID:            uuid.New(),
			FirstName:     "Marcus",
			LastName:      "Blade",
			Service:       "Mobile Barber & Beard Specialist",
			Bio:           "Traveling haircuts, skin fade, hot towel shave, and beard sculpting right at your location.",
			AverageRating: 5.0,
			TotalReviews:  28,
		},
		{
			ID:            uuid.New(),
			FirstName:     "Sam",
			LastName:      "Mover",
			Service:       "Debris Removal Service & Hauling",
			Bio:           "Junk removal and cleanout.",
		},
	}

	query := "Need a haircut and beard fade at my home"
	results := RankProviders(query, candidates)

	if len(results) == 0 || results[0].Profile.FirstName != "Marcus" {
		t.Fatalf("Expected Marcus (Mobile Barber) as top match")
	}
	t.Logf("Barber match: %s - %d%% (%s)", results[0].Profile.FirstName, results[0].MatchPercentage, results[0].MatchReason)
}

func TestVectorMatchEldercareAndPet(t *testing.T) {
	candidates := []entity.Profile{
		{
			ID:                 uuid.New(),
			FirstName:          "Sarah",
			LastName:           "Care",
			Service:            "Adult Homecare & Senior Companion",
			Bio:                "Certified caregiver assisting seniors with daily activities, companionship, and medication reminders.",
			IsIdentityVerified: true,
			AverageRating:      4.95,
		},
		{
			ID:            uuid.New(),
			FirstName:     "Mike",
			LastName:      "Paws",
			Service:       "Mobile Dog Grooming & Puppy Training",
			Bio:           "Hydrobath, dog nail clipping, fur trim, and behavioral obedience.",
			AverageRating: 4.8,
		},
	}

	elderQuery := "Looking for adult homecare assistance for elderly parent"
	elderResults := RankProviders(elderQuery, candidates)
	if len(elderResults) == 0 || elderResults[0].Profile.FirstName != "Sarah" {
		t.Fatalf("Expected Sarah for adult homecare")
	}

	petQuery := "Puppy needs mobile dog grooming and nail trim"
	petResults := RankProviders(petQuery, candidates)
	if len(petResults) == 0 || petResults[0].Profile.FirstName != "Mike" {
		t.Fatalf("Expected Mike for pet grooming")
	}
}

func TestVectorMatchNewServices(t *testing.T) {
	candidates := []entity.Profile{
		{
			ID:        uuid.New(),
			FirstName: "Gordon",
			Service:   "Personal Chef & Private Catering",
			Bio:       "Custom weekly meal prep and multi-course private dinners.",
		},
		{
			ID:        uuid.New(),
			FirstName: "Victor",
			Service:   "Concierge Car Buyer & Auto Consultant",
			Bio:       "Negotiating dealership prices and vehicle inspection for car buyers.",
		},
		{
			ID:        uuid.New(),
			FirstName: "Claire",
			Service:   "Dress Maker & Alterations Specialist",
			Bio:       "Custom tailoring, dress alterations, hems, and fashion styling.",
		},
		{
			ID:        uuid.New(),
			FirstName: "Arthur",
			Service:   "Certified Home Inspection Services",
			Bio:       "Full property foundation, roof, plumbing, and electrical pre-purchase inspections.",
		},
	}

	r1 := RankProviders("Looking for a personal chef for anniversary dinner", candidates)
	if len(r1) == 0 || r1[0].Profile.FirstName != "Gordon" {
		t.Errorf("Expected Gordon for chef query")
	}

	r2 := RankProviders("Need car buyer concierge to help buy a vehicle", candidates)
	if len(r2) == 0 || r2[0].Profile.FirstName != "Victor" {
		t.Errorf("Expected Victor for car buyer query")
	}

	r3 := RankProviders("Wedding dress alteration and hem adjustment", candidates)
	if len(r3) == 0 || r3[0].Profile.FirstName != "Claire" {
		t.Errorf("Expected Claire for dress alteration query")
	}

	r4 := RankProviders("Pre-purchase home inspection before buying house", candidates)
	if len(r4) == 0 || r4[0].Profile.FirstName != "Arthur" {
		t.Errorf("Expected Arthur for home inspection query")
	}
}

func TestSentimentAnalysisEngine(t *testing.T) {
	// Test 1: Emergency & Urgency Sentiment
	s1 := AnalyzeSentiment("Emergency! Pipe is bursting and flooding my house right now asap!")
	if s1.PrimarySentiment != "URGENT" || s1.UrgencyScore < 0.5 {
		t.Errorf("Expected URGENT sentiment with high urgency score, got %+v", s1)
	}

	// Test 2: Quality & Master Expectation
	s2 := AnalyzeSentiment("Looking for the best top rated master craftsman for meticulous luxury woodwork")
	if s2.PrimarySentiment != "QUALITY" || s2.QualityScore < 0.5 {
		t.Errorf("Expected QUALITY sentiment, got %+v", s2)
	}

	// Test 3: Budget & Price Sensitivity
	s3 := AnalyzeSentiment("Need an affordable student budget friendly moving helper with cheap quote")
	if s3.PrimarySentiment != "BUDGET" || s3.BudgetScore < 0.5 {
		t.Errorf("Expected BUDGET sentiment, got %+v", s3)
	}

	// Test 4: Trust, Safety & Background Checked
	s4 := AnalyzeSentiment("Need a trustworthy background checked safe caregiver for my elderly mother")
	if s4.PrimarySentiment != "TRUST" || s4.TrustScore < 0.5 {
		t.Errorf("Expected TRUST sentiment, got %+v", s4)
	}

	// Test 5: Customer Distress / Frustration
	s5 := AnalyzeSentiment("Complete nightmare disaster! Everything is botched and ruined please help me")
	if s5.PrimarySentiment != "DISTRESS" || s5.FrustrationScore < 0.5 || s5.Polarity != "NEGATIVE" {
		t.Errorf("Expected DISTRESS sentiment with NEGATIVE polarity, got %+v", s5)
	}
}

func TestReviewSentimentAndProviderIndex(t *testing.T) {
	// 1. Positive Review Analysis
	rev1 := AnalyzeReviewSentiment(5.0, "Marcus was extremely punctual, clean, polite, and did a fantastic master haircut! Highly recommend.")
	if rev1.Polarity != "POSITIVE" || rev1.SentimentScore < 0.5 || len(rev1.PraiseSignals) == 0 {
		t.Errorf("Expected positive review sentiment with praise signals, got %+v", rev1)
	}

	// 2. Negative Review Analysis
	rev2 := AnalyzeReviewSentiment(1.0, "He was late, very rude, left a mess everywhere and overcharged me for botched plumbing work.")
	if rev2.Polarity != "NEGATIVE" || rev2.SentimentScore > -0.2 || len(rev2.ComplaintSignals) == 0 {
		t.Errorf("Expected negative review sentiment with complaint signals, got %+v", rev2)
	}

	// 3. Provider Sentiment Aggregation & Indexing
	pID := uuid.New()
	uID := uuid.New()
	provider := entity.Profile{
		ID:                 pID,
		UserID:             uID,
		FirstName:          "Marcus",
		LastName:           "MasterBarber",
		Service:            "Mobile Barber & Beard Specialist",
		AverageRating:      4.9,
		TotalReviews:       2,
		IsIdentityVerified: true,
		SubscriptionTier:   "PLATINUM",
	}

	reviews := []entity.Review{
		{
			ID:         uuid.New(),
			ProviderID: uID,
			Rating:     5.0,
			Comment:    "Best barber in town, super clean and on time!",
		},
		{
			ID:         uuid.New(),
			ProviderID: uID,
			Rating:     5.0,
			Comment:    "Top notch expert haircut, very professional and friendly.",
		},
	}

	summary := SummarizeProviderReviews(reviews, provider)
	if summary.PositiveRatio != 1.0 || summary.PerformanceScore < 70.0 {
		t.Errorf("Expected 100%% positive ratio and high performance score, got %+v", summary)
	}

	// 4. Test In-Memory Global Index Fast Match
	revMap := map[uuid.UUID][]entity.Review{uID: reviews}
	GlobalIndex.BulkIndex([]entity.Profile{provider}, revMap)

	sum, ok := GlobalIndex.GetSentimentSummary(pID)
	if !ok || sum.PositiveRatio != 1.0 {
		t.Errorf("Expected cached sentiment summary from GlobalIndex, got ok=%v, sum=%+v", ok, sum)
	}

	fastMatches := GlobalIndex.FastRankProviders("Looking for top notch expert master barber", []entity.Profile{provider})
	if len(fastMatches) == 0 || fastMatches[0].MatchPercentage < 75 {
		t.Errorf("Expected high percentage fast match from GlobalIndex, got %+v", fastMatches)
	}
}
