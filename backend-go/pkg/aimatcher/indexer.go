package aimatcher

import (
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

// ReviewSentiment holds NLP sentiment polarity and signals for an individual review.
type ReviewSentiment struct {
	Rating           float64  `json:"rating"`
	Comment          string   `json:"comment"`
	Polarity         string   `json:"polarity"` // "POSITIVE", "NEGATIVE", "NEUTRAL"
	SentimentScore   float64  `json:"sentiment_score"` // -1.0 to 1.0
	PraiseSignals    []string `json:"praise_signals"`
	ComplaintSignals []string `json:"complaint_signals"`
}

// ProviderSentimentSummary aggregates sentiment statistics across all reviews for a provider.
type ProviderSentimentSummary struct {
	TotalReviews          int               `json:"total_reviews"`
	PositiveCount         int               `json:"positive_count"`
	NegativeCount         int               `json:"negative_count"`
	NeutralCount          int               `json:"neutral_count"`
	PositiveRatio         float64           `json:"positive_ratio"` // 0.0 to 1.0 (e.g. 0.95 = 95% positive)
	AverageSentimentScore float64           `json:"avg_sentiment_score"` // -1.0 to 1.0
	PerformanceScore      float64           `json:"performance_score"` // 0 to 100
	SentimentBadge        string            `json:"sentiment_badge"`
	TopPraiseBadges       []string          `json:"top_praise_badges"`
	CriticalWarnings      []string          `json:"critical_warnings"`
	RecentSentiments      []ReviewSentiment `json:"recent_sentiments,omitempty"`
}

// IndexedProvider contains pre-computed document vectors, sentiment summaries, and ranking weights.
type IndexedProvider struct {
	Profile          entity.Profile
	DocVector        map[string]float64
	SentimentSummary ProviderSentimentSummary
	ConceptKeys      []string
	PopularityScore  float64
	LastIndexedAt    time.Time
}

// ProviderIndex is a high-speed, thread-safe in-memory cache for pre-indexed paid providers.
type ProviderIndex struct {
	mu        sync.RWMutex
	providers map[uuid.UUID]*IndexedProvider
}

var GlobalIndex = NewProviderIndex()

func NewProviderIndex() *ProviderIndex {
	return &ProviderIndex{
		providers: make(map[uuid.UUID]*IndexedProvider),
	}
}

// DynamicOntologyStore holds dynamically learned keywords, query patterns, and sentiment associations from PostgreSQL.
type DynamicOntologyStore struct {
	mu             sync.RWMutex
	concepts       map[string][]string
	customKeywords map[string]float64
	queryPatterns  []entity.CatalogKnowledgeIndex
	praiseKeywords map[string]float64
	complaintWords map[string]float64
}

var GlobalKnowledge = NewDynamicOntologyStore()

func NewDynamicOntologyStore() *DynamicOntologyStore {
	store := &DynamicOntologyStore{
		concepts:       make(map[string][]string),
		customKeywords: make(map[string]float64),
		queryPatterns:  make([]entity.CatalogKnowledgeIndex, 0),
		praiseKeywords: make(map[string]float64),
		complaintWords: make(map[string]float64),
	}
	// Copy baseline reviews lexicons
	for k, v := range reviewPraiseKeywords {
		store.praiseKeywords[k] = v
	}
	for k, v := range reviewComplaintKeywords {
		store.complaintWords[k] = v
	}
	return store
}

// LoadFromDatabase incorporates learned knowledge from database records into the in-memory AI engine.
func (s *DynamicOntologyStore) LoadFromDatabase(items []entity.CatalogKnowledgeIndex) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.queryPatterns = items

	for _, item := range items {
		concept := strings.ToLower(strings.TrimSpace(item.TradeConcept))
		kw := strings.ToLower(strings.TrimSpace(item.Keyword))

		if concept != "" && kw != "" {
			existing := s.concepts[concept]
			found := false
			for _, e := range existing {
				if strings.EqualFold(e, kw) {
					found = true
					break
				}
			}
			if !found {
				s.concepts[concept] = append(s.concepts[concept], kw)
			}
			s.customKeywords[kw] = item.ConfidenceScore
		}

		// Learn emergent praise/complaint signals
		for _, p := range item.PraiseSignals {
			pLower := strings.ToLower(strings.TrimSpace(p))
			if pLower != "" {
				s.praiseKeywords[pLower] = math.Max(s.praiseKeywords[pLower], item.ConfidenceScore)
			}
		}
		for _, c := range item.ComplaintSignals {
			cLower := strings.ToLower(strings.TrimSpace(c))
			if cLower != "" {
				s.complaintWords[cLower] = math.Max(s.complaintWords[cLower], item.ConfidenceScore)
			}
		}
	}
}

