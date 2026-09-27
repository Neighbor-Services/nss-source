package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/pkg/aimatcher"
	"backend-go/pkg/recovery"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// AICatalogKnowledgeWorker analyzes the service catalog, customer queries, and reviews in the database.
// It generates dynamic keywords, natural search query patterns, and sentiment associations, storing them in PostgreSQL
// and updating the in-memory AI Matcher so the platform continually self-improves.
type AICatalogKnowledgeWorker struct {
	db       *gorm.DB
	cfg      *config.Config
	stopChan chan struct{}
}

func NewAICatalogKnowledgeWorker(db *gorm.DB, cfg *config.Config) *AICatalogKnowledgeWorker {
	return &AICatalogKnowledgeWorker{
		db:       db,
		cfg:      cfg,
		stopChan: make(chan struct{}),
	}
}

func (w *AICatalogKnowledgeWorker) Start() {
	slog.Info("🤖 Starting AI Catalog Knowledge & Continuous Learning Background Worker...")

	go func() {
		// Run every 30 minutes to dynamically assimilate new services, reviews, and search patterns
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		// Initial learning run on server boot
		recovery.SafeRun("AICatalogKnowledgeWorker.Initial", w.analyzeCatalogAndLearn)

		for {
			select {
			case <-w.stopChan:
				slog.Info("🤖 Stopping AI Catalog Knowledge Worker...")
				return
			case <-ticker.C:
				recovery.SafeRun("AICatalogKnowledgeWorker.Tick", w.analyzeCatalogAndLearn)
			}
		}
	}()
}

func (w *AICatalogKnowledgeWorker) Stop() {
	close(w.stopChan)
}

// AnalyzeCatalogAndLearn manually or automatically analyzes database catalog and updates knowledge index.
func (w *AICatalogKnowledgeWorker) AnalyzeCatalogAndLearn() {
	w.analyzeCatalogAndLearn()
}

