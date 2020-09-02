package squelch

import (
	"github.com/stretchr/testify/mock"
)

type chooseFunc func(matchID, gameID, turnID string, dieValues string, options []ScoringOption) (*PlayerChoice, error)

type MockPlayer struct {
	mock.Mock
	name string

	chooseFn chooseFunc
}

func NewMockPlayer(name string, chooseFn chooseFunc) *MockPlayer {
	return &MockPlayer{name: name, chooseFn: chooseFn}
}

func (p *MockPlayer) Info() (*PlayerInfo, error) {
	return &PlayerInfo{p.name}, nil
}

func (p *MockPlayer) MatchStart(matchID string, dieCount, maxPoints, gameCount, yourBotIndex int, botNames []string) error {
	ret := p.Mock.Called(matchID, dieCount, maxPoints, gameCount, yourBotIndex, botNames)
	return ret.Error(0)
}

func (p *MockPlayer) MatchEnd(matchID string, winsByBotIndex []int) error {
	ret := p.Mock.Called(matchID, winsByBotIndex)
	return ret.Error(0)
}

func (p *MockPlayer) GameStart(matchID, gameID string) error {
	ret := p.Mock.Called(matchID, gameID)
	return ret.Error(0)
}

func (p *MockPlayer) GameEnd(matchID, gameID string, finalPlayerTurns []PlayerTurn, winnerBotIndex int) error {
	ret := p.Mock.Called(matchID, gameID, finalPlayerTurns, winnerBotIndex)
	return ret.Error(0)
}

func (p *MockPlayer) TurnStart(matchID, gameID, turnID string, startPoints int, otherPlayerTurns []PlayerTurn, isFinalRound bool) error {
	ret := p.Mock.Called(matchID, gameID, turnID, startPoints, otherPlayerTurns, isFinalRound)
	return ret.Error(0)
}

func (p *MockPlayer) Choose(matchID, gameID, turnID string, dieValues string, options []ScoringOption) (*PlayerChoice, error) {
	// document the func was called for asserts
	p.Mock.Called(matchID, gameID, turnID, dieValues, options)
	return p.chooseFn(matchID, gameID, turnID, dieValues, options)
}

func (p *MockPlayer) Squelch(matchID, gameID, turnID string, dieValues string) error {
	ret := p.Mock.Called(matchID, gameID, turnID, dieValues)
	return ret.Error(0)
}