// GetConceptKeywords returns all keywords associated with a trade concept (combining static and dynamic DB knowledge).
func (s *DynamicOntologyStore) GetConceptKeywords(concept string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.concepts[strings.ToLower(concept)]
}

// GetAllLearnedConcepts returns a snapshot of all concepts in the knowledge store.
func (s *DynamicOntologyStore) GetAllLearnedConcepts() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make(map[string][]string, len(s.concepts))
	for k, v := range s.concepts {
		copied := make([]string, len(v))
		copy(copied, v)
		res[k] = copied
	}
	return res
}

func (s *DynamicOntologyStore) AddConceptSynonyms(concept string, keywords []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	conceptLower := strings.ToLower(strings.TrimSpace(concept))
	for _, kw := range keywords {
		kwClean := strings.ToLower(strings.TrimSpace(kw))
		if kwClean == "" {
			continue
		}
		exists := false
		for _, e := range s.concepts[conceptLower] {
			if e == kwClean {
				exists = true
				break
			}
		}
		if !exists {
			s.concepts[conceptLower] = append(s.concepts[conceptLower], kwClean)
		}
	}
}

func (s *DynamicOntologyStore) RemoveConcept(concept string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.concepts, strings.ToLower(strings.TrimSpace(concept)))
}

func (s *DynamicOntologyStore) GetAllConceptsMerged() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make(map[string][]string)
	// Base static ontology
	for k, v := range conceptOntology {
		copied := make([]string, len(v))
		copy(copied, v)
		res[k] = copied
	}
	// Dynamic ontology overlay
	for k, v := range s.concepts {
		res[k] = append(res[k], v...)
	}
	return res
}

// GetQueryPatterns returns all active search query patterns learned from the catalog.
func (s *DynamicOntologyStore) GetQueryPatterns() []entity.CatalogKnowledgeIndex {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]entity.CatalogKnowledgeIndex, len(s.queryPatterns))
	copy(res, s.queryPatterns)
	return res
}

// GetPraiseSignals returns praise signals map.
func (s *DynamicOntologyStore) GetPraiseSignals() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make(map[string]float64, len(s.praiseKeywords))
	for k, v := range s.praiseKeywords {
		res[k] = v
	}
	return res
}

// GetComplaintSignals returns complaint signals map.
func (s *DynamicOntologyStore) GetComplaintSignals() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make(map[string]float64, len(s.complaintWords))
	for k, v := range s.complaintWords {
		res[k] = v
	}
	return res
}

// Praise & Complaint Lexicons for Review Sentiment Classification
var reviewPraiseKeywords = map[string]float64{
	"punctual": 1.0, "on time": 1.0, "clean": 0.8, "professional": 1.0,
	"courteous": 0.9, "skilled": 1.0, "quick": 0.8, "honest": 1.0,
	"fair": 0.9, "reasonable": 0.8, "highly recommend": 1.0, "amazing": 1.0,
	"great job": 1.0, "flawless": 1.0, "polite": 0.8, "reliable": 1.0,
	"trustworthy": 1.0, "5 star": 1.0, "master": 1.0, "expert": 1.0,
	"fantastic": 1.0, "excellent": 1.0, "best": 0.9, "superb": 1.0,
	"top notch": 1.0, "efficient": 0.9, "careful": 0.9, "neat": 0.8,
}

var reviewComplaintKeywords = map[string]float64{
	"late": 0.9, "rude": 1.0, "unprofessional": 1.0, "damaged": 1.0,
	"mess": 0.8, "dirty": 0.8, "overpriced": 0.9, "ripoff": 1.0,
	"scam": 1.0, "shoddy": 1.0, "slow": 0.8, "ghosted": 1.0,
	"did not show": 1.0, "no show": 1.0, "terrible": 1.0, "worst": 1.0,
	"unreliable": 1.0, "botched": 1.0, "broke": 0.9, "poor": 0.9,
	"disappointed": 0.8, "frustrated": 0.8, "waste of money": 1.0,
}