func (w *AICatalogKnowledgeWorker) analyzeCatalogAndLearn() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	slog.Info("🧠 [AI Catalog Worker] Starting catalog semantic analysis and knowledge indexing...")

	// 1. Fetch all categories and catalog services from PostgreSQL
	var categories []entity.Category
	if err := w.db.WithContext(ctx).Preload("Services").Find(&categories).Error; err != nil {
		slog.Error("❌ [AI Catalog Worker] Failed to query categories", "err", err)
		return
	}

	var services []entity.CatalogService
	if err := w.db.WithContext(ctx).Preload("Category").Find(&services).Error; err != nil {
		slog.Error("❌ [AI Catalog Worker] Failed to query catalog services", "err", err)
		return
	}

	var generatedItems []entity.CatalogKnowledgeIndex

	// 2. Map & Synthesize Knowledge for each Catalog Service
	for _, s := range services {
		categoryName := ""
		if s.Category != nil {
			categoryName = s.Category.Name
		}

		concept := deriveTradeConcept(s.Name, categoryName, s.Description)
		queries := generateSearchQueryPatterns(s.Name, categoryName)
		sentimentIntent := deriveSentimentIntent(s.Name, s.Description)

		serviceID := s.ID
		for _, q := range queries {
			item := entity.CatalogKnowledgeIndex{
				ID:                 uuid.New(),
				TradeConcept:       concept,
				Keyword:            strings.ToLower(s.Name),
				CategoryName:       categoryName,
				ServiceID:          &serviceID,
				ServiceName:        s.Name,
				SearchQueryPattern: q,
				SentimentIntent:    sentimentIntent,
				PraiseSignals:      pq.StringArray{"punctual", "clean work", "skilled", "professional", "fair price"},
				ComplaintSignals:   pq.StringArray{"delayed", "unprofessional", "overpriced", "mess"},
				ConfidenceScore:    1.0,
				Source:             "CATALOG_ANALYZER",
			}
			generatedItems = append(generatedItems, item)
		}
	}

	// 3. Synthesize Category-level Knowledge
	for _, c := range categories {
		concept := deriveTradeConcept(c.Name, c.Name, c.Description)
		sampleQueries := generateCategoryQueries(c.Name)

		for _, q := range sampleQueries {
			item := entity.CatalogKnowledgeIndex{
				ID:                 uuid.New(),
				TradeConcept:       concept,
				Keyword:            strings.ToLower(c.Name),
				CategoryName:       c.Name,
				SearchQueryPattern: q,
				SentimentIntent:    "STANDARD",
				PraiseSignals:      pq.StringArray{"expert", "licensed", "reliable", "on time"},
				ComplaintSignals:   pq.StringArray{"no show", "rude"},
				ConfidenceScore:    0.95,
				Source:             "CATALOG_ANALYZER",
			}
			generatedItems = append(generatedItems, item)
		}
	}

	// 4. Mine Recent Reviews for emergent Praise & Complaint Signals
	var recentReviews []entity.Review
	if err := w.db.WithContext(ctx).Where("comment != ''").Order("created_at DESC").Limit(200).Find(&recentReviews).Error; err == nil {
		for _, rev := range recentReviews {
			sentiment := aimatcher.AnalyzeReviewSentiment(rev.Rating, rev.Comment)
			if len(sentiment.PraiseSignals) > 0 || len(sentiment.ComplaintSignals) > 0 {
				item := entity.CatalogKnowledgeIndex{
					ID:                 uuid.New(),
					TradeConcept:       "customer_feedback_signal",
					Keyword:            fmt.Sprintf("review_%s", rev.ID.String()[:8]),
					SearchQueryPattern: "",
					SentimentIntent:    sentiment.Polarity,
					PraiseSignals:      pq.StringArray(sentiment.PraiseSignals),
					ComplaintSignals:   pq.StringArray(sentiment.ComplaintSignals),
					ConfidenceScore:    mathMin(1.0, mathMax(0.5, sentiment.SentimentScore)),
					Source:             "SENTIMENT_MINER",
				}
				generatedItems = append(generatedItems, item)
			}
		}
	}

	// 5. Upsert all generated knowledge items into PostgreSQL
	savedCount := 0
	for _, item := range generatedItems {
		var existing entity.CatalogKnowledgeIndex
		err := w.db.WithContext(ctx).
			Where("trade_concept = ? AND keyword = ? AND search_query_pattern = ?", item.TradeConcept, item.Keyword, item.SearchQueryPattern).
			First(&existing).Error

		if err == nil {
			// Update existing record
			w.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
				"category_name":    item.CategoryName,
				"service_name":     item.ServiceName,
				"service_id":       item.ServiceID,
				"sentiment_intent": item.SentimentIntent,
				"praise_signals":   item.PraiseSignals,
				"complaint_signals": item.ComplaintSignals,
				"confidence_score": item.ConfidenceScore,
				"updated_at":       time.Now(),
			})
			savedCount++
		} else {
			if err := w.db.WithContext(ctx).Create(&item).Error; err == nil {
				savedCount++
			}
		}
	}

	// 6. Query all stored knowledge from DB and update the in-memory AI Matcher Engine
	var allKnowledge []entity.CatalogKnowledgeIndex
	if err := w.db.WithContext(ctx).Find(&allKnowledge).Error; err == nil {
		aimatcher.GlobalKnowledge.LoadFromDatabase(allKnowledge)
		slog.Info("✅ [AI Catalog Worker] Successfully indexed catalog knowledge and updated dynamic ontology",
			"generated", len(generatedItems),
			"saved_or_updated", savedCount,
			"total_active_knowledge", len(allKnowledge),
		)
	}
}

