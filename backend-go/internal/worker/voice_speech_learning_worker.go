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

// VoiceSpeechLearningWorker periodically analyzes voice transcripts and user project requests.
// It extracts emerging trade vocabulary, colloquial terminology, sentiment indicators, and updates
// the AI knowledge base in PostgreSQL so voice parsing and matching continually self-improve.
type VoiceSpeechLearningWorker struct {
	db       *gorm.DB
	cfg      *config.Config
	stopChan chan struct{}
}

func NewVoiceSpeechLearningWorker(db *gorm.DB, cfg *config.Config) *VoiceSpeechLearningWorker {
	return &VoiceSpeechLearningWorker{
		db:       db,
		cfg:      cfg,
		stopChan: make(chan struct{}),
	}
}

func (w *VoiceSpeechLearningWorker) Start() {
	slog.Info("🎙️ Starting AI Voice Speech & Sentiment Learning Background Worker...")

	go func() {
		// Run every 15 minutes
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		// Run immediately on worker start
		recovery.SafeRun("VoiceSpeechLearningWorker.Initial", w.analyzeVoiceSpeechAndLearn)

		for {
			select {
			case <-w.stopChan:
				slog.Info("🎙️ Stopping AI Voice Speech Learning Worker...")
				return
			case <-ticker.C:
				recovery.SafeRun("VoiceSpeechLearningWorker.Tick", w.analyzeVoiceSpeechAndLearn)
			}
		}
	}()
}

func (w *VoiceSpeechLearningWorker) Stop() {
	close(w.stopChan)
}

func (w *VoiceSpeechLearningWorker) AnalyzeNow() {
	w.analyzeVoiceSpeechAndLearn()
}

func (w *VoiceSpeechLearningWorker) analyzeVoiceSpeechAndLearn() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	slog.Info("[VoiceSpeechLearningWorker] Analyzing voice speech logs and project conversations...")

	// 1. Fetch recent voice speech logs (last 7 days)
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	var logs []entity.VoiceSpeechLog
	err := w.db.WithContext(ctx).
		Preload("CatalogService").
		Where("created_at >= ?", sevenDaysAgo).
		Order("created_at DESC").
		Limit(200).
		Find(&logs).Error

	if err != nil {
		slog.Error("[VoiceSpeechLearningWorker] Failed to fetch voice speech logs", "err", err)
		return
	}

	if len(logs) == 0 {
		slog.Debug("[VoiceSpeechLearningWorker] No recent voice speech logs to analyze.")
		return
	}

	// 2. Compute Sentiment Distribution
	var urgentCount, positiveCount, frustratedCount, neutralCount int
	keywordFrequency := make(map[string]int)
	conceptPhrases := make(map[string][]string)

	for _, log := range logs {
		switch log.Sentiment {
		case "URGENT":
			urgentCount++
		case "POSITIVE":
			positiveCount++
		case "FRUSTRATED":
			frustratedCount++
		default:
			neutralCount++
		}

		// Count keyword co-occurrences
		for _, kw := range log.Keywords {
			cleaned := strings.ToLower(strings.TrimSpace(kw))
			if len(cleaned) > 2 {
				keywordFrequency[cleaned]++
			}
		}

		// Group natural query patterns by service type
		tradeKey := strings.ToLower(strings.TrimSpace(log.ServiceType))
		if tradeKey == "" && log.CatalogService != nil {
			tradeKey = strings.ToLower(log.CatalogService.Name)
		}
		if tradeKey != "" && log.RawTranscript != "" {
			conceptPhrases[tradeKey] = append(conceptPhrases[tradeKey], log.RawTranscript)
		}
	}

	slog.Info("[VoiceSpeechLearningWorker] Speech Sentiment Analysis",
		"total_logs", len(logs),
		"urgent", urgentCount,
		"positive", positiveCount,
		"frustrated", frustratedCount,
		"neutral", neutralCount,
	)

	// 3. Extract New Learned Knowledge Indices
	learnedCount := 0
	for tradeConcept, phrases := range conceptPhrases {
		if len(phrases) == 0 {
			continue
		}

		// Find most representative phrase
		for _, rawPhrase := range phrases {
			cleanPhrase := strings.TrimSpace(rawPhrase)
			if len(cleanPhrase) < 5 || len(cleanPhrase) > 200 {
				continue
			}

			// Determine sentiment intent
			intent := "STANDARD"
			lower := strings.ToLower(cleanPhrase)
			if strings.Contains(lower, "emergency") || strings.Contains(lower, "urgent") || strings.Contains(lower, "asap") || strings.Contains(lower, "leak") || strings.Contains(lower, "burst") {
				intent = "URGENT"
			} else if strings.Contains(lower, "remodel") || strings.Contains(lower, "upgrade") || strings.Contains(lower, "custom") {
				intent = "QUALITY"
			}

			// Check if knowledge already exists in DB
			var existing entity.CatalogKnowledgeIndex
			findErr := w.db.WithContext(ctx).
				Where("trade_concept = ? AND search_query_pattern = ?", tradeConcept, cleanPhrase).
				First(&existing).Error

			if findErr == gorm.ErrRecordNotFound {
				newIndex := entity.CatalogKnowledgeIndex{
					ID:                 uuid.New(),
					TradeConcept:       tradeConcept,
					Keyword:            tradeConcept,
					CategoryName:       "Voice Learned Service",
					SearchQueryPattern: cleanPhrase,
					SentimentIntent:    intent,
					PraiseSignals:      pq.StringArray{"verified voice request", "high user intent"},
					ConfidenceScore:    1.2,
					SearchCount:        1,
					Source:             "VOICE_SPEECH_MINER",
					CreatedAt:          time.Now(),
					UpdatedAt:          time.Now(),
				}
				if err := w.db.WithContext(ctx).Create(&newIndex).Error; err == nil {
					learnedCount++
				}
			} else if findErr == nil {
				// Increment usage count and boost confidence
				_ = w.db.WithContext(ctx).Model(&existing).
					Updates(map[string]interface{}{
						"search_count":     gorm.Expr("search_count + 1"),
						"confidence_score": gorm.Expr("confidence_score + 0.05"),
						"updated_at":       time.Now(),
					}).Error
			}
		}
	}

	// 4. Reload in-memory matcher with updated knowledge
	var allKnowledge []entity.CatalogKnowledgeIndex
	if err := w.db.WithContext(ctx).Find(&allKnowledge).Error; err == nil && len(allKnowledge) > 0 {
		aimatcher.GlobalKnowledge.LoadFromDatabase(allKnowledge)
	}

	slog.Info(fmt.Sprintf("[VoiceSpeechLearningWorker] Completed speech learning pass. Learned %d new query patterns from speech data.", learnedCount))
}
