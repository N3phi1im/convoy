package models

import (
	"time"

	"github.com/google/uuid"
)

type ParticipantStatus string

const (
	ParticipantStatusActive  ParticipantStatus = "active"
	ParticipantStatusLeft    ParticipantStatus = "left"
	ParticipantStatusRemoved ParticipantStatus = "removed"
)

type Participant struct {
	ID       uuid.UUID         `json:"id" db:"id"`
	RouteID  uuid.UUID         `json:"route_id" db:"route_id"`
	UserID   uuid.UUID         `json:"user_id" db:"user_id"`
	Status   ParticipantStatus `json:"status" db:"status"`
	JoinedAt time.Time         `json:"joined_at" db:"joined_at"`
	LeftAt   *time.Time        `json:"left_at,omitempty" db:"left_at"`

	// Relations
	User *UserResponse `json:"user,omitempty" db:"-"`
}

// ParticipantWithUser includes participant with user details
type ParticipantWithUser struct {
	Participant
	User UserResponse `json:"user"`
}
