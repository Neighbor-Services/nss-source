package aimatcher

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"backend-go/internal/domain/entity"
)

// MatchResult encapsulates a matched provider with vector similarity analytics.
type MatchResult struct {
	Profile         entity.Profile `json:"profile"`
	Score           float64        `json:"score"`            // 0.0 to 1.0
	MatchPercentage int            `json:"match_percentage"` // 0 to 100
	MatchReason     string         `json:"match_reason"`     // Human readable AI justification
	MatchedConcepts []string       `json:"matched_concepts"`
}

// Stop words to remove noise during tokenization
var stopWords = map[string]bool{
	"a": true, "about": true, "above": true, "after": true, "again": true, "against": true,
	"all": true, "am": true, "an": true, "and": true, "any": true, "are": true, "aren't": true,
	"as": true, "at": true, "be": true, "because": true, "been": true, "before": true, "being": true,
	"below": true, "between": true, "both": true, "but": true, "by": true, "can": true, "can't": true,
	"cannot": true, "could": true, "couldn't": true, "did": true, "didn't": true, "do": true, "does": true,
	"doesn't": true, "doing": true, "don't": true, "down": true, "during": true, "each": true, "few": true,
	"for": true, "from": true, "further": true, "had": true, "hadn't": true, "has": true, "hasn't": true,
	"have": true, "haven't": true, "having": true, "he": true, "he'd": true, "he'll": true, "he's": true,
	"her": true, "here": true, "here's": true, "hers": true, "herself": true, "him": true, "himself": true,
	"his": true, "how": true, "how's": true, "i": true, "i'd": true, "i'll": true, "i'm": true, "i've": true,
	"if": true, "in": true, "into": true, "is": true, "isn't": true, "it": true, "it's": true, "its": true,
	"itself": true, "let's": true, "me": true, "more": true, "most": true, "mustn't": true, "my": true,
	"myself": true, "no": true, "nor": true, "not": true, "of": true, "off": true, "on": true, "once": true,
	"only": true, "or": true, "other": true, "ought": true, "our": true, "ours": true, "ourselves": true,
	"out": true, "over": true, "own": true, "same": true, "shan't": true, "she": true, "she'd": true,
	"she'll": true, "she's": true, "should": true, "shouldn't": true, "so": true, "some": true, "such": true,
	"than": true, "that": true, "that's": true, "the": true, "their": true, "theirs": true, "them": true,
	"themselves": true, "then": true, "there": true, "there's": true, "these": true, "they": true, "they'd": true,
	"they'll": true, "they're": true, "they've": true, "this": true, "those": true, "through": true, "to": true,
	"too": true, "under": true, "until": true, "up": true, "very": true, "was": true, "wasn't": true, "we": true,
	"we'd": true, "we'll": true, "we're": true, "we've": true, "were": true, "weren't": true, "what": true,
	"what's": true, "when": true, "when's": true, "where": true, "where's": true, "which": true, "while": true,
	"who": true, "who's": true, "whom": true, "why": true, "why's": true, "with": true, "won't": true,
	"would": true, "wouldn't": true, "you": true, "you'd": true, "you'll": true, "you're": true, "you've": true,
	"your": true, "yours": true, "yourself": true, "yourselves": true, "need": true, "needs": true, "needed": true,
	"want": true, "wants": true, "wanted": true, "looking": true, "look": true, "urgent": true, "immediately": true,
	"immediate": true, "help": true, "please": true, "someone": true, "somebody": true,
}