// AnalyzeReviewSentiment computes sentiment score, polarity, and key praise/complaint signals from a review.
func AnalyzeReviewSentiment(rating float64, comment string) ReviewSentiment {
	commentLower := strings.ToLower(comment)
	var praise []string
	var complaints []string

	praiseSum := 0.0
	for kw, weight := range reviewPraiseKeywords {
		if strings.Contains(commentLower, kw) {
			praiseSum += weight
			praise = append(praise, strings.Title(kw))
		}
	}

	complaintSum := 0.0
	for kw, weight := range reviewComplaintKeywords {
		if strings.Contains(commentLower, kw) {
			complaintSum += weight
			complaints = append(complaints, strings.Title(kw))
		}
	}

	// Calculate base sentiment from star rating (-1.0 to 1.0)
	// 5.0 -> +1.0, 4.0 -> +0.5, 3.0 -> 0.0, 2.0 -> -0.5, 1.0 -> -1.0
	starSentiment := (rating - 3.0) / 2.0

	// Text modifier (-0.5 to +0.5)
	textModifier := math.Min(math.Max((praiseSum-complaintSum)*0.2, -0.5), 0.5)

	finalScore := math.Min(math.Max(starSentiment+textModifier, -1.0), 1.0)

	polarity := "NEUTRAL"
	if finalScore >= 0.25 || rating >= 4.0 {
		polarity = "POSITIVE"
	} else if finalScore <= -0.25 || rating <= 2.0 {
		polarity = "NEGATIVE"
	}

	return ReviewSentiment{
		Rating:           rating,
		Comment:          comment,
		Polarity:         polarity,
		SentimentScore:   finalScore,
		PraiseSignals:    praise,
		ComplaintSignals: complaints,
	}
}

// SummarizeProviderReviews aggregates sentiment analysis across all reviews for a provider.
func SummarizeProviderReviews(reviews []entity.Review, profile entity.Profile) ProviderSentimentSummary {
	total := len(reviews)
	if total == 0 {
		// If no reviews yet, synthesize baseline from profile attributes
		basePerf := 50.0
		if profile.IsIdentityVerified {
			basePerf += 15.0
		}
		if profile.SubscriptionTier == "PLATINUM" || profile.SubscriptionTier == "GOLD" {
			basePerf += 10.0
		}
		return ProviderSentimentSummary{
			TotalReviews:          0,
			PositiveCount:         0,
			NegativeCount:         0,
			NeutralCount:          0,
			PositiveRatio:         1.0,
			AverageSentimentScore: 0.5,
			PerformanceScore:      basePerf,
			SentimentBadge:        "New Verified Provider",
		}
	}

	posCount := 0
	negCount := 0
	neutCount := 0
	totalScore := 0.0
	praiseMap := make(map[string]int)
	complaintMap := make(map[string]int)
	var recentSentiments []ReviewSentiment

	for _, r := range reviews {
		sent := AnalyzeReviewSentiment(r.Rating, r.Comment)
		totalScore += sent.SentimentScore
		recentSentiments = append(recentSentiments, sent)

		switch sent.Polarity {
		case "POSITIVE":
			posCount++
		case "NEGATIVE":
			negCount++
		default:
			neutCount++
		}

		for _, p := range sent.PraiseSignals {
			praiseMap[p]++
		}
		for _, c := range sent.ComplaintSignals {
			complaintMap[c]++
		}
	}

	avgScore := totalScore / float64(total)
	posRatio := float64(posCount) / float64(total)

	// Top Praise Badges
	type badgeCount struct {
		badge string
		count int
	}
	var praiseList []badgeCount
	for b, c := range praiseMap {
		praiseList = append(praiseList, badgeCount{badge: b, count: c})
	}
	sort.Slice(praiseList, func(i, j int) bool {
		return praiseList[i].count > praiseList[j].count
	})
	var topPraise []string
	for i := 0; i < len(praiseList) && i < 4; i++ {
		topPraise = append(topPraise, praiseList[i].badge)
	}

	// Critical Warnings
	var complaintList []badgeCount
	for b, c := range complaintMap {
		complaintList = append(complaintList, badgeCount{badge: b, count: c})
	}
	sort.Slice(complaintList, func(i, j int) bool {
		return complaintList[i].count > complaintList[j].count
	})
	var warnings []string
	for i := 0; i < len(complaintList) && i < 2; i++ {
		warnings = append(warnings, complaintList[i].badge)
	}

	// Calculate Composite Performance Score (0 to 100)
	// Rating component (40 pts)
	ratingPart := (profile.AverageRating / 5.0) * 40.0
	// Positive Sentiment Ratio (30 pts)
	sentimentPart := posRatio * 30.0
	// Review volume log dampening (15 pts)
	volumePart := math.Min(math.Log10(float64(total+1))*15.0, 15.0)
	// Verified pro bonus (10 pts)
	verifiedPart := 0.0
	if profile.IsIdentityVerified {
		verifiedPart = 10.0
	}
	// Subscription Tier bonus (5 pts)
	tierPart := 0.0
	switch profile.SubscriptionTier {
	case "PLATINUM", "DIAMOND":
		tierPart = 5.0
	case "GOLD":
		tierPart = 3.0
	case "SILVER":
		tierPart = 1.5
	}

	performanceScore := math.Min(100.0, math.Max(0.0, ratingPart+sentimentPart+volumePart+verifiedPart+tierPart))

	// Generate Sentiment Badge
	badge := "👍 Reliable Service"
	if posRatio >= 0.90 && total >= 3 {
		badge = "🌟 Highly Recommended (Top Customer Praise)"
	} else if posRatio >= 0.80 {
		badge = "⭐ 80%+ Positive Client Satisfaction"
	} else if negCount > posCount {
		badge = "⚠️ Needs Service Quality Attention"
	}

	return ProviderSentimentSummary{
		TotalReviews:          total,
		PositiveCount:         posCount,
		NegativeCount:         negCount,
		NeutralCount:          neutCount,
		PositiveRatio:         posRatio,
		AverageSentimentScore: avgScore,
		PerformanceScore:      performanceScore,
		SentimentBadge:        badge,
		TopPraiseBadges:       topPraise,
		CriticalWarnings:      warnings,
		RecentSentiments:      recentSentiments,
	}
}

