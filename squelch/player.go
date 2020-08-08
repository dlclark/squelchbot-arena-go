package squelch

// Player represents a single squelch player.
type Player interface {
	Info() (*PlayerInfo, error)

	MatchStart(matchID string, dieCount, maxPoints, gameCount, yourBotIndex int, botNames []string) error
	MatchEnd(matchID string, winsByBotIndex []int) error

	GameStart(matchID, gameID string) error
	GameEnd(matchID, gameID string, finalPlayerTurns []PlayerTurn, winnerBotIndex int) error

	TurnStart(matchID, gameID, turnID string, startPoints int, otherPlayerTurns []PlayerTurn, isFinalRound bool) error
	Choose(matchID, gameID, turnID string, dieValues string, options []ScoringOption) (*PlayerChoice, error)
	Squelch(matchID, gameID, turnID string, dieValues string) error
}

// PlayerTurn is a catalog of the turn choices made by a player
type PlayerTurn struct {
	BotIndex               int
	StartPoints, EndPoints int
	Rolls                  []PlayerRoll
}

// PlayerRoll is a single roll and selection made by a player
type PlayerRoll struct {
	DieValues string
	Take      string
	Points    int
}

// PlayerInfo is the basic information about a player
type PlayerInfo struct {
	Name string
}

// PlayerChoice is the option selected after a roll and if the player wants to keep rolling
type PlayerChoice struct {
	TakeOptionID string
	Stay         bool
}
