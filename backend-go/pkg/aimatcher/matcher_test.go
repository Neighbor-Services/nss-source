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