func deriveTradeConcept(name, category, desc string) string {
	combined := strings.ToLower(name + " " + category + " " + desc)

	switch {
	case strings.Contains(combined, "plumb") || strings.Contains(combined, "pipe") || strings.Contains(combined, "leak") || strings.Contains(combined, "drain") || strings.Contains(combined, "faucet") || strings.Contains(combined, "toilet"):
		return "plumbing"
	case strings.Contains(combined, "electric") || strings.Contains(combined, "wiring") || strings.Contains(combined, "panel") || strings.Contains(combined, "circuit") || strings.Contains(combined, "lighting") || strings.Contains(combined, "outlet"):
		return "electrical"
	case strings.Contains(combined, "barber") || strings.Contains(combined, "hair") || strings.Contains(combined, "fade") || strings.Contains(combined, "braid") || strings.Contains(combined, "cut") || strings.Contains(combined, "salon"):
		return "beauty_barber_hair"
	case strings.Contains(combined, "nail") || strings.Contains(combined, "manicure") || strings.Contains(combined, "pedicure") || strings.Contains(combined, "lash") || strings.Contains(combined, "facial") || strings.Contains(combined, "makeup"):
		return "beauty_nails_skincare"
	case strings.Contains(combined, "mechanic") || strings.Contains(combined, "auto") || strings.Contains(combined, "car") || strings.Contains(combined, "brake") || strings.Contains(combined, "oil change") || strings.Contains(combined, "engine") || strings.Contains(combined, "tire"):
		return "auto_mobile_mechanic"
	case strings.Contains(combined, "clean") || strings.Contains(combined, "maid") || strings.Contains(combined, "housekeeping") || strings.Contains(combined, "deep clean") || strings.Contains(combined, "carpet"):
		return "cleaning_maid_housekeeping"
	case strings.Contains(combined, "ac ") || strings.Contains(combined, "hvac") || strings.Contains(combined, "air condition") || strings.Contains(combined, "heating") || strings.Contains(combined, "furnace") || strings.Contains(combined, "cooling"):
		return "hvac"
	case strings.Contains(combined, "lock") || strings.Contains(combined, "key") || strings.Contains(combined, "lockout"):
		return "locksmith"
	case strings.Contains(combined, "handyman") || strings.Contains(combined, "mount") || strings.Contains(combined, "tv mount") || strings.Contains(combined, "assemble") || strings.Contains(combined, "fix"):
		return "handyman_mounting"
	case strings.Contains(combined, "paint") || strings.Contains(combined, "drywall") || strings.Contains(combined, "wall"):
		return "painting_drywall"
	case strings.Contains(combined, "roof") || strings.Contains(combined, "gutter") || strings.Contains(combined, "shingle"):
		return "roofing_gutters"
	case strings.Contains(combined, "dog") || strings.Contains(combined, "cat") || strings.Contains(combined, "pet") || strings.Contains(combined, "grooming"):
		return "pet_care_grooming"
	case strings.Contains(combined, "nurse") || strings.Contains(combined, "medical") || strings.Contains(combined, "therapy") || strings.Contains(combined, "health") || strings.Contains(combined, "caregiver") || strings.Contains(combined, "elder"):
		return "healthcare_medical"
	case strings.Contains(combined, "lawn") || strings.Contains(combined, "landscape") || strings.Contains(combined, "tree") || strings.Contains(combined, "mow") || strings.Contains(combined, "garden"):
		return "landscaping_lawn_tree"
	case strings.Contains(combined, "pest") || strings.Contains(combined, "bug") || strings.Contains(combined, "roach") || strings.Contains(combined, "termite") || strings.Contains(combined, "fumigat"):
		return "pest_control"
	case strings.Contains(combined, "junk") || strings.Contains(combined, "debris") || strings.Contains(combined, "trash") || strings.Contains(combined, "moving") || strings.Contains(combined, "haul"):
		return "debris_junk_moving"
	case strings.Contains(combined, "massage") || strings.Contains(combined, "spa") || strings.Contains(combined, "wellness"):
		return "wellness_massage_spa"
	case strings.Contains(combined, "chef") || strings.Contains(combined, "cook") || strings.Contains(combined, "catering") || strings.Contains(combined, "meal"):
		return "culinary_personal_chef"
	default:
		return "general_service"
	}
}

