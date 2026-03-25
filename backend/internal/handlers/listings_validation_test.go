package handlers

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"resellution/backend/internal/models"
)

func ptrF(v float64) *float64 { return &v }

func TestValidateListingCreate(t *testing.T) {
	cat := uuid.NewString()
	in := models.ListingCreate{
		Title:       "T",
		Description: "D",
		Condition:   "new",
		Price:       10,
		Currency:    "INR",
		City:        "C",
	}
	if err := validateListingCreate(&in); err != nil {
		t.Fatal(err)
	}
	if in.Currency != "INR" {
		t.Fatal("currency should remain INR")
	}

	in2 := models.ListingCreate{
		Title: "T", Description: "D", Condition: "broken", Price: 1, City: "X",
	}
	if err := validateListingCreate(&in2); err == nil {
		t.Fatal("expected bad condition")
	}

	in3 := models.ListingCreate{
		Title: "T", Description: "D", Condition: "new", Price: -1, City: "X",
	}
	if err := validateListingCreate(&in3); err == nil {
		t.Fatal("expected bad price")
	}

	in4 := models.ListingCreate{
		Title: "T", Description: "D", Condition: "new", Price: 1, City: "X",
		CategoryID: &cat,
	}
	if err := validateListingCreate(&in4); err != nil {
		t.Fatal(err)
	}

	badCat := "not-uuid"
	in5 := models.ListingCreate{
		Title: "T", Description: "D", Condition: "new", Price: 1, City: "X",
		CategoryID: &badCat,
	}
	if err := validateListingCreate(&in5); err == nil {
		t.Fatal("expected bad category uuid")
	}

	in6 := models.ListingCreate{
		Title: "T", Description: "D", Condition: "new", Price: 1, City: "X",
		Latitude: ptrF(10),
	}
	if err := validateListingCreate(&in6); err == nil {
		t.Fatal("expected lat without lng")
	}
}

func TestValidateListingCreateDefaultCurrency(t *testing.T) {
	in := models.ListingCreate{
		Title: "T", Description: "D", Condition: "good", Price: 0, Currency: "", City: "Here",
	}
	if err := validateListingCreate(&in); err != nil {
		t.Fatal(err)
	}
	if in.Currency != "INR" {
		t.Fatalf("got %q", in.Currency)
	}
}

func TestValidateListingPatch(t *testing.T) {
	title := "  Hello  "
	p := models.ListingPatch{Title: &title}
	if err := validateListingPatch(&p); err != nil {
		t.Fatal(err)
	}
	if *p.Title != "Hello" {
		t.Fatalf("got %q", *p.Title)
	}

	cur := "usd"
	p2 := models.ListingPatch{Currency: &cur}
	if err := validateListingPatch(&p2); err != nil {
		t.Fatal(err)
	}
	if *p2.Currency != "USD" {
		t.Fatalf("got %q", *p2.Currency)
	}

	badCur := "XX"
	p3 := models.ListingPatch{Currency: &badCur}
	if err := validateListingPatch(&p3); err == nil {
		t.Fatal("expected bad currency length")
	}
}

func TestValidateCoordinates(t *testing.T) {
	if err := validateCoordinates(nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := validateCoordinates(ptrF(0), ptrF(0)); err != nil {
		t.Fatal(err)
	}
	if err := validateCoordinates(ptrF(10), nil); err == nil {
		t.Fatal("expected mismatch")
	}
	if err := validateCoordinates(ptrF(100), ptrF(0)); err == nil {
		t.Fatal("expected bad latitude")
	}
	if err := validateCoordinates(ptrF(0), ptrF(200)); err == nil {
		t.Fatal("expected bad longitude")
	}
}

func TestParseNullableCoordinate(t *testing.T) {
	v, err := parseNullableCoordinate(json.RawMessage(`null`))
	if err != nil || v != nil {
		t.Fatalf("null: %v %v", v, err)
	}
	v2, err := parseNullableCoordinate(json.RawMessage(`12.5`))
	if err != nil || v2 == nil || *v2 != 12.5 {
		t.Fatalf("number: %v %v", v2, err)
	}
	_, err = parseNullableCoordinate(json.RawMessage(`"x"`))
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestValidateImageURL(t *testing.T) {
	if err := validateImageURL("https://example.com/a.png"); err != nil {
		t.Fatal(err)
	}
	if err := validateImageURL(""); err == nil {
		t.Fatal("empty")
	}
	if err := validateImageURL("ftp://x.com/a"); err == nil {
		t.Fatal("ftp")
	}
	if err := validateImageURL("not-a-url"); err == nil {
		t.Fatal("relative")
	}
}
