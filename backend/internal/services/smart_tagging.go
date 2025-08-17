package services

import (
	"strings"
	"unicode"
)

// SmartTaggingService handles intelligent video tagging based on title analysis
type SmartTaggingService struct {
	// Common articles and words to filter out
	articles map[string]bool
}

// NewSmartTaggingService creates a new smart tagging service
func NewSmartTaggingService() *SmartTaggingService {
	return &SmartTaggingService{
		articles: map[string]bool{
			"and": true, "the": true, "of": true, "for": true, "a": true, "an": true,
			"in": true, "on": true, "at": true, "to": true, "with": true, "by": true,
			"from": true, "into": true, "during": true, "including": true, "until": true,
			"against": true, "among": true, "throughout": true, "despite": true,
			"towards": true, "upon": true, "concerning": true, "excepting": true,
			"excluding": true, "following": true, "inside": true, "outside": true,
			"over": true, "past": true, "since": true, "under": true, "within": true,
			"without": true, "about": true, "above": true, "across": true, "after": true,
			"along": true, "around": true, "before": true, "behind": true, "below": true,
			"beneath": true, "beside": true, "between": true, "beyond": true,
			"down": true, "except": true, "near": true, "off": true, "onto": true,
			"out": true, "through": true, "toward": true, "underneath": true,
			"up": true, "1080p": true, "720p": true, "480p": true, "360p": true, "240p": true, "144p": true,
			"are": true, "how": true, "is": true, "29fps": true, "your": true, "why": true, "what": true,
			"when": true, "where": true, "who": true, "which": true, "that": true, "this": true, "these": true,
			"they": true, "them": true, "their": true, "they're": true, "they've": true, "they'll": true,
		},
	}
}

// TaggingResult represents the result of smart tagging
type TaggingResult struct {
	Name           string   `json:"name"`
	Tags           []string `json:"tags"`
	OriginalTitle  string   `json:"original_title"`
	ProcessedTitle string   `json:"processed_title"`
}

// GenerateTagsFromTitle implements the smart tagging algorithm
// 0. Checks video "tagged" boolean value from our master_video_list
// 1. Takes the title value and leaves the title value intact
// 2. Takes the characters including spaces before "-" as a name, first entry of the tags array
// 3. Parses the rest of the words, eliminating article word types with spaces before and after
// 4. Puts each word left over as an individual tag
// 5. Sends Tag array to the tags column of our master_video_list and sets tagged column true
func (s *SmartTaggingService) GenerateTagsFromTitle(title string) *TaggingResult {
	if title == "" {
		return &TaggingResult{
			Name:           "",
			Tags:           []string{},
			OriginalTitle:  title,
			ProcessedTitle: "",
		}
	}

	// Step 1: Keep title intact
	originalTitle := title

	// Step 2: Extract name (characters before first "-")
	var name string
	var remainingTitle string

	if dashIndex := strings.Index(title, "-"); dashIndex != -1 {
		name = strings.TrimSpace(title[:dashIndex])
		remainingTitle = strings.TrimSpace(title[dashIndex+1:])
	} else {
		// If no dash, use the first word as name
		words := strings.Fields(title)
		if len(words) > 0 {
			name = words[0]
			remainingTitle = strings.TrimSpace(strings.TrimPrefix(title, words[0]))
		}
	}

	// Step 3: Process remaining title to remove articles
	var tags []string

	// Add name as first tag if it exists
	if name != "" {
		tags = append(tags, name)
	}

	// Process remaining title
	if remainingTitle != "" {
		// Split into words
		words := strings.Fields(remainingTitle)

		for _, word := range words {
			// Clean the word (remove punctuation, convert to lowercase)
			cleanWord := s.cleanWord(word)

			// Skip empty words and articles
			if cleanWord == "" || s.articles[cleanWord] {
				continue
			}

			// Skip very short words (less than 2 characters)
			if len(cleanWord) < 2 {
				continue
			}

			// Add to tags if not already present
			if !s.containsTag(tags, cleanWord) {
				tags = append(tags, cleanWord)
			}
		}
	}

	// If we still don't have tags, try to extract meaningful words from the full title
	if len(tags) <= 1 {
		words := strings.Fields(title)
		for _, word := range words {
			cleanWord := s.cleanWord(word)
			if cleanWord != "" && !s.articles[cleanWord] && len(cleanWord) >= 3 {
				if !s.containsTag(tags, cleanWord) {
					tags = append(tags, cleanWord)
				}
			}
		}
	}

	return &TaggingResult{
		Name:           name,
		Tags:           tags,
		OriginalTitle:  originalTitle,
		ProcessedTitle: remainingTitle,
	}
}

