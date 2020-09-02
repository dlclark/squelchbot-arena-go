package squelch

import (
	"container/ring"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strconv"
)

// Game is a single game of squelch
type Game struct {
	players     *ring.Ring
	playerCount int
	targetScore int
	matchID     string
	gameID      string
	roll        func(diceCount int) string
	lastTurnID  int
}

// GameResult is the return of the Run method
type GameResult struct {
	WinnerIndex int
	ErrIndex    int
}
type gamePlayer struct {
	Player
	score      int
	playedInOT bool
	info       *PlayerInfo
	index      int
	lastTurn   *PlayerTurn
}

func (g gamePlayer) name() string {
	return fmt.Sprintf("%s %v", g.info.Name, g.index)
}

// NewGame receives starting information for creating a game
func NewGame(players []Player, targetScore int, matchID, gameID string, startPlayerIndex int) *Game {
	g := &Game{
		targetScore: targetScore,
		matchID:     matchID,
		gameID:      gameID,
		roll:        rollDice,
		playerCount: len(players),
		lastTurnID:  0,
	}

	r := ring.New(g.playerCount)
	var start *ring.Ring
	// setup our ring of players
	for i, p := range players {
		info, _ := p.Info()
		r.Value = &gamePlayer{
			Player: p,
			info:   info,
			index:  i,
		}
		if i == startPlayerIndex {
			start = r
		}
		r = r.Next()
	}

	g.players = start

	return g
}

// Run executes a game of squelch and returns a GameResult.
func (g *Game) Run() (GameResult, error) {
	//  notify all players the game is starting
	g.notifyAllPlayers(func(p *gamePlayer) error {
		return p.GameStart(g.matchID, g.gameID)
	})

	isOvertime := false
	var currentWinner *gamePlayer

	// run through our players until we have a winner
	for {
		p := g.players.Value.(*gamePlayer)

		debug("%v - Start Turn", p.name())

		otherPlayerTurns := g.getOtherPlayerTurns()
		// first, our game end conditions:
		// 	we're in the "overtime" round (after someone gets the threshold score w/o squelching)
		//  the player has already gone in the OT round
		if isOvertime {
			debug("%v - OT round", p.name())
			if p.playedInOT {
				//notify all players game ended with the result
				g.notifyAllPlayers(func(p *gamePlayer) error {
					return p.GameEnd(g.matchID, g.gameID, otherPlayerTurns, currentWinner.index)
				})
				return GameResult{WinnerIndex: currentWinner.index}, nil
			}
			// tag that we've had our shot in OT
			p.playedInOT = true
		}

		//TURN START
		p.lastTurn = &PlayerTurn{
			BotIndex:    p.index,
			StartPoints: p.score,
			EndPoints:   p.score,
		}
		turnID := g.nextTurnID()
		if err := p.TurnStart(g.matchID, g.gameID, turnID, p.score, otherPlayerTurns, isOvertime); err != nil {
			return GameResult{ErrIndex: p.index}, err
		}
		diceCount := 6
		points := 0
		turnOptionCount := 0
		for {
			// 1. if there are 0 dice in the pool then reset to 6 dice
			if diceCount == 0 {
				debug("%v - Rollover", p.name())
				// called a "rollover"
				diceCount = 6
			}

			// roll the dice for the player
			rawRoll := g.roll(diceCount)
			options := getDiceOptions(turnOptionCount, rawRoll)
			turnOptionCount += len(options)

			// if there are 0 options, it's a squelch, no points, turn over
			if len(options) == 0 {
				debug("%v - Squelch", p.name())
				p.lastTurn.Rolls = append(p.lastTurn.Rolls, PlayerRoll{rawRoll, "", 0})
				err := p.Squelch(g.matchID, g.gameID, turnID, rawRoll)
				if err != nil {
					return GameResult{ErrIndex: p.index}, err
				}
				break
			}

			// let the player choose which point option to take
			// and if to keep rolling the remaining dice or hold
			choice, err := p.Choose(g.matchID, g.gameID, turnID, rawRoll, options)
			if err != nil {
				return GameResult{ErrIndex: p.index}, err
			}

			opt := getOption(options, choice.TakeOptionID)
			// confirm the options contains the ID requested
			if opt == nil {
				// bad selection, error, bad player
				return GameResult{ErrIndex: p.index}, fmt.Errorf("Invalid option %v", choice.TakeOptionID)
			}

			points += opt.Points

			p.lastTurn.Rolls = append(p.lastTurn.Rolls, PlayerRoll{rawRoll, opt.DieValues, opt.Points})

			debug("%v - %v points so far", p.name(), points)
			// if the player chooses to hold, add running points to player score, turn over.
			if choice.Stay {
				p.score += points
				p.lastTurn.EndPoints = p.score
				debug("%v - hold, %v total points", p.name(), p.score)
				// figure out current winner
				if currentWinner == nil || p.score > currentWinner.score {
					currentWinner = p
				}

				// if we're not in OT and someone went over, then we're in OT
				// and the player that goes over is done
				if !isOvertime && p.score >= g.targetScore {
					isOvertime = true
					p.playedInOT = true
				}
				break
			}

			// if the player chooses to keep going then remove "point dice" and goto 1
			diceCount -= len(opt.DieValues)
		}

		// turn over, next player
		g.players = g.players.Next()
	}
}

func (g *Game) nextTurnID() string {
	g.lastTurnID++
	return strconv.Itoa(g.lastTurnID)
}

func getOption(options []ScoringOption, id string) *ScoringOption {
	for _, o := range options {
		if o.ID == id {
			return &o
		}
	}

	return nil
}

func (g *Game) getOtherPlayerTurns() []PlayerTurn {
	turns := make([]PlayerTurn, g.playerCount-1)
	// skip us, just iterate and gather the other players
	for p := g.players.Next(); p != g.players; p = p.Next() {
		if p.Value.(*gamePlayer).lastTurn != nil {
			turns = append(turns, *p.Value.(*gamePlayer).lastTurn)
		}
	}

	return turns
}

func (g *Game) notifyAllPlayers(f func(p *gamePlayer) error) error {
	var err error

	err = f(g.players.Value.(*gamePlayer))
	if err != nil {
		return err
	}

	for p := g.players.Next(); p != g.players; p = p.Next() {
		err2 := f(p.Value.(*gamePlayer))

		// we'll just save the first error that happens
		if err == nil {
			err = err2
		}
	}

	return err
}

func debug(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func rollDice(diceCount int) string {
	// give a number of d6, randomly generate a string and get our list
	// of options from the pre-generated table
	b := make([]rune, diceCount)
	for i := range b {
		b[i] = rune('1' + rand.Intn(6))
	}

	//sort our dice
	sort.Slice(b, func(i, j int) bool {
		return b[i] < b[j]
	})

	return string(b)
}

func getDiceOptions(turnOptionCount int, sortedDice string) []ScoringOption {
	//clone it and set IDs
	opts := append([]ScoringOption{}, allOptions[sortedDice]...)
	for i := range opts {
		opts[i].ID = strconv.Itoa(i + turnOptionCount)
	}

	return opts
}
