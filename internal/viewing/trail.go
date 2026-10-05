package viewing

import (
	"errors"
	"strings"
	"time"
)

// DecisionStatus is where a rental stands in the user's decision.
type DecisionStatus string

// The five decision statuses.
const (
	StatusUnreviewed       DecisionStatus = "unreviewed"
	StatusViewingPlanned   DecisionStatus = "viewing_planned"
	StatusConsidering      DecisionStatus = "considering"
	StatusRejected         DecisionStatus = "rejected"
	StatusReadyToNegotiate DecisionStatus = "ready_to_negotiate"
)

// Errors returned by DecisionTrail mutators.
var (
	ErrInvalidStatus = errors.New("viewing: invalid decision status")
	ErrEmptyAnswer   = errors.New("viewing: answer is empty")
)

// Valid reports whether s is one of the five known statuses.
func (s DecisionStatus) Valid() bool {
	switch s {
	case StatusUnreviewed, StatusViewingPlanned, StatusConsidering, StatusRejected, StatusReadyToNegotiate:
		return true
	default:
		return false
	}
}

// DecisionTrail is the user's recorded progress on one rental. Answers are
// keyed by item ID so regenerating the checklist never overwrites them.
type DecisionTrail struct {
	RentalID   int64
	Status     DecisionStatus
	NextAction *time.Time
	Answers    map[ItemID]string
}

// NewTrail returns an empty Unreviewed trail for a rental.
func NewTrail(rentalID int64) DecisionTrail {
	return DecisionTrail{RentalID: rentalID, Status: StatusUnreviewed, Answers: map[ItemID]string{}}
}

// SetAnswer records or edits the answer for an item. Blank text is rejected;
// use ClearAnswer to remove an answer.
func (t *DecisionTrail) SetAnswer(id ItemID, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return ErrEmptyAnswer
	}

	if t.Answers == nil {
		t.Answers = map[ItemID]string{}
	}

	t.Answers[id] = text

	return nil
}

// ClearAnswer removes the answer for an item, if any.
func (t *DecisionTrail) ClearAnswer(id ItemID) {
	delete(t.Answers, id)
}

// SetStatus changes the status and the optional next-action date. Any known
// status may follow any other, so a rejected rental can be reopened. A nil
// nextAction clears the date; an unknown status changes nothing.
func (t *DecisionTrail) SetStatus(status DecisionStatus, nextAction *time.Time) error {
	if !status.Valid() {
		return ErrInvalidStatus
	}

	t.Status = status

	if nextAction == nil {
		t.NextAction = nil

		return nil
	}

	due := *nextAction
	t.NextAction = &due

	return nil
}