// BulkIndex loads and pre-indexes active paid providers and their review sentiments in memory.
func (pi *ProviderIndex) BulkIndex(providers []entity.Profile, reviewsMap map[uuid.UUID][]entity.Review) {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	now := time.Now().UTC()
	for _, p := range providers {
		docVec := BuildProviderDocumentVector(&p)
		summary := SummarizeProviderReviews(reviewsMap[p.UserID], p)

		// Concept extraction
		tokens, tf := Tokenize(p.Service)
		for _, cs := range p.CatalogServices {
			t2, tf2 := Tokenize(cs.Name)
			tokens = append(tokens, t2...)
			for k, v := range tf2 {
				tf[k] += v
			}
		}
		concepts := DetectConcepts(tokens, tf)

		pi.providers[p.ID] = &IndexedProvider{
			Profile:          p,
			DocVector:        docVec,
			SentimentSummary: summary,
			ConceptKeys:      concepts,
			PopularityScore:  summary.PerformanceScore,
			LastIndexedAt:    now,
		}
	}
}

// GetSentimentSummary returns the pre-computed review sentiment summary for a provider.
func (pi *ProviderIndex) GetSentimentSummary(profileID uuid.UUID) (ProviderSentimentSummary, bool) {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	if entry, exists := pi.providers[profileID]; exists {
		return entry.SentimentSummary, true
	}
	return ProviderSentimentSummary{}, false
}

// GetIndexed returns the cached provider index entry.
func (pi *ProviderIndex) GetIndexed(profileID uuid.UUID) (*IndexedProvider, bool) {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	if entry, exists := pi.providers[profileID]; exists {
		return entry, true
	}
	return nil, false
}