// cleanWord removes punctuation and converts to lowercase
func (s *SmartTaggingService) cleanWord(word string) string {
	// Remove punctuation from start and end
	word = strings.TrimFunc(word, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	// Convert to lowercase
	word = strings.ToLower(word)

	return word
}

// containsTag checks if a tag already exists in the tags slice
func (s *SmartTaggingService) containsTag(tags []string, tag string) bool {
	for _, existingTag := range tags {
		if strings.EqualFold(existingTag, tag) {
			return true
		}
	}
	return false
}

// GetTagCategories returns the available tag categories
func (s *SmartTaggingService) GetTagCategories() []string {
	return []string{
		"Archaeology", "Geography", "DNA Research", "Linguistics",
		"Historical Evidence", "Cultural Studies", "Religious Studies",
		"Documentary", "Lecture", "Interview", "Presentation", "Virtual Tour",
	}
}

// CategorizeTag attempts to automatically categorize a tag
func (s *SmartTaggingService) CategorizeTag(tag string) string {
	tag = strings.ToLower(tag)

	// Archaeology terms
	archaeologyTerms := []string{"archaeology", "archaeological", "excavation", "artifact", "ruins", "temple", "pyramid", "tomb", "burial", "settlement", "pottery", "ceramic", "stone", "monument", "inscription", "hieroglyph", "maya", "aztec", "inca", "mesoamerica", "ancient", "prehistoric"}
	for _, term := range archaeologyTerms {
		if strings.Contains(tag, term) {
			return "Archaeology"
		}
	}

	// Geography terms
	geographyTerms := []string{"geography", "geographic", "location", "region", "area", "territory", "landscape", "mountain", "river", "valley", "coast", "island", "continent", "america", "mexico", "guatemala", "peru", "bolivia", "ecuador", "colombia", "central", "south", "north"}
	for _, term := range geographyTerms {
		if strings.Contains(tag, term) {
			return "Geography"
		}
	}

	// DNA Research terms
	dnaTerms := []string{"dna", "genetic", "genetics", "genome", "chromosome", "mutation", "haplogroup", "mitochondrial", "nuclear", "inheritance", "ancestry", "lineage", "population", "migration", "evolution", "biological", "molecular"}
	for _, term := range dnaTerms {
		if strings.Contains(tag, term) {
			return "DNA Research"
		}
	}

	// Linguistics terms
	linguisticsTerms := []string{"language", "linguistic", "translation", "script", "writing", "text", "inscription", "alphabet", "syllable", "phonetic", "grammar", "vocabulary", "dialect", "hebrew", "egyptian", "maya", "aztec", "quipu", "hieroglyph"}
	for _, term := range linguisticsTerms {
		if strings.Contains(tag, term) {
			return "Linguistics"
		}
	}

	// Historical Evidence terms
	historicalTerms := []string{"history", "historical", "evidence", "document", "record", "chronicle", "manuscript", "scroll", "codex", "inscription", "monument", "artifact", "dating", "carbon", "radiometric", "stratigraphy", "context", "provenance"}
	for _, term := range historicalTerms {
		if strings.Contains(tag, term) {
			return "Historical Evidence"
		}
	}

	// Cultural Studies terms
	culturalTerms := []string{"culture", "cultural", "society", "civilization", "tradition", "custom", "belief", "religion", "mythology", "folklore", "ceremony", "ritual", "festival", "celebration", "community", "social", "anthropology", "ethnography"}
	for _, term := range culturalTerms {
		if strings.Contains(tag, term) {
			return "Cultural Studies"
		}
	}

	// Religious Studies terms
	religiousTerms := []string{"religion", "religious", "theology", "spiritual", "sacred", "divine", "prophet", "scripture", "canon", "doctrine", "faith", "worship", "prayer", "meditation", "mormon", "christian", "bible", "book", "mormon", "latter", "saint", "church"}
	for _, term := range religiousTerms {
		if strings.Contains(tag, term) {
			return "Religious Studies"
		}
	}

	// Documentary terms
	documentaryTerms := []string{"documentary", "film", "video", "recording", "footage", "interview", "narration", "story", "narrative", "production", "filmmaking", "cinematography", "editing", "director", "producer", "camera"}
	for _, term := range documentaryTerms {
		if strings.Contains(tag, term) {
			return "Documentary"
		}
	}

	// Lecture terms
	lectureTerms := []string{"lecture", "presentation", "talk", "speech", "address", "discourse", "seminar", "workshop", "conference", "symposium", "academic", "scholarly", "research", "study", "analysis", "examination", "investigation"}
	for _, term := range lectureTerms {
		if strings.Contains(tag, term) {
			return "Lecture"
		}
	}

	// Interview terms
	interviewTerms := []string{"interview", "conversation", "discussion", "dialogue", "exchange", "question", "answer", "response", "testimony", "witness", "account", "statement", "declaration", "narrative", "story", "experience"}
	for _, term := range interviewTerms {
		if strings.Contains(tag, term) {
			return "Interview"
		}
	}

	// Presentation terms
	presentationTerms := []string{"presentation", "slide", "visual", "graphic", "chart", "diagram", "illustration", "image", "picture", "photo", "photograph", "map", "drawing", "sketch", "design", "layout"}
	for _, term := range presentationTerms {
		if strings.Contains(tag, term) {
			return "Presentation"
		}
	}

	// Virtual Tour terms
	tourTerms := []string{"tour", "visit", "explore", "journey", "expedition", "adventure", "discovery", "exploration", "travel", "trip", "voyage", "pilgrimage", "walk", "walkthrough", "guided", "virtual", "3d", "reconstruction"}
	for _, term := range tourTerms {
		if strings.Contains(tag, term) {
			return "Virtual Tour"
		}
	}

	// Default category
	return "General"
}
