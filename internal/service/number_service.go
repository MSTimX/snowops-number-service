package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"

	"number-service/internal/repository"
	"number-service/internal/utils"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound      = errors.New("not found")
)

type NumberService struct {
	repo *repository.NumberRepository
	log  zerolog.Logger
}

func NewNumberService(repo *repository.NumberRepository, log zerolog.Logger) *NumberService {
	return &NumberService{
		repo: repo,
		log:  log,
	}
}

func (s *NumberService) NormalizeAndCheck(ctx context.Context, plateNumber string) (*CheckResult, error) {
	normalized := utils.NormalizePlate(plateNumber)
	if normalized == "" {
		return nil, fmt.Errorf("%w: plate number cannot be empty", ErrInvalidInput)
	}

	plateID, err := s.repo.GetOrCreatePlate(ctx, normalized, plateNumber)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get or create plate")
		return nil, fmt.Errorf("failed to get or create plate: %w", err)
	}

	hits, err := s.repo.FindListsForPlate(ctx, plateID)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to find lists for plate")
		return nil, fmt.Errorf("failed to find lists for plate: %w", err)
	}

	return &CheckResult{
		PlateID:  plateID,
		Plate:    normalized,
		Original: plateNumber,
		Hits:     hits,
	}, nil
}

func (s *NumberService) AddToWhitelist(ctx context.Context, plateNumber string, note *string) error {
	return s.addToList(ctx, plateNumber, "default_whitelist", note)
}

func (s *NumberService) AddToBlacklist(ctx context.Context, plateNumber string, note *string) error {
	return s.addToList(ctx, plateNumber, "default_blacklist", note)
}

func (s *NumberService) RemoveFromWhitelist(ctx context.Context, plateNumber string) error {
	return s.removeFromList(ctx, plateNumber, "default_whitelist")
}

func (s *NumberService) RemoveFromBlacklist(ctx context.Context, plateNumber string) error {
	return s.removeFromList(ctx, plateNumber, "default_blacklist")
}

func (s *NumberService) addToList(ctx context.Context, plateNumber, listName string, note *string) error {
	normalized := utils.NormalizePlate(plateNumber)
	if normalized == "" {
		return fmt.Errorf("%w: plate number cannot be empty", ErrInvalidInput)
	}

	plateID, err := s.repo.GetOrCreatePlate(ctx, normalized, plateNumber)
	if err != nil {
		return fmt.Errorf("failed to get or create plate: %w", err)
	}

	list, err := s.repo.GetListByName(ctx, listName)
	if err != nil {
		return fmt.Errorf("failed to get list: %w", err)
	}

	if err := s.repo.AddPlateToList(ctx, list.ID, plateID, note); err != nil {
		return fmt.Errorf("failed to add plate to list: %w", err)
	}

	return nil
}

func (s *NumberService) removeFromList(ctx context.Context, plateNumber, listName string) error {
	normalized := utils.NormalizePlate(plateNumber)
	if normalized == "" {
		return fmt.Errorf("%w: plate number cannot be empty", ErrInvalidInput)
	}

	plates, err := s.repo.FindPlatesByNormalized(ctx, normalized)
	if err != nil {
		return fmt.Errorf("failed to find plate: %w", err)
	}
	if len(plates) == 0 {
		return fmt.Errorf("%w: plate not found", ErrNotFound)
	}

	list, err := s.repo.GetListByName(ctx, listName)
	if err != nil {
		return fmt.Errorf("failed to get list: %w", err)
	}

	if err := s.repo.RemovePlateFromList(ctx, list.ID, plates[0].ID); err != nil {
		return fmt.Errorf("failed to remove plate from list: %w", err)
	}

	return nil
}

type CheckResult struct {
	PlateID  int64                `json:"plate_id"`
	Plate    string               `json:"plate"`
	Original string               `json:"original"`
	Hits     []repository.ListHit `json:"hits"`
}

