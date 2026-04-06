package core

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReverseMultiple(t *testing.T) {
	p := Postings{
		{
			Source:      "world",
			Destination: "users:081",
			Amount:      100,
			Asset:       "VOLTIS",
		},
		{
			Source:      "users:081",
			Destination: "payments:081",
			Amount:      100,
			Asset:       "VOLTIS",
		},
	}

	expected := Postings{
		{
			Source:      "payments:081",
			Destination: "users:081",
			Amount:      100,
			Asset:       "VOLTIS",
		},
		{
			Source:      "users:081",
			Destination: "world",
			Amount:      100,
			Asset:       "VOLTIS",
		},
	}

	p.Reverse()

	if diff := cmp.Diff(expected, p); diff != "" {
		t.Errorf("Reverse() mismatch (-want +got):\n%s", diff)
	}
}

func TestReverseSingle(t *testing.T) {
	p := Postings{
		{
			Source:      "world",
			Destination: "users:081",
			Amount:      100,
			Asset:       "VOLTIS",
		},
	}

	expected := Postings{
		{
			Source:      "users:081",
			Destination: "world",
			Amount:      100,
			Asset:       "VOLTIS",
		},
	}

	p.Reverse()

	if diff := cmp.Diff(expected, p); diff != "" {
		t.Errorf("Reverse() mismatch (-want +got):\n%s", diff)
	}
}
