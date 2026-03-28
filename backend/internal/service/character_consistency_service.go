package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"storyflow/internal/model"
	"storyflow/internal/repository"
	"storyflow/pkg/ai"
)

// CharacterConsistencyService handles character consistency for image generation
type CharacterConsistencyService struct {
	imageGenerator ai.ImageGenerator
	repo           *repository.StoryRepository
}

// NewCharacterConsistencyService creates a new character consistency service
func NewCharacterConsistencyService(imageGenerator ai.ImageGenerator, repo *repository.StoryRepository) *CharacterConsistencyService {
	return &CharacterConsistencyService{
		imageGenerator: imageGenerator,
		repo:           repo,
	}
}

// GenerateReferenceImage generates a reference image for a character
func (s *CharacterConsistencyService) GenerateReferenceImage(
	ctx context.Context,
	character *model.Character,
	style string,
) (*model.Character, error) {
	// 1. Generate visual prompt if not exists
	if character.VisualPrompt == "" {
		character.VisualPrompt = s.generateVisualPrompt(character)
	}

	// 2. Assign fixed seed if not exists
	if character.Seed == 0 {
		character.Seed = time.Now().UnixNano()
	}

	// 3. Call image generator
	// Note: Volcengine requires minimum 3686400 pixels (e.g., 1920x1920)
	req := &ai.ImageRequest{
		Prompt: fmt.Sprintf("%s style, %s, character reference sheet, front view, simple background",
			style, character.VisualPrompt),
		Width:  2048,
		Height: 2048,
		Style:  style,
		Seed:   character.Seed,
	}

	result, err := s.imageGenerator.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reference image: %w", err)
	}

	// 4. Update character record
	character.ReferenceImageURL = result.ImageURL
	character.ReferenceImageID = result.ID

	return character, nil
}

// generateVisualPrompt creates a detailed visual prompt from character attributes
func (s *CharacterConsistencyService) generateVisualPrompt(character *model.Character) string {
	var parts []string

	// Gender and age
	if character.Gender != "" {
		if character.Gender == "male" {
			parts = append(parts, "1boy")
		} else if character.Gender == "female" {
			parts = append(parts, "1girl")
		} else {
			parts = append(parts, character.Gender)
		}
	}

	if character.Age != "" {
		parts = append(parts, character.Age)
	}

	// Physical features
	if character.HairColor != "" {
		parts = append(parts, character.HairColor+" hair")
	}
	if character.EyeColor != "" {
		parts = append(parts, character.EyeColor+" eyes")
	}
	if character.BodyType != "" {
		parts = append(parts, character.BodyType)
	}

	// Clothing
	if character.Clothing != "" {
		parts = append(parts, character.Clothing)
	}

	// Description fallback
	if len(parts) == 0 && character.Description != "" {
		parts = append(parts, character.Description)
	}

	return strings.Join(parts, ", ")
}

// EnhanceScenePrompt enhances a scene prompt with character descriptions
func (s *CharacterConsistencyService) EnhanceScenePrompt(
	scene *model.Scene,
	characters []model.Character,
) string {
	// Build character map for quick lookup
	charMap := make(map[uuid.UUID]model.Character)
	for _, c := range characters {
		charMap[c.ID] = c
	}

	// Collect visual prompts for characters in this scene
	var charPrompts []string
	for _, charIDStr := range scene.CharacterIDs {
		charID, err := uuid.Parse(charIDStr)
		if err != nil {
			continue
		}
		if c, ok := charMap[charID]; ok && c.VisualPrompt != "" {
			charPrompts = append(charPrompts, c.VisualPrompt)
		}
	}

	// Combine with scene prompt
	if len(charPrompts) > 0 {
		return fmt.Sprintf("%s, %s",
			strings.Join(charPrompts, ", "),
			scene.ImagePrompt)
	}

	return scene.ImagePrompt
}

// GenerateSceneWithConsistency generates a scene image with character consistency
func (s *CharacterConsistencyService) GenerateSceneWithConsistency(
	ctx context.Context,
	scene *model.Scene,
	characters []model.Character,
	style string,
) (*ai.ImageResult, error) {
	// 1. Get characters in this scene
	var sceneChars []model.Character
	charMap := make(map[uuid.UUID]model.Character)
	for _, c := range characters {
		charMap[c.ID] = c
	}
	for _, charIDStr := range scene.CharacterIDs {
		charID, err := uuid.Parse(charIDStr)
		if err != nil {
			continue
		}
		if c, ok := charMap[charID]; ok {
			sceneChars = append(sceneChars, c)
		}
	}

	// 2. Enhance prompt
	enhancedPrompt := s.EnhanceScenePrompt(scene, characters)

	// 3. Calculate average seed for consistency
	var avgSeed int64
	hasReference := false
	var refImageURL string

	for _, c := range sceneChars {
		avgSeed += c.Seed
		if c.ReferenceImageURL != "" {
			hasReference = true
			refImageURL = c.ReferenceImageURL // Use first reference image
		}
	}
	if len(sceneChars) > 0 {
		avgSeed /= int64(len(sceneChars))
	}

	// 4. Create request
	req := &ai.ImageRequest{
		Prompt: fmt.Sprintf("%s style, %s", style, enhancedPrompt),
		Width:  2048,
		Height: 2048,
		Style:  style,
		Seed:   avgSeed,
	}

	// 5. Use img2img if supported and reference exists
	if hasReference {
		if img2img, ok := s.imageGenerator.(ai.ImageToImageGenerator); ok {
			return img2img.GenerateFromImage(ctx, req, refImageURL)
		}
	}

	return s.imageGenerator.Generate(ctx, req)
}

// SupportsImageToImage checks if the image generator supports img2img
func (s *CharacterConsistencyService) SupportsImageToImage() bool {
	_, ok := s.imageGenerator.(ai.ImageToImageGenerator)
	return ok
}

// GenerateAllReferenceImages generates reference images for all characters in a story
func (s *CharacterConsistencyService) GenerateAllReferenceImages(
	ctx context.Context,
	storyID uuid.UUID,
	style string,
) ([]model.Character, error) {
	// Get all characters
	characters, err := s.repo.GetCharactersByStoryID(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get characters: %w", err)
	}

	var updated []model.Character
	for i := range characters {
		updatedChar, err := s.GenerateReferenceImage(ctx, &characters[i], style)
		if err != nil {
			// Log error but continue with other characters
			fmt.Printf("Warning: failed to generate reference for %s: %v\n", characters[i].Name, err)
			continue
		}

		// Save to database
		if err := s.repo.UpdateCharacter(ctx, updatedChar); err != nil {
			fmt.Printf("Warning: failed to update character %s: %v\n", characters[i].Name, err)
			continue
		}

		updated = append(updated, *updatedChar)
	}

	return updated, nil
}