package service

import (
	"context"
	"fmt"

	"convoy/internal/models"

	"github.com/google/uuid"
)

type ParticipantService struct {
	participantRepo ParticipantRepository
	routeRepo       RouteRepository
}

func NewParticipantService(participantRepo ParticipantRepository, routeRepo RouteRepository) *ParticipantService {
	return &ParticipantService{
		participantRepo: participantRepo,
		routeRepo:       routeRepo,
	}
}

// JoinRoute allows a user to join a route
func (s *ParticipantService) JoinRoute(ctx context.Context, routeID, userID uuid.UUID) error {
	route, err := s.routeRepo.GetByID(ctx, routeID.String())
	if err != nil {
		return fmt.Errorf("failed to get route: %w", err)
	}

	if route.Status != models.RouteStatusPlanned {
		return fmt.Errorf("cannot join route with status: %s", route.Status)
	}

	existing, err := s.participantRepo.GetByRouteAndUser(ctx, routeID.String(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to check existing participant: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("already joined this route")
	}

	if route.MaxParticipants != nil {
		count, err := s.participantRepo.CountByRoute(ctx, routeID.String())
		if err != nil {
			return fmt.Errorf("failed to count participants: %w", err)
		}
		if count >= *route.MaxParticipants {
			return fmt.Errorf("route is full")
		}
	}

	participant := &models.Participant{
		ID:      uuid.New(),
		RouteID: routeID,
		UserID:  userID,
		Status:  models.ParticipantStatusActive,
	}

	if err := s.participantRepo.Create(ctx, participant); err != nil {
		return fmt.Errorf("failed to create participant: %w", err)
	}

	return nil
}

// LeaveRoute allows a user to leave a route
func (s *ParticipantService) LeaveRoute(ctx context.Context, routeID, userID uuid.UUID) error {
	route, err := s.routeRepo.GetByID(ctx, routeID.String())
	if err != nil {
		return fmt.Errorf("failed to get route: %w", err)
	}

	if route.CreatorID == userID {
		return fmt.Errorf("route creator cannot leave the route")
	}

	participant, err := s.participantRepo.GetByRouteAndUser(ctx, routeID.String(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to check participant: %w", err)
	}
	if participant == nil {
		return fmt.Errorf("not a participant of this route")
	}

	if err := s.participantRepo.DeleteByRouteAndUser(ctx, routeID.String(), userID.String()); err != nil {
		return fmt.Errorf("failed to leave route: %w", err)
	}

	return nil
}

// ListParticipants returns all participants for a route
func (s *ParticipantService) ListParticipants(ctx context.Context, routeID uuid.UUID) ([]models.ParticipantWithUser, error) {
	participants, err := s.participantRepo.ListByRoute(ctx, routeID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list participants: %w", err)
	}

	return participants, nil
}

// RemoveParticipant allows a route creator to remove a participant
func (s *ParticipantService) RemoveParticipant(ctx context.Context, routeID, userID, creatorID uuid.UUID) error {
	route, err := s.routeRepo.GetByID(ctx, routeID.String())
	if err != nil {
		return fmt.Errorf("failed to get route: %w", err)
	}

	if route.CreatorID != creatorID {
		return fmt.Errorf("only route creator can remove participants")
	}

	if userID == creatorID {
		return fmt.Errorf("cannot remove route creator")
	}

	participant, err := s.participantRepo.GetByRouteAndUser(ctx, routeID.String(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to check participant: %w", err)
	}
	if participant == nil {
		return fmt.Errorf("user is not a participant of this route")
	}

	if err := s.participantRepo.DeleteByRouteAndUser(ctx, routeID.String(), userID.String()); err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	return nil
}
