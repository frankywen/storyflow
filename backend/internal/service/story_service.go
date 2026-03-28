package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"storyflow/internal/agent"
	"storyflow/internal/model"
	"storyflow/internal/repository"
	"storyflow/pkg/ai"
)

// StoryService handles story-related business logic
type StoryService struct {
	repo           *repository.StoryRepository
	parserAgent    *agent.StoryParserAgent
	imageGenerator ai.ImageGenerator
}

// NewStoryService creates a new story service
func NewStoryService(repo *repository.StoryRepository, llmProvider ai.LLMProvider, imageGenerator ai.ImageGenerator) *StoryService {
	return &StoryService{
		repo:           repo,
		parserAgent:    agent.NewStoryParserAgent(llmProvider),
		imageGenerator: imageGenerator,
	}
}

// CreateStory creates a new story
func (s *StoryService) CreateStory(ctx context.Context, title, content string) (*model.Story, error) {
	story := &model.Story{
		ID:      uuid.New(),
		Title:   title,
		Content: content,
		Status:  "pending",
	}

	if err := s.repo.Create(ctx, story); err != nil {
		return nil, fmt.Errorf("failed to create story: %w", err)
	}

	return story, nil
}

// GetStory retrieves a story by ID
func (s *StoryService) GetStory(ctx context.Context, id uuid.UUID) (*model.Story, error) {
	story, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get story: %w", err)
	}
	return story, nil
}

// GetStoryWithRelations retrieves a story with all related data
func (s *StoryService) GetStoryWithRelations(ctx context.Context, id uuid.UUID) (*model.Story, error) {
	story, err := s.repo.GetWithRelations(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get story: %w", err)
	}
	return story, nil
}

// ListStories lists stories with pagination
func (s *StoryService) ListStories(ctx context.Context, page, pageSize int) ([]model.Story, int64, error) {
	offset := (page - 1) * pageSize
	stories, total, err := s.repo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list stories: %w", err)
	}
	return stories, total, nil
}

// ParseStory parses a story and extracts characters and scenes
func (s *StoryService) ParseStory(ctx context.Context, id uuid.UUID, style string) (*model.Story, error) {
	// Get the story
	story, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get story: %w", err)
	}

	// Update status
	story.Status = "parsing"
	s.repo.Update(ctx, story)

	// Parse the story
	result, err := s.parserAgent.ParseWithStyle(ctx, story.Content, style)
	if err != nil {
		story.Status = "error"
		s.repo.Update(ctx, story)
		return nil, fmt.Errorf("failed to parse story: %w", err)
	}

	// Update story with parsed data
	story.Summary = result.Summary
	story.Genre = result.Genre
	story.Status = "parsed"

	// Save characters and build name-to-ID mapping
	charNameToID := make(map[string]uuid.UUID)
	for i := range result.Characters {
		char := &model.Character{
			ID:             uuid.New(),
			StoryID:        story.ID,
			Name:           result.Characters[i].Name,
			Description:    result.Characters[i].Description,
			Gender:         result.Characters[i].Gender,
			Age:            result.Characters[i].Age,
			PromptTemplate: result.Characters[i].VisualPrompt,
		}
		if err := s.repo.CreateCharacter(ctx, char); err != nil {
			return nil, fmt.Errorf("failed to create character: %w", err)
		}
		story.Characters = append(story.Characters, *char)
		charNameToID[char.Name] = char.ID
	}

	// Save scenes with character associations
	for i := range result.Scenes {
		// Convert character names to IDs
		var characterIDs []string
		for _, charName := range result.Scenes[i].Characters {
			if charID, ok := charNameToID[charName]; ok {
				characterIDs = append(characterIDs, charID.String())
			}
		}

		scene := &model.Scene{
			ID:           uuid.New(),
			StoryID:      story.ID,
			Sequence:     result.Scenes[i].Sequence,
			Title:        result.Scenes[i].Title,
			Description:  result.Scenes[i].Description,
			Location:     result.Scenes[i].Location,
			TimeOfDay:    result.Scenes[i].TimeOfDay,
			Mood:         result.Scenes[i].Mood,
			Dialogue:     result.Scenes[i].Dialogue,
			Narration:    result.Scenes[i].Narration,
			ImagePrompt:  result.Scenes[i].ImagePrompt,
			CharacterIDs: characterIDs,
			Status:       "pending",
		}
		if err := s.repo.CreateScene(ctx, scene); err != nil {
			return nil, fmt.Errorf("failed to create scene: %w", err)
		}
		story.Scenes = append(story.Scenes, *scene)
	}

	// Update story
	if err := s.repo.Update(ctx, story); err != nil {
		return nil, fmt.Errorf("failed to update story: %w", err)
	}

	return story, nil
}

// DeleteStory deletes a story
func (s *StoryService) DeleteStory(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete story: %w", err)
	}
	return nil
}

// GetCharacters returns characters for a story
func (s *StoryService) GetCharacters(ctx context.Context, storyID uuid.UUID) ([]model.Character, error) {
	return s.repo.GetCharactersByStoryID(ctx, storyID)
}

// GetScenes returns scenes for a story
func (s *StoryService) GetScenes(ctx context.Context, storyID uuid.UUID) ([]model.Scene, error) {
	return s.repo.GetScenesByStoryID(ctx, storyID)
}