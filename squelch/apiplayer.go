package squelch

import "net/url"

var _ Player = &ApiPlayer{}

type ApiPlayer struct {
	baseURL url.URL
}

func NewApiPlayer(baseURL url.URL) *ApiPlayer {
	return &ApiPlayer{
		baseURL: baseURL,
	}
}

func (p *ApiPlayer) Info() (*PlayerInfo, error) {
	// TODO: make API call
	return nil, nil
}

func (p *ApiPlayer) MatchStart(matchId string, dieCount, maxPoints, gameCount, yourBotIndex int, botNames []string) error {
	return nil
}

func (p *ApiPlayer) MatchEnd(matchId string, winsByBotIndex []int) error {
	return nil
}

func (p *ApiPlayer) GameStart(matchId, gameId string) error {
	return nil
}

func (p *ApiPlayer) GameEnd(matchId, gameId string, finalPlayerTurns []PlayerTurn, winnerBotIndex int) error {
	return nil
}

func (p *ApiPlayer) TurnStart(matchId, gameId, turnId string, startPoints int, otherPlayerTurns []PlayerTurn, isFinalRound bool) error {
	return nil
}

func (p *ApiPlayer) Choose(matchId, gameId, turnId string, dieValues string, options []ScoringOption) (*PlayerChoice, error) {
	return nil, nil
}

func (p *ApiPlayer) Squelch(matchId, gameId, turnId string, dieValues string) error {
	return nil
}
