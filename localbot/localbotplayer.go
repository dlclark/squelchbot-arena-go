package localbot

import (
	"sync"

	"github.com/dlclark/squelchbot-arena-go/squelch"
)

var _ squelch.Player = &LocalBotPlayer{}

type LocalBotPlayer struct {
	data map[string]*turn
	sync *sync.Mutex
}

func NewLocalBotPlayer() *LocalBotPlayer {
	return &LocalBotPlayer{
		data: make(map[string]*turn),
		sync: &sync.Mutex{},
	}
}

func (p *LocalBotPlayer) Info() (*squelch.PlayerInfo, error) {
	return &squelch.PlayerInfo{"LocalBot"}, nil
}

func (p *LocalBotPlayer) MatchStart(matchID string, dieCount, maxPoints, gameCount, yourBotIndex int, botNames []string) error {
	return nil
}

func (p *LocalBotPlayer) MatchEnd(matchID string, winsByBotIndex []int) error {
	return nil
}

func (p *LocalBotPlayer) GameStart(matchID, gameID string) error {
	return nil
}

func (p *LocalBotPlayer) GameEnd(matchID, gameID string, finalPlayerTurns []squelch.PlayerTurn, winnerBotIndex int) error {
	return nil
}

func (p *LocalBotPlayer) TurnStart(matchID, gameID, turnID string, startPoints int, otherPlayerTurns []squelch.PlayerTurn, isFinalRound bool) error {
	state := p.getState(matchID, gameID, turnID)
	state.isFinalRound = isFinalRound

	for _, t := range otherPlayerTurns {
		if t.EndPoints > state.currentHighScore {
			state.currentHighScore = t.EndPoints
		}
	}

	return nil
}

func (p *LocalBotPlayer) Choose(matchID, gameID, turnID string, dieValues string, options []squelch.ScoringOption) (*squelch.PlayerChoice, error) {
	// choose the highest scoring option every time, stay if > turn total 300 points
	// get turn metadata
	state := p.getState(matchID, gameID, turnID)

	opt := options[0]
	state.points += opt.Points

	stay := state.points > 300

	if state.isFinalRound {
		// keep rolling until we're #1 or bust
		stay = state.points > state.currentHighScore
	}

	return &squelch.PlayerChoice{
		TakeOptionID: opt.ID,
		Stay:         stay,
	}, nil
}

func (p *LocalBotPlayer) Squelch(matchID, gameID, turnID string, dieValues string) error {
	return nil
}

func (p *LocalBotPlayer) getState(matchID, gameID, turnID string) *turn {
	key := matchID + gameID + turnID
	p.sync.Lock()
	v, ok := p.data[key]
	if !ok {
		v = &turn{}
		p.data[key] = v
	}
	p.sync.Unlock()
	return v
}

type turn struct {
	isFinalRound     bool
	currentHighScore int
	points           int
}
