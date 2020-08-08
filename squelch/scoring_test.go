package squelch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateLists(t *testing.T) {
	c := generateLists("1123")
	assert.Equal(t, "1", <-c, "1")
	assert.Equal(t, "11", <-c, "2")
	assert.Equal(t, "2", <-c, "3")
	assert.Equal(t, "12", <-c, "4")
	assert.Equal(t, "112", <-c, "5")
	assert.Equal(t, "3", <-c, "6")
	assert.Equal(t, "13", <-c, "7")
	assert.Equal(t, "113", <-c, "8")
	assert.Equal(t, "23", <-c, "9")
	assert.Equal(t, "123", <-c, "10")
	assert.Equal(t, "1123", <-c, "11")

}

func TestScoreUseAll_Run(t *testing.T) {
	s := scoreUseAll("123456")
	assert.Equal(t, []ScoringOption{{DieValues: "123456", Points: 1500}}, s, "Score")
}

func TestScoreUseAll_None(t *testing.T) {
	s := scoreUseAll("12345")
	assert.Equal(t, []ScoringOption{}, s, "score")
}

func TestScore_12345(t *testing.T) {
	s := score("12345")
	assert.Equal(t, []ScoringOption{
		{DieValues: "15", Points: 150},
		{DieValues: "1", Points: 100},
		{DieValues: "5", Points: 50},
	}, s, "score")
}

func TestScore_111555(t *testing.T) {
	s := score("111555")
	assert.Equal(t, []ScoringOption{
		{DieValues: "111555", Points: 1500},
		{DieValues: "11155", Points: 1100},
		{DieValues: "1115", Points: 1050},
		{DieValues: "111", Points: 1000},
		{DieValues: "11555", Points: 700},
		{DieValues: "1555", Points: 600},
		{DieValues: "555", Points: 500},
		{DieValues: "1155", Points: 300},
		{DieValues: "115", Points: 250},
		{DieValues: "11", Points: 200},
		{DieValues: "155", Points: 200},
		{DieValues: "15", Points: 150},
		{DieValues: "1", Points: 100},
		{DieValues: "55", Points: 100},
		{DieValues: "5", Points: 50},
	}, s, "score")
}

func TestScore_Squelch(t *testing.T) {
	s := score("223466")
	assert.Equal(t, []ScoringOption{}, s, "score")
}

func TestScore_TripleDouble(t *testing.T) {
	s := score("223366")
	assert.Equal(t, []ScoringOption{{DieValues: "223366", Points: 750}}, s, "score")
}

func TestScore_SixOfAKind(t *testing.T) {
	s := score("333333")
	assert.Equal(t, []ScoringOption{
		{DieValues: "333333", Points: 750},
		{DieValues: "333333", Points: 600},
		{DieValues: "333", Points: 300},
	}, s, "score")
}

func TestGetAllOptions_Basic(t *testing.T) {
	all := getAllOptions()

	if all == nil {
		t.Fatalf("no options returned")
	}

	if want, got := 923, len(all); want != got {
		t.Fatalf("missing options by count, want=%v, got=%v", want, got)
	}

	if opts := all["123456"]; len(opts) != 4 {
		t.Fatalf("123456 has invalid length, got %v", len(opts))
	} else if opts[0].DieValues != "123456" || opts[0].Points != 1500 {
		t.Fatalf("123456 has invalid 0 index value or points, got %v", opts[0])
	} else if opts[1].DieValues != "15" || opts[1].Points != 150 {
		t.Fatalf("123456 has invalid 1 index value or points, got %v", opts[1])
	} else if opts[2].DieValues != "1" || opts[2].Points != 100 {
		t.Fatalf("123456 has invalid 2 index value or points, got %v", opts[2])
	} else if opts[3].DieValues != "5" || opts[3].Points != 50 {
		t.Fatalf("123456 has invalid 3 index value or points, got %v", opts[3])
	}
}
