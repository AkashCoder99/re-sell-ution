package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"resellution/backend/internal/models"
	"resellution/backend/internal/observability"
)

const (
	searchQueryMaxLen   = 120
	searchMinTokenLen   = 2
	searchMaxTokens     = 8
	searchDefaultLimit  = 12
	searchMaxLimit      = 20
	searchDefaultRadius = 25.0
	searchMaxRadiusKM   = 200.0
	searchDefaultSort   = "relevance"
)

func (h ListingHandler) Search(w http.ResponseWriter, r *http.Request) {
	params, err := parseListingSearchParams(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	result, err := h.Listings.Search(r.Context(), params)
	if err != nil {
		observability.Error(r.Context(), "listings.search.failed", map[string]any{
			"q":     params.Query,
			"error": err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to search listings"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"listings":    result.Listings,
		"total":       result.Total,
		"page":        result.Page,
		"limit":       result.Limit,
		"total_pages": result.TotalPages,
	})
}

func parseListingSearchParams(r *http.Request) (models.ListingSearchParams, error) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	normalizedQuery, tsQuery, err := buildSafeTSQuery(query)
	if err != nil {
		return models.ListingSearchParams{}, err
	}

	page, err := parsePositiveIntQuery(r, "page", 1)
	if err != nil {
		return models.ListingSearchParams{}, err
	}
	limit, err := parsePositiveIntQuery(r, "limit", searchDefaultLimit)
	if err != nil {
		return models.ListingSearchParams{}, err
	}
	if limit > searchMaxLimit {
		return models.ListingSearchParams{}, errors.New("limit must be 20 or less")
	}

	city := strings.TrimSpace(r.URL.Query().Get("city"))
	if utf8.RuneCountInString(city) > maxListingCityLen {
		return models.ListingSearchParams{}, errors.New("city is too long")
	}

	var categoryID *string
	if rawCategoryID := strings.TrimSpace(r.URL.Query().Get("category_id")); rawCategoryID != "" {
		if _, err := uuid.Parse(rawCategoryID); err != nil {
			return models.ListingSearchParams{}, errors.New("invalid category_id")
		}
		categoryID = &rawCategoryID
	}

	var condition *string
	if rawCondition := strings.TrimSpace(r.URL.Query().Get("condition")); rawCondition != "" {
		if _, ok := allowedConditions[rawCondition]; !ok {
			return models.ListingSearchParams{}, errors.New("invalid condition")
		}
		condition = &rawCondition
	}

	minPrice, err := parseOptionalNonNegativeFloatQuery(r, "min_price")
	if err != nil {
		return models.ListingSearchParams{}, errors.New("invalid min_price")
	}
	maxPrice, err := parseOptionalNonNegativeFloatQuery(r, "max_price")
	if err != nil {
		return models.ListingSearchParams{}, errors.New("invalid max_price")
	}
	if minPrice != nil && maxPrice != nil && *minPrice > *maxPrice {
		return models.ListingSearchParams{}, errors.New("min_price cannot be greater than max_price")
	}

	sortBy := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sortBy == "" {
		sortBy = searchDefaultSort
	}
	switch sortBy {
	case "relevance", "created_at_desc", "created_at_asc", "price_asc", "price_desc":
	default:
		return models.ListingSearchParams{}, errors.New("invalid sort")
	}

	latitude, err := parseOptionalFloatQuery(r, "lat")
	if err != nil {
		return models.ListingSearchParams{}, errors.New("invalid lat")
	}
	longitude, err := parseOptionalFloatQuery(r, "lng")
	if err != nil {
		return models.ListingSearchParams{}, errors.New("invalid lng")
	}
	if err := validateCoordinates(latitude, longitude); err != nil {
		if latitude != nil || longitude != nil {
			return models.ListingSearchParams{}, err
		}
	}

	var radiusKM *float64
	if latitude != nil && longitude != nil {
		radiusValue := searchDefaultRadius
		if rawRadius := strings.TrimSpace(r.URL.Query().Get("radius_km")); rawRadius != "" {
			parsed, err := strconv.ParseFloat(rawRadius, 64)
			if err != nil {
				return models.ListingSearchParams{}, errors.New("invalid radius_km")
			}
			radiusValue = parsed
		}
		if radiusValue <= 0 || radiusValue > searchMaxRadiusKM {
			return models.ListingSearchParams{}, errors.New("radius_km must be greater than 0 and at most 200")
		}
		radiusKM = &radiusValue
	} else if strings.TrimSpace(r.URL.Query().Get("radius_km")) != "" {
		return models.ListingSearchParams{}, errors.New("lat and lng are required when radius_km is provided")
	}

	return models.ListingSearchParams{
		Query:      normalizedQuery,
		TSQuery:    tsQuery,
		City:       city,
		CategoryID: categoryID,
		Condition:  condition,
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		Sort:       sortBy,
		Latitude:   latitude,
		Longitude:  longitude,
		RadiusKM:   radiusKM,
		Page:       page,
		Limit:      limit,
	}, nil
}

func buildSafeTSQuery(raw string) (string, string, error) {
	normalized := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
	if normalized == "" {
		return "", "", errors.New("q is required")
	}
	if utf8.RuneCountInString(normalized) > searchQueryMaxLen {
		return "", "", errors.New("q is too long")
	}

	tokens := splitSearchTokens(normalized)
	if len(tokens) == 0 {
		return "", "", errors.New("q must contain letters or numbers")
	}

	seen := make(map[string]struct{}, len(tokens))
	lexemes := make([]string, 0, len(tokens))
	for _, token := range tokens {
		lower := strings.ToLower(token)
		if utf8.RuneCountInString(lower) < searchMinTokenLen {
			continue
		}
		if _, ok := seen[lower]; ok {
			continue
		}
		seen[lower] = struct{}{}
		lexemes = append(lexemes, lower)
		if len(lexemes) == searchMaxTokens {
			break
		}
	}

	if len(lexemes) == 0 {
		return "", "", errors.New("q must contain at least one word with 2 or more letters or numbers")
	}

	terms := make([]string, 0, len(lexemes))
	for _, lexeme := range lexemes {
		terms = append(terms, lexeme+":*")
	}

	return strings.Join(lexemes, " "), strings.Join(terms, " & "), nil
}

func splitSearchTokens(raw string) []string {
	return strings.FieldsFunc(raw, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func parsePositiveIntQuery(r *http.Request, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, errors.New(key + " must be a positive integer")
	}
	return value, nil
}

func parseOptionalFloatQuery(r *http.Request, key string) (*float64, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parseOptionalNonNegativeFloatQuery(r *http.Request, key string) (*float64, error) {
	value, err := parseOptionalFloatQuery(r, key)
	if err != nil {
		return nil, err
	}
	if value != nil && *value < 0 {
		return nil, errors.New(key + " must be non-negative")
	}
	return value, nil
}
