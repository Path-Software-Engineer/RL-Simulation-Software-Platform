package httpapi

import "testing"

func TestOperatorTokenComparison(t *testing.T) {
	server := &Server{operatorToken: "registered-operator-token"}

	if !server.matchesToken("registered-operator-token") {
		t.Fatal("registered operator token was rejected")
	}
	if server.matchesToken("registered-operator-taken") {
		t.Fatal("different operator token was accepted")
	}
	if server.matchesToken("short") {
		t.Fatal("different-length operator token was accepted")
	}
}

func TestPaginationParsersRejectInvalidValues(t *testing.T) {
	if _, err := parsePositiveInt("0", "limit"); err == nil {
		t.Fatal("zero page size was accepted")
	}
	if _, err := parseNonNegativeInt("-1", "afterStep", -1); err == nil {
		t.Fatal("negative explicit cursor was accepted")
	}
	if _, err := parseTimeCursor("not-a-time", "before"); err == nil {
		t.Fatal("invalid timestamp cursor was accepted")
	}
}

func TestPaginationParsersAcceptDefaults(t *testing.T) {
	limit, err := parsePositiveInt("", "limit")
	if err != nil || limit != 0 {
		t.Fatalf("default page size = %d, %v; want 0, nil", limit, err)
	}
	step, err := parseNonNegativeInt("", "afterStep", -1)
	if err != nil || step != -1 {
		t.Fatalf("default step cursor = %d, %v; want -1, nil", step, err)
	}
}
