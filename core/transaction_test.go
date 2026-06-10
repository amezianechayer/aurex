package core

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHash(t *testing.T) {
	a := Transaction{
		ID: 0,
		Postings: []Posting{
			{
				Source:      "world",
				Destination: "users:081",
				Amount:      100,
				Asset:       "VOLTIS",
			},
		},
	}

	b := Transaction{
		ID: 1,
		Postings: []Posting{
			{
				Source:      "world",
				Destination: "users:081",
				Amount:      100,
				Asset:       "VOLTIS",
			},
		},
	}

	h1 := Hash(nil, &a)

	// golden values recomputed after the Transaction JSON shape changed
	// (Reference/Metadata fields); the hash only needs to be stable and
	// chained, not equal to a historical value
	if h1 != "8aa09549e0fc03174bb3e084765750baf8c23489c1361a3f55cb0c3695c685c6" {
		t.Fatalf("unexpected h1: %s", h1)
	}

	a.Hash = h1
	h2 := Hash(&a, &b)

	if h2 != "157b62b06826de9b6181e55e92c2686ca532938353ecb5ec7a6be512ad09b3b6" {
		t.Fatalf("unexpected h2: %s", h2)
	}
}

func TestReverseTransaction(t *testing.T) {
	tx := &Transaction{
		Postings: Postings{
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
		},
		Reference: "foo",
	}

	expected := Transaction{
		Postings: Postings{
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
		},
		Reference: "revert_foo",
	}

	if diff := cmp.Diff(expected, tx.Reverse()); diff != "" {
		t.Errorf("Reverse() mismatch (-want +got):\n%s", diff)
	}
}