func generateSearchQueryPatterns(serviceName, categoryName string) []string {
	nameLower := strings.ToLower(serviceName)
	var patterns []string

	patterns = append(patterns, fmt.Sprintf("Need someone for %s", nameLower))
	patterns = append(patterns, fmt.Sprintf("Looking for %s near me", nameLower))

	switch {
	case strings.Contains(nameLower, "pipe") || strings.Contains(nameLower, "leak") || strings.Contains(nameLower, "plumb"):
		patterns = append(patterns, "My bathroom pipe is leaking under the sink and needs immediate repair")
		patterns = append(patterns, "Emergency plumber to fix a dripping drain and water leak")
	case strings.Contains(nameLower, "barber") || strings.Contains(nameLower, "fade") || strings.Contains(nameLower, "hair"):
		patterns = append(patterns, "Mobile barber for fresh fade and beard lineup at my house")
		patterns = append(patterns, "Top rated hairstylist for hair styling and braiding")
	case strings.Contains(nameLower, "brake") || strings.Contains(nameLower, "mechanic") || strings.Contains(nameLower, "oil"):
		patterns = append(patterns, "Mobile mechanic to change squeaking front brake pads today")
		patterns = append(patterns, "Car won't start, need diagnostic and battery replacement")
	case strings.Contains(nameLower, "clean"):
		patterns = append(patterns, "Deep clean 3-bedroom apartment before moving out this weekend")
		patterns = append(patterns, "Thorough house cleaning service with great reviews")
	case strings.Contains(nameLower, "electric") || strings.Contains(nameLower, "fan") || strings.Contains(nameLower, "light"):
		patterns = append(patterns, "Certified electrician to install modern ceiling fan and test wiring")
		patterns = append(patterns, "Circuit breaker keeps tripping, need fast electrical inspection")
	case strings.Contains(nameLower, "ac ") || strings.Contains(nameLower, "hvac") || strings.Contains(nameLower, "cool"):
		patterns = append(patterns, "AC unit blowing warm air on hot summer day, urgent HVAC repair")
	case strings.Contains(nameLower, "lock") || strings.Contains(nameLower, "key"):
		patterns = append(patterns, "Locked out of front door, need emergency locksmith right now")
	case strings.Contains(nameLower, "pet") || strings.Contains(nameLower, "dog") || strings.Contains(nameLower, "groom"):
		patterns = append(patterns, "Mobile dog groomer for large golden retriever bath and brush")
	case strings.Contains(nameLower, "paint") || strings.Contains(nameLower, "wall"):
		patterns = append(patterns, "Professional painter to paint living room walls and smooth drywall")
	case strings.Contains(nameLower, "chef") || strings.Contains(nameLower, "dinner"):
		patterns = append(patterns, "Personal chef to prepare intimate 3-course anniversary dinner")
	default:
		patterns = append(patterns, fmt.Sprintf("Professional %s specialist with excellent customer ratings", nameLower))
	}

	return patterns
}

func generateCategoryQueries(categoryName string) []string {
	catLower := strings.ToLower(categoryName)
	return []string{
		fmt.Sprintf("Find best %s pros in my area", catLower),
		fmt.Sprintf("Top rated verified %s specialists", catLower),
	}
}

func deriveSentimentIntent(name, desc string) string {
	combined := strings.ToLower(name + " " + desc)
	switch {
	case strings.Contains(combined, "emergency") || strings.Contains(combined, "urgent") || strings.Contains(combined, "leak") || strings.Contains(combined, "lockout") || strings.Contains(combined, "breakdown"):
		return "URGENT"
	case strings.Contains(combined, "luxury") || strings.Contains(combined, "master") || strings.Contains(combined, "custom") || strings.Contains(combined, "wedding") || strings.Contains(combined, "gourmet"):
		return "QUALITY"
	case strings.Contains(combined, "cheap") || strings.Contains(combined, "budget") || strings.Contains(combined, "discount") || strings.Contains(combined, "basic"):
		return "BUDGET"
	case strings.Contains(combined, "elder") || strings.Contains(combined, "child") || strings.Contains(combined, "safety") || strings.Contains(combined, "certified"):
		return "TRUST"
	default:
		return "STANDARD"
	}
}

func mathMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func mathMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