// Concept ontology for comprehensive multi-trade semantic expansion & synonym mapping
var conceptOntology = map[string][]string{
	// ─── 1. BEAUTY & GROOMING ───
	"beauty_barber_hair": {
		"mobile barber", "barbershop", "barber shop", "barber", "haircut", "hair cut",
		"fade", "beard trim", "shave", "razor", "mobile hairstylist", "mobile hair stylist",
		"hair stylist", "hair styling", "braiding", "braids", "braider", "loctician", "locs",
		"dreadlocks", "hair extensions", "blowout", "hair color", "hair salon", "salon",
	},
	"beauty_nails_skincare": {
		"mobile nail technician", "mobile nail tech", "nail salon", "nail technician", "manicure",
		"pedicure", "acrylic nails", "gel nails", "nail art", "mobile makeup artist", "makeup artist",
		"mobile make-up specialist", "makeup specialist", "cosmetics", "mobile beautician", "beautician",
		"waxing studio", "waxing", "facial", "skincare", "lash extensions", "lashes", "eyebrows",
	},
	"wellness_massage_spa": {
		"mobile massage", "mobile massage therapist", "massage therapist", "massage parlor",
		"massage", "mobile spa", "spa", "deep tissue", "swedish massage", "hot stone",
		"reflexology", "sports massage", "bodywork", "in-home massage",
	},

	// ─── 2. HEALTHCARE, WELLNESS & ELDERCARE ───
	"healthcare_medical": {
		"mobile primary care", "primary care", "nurse practitioner", "physician assistant",
		"mobile urgent care", "urgent care", "mobile dental", "dental clinic", "dentist",
		"mobile physical therapy", "physical therapy", "occupational therapy", "chiropractic",
		"mobile phlebotomy", "phlebotomy", "blood draw", "lab work", "mobile iv hydration",
		"iv hydration therapy", "iv drip", "mobile immunization", "vaccine", "vaccination",
		"mobile imaging", "x-ray", "ultrasound", "imaging center", "pharmacy", "refills",
		"mental health", "counseling", "therapy", "telehealth", "doctor",
	},
	"childcare_eldercare": {
		"baby sitter", "babysitter", "babysitting", "child care", "childcare", "daycare",
		"preschool", "nanny", "adult homecare", "adult home care", "eldercare", "senior care",
		"in-home caregiver", "caregiver", "companion care", "home health aide",
	},

	// ─── 3. PET SERVICES ───
	"pet_care_grooming": {
		"mobile dog grooming", "mobile pet grooming", "dog grooming", "pet groomer",
		"dog walker", "dog walking", "dog sitter", "pet sitter", "pet sitting",
		"dog training", "puppy training", "dog obedience", "cat care", "pet boarding",
	},

	// ─── 4. AUTOMOTIVE & TRANSPORTATION ───
	"auto_mobile_mechanic": {
		"mobile mechanic", "mechanic", "mobile oil change", "oil change", "mobile tire service",
		"tire service", "tire shop", "flat tire", "mobile windshield repair", "windshield repair",
		"auto glass", "auto repair", "auto repair shop", "brakes", "brake pads", "battery replacement",
		"car battery", "alternator", "starter", "diagnostic", "engine repair", "transmission",
		"car dealership", "concierge car buyer", "car buyer", "car buying", "auto consultant",
		"vehicle inspection", "vehicle", "automobile",
	},
	"auto_detailing_wash": {
		"mobile car wash", "car wash", "car detailing", "auto detailing", "interior shampoo",
		"waxing", "buffing", "ceramic coating", "vehicle wash", "fleet washing",
	},
	"transportation_chauffeur_delivery": {
		"personal chauffeur", "chauffeur", "personal driver", "private driver", "executive driver",
		"delivery driver", "delivery", "courier", "freight", "hauling", "package delivery",
	},

	// ─── 5. FOOD & HOSPITALITY ───
	"culinary_food_hospitality": {
		"personal chef", "personal cheff", "private chef", "chef", "catering", "caterer", "catering services",
		"mobile food truck", "food truck", "mobile coffee van", "coffee van", "mobile ice cream truck",
		"ice cream truck", "restaurant", "café", "cafe", "bakery", "dessert shop", "pastry", "bar", "lounge",
		"cloud kitchen", "catering kitchen", "grocery store", "super market", "supermarket",
		"meal prep", "private dinner", "dinner", "cuisine", "anniversary dinner",
	},

	// ─── 6. HOME TRADES, REPAIRS & INSPECTIONS ───
	"plumbing": {
		"plumbing", "plumber", "pipe", "pipes", "leak", "leaks", "leaking", "sink", "faucet",
		"drain", "drainage", "clog", "clogged", "toilet", "sewer", "shower", "bathtub",
		"water heater", "garbage disposal", "spigot", "valve", "p-trap", "sewage", "boiler",
	},
	"electrical": {
		"electrical", "electrician", "wiring", "wire", "wires", "switch", "switches", "socket",
		"outlet", "breaker", "circuit", "panel", "ceiling fan", "light", "lighting", "fixture",
		"chandelier", "short circuit", "voltage", "amperage", "re-wiring", "fuse", "generator",
	},
	"hvac": {
		"hvac", "ac", "air condition", "air conditioning", "ac repair", "cooling", "heat", "heater",
		"heating", "furnace", "thermostat", "duct", "ductwork", "freon", "compressor",
		"ventilation", "filter", "radiator", "heat pump", "central air",
	},
	"roofing_gutters": {
		"roof repair", "roofing", "roofer", "roof", "roof leak", "shingles", "gutter",
		"gutters", "gutter cleaning", "flashing", "skylight", "siding", "downspout", "soffit", "fascia",
	},
	"locksmith": {
		"locksmith", "lock smith", "lock", "locks", "deadbolt", "rekey", "rekeying", "lockout",
		"key", "smart lock", "keyless", "keypad", "padlock", "door lock repair",
	},
	"home_inspection": {
		"home inspection", "home inspector", "property inspection", "house inspection", "building inspection",
		"roof inspection", "foundation inspection", "mold inspection", "radon testing", "pre-purchase inspection",
	},
	"handyman_mounting": {
		"handyman", "mobile handyman", "mounting", "tv mount", "tv mounting", "hang", "hanging",
		"curtain rod", "mirror", "shelf", "drilling", "patch", "patching", "drywall repair",
		"tile", "tiling", "grout", "caulk", "caulking", "general repair",
	},
	"carpentry_furniture": {
		"carpentry", "carpenter", "mobile furniture repair", "furniture repair", "furniture assembly",
		"assemble", "assembly", "cabinet", "cabinets", "desk", "table", "chair", "bed", "wardrobe",
		"bookshelf", "shelving", "woodwork", "wood", "framing", "door", "deck", "patio", "ikea",
	},
	"painting": {
		"painting", "painter", "paint", "wall", "walls", "wall painting", "interior paint",
		"exterior paint", "primer", "stain", "staining", "drywall", "plaster", "trim", "roller",
	},

	// ─── 7. CLEANING, DEBRIS & JUNK ───
	"cleaning_sanitation": {
		"mobile cleaning", "cleaning & sanitation", "house cleaning", "clean", "cleaner",
		"maid", "maid service", "housekeeping", "deep house clean", "deep clean", "deep cleaning",
		"mop", "vacuum", "sanitize", "sanitization", "carpet cleaning", "window washer",
		"window washing", "window cleaning", "move in move out", "janitorial", "pressure wash",
	},
	"debris_junk_moving": {
		"debris removal service", "debris removal", "junk removal", "trash haul", "hauling",
		"haul", "moving", "mover", "movers", "relocation", "loading", "unloading", "truck",
		"heavy lifting", "furniture move", "warehouse storage",
	},

	// ─── 8. OUTDOOR, LANDSCAPING & PEST CONTROL ───
	"landscaping_lawn_tree": {
		"mobile landscaping", "landscaping", "lawn care", "lawn", "mowing", "lawn mowing",
		"mower", "grass", "yard", "garden", "tree service", "tree trimming", "tree removal",
		"tree", "hedges", "mulch", "mulching", "sprinkler", "irrigation", "weed", "weeding",
		"leaf removal", "stump grinding",
	},
	"pest_control": {
		"pest control", "pest", "exterminator", "bugs", "insects", "ants", "roaches",
		"cockroaches", "termites", "bed bugs", "rodents", "mice", "rats", "mosquitoes",
		"wasps", "hornets", "fumigation",
	},

	// ─── 9. APPAREL, FASHION & SHOPPING ───
	"apparel_fashion_shopping": {
		"dress maker", "dressmaker", "alteration", "alterations", "seamstress", "tailor",
		"tailoring", "clothing repair", "hem", "hems", "wedding dress", "dry cleaning", "dry cleaner",
		"fashion consultant", "wardrobe stylist", "personal shopper", "shopper", "concierge shopping",
	},

	// ─── 10. REAL ESTATE & PROFESSIONAL ───
	"real_estate": {
		"realtor", "real estate agent", "real estate", "property agent", "broker",
		"home buying", "home selling", "house listing", "apartment rental", "lease",
	},
	"professional_accounting_notary_it": {
		"mobile notary", "notary public", "notary", "mobile accounting", "tax prep",
		"accounting", "bookkeeping", "cpa", "mobile it support", "it support", "computer repair",
		"wifi setup", "network setup", "tech support",
	},

	// ─── 11. EDUCATION, FITNESS & DIGITAL ───
	"education_tutoring": {
		"tutoring", "tutor", "test prep", "sat prep", "act prep", "math tutor", "science tutor",
		"music school", "music lessons", "piano lessons", "guitar lessons", "dance studio",
		"dance school", "art school", "trade school", "vocational training",
	},
	"fitness_wellness": {
		"gym", "fitness center", "yoga studio", "yoga", "martial arts", "boxing gym",
		"boxing", "personal trainer", "fitness coaching",
	},
	"digital_remote_services": {
		"online consulting", "consulting", "digital marketing", "social media management",
		"graphic design", "branding", "logo design", "web development", "web design",
		"app development", "virtual assistant", "online course creation", "e-book publishing",
		"print on demand", "dropshipping", "ecommerce store",
	},
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9\s\-]+`)

// Tokenize processes raw text into clean, stemmed/normalized keywords and N-grams.
func Tokenize(text string) ([]string, map[string]int) {
	text = strings.ToLower(text)
	text = nonAlphanumericRegex.ReplaceAllString(text, " ")

	fields := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || r == '-' || r == '/' || r == '_'
	})

	var validTokens []string
	termFreq := make(map[string]int)

	for _, w := range fields {
		w = strings.TrimSpace(w)
		if len(w) <= 1 || stopWords[w] {
			continue
		}
		// Basic suffix normalization
		w = normalizeToken(w)
		if len(w) > 1 && !stopWords[w] {
			validTokens = append(validTokens, w)
			termFreq[w]++
		}
	}

	// Extract Bi-grams and Tri-grams for phrase context
	for i := 0; i < len(validTokens)-1; i++ {
		bigram := validTokens[i] + " " + validTokens[i+1]
		termFreq[bigram] += 2
		if i < len(validTokens)-2 {
			trigram := validTokens[i] + " " + validTokens[i+1] + " " + validTokens[i+2]
			termFreq[trigram] += 3
		}
	}

	return validTokens, termFreq
}

func normalizeToken(w string) string {
	if strings.HasSuffix(w, "ies") && len(w) > 4 {
		return w[:len(w)-3] + "y"
	}
	if strings.HasSuffix(w, "ing") && len(w) > 5 {
		return w[:len(w)-3]
	}
	if strings.HasSuffix(w, "ed") && len(w) > 4 {
		return w[:len(w)-2]
	}
	if strings.HasSuffix(w, "es") && len(w) > 4 {
		return w[:len(w)-2]
	}
	if strings.HasSuffix(w, "s") && len(w) > 3 && !strings.HasSuffix(w, "ss") {
		return w[:len(w)-1]
	}
	return w
}

// DetectConcepts identifies which high-level semantic trade domains match the given tokens and term frequencies.
func DetectConcepts(tokens []string, tf map[string]int) []string {
	detected := make(map[string]int)

	for concept, keywords := range conceptOntology {
		for _, kw := range keywords {
			kwNorm := normalizeToken(kw)
			if tf[kw] > 0 || tf[kwNorm] > 0 {
				detected[concept] += 5
			}
			for _, t := range tokens {
				if t == kw || t == kwNorm {
					detected[concept] += 3
				}
			}
		}
	}

	type cScore struct {
		concept string
		score   int
	}
	var list []cScore
	for c, score := range detected {
		if score >= 3 {
			list = append(list, cScore{concept: c, score: score})
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].score > list[j].score
	})

	var results []string
	for _, item := range list {
		results = append(results, item.concept)
	}
	return results
}

// BuildQueryVector generates a weighted TF-IDF vector with semantic concept expansion.
func BuildQueryVector(queryText string) (map[string]float64, []string) {
	tokens, tf := Tokenize(queryText)
	concepts := DetectConcepts(tokens, tf)

	vec := make(map[string]float64)
	for term, count := range tf {
		vec[term] = float64(count) * 1.5
	}

	// Concept expansion bonus: seed domain synonyms into vector with weighted relevance
	for _, c := range concepts {
		vec["concept:"+c] = 4.0
		for _, kw := range conceptOntology[c] {
			vec[normalizeToken(kw)] += 1.2
		}
	}

	normalizeVector(vec)
	return vec, concepts
}

// BuildProviderDocumentVector creates a rich semantic document vector from provider profile features.
func BuildProviderDocumentVector(p *entity.Profile) map[string]float64 {
	vec := make(map[string]float64)

	// 1. Primary Service Title (4.0x weight)
	if p.Service != "" {
		tokens, tf := Tokenize(p.Service)
		for t, c := range tf {
			vec[t] += float64(c) * 4.0
		}
		for _, c := range DetectConcepts(tokens, tf) {
			vec["concept:"+c] += 5.0
		}
	}

	// 2. Catalog Services (3.5x weight)
	for _, cs := range p.CatalogServices {
		if cs.Name != "" {
			tokens, tf := Tokenize(cs.Name)
			for t, c := range tf {
				vec[t] += float64(c) * 3.5
			}
			for _, con := range DetectConcepts(tokens, tf) {
				vec["concept:"+con] += 4.5
			}
		}
		if cs.Description != "" {
			tokens, tf := Tokenize(cs.Description)
			for t, c := range tf {
				vec[t] += float64(c) * 2.0
			}
			for _, con := range DetectConcepts(tokens, tf) {
				vec["concept:"+con] += 2.5
			}
		}
		if cs.Category != nil && cs.Category.Name != "" {
			tokens, tf := Tokenize(cs.Category.Name)
			for t, c := range tf {
				vec[t] += float64(c) * 3.0
			}
			for _, con := range DetectConcepts(tokens, tf) {
				vec["concept:"+con] += 4.0
			}
		}
	}

	// 3. Service Packages (2.5x weight)
	for _, pkg := range p.ServicePackages {
		if pkg.Name != "" {
			_, tf := Tokenize(pkg.Name)
			for t, c := range tf {
				vec[t] += float64(c) * 2.5
			}
		}
		if pkg.Description != "" {
			_, tf := Tokenize(pkg.Description)
			for t, c := range tf {
				vec[t] += float64(c) * 1.5
			}
		}
		for _, f := range pkg.Features {
			_, tf := Tokenize(f)
			for t, c := range tf {
				vec[t] += float64(c) * 1.5
			}
		}
	}

	// 4. Portfolio Items (2.0x weight)
	for _, item := range p.PortfolioItems {
		if item.Description != "" {
			_, tf := Tokenize(item.Description)
			for t, c := range tf {
				vec[t] += float64(c) * 1.8
			}
		}
		for _, tag := range item.Tags {
			_, tf := Tokenize(tag)
			for t, c := range tf {
				vec[t] += float64(c) * 2.0
			}
		}
	}

	// 5. Bio / About (1.5x weight)
	if p.Bio != "" {
		tokens, tf := Tokenize(p.Bio)
		for t, c := range tf {
			vec[t] += float64(c) * 1.5
		}
		for _, con := range DetectConcepts(tokens, tf) {
			vec["concept:"+con] += 2.5
		}
	}

	// 6. City & State (Location context)
	if p.City != "" {
		vec["loc:"+strings.ToLower(p.City)] = 1.0
	}
	if p.State != "" {
		vec["loc:"+strings.ToLower(p.State)] = 0.5
	}

	normalizeVector(vec)
	return vec
}

// ComputeCosineSimilarity calculates the dot product between two normalized unit vectors.
func ComputeCosineSimilarity(v1, v2 map[string]float64) float64 {
	if len(v1) == 0 || len(v2) == 0 {
		return 0.0
	}
	dotProduct := 0.0
	for term, val1 := range v1 {
		if val2, exists := v2[term]; exists {
			dotProduct += val1 * val2
		}
	}
	if dotProduct > 1.0 {
		return 1.0
	}
	return dotProduct
}

func normalizeVector(vec map[string]float64) {
	norm := 0.0
	for _, val := range vec {
		norm += val * val
	}
	if norm == 0 {
		return
	}
	sqrtNorm := math.Sqrt(norm)
	for term := range vec {
		vec[term] /= sqrtNorm
	}
}

// RankProviders matches candidate providers against the query using Vector Space Cosine Similarity and Multi-Signal quality scoring.
func RankProviders(queryText string, candidates []entity.Profile) []MatchResult {
	if len(candidates) == 0 {
		return nil
	}

	queryVec, queryConcepts := BuildQueryVector(queryText)

	var results []MatchResult

	for _, p := range candidates {
		docVec := BuildProviderDocumentVector(&p)
		cosineSim := ComputeCosineSimilarity(queryVec, docVec)

		// Check direct concept alignment bonus
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

		// Direct substring matching on service title, catalog services, or bio
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

		// Handyman cross-domain semantic bridge for general minor repairs (only if no direct concept conflict)
		isExplicitHandyman := strings.Contains(strings.ToLower(p.Service), "handyman")
		if !hasDirectConceptMatch && isExplicitHandyman && len(queryConcepts) > 0 {
			switch queryConcepts[0] {
			case "plumbing", "electrical", "carpentry_furniture", "painting", "roofing_gutters":
				conceptOverlap += 0.15
			}
		}

		// Base semantic relevance (must be non-zero to qualify as a match)
		semanticScore := (cosineSim * 0.60) + conceptOverlap + directMatchBonus

		if semanticScore < 0.14 {
			// No meaningful semantic match for this query -> Skip provider completely
			continue
		}

		// Quality signals (only applied to semantically relevant candidates)
		ratingBonus := 0.0
		if p.AverageRating > 0 {
			ratingBonus = (p.AverageRating / 5.0) * 0.08
		}
		reviewBonus := 0.0
		if p.TotalReviews > 0 {
			reviewBonus = math.Min(math.Log10(float64(p.TotalReviews+1))/2.0, 1.0) * 0.05
		}
		tierBonus := 0.0
		switch p.SubscriptionTier {
		case "PLATINUM":
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

		// Composite final match score (scaled smoothly to 60%-99% for real matches)
		finalScore := (semanticScore * 0.70) + ratingBonus + reviewBonus + tierBonus + verifiedBonus
		if finalScore > 0.99 {
			finalScore = 0.99
		}

		// Map to a calibrated, realistic percentage (65% to 99% for true matches)
		matchPct := int(math.Round(60.0 + (finalScore * 39.0)))
		if matchPct < 55 {
			matchPct = 55
		}
		if matchPct > 99 {
			matchPct = 99
		}

		// Generate human-friendly reasoning
		reason := generateMatchReason(p, matchPct, queryConcepts, cosineSim)

		results = append(results, MatchResult{
			Profile:         p,
			Score:           finalScore,
			MatchPercentage: matchPct,
			MatchReason:     reason,
			MatchedConcepts: queryConcepts,
		})
	}

	// Sort by match score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

func generateMatchReason(p entity.Profile, matchPct int, queryConcepts []string, cosineSim float64) string {
	svcName := p.Service
	if svcName == "" && len(p.CatalogServices) > 0 {
		svcName = p.CatalogServices[0].Name
	}
	if svcName == "" {
		svcName = "Professional Service"
	}

	var sb strings.Builder
	sb.WriteString(svcName)

	if p.AverageRating >= 4.5 && p.TotalReviews >= 5 {
		sb.WriteString(" • Top Rated")
	} else if p.IsIdentityVerified {
		sb.WriteString(" • Verified Pro")
	} else if p.SubscriptionTier == "PLATINUM" || p.SubscriptionTier == "GOLD" {
		sb.WriteString(" • Featured Pro")
	}

	if len(queryConcepts) > 0 {
		tradeName := formatConceptName(queryConcepts[0])
		sb.WriteString(" (" + tradeName + " Match)")
	}

	return sb.String()
}

func formatConceptName(c string) string {
	switch c {
	case "beauty_barber_hair":
		return "Barber & Hair"
	case "beauty_nails_skincare":
		return "Nails & Skincare"
	case "wellness_massage_spa":
		return "Massage & Spa"
	case "healthcare_medical":
		return "Healthcare"
	case "childcare_eldercare":
		return "Caregiving"
	case "pet_care_grooming":
		return "Pet Care & Grooming"
	case "auto_mobile_mechanic":
		return "Mobile Mechanic"
	case "auto_detailing_wash":
		return "Auto Detailing"
	case "transportation_chauffeur_delivery":
		return "Chauffeur & Delivery"
	case "culinary_food_hospitality":
		return "Culinary & Chef"
	case "plumbing":
		return "Plumbing"
	case "electrical":
		return "Electrical"
	case "hvac":
		return "HVAC & AC"
	case "roofing_gutters":
		return "Roofing & Gutters"
	case "locksmith":
		return "Locksmith"
	case "home_inspection":
		return "Home Inspection"
	case "handyman_mounting":
		return "Handyman & Mounting"
	case "carpentry_furniture":
		return "Carpentry & Assembly"
	case "painting":
		return "Painting"
	case "cleaning_sanitation":
		return "Cleaning & Sanitation"
	case "debris_junk_moving":
		return "Debris & Moving"
	case "landscaping_lawn_tree":
		return "Landscaping & Lawn"
	case "pest_control":
		return "Pest Control"
	case "apparel_fashion_shopping":
		return "Fashion & Alterations"
	case "real_estate":
		return "Real Estate"
	case "professional_accounting_notary_it":
		return "Professional & IT"
	case "education_tutoring":
		return "Education & Tutoring"
	case "fitness_wellness":
		return "Fitness & Wellness"
	case "digital_remote_services":
		return "Digital & Remote"
	default:
		return strings.Title(strings.ReplaceAll(c, "_", " "))
	}
}