// FastRankProviders matches candidate providers using pre-indexed vectors and sentiment scores.
func (pi *ProviderIndex) FastRankProviders(queryText string, candidates []entity.Profile) []MatchResult {
	if len(candidates) == 0 {
		return nil
	}

	queryVec, queryConcepts := BuildQueryVector(queryText)
	sentiment := AnalyzeSentiment(queryText)

	var results []MatchResult

	pi.mu.RLock()
	defer pi.mu.RUnlock()

	for _, p := range candidates {
		var docVec map[string]float64
		var summary ProviderSentimentSummary

		if indexed, ok := pi.providers[p.ID]; ok {
			docVec = indexed.DocVector
			summary = indexed.SentimentSummary
		} else {
			docVec = BuildProviderDocumentVector(&p)
			summary = SummarizeProviderReviews(nil, p)
		}

		cosineSim := ComputeCosineSimilarity(queryVec, docVec)

		// Direct concept alignment bonus
		conceptOverlap := 0.0
		hasDirectConceptMatch := false
		for _, qc := range queryConcepts {
			if docVec["concept:"+qc] > 0 {
				conceptOverlap += 0.40
				hasDirectConceptMatch = true
			}
		}
		if conceptOverlap > 0.50 {
			conceptOverlap = 0.50
		}

		// Direct substring matching
		directMatchBonus := 0.0
		queryLower := strings.ToLower(queryText)
		if p.Service != "" && strings.Contains(queryLower, strings.ToLower(p.Service)) {
			directMatchBonus += 0.35
		}
		for _, cs := range p.CatalogServices {
			csLower := strings.ToLower(cs.Name)
			if cs.Name != "" && (strings.Contains(queryLower, csLower) || strings.Contains(csLower, queryLower)) {
				directMatchBonus += 0.40
				break
			}
		}

		// Handyman bridge
		isExplicitHandyman := strings.Contains(strings.ToLower(p.Service), "handyman")
		if !hasDirectConceptMatch && isExplicitHandyman && len(queryConcepts) > 0 {
			switch queryConcepts[0] {
			case "plumbing", "electrical", "carpentry_furniture", "painting", "roofing_gutters":
				conceptOverlap += 0.15
			}
		}

		semanticScore := (cosineSim * 0.60) + conceptOverlap + directMatchBonus
		if semanticScore < 0.14 {
			continue
		}

		// Sentiment & Performance adjustments
		ratingBonus := (p.AverageRating / 5.0) * 0.08
		sentimentReviewBonus := (summary.PositiveRatio * 0.06)
		tierBonus := 0.0
		switch p.SubscriptionTier {
		case "PLATINUM", "DIAMOND":
			tierBonus = 0.05
		case "GOLD":
			tierBonus = 0.03
		case "SILVER":
			tierBonus = 0.01
		}
		verifiedBonus := 0.0
		if p.IsIdentityVerified {
			verifiedBonus = 0.04
		}

		// Intent-sentiment alignment bonus
		sentimentBonus := 0.0
		if sentiment.UrgencyScore > 0 {
			if p.IsOnline {
				sentimentBonus += 0.06 * sentiment.UrgencyScore
			}
			if p.StreakCount > 3 {
				sentimentBonus += 0.03 * sentiment.UrgencyScore
			}
		}
		if sentiment.QualityScore > 0 {
			if p.AverageRating >= 4.7 && summary.PositiveRatio >= 0.85 {
				sentimentBonus += 0.08 * sentiment.QualityScore
			}
		}
		if sentiment.TrustScore > 0 {
			if p.IsIdentityVerified && summary.PositiveRatio >= 0.80 {
				sentimentBonus += 0.08 * sentiment.TrustScore
			}
		}
		if sentiment.BudgetScore > 0 {
			if len(p.ServicePackages) > 0 {
				sentimentBonus += 0.06 * sentiment.BudgetScore
			}
		}
		if sentiment.FrustrationScore > 0 {
			if p.AverageRating >= 4.5 && p.IsIdentityVerified {
				sentimentBonus += 0.08 * sentiment.FrustrationScore
			}
		}

		finalScore := (semanticScore * 0.60) + ratingBonus + sentimentReviewBonus + tierBonus + verifiedBonus + sentimentBonus
		if finalScore > 0.99 {
			finalScore = 0.99
		}

		matchPct := int(math.Round(60.0 + (finalScore * 39.0)))
		if matchPct < 55 {
			matchPct = 55
		}
		if matchPct > 99 {
			matchPct = 99
		}

		reason := generateMatchReason(p, matchPct, queryConcepts, cosineSim, sentiment)

		results = append(results, MatchResult{
			Profile:           p,
			Score:             finalScore,
			MatchPercentage:   matchPct,
			MatchReason:       reason,
			MatchedConcepts:   queryConcepts,
			SentimentAnalysis: sentiment,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
