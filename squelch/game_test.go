package squelch

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestGame_Basics(t *testing.T) {
	p1 := getMockPlayerTakeHighestXTimes(t, "1", 1)
	p2 := getMockPlayerTakeHighestXTimes(t, "2", 1)
	g := NewGame([]Player{p1, p2}, 2000, "m", "g", 0)
	g.roll = getRollFunc(t, []string{"123456", "111111", "123446"})

	res, err := g.Run()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if want, got := 1, res.WinnerIndex; want != got {
		t.Errorf("Winner incorrect, want %v got %v", want, got)
	}
}

func TestGame_MultipleRolls(t *testing.T) {
	// player rolls up to 2 times per turn --
	// player 0 rolls: 111223, 122 | 122466 66644 | 123456
	// player 1 rolls: 223466 | 223466 | 223466 |
	p1 := getMockPlayerTakeHighestXTimes(t, "1", 2)
	p2 := getMockPlayerTakeHighestXTimes(t, "2", 2)
	g := NewGame([]Player{p1, p2}, 2000, "m", "g", 0)
	g.roll = getRollFunc(t, []string{
		"111223", "122",
		"223466",
		"122466", "66644",
		"223466",
		"123456", "122346",
		"223466",
	})

	res, err := g.Run()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if want, got := 0, res.WinnerIndex; want != got {
		t.Errorf("Winner incorrect, want %v got %v", want, got)
	}

}

func getMockPlayerTakeHighestXTimes(t *testing.T, name string, x int) *MockPlayer {
	turns := make(map[string]map[string]struct{})
	choiceNum := 0
	p := NewMockPlayer(name, func(matchID, gameID, turnID string, dieValues string, options []ScoringOption) (*PlayerChoice, error) {
		// do a bit of asserting for our inputs

		if choiceNum == 0 {
			// if it's a new turn then the turn ID should be unique
			if _, ok := turns[turnID]; ok {
				t.Fatalf("bot %v, turnID '%v' already happened!", name, turnID)
			}
			turns[turnID] = make(map[string]struct{})
		}

		// every optionID should be unique for this turn
		optIDs := turns[turnID]
		for _, o := range options {
			if _, ok := optIDs[o.ID]; ok {
				t.Fatalf("bot %v, turnID '%v', optionID '%v' already happened this turn:\n %+v", name, turnID, o.ID, options)
			}
			optIDs[o.ID] = struct{}{}
		}

		choiceNum++
		if choiceNum == x {
			// we're done, reset for next turn
			choiceNum = 0
			return &PlayerChoice{options[0].ID, true}, nil
		}

		// choose the first ID and say our turn is NOT over
		return &PlayerChoice{options[0].ID, false}, nil
	})

	//TODO:
	//p.On("MatchStart")
	//p.On("MatchEnd")
	p.On("GameStart", "m", "g").Return(nil)
	p.On("GameEnd", "m", "g", mock.Anything, mock.Anything).Return(nil)
	p.On("TurnStart", "m", "g", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	p.On("Choose", "m", "g", mock.Anything, mock.Anything, mock.Anything)
	p.On("Squelch", "m", "g", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		choiceNum = 0
	})

	return p
}

func getRollFunc(t *testing.T, rolls []string) func(int) string {
	idx := 0

	return func(dieCount int) string {
		if idx >= len(rolls) {
			t.Fatalf("ran out of rolls at index %v", idx)
		}
		val := rolls[idx]

		if want, got := dieCount, len(val); want != got {
			t.Fatalf("Wanted to roll %v dice, but got %v instead", want, got)
		}

		idx++
		return val
	}
}
