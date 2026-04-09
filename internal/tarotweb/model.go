package tarotweb

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	mrand "math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

type Card struct {
	ID      string
	Name    string
	Arcana  string
	Suit    string
	Rank    string
	Element string
	Meaning string
	Tone    string
	Order   int
}

type Draw struct {
	Position string
	Card     Card
	Reversed bool
	Insight  string
}

type Reading struct {
	ID        string
	Question  string
	Spread    string
	Cards     []Draw
	Summary   string
	CreatedAt time.Time
}

type Library struct {
	mu       sync.RWMutex
	cards    []Card
	readings map[string]Reading
	rng      *mrand.Rand
}

func NewLibrary() *Library {
	seed := time.Now().UnixNano()
	if n, err := secureSeed(); err == nil {
		seed = int64(n)
	}
	return &Library{
		cards:    seedDeck(),
		readings: map[string]Reading{},
		rng:      mrand.New(mrand.NewSource(seed)),
	}
}

func secureSeed() (uint64, error) {
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}

func (l *Library) Cards() []Card {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Card, len(l.cards))
	copy(out, l.cards)
	return out
}

func (l *Library) Reading(question string) Reading {
	l.mu.Lock()
	defer l.mu.Unlock()

	if strings.TrimSpace(question) == "" {
		question = "What should I focus on right now?"
	}
	positions := []string{"Past", "Present", "Future"}
	deck := make([]Card, len(l.cards))
	copy(deck, l.cards)
	l.rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	draws := make([]Draw, 0, len(positions))
	for i, pos := range positions {
		card := deck[i]
		reversed := l.rng.Intn(2) == 0
		draws = append(draws, Draw{
			Position: pos,
			Card:     card,
			Reversed: reversed,
			Insight:  cardInsight(card, pos, reversed),
		})
	}

	id := fmt.Sprintf("reading-%d", time.Now().UnixNano())
	reading := Reading{
		ID:        id,
		Question:  question,
		Spread:    "Three-card spread",
		Cards:     draws,
		Summary:   summarizeReading(question, draws),
		CreatedAt: time.Now().UTC(),
	}
	l.readings[id] = reading
	return reading
}

func (l *Library) ReadingByID(id string) (Reading, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	r, ok := l.readings[id]
	return r, ok
}

func seedDeck() []Card {
	cards := make([]Card, 0, 78)
	majors := []struct {
		name    string
		meaning string
		tone    string
	}{
		{"The Fool", "New beginnings and trust", "open"},
		{"The Magician", "Skill, focus, and manifestation", "confident"},
		{"The High Priestess", "Intuition and hidden truth", "mystic"},
		{"The Empress", "Nurture, creativity, and abundance", "lush"},
		{"The Emperor", "Structure and authority", "grounded"},
		{"The Hierophant", "Tradition and guidance", "steady"},
		{"The Lovers", "Alignment and choice", "harmonious"},
		{"The Chariot", "Momentum and victory", "driven"},
		{"Strength", "Compassionate courage", "resolute"},
		{"The Hermit", "Reflection and inner light", "quiet"},
		{"Wheel of Fortune", "Cycles and turning points", "dynamic"},
		{"Justice", "Balance and truth", "clear"},
		{"The Hanged Man", "Surrender and a new perspective", "paused"},
		{"Death", "Transformation and release", "waking"},
		{"Temperance", "Integration and moderation", "flowing"},
		{"The Devil", "Attachment and shadow", "intense"},
		{"The Tower", "Disruption and revelation", "electric"},
		{"The Star", "Hope and renewal", "radiant"},
		{"The Moon", "Dreams and uncertainty", "lunar"},
		{"The Sun", "Clarity and joy", "golden"},
		{"Judgement", "Awakening and decision", "echoing"},
		{"The World", "Completion and wholeness", "complete"},
	}
	for i, major := range majors {
		cards = append(cards, Card{
			ID:      fmt.Sprintf("major-%02d", i),
			Name:    major.name,
			Arcana:  "Major",
			Meaning: major.meaning,
			Tone:    major.tone,
			Order:   i,
		})
	}

	suits := []struct {
		name    string
		element string
		meaning string
	}{
		{"Wands", "Fire", "drive, passion, and creative action"},
		{"Cups", "Water", "emotion, intuition, and connection"},
		{"Swords", "Air", "thought, truth, and conflict"},
		{"Pentacles", "Earth", "work, body, and resources"},
	}
	ranks := []string{"Ace", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine", "Ten", "Page", "Knight", "Queen", "King"}
	order := 100
	for _, suit := range suits {
		for _, rank := range ranks {
			cards = append(cards, Card{
				ID:      fmt.Sprintf("%s-%s", strings.ToLower(suit.name), strings.ToLower(rank)),
				Name:    fmt.Sprintf("%s of %s", rank, suit.name),
				Arcana:  "Minor",
				Suit:    suit.name,
				Rank:    rank,
				Element: suit.element,
				Meaning: fmt.Sprintf("%s in %s", rank, suit.meaning),
				Tone:    strings.ToLower(suit.name),
				Order:   order,
			})
			order++
		}
	}
	sort.Slice(cards, func(i, j int) bool { return cards[i].Order < cards[j].Order })
	return cards
}

func cardInsight(card Card, position string, reversed bool) string {
	state := "upright"
	if reversed {
		state = "reversed"
	}
	return fmt.Sprintf("%s in %s position (%s): %s", card.Name, position, state, card.Meaning)
}

func summarizeReading(question string, draws []Draw) string {
	parts := make([]string, 0, len(draws))
	for _, draw := range draws {
		parts = append(parts, fmt.Sprintf("%s suggests %s", draw.Position, draw.Card.Tone))
	}
	return fmt.Sprintf("For '%s', %s.", question, strings.Join(parts, ", "))
}
