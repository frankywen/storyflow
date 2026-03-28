package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"storyflow/pkg/ai"
)

// StoryParserAgent parses stories and extracts structured data
type StoryParserAgent struct {
	llmProvider ai.LLMProvider
	style       string
}

// NewStoryParserAgent creates a new story parser agent
func NewStoryParserAgent(llmProvider ai.LLMProvider) *StoryParserAgent {
	return &StoryParserAgent{
		llmProvider: llmProvider,
		style:       "manga", // default style
	}
}

// SetStyle sets the visual style for image prompts
func (a *StoryParserAgent) SetStyle(style string) {
	a.style = style
}

// ParseResult represents the parsed story structure
type ParseResult struct {
	Summary    string      `json:"summary"`
	Genre      string      `json:"genre"`
	Tone       string      `json:"tone"`
	Setting    string      `json:"setting"` // story background/setting
	Characters []Character `json:"characters"`
	Scenes     []Scene     `json:"scenes"`
	Questions  []string    `json:"questions"`
}

// Character represents a parsed character
type Character struct {
	Name        string `json:"name"`
	Role        string `json:"role"` // protagonist, antagonist, supporting
	Gender      string `json:"gender"`
	Age         string `json:"age"`
	Description string `json:"description"`
	Appearance  string `json:"appearance"`
	Personality string `json:"personality"`
	// For image generation
	VisualPrompt string `json:"visual_prompt"` // English prompt for consistent generation
}

// Scene represents a parsed scene
type Scene struct {
	Sequence    int      `json:"sequence"`
	Title       string   `json:"title"`
	Location    string   `json:"location"`
	TimeOfDay   string   `json:"time_of_day"`
	Mood        string   `json:"mood"`
	Description string   `json:"description"`
	Characters  []string `json:"characters"`
	Dialogue    string   `json:"dialogue"`
	Narration   string   `json:"narration"`
	Action      string   `json:"action"` // character actions
	ImagePrompt string   `json:"image_prompt"`
}

// Parse parses a story text and extracts structured data
func (a *StoryParserAgent) Parse(ctx context.Context, storyText string) (*ParseResult, error) {
	systemPrompt := a.getSystemPrompt()
	userPrompt := fmt.Sprintf("请分析以下故事：\n\n%s", storyText)

	resp, err := a.llmProvider.SendMessage(ctx, systemPrompt, userPrompt, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	text := resp.Content

	// Extract JSON from response
	jsonStr := extractJSON(text)
	if jsonStr == "" {
		return nil, fmt.Errorf("failed to extract JSON from response: %s", text)
	}

	var result ParseResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Post-process: generate visual prompts for characters
	for i := range result.Characters {
		if result.Characters[i].VisualPrompt == "" {
			result.Characters[i].VisualPrompt = a.generateCharacterPrompt(result.Characters[i])
		}
	}

	// Post-process: enhance scene image prompts
	for i := range result.Scenes {
		if result.Scenes[i].ImagePrompt == "" {
			result.Scenes[i].ImagePrompt = a.generateScenePrompt(result.Scenes[i], result.Characters)
		}
	}

	return &result, nil
}

// ParseWithStyle parses with a specific visual style
func (a *StoryParserAgent) ParseWithStyle(ctx context.Context, storyText, style string) (*ParseResult, error) {
	a.style = style
	return a.Parse(ctx, storyText)
}

// ExtractCharacters extracts only characters from story
func (a *StoryParserAgent) ExtractCharacters(ctx context.Context, storyText string) ([]Character, error) {
	result, err := a.Parse(ctx, storyText)
	if err != nil {
		return nil, err
	}
	return result.Characters, nil
}

// GenerateScenePrompts generates image prompts for scenes
func (a *StoryParserAgent) GenerateScenePrompts(ctx context.Context, scenes []Scene, characters []Character, style string) ([]string, error) {
	prompts := make([]string, len(scenes))

	charMap := make(map[string]Character)
	for _, c := range characters {
		charMap[c.Name] = c
	}

	for i, scene := range scenes {
		// Build prompt with character visual descriptions
		var charPrompts []string
		for _, name := range scene.Characters {
			if c, ok := charMap[name]; ok {
				charPrompts = append(charPrompts, c.VisualPrompt)
			}
		}

		scenePrompt := scene.ImagePrompt
		if len(charPrompts) > 0 {
			scenePrompt = fmt.Sprintf("%s, %s", strings.Join(charPrompts, ", "), scenePrompt)
		}

		// Add style prefix
		prompts[i] = fmt.Sprintf("%s style, %s, high quality, detailed", style, scenePrompt)
	}

	return prompts, nil
}

func (a *StoryParserAgent) getSystemPrompt() string {
	return `你是一个专业的小说分析师和分镜设计师。你的任务是将小说文本转换为结构化的视觉分镜脚本。

## 输出要求
你必须以JSON格式输出分析结果，格式如下：
{
  "summary": "故事摘要（一句话）",
  "genre": "类型（言情/悬疑/奇幻/都市/科幻等）",
  "tone": "基调（轻松/沉重/温馨/紧张等）",
  "setting": "故事背景设定",
  "characters": [
    {
      "name": "角色名",
      "role": "protagonist/antagonist/supporting",
      "gender": "male/female",
      "age": "年龄描述如 20s, 30s, young, middle-aged",
      "description": "角色描述",
      "appearance": "外貌描述（中文）",
      "personality": "性格特点",
      "visual_prompt": "英文视觉描述，用于AI生图，包含发型发色、眼睛颜色、肤色、身材、穿着等详细描述"
    }
  ],
  "scenes": [
    {
      "sequence": 1,
      "title": "场景标题",
      "location": "地点",
      "time_of_day": "morning/afternoon/evening/night",
      "mood": "情绪氛围",
      "description": "场景描述",
      "characters": ["出现的角色名"],
      "dialogue": "关键对白",
      "narration": "旁白说明",
      "action": "角色动作",
      "image_prompt": "英文场景描述，用于AI生图，包含角色位置、动作、表情、场景、光线、构图等"
    }
  ],
  "questions": ["需要进一步确认的问题（如有）"]
}

## 重要规则
1. 每个场景应该是一个独立的视觉画面，可以转化为一张图片
2. visual_prompt 和 image_prompt 必须是英文，因为AI生图工具主要支持英文
3. 角色的 visual_prompt 要足够详细，确保在不同场景中保持一致性
4. image_prompt 要描述画面中的所有元素：角色、动作、表情、场景、光线
5. 按故事发展顺序排列场景，一般5-15个场景为宜
6. 不要遗漏重要情节转折点

## 视觉风格参考
- manga: 日式漫画风格
- manhwa: 韩式漫画风格
- western_comic: 美式漫画风格
- realistic: 写实风格
- anime: 动漫风格`
}

func (a *StoryParserAgent) generateCharacterPrompt(c Character) string {
	var parts []string

	if c.Gender != "" {
		if c.Gender == "female" {
			parts = append(parts, "1girl")
		} else {
			parts = append(parts, "1boy")
		}
	}

	if c.Appearance != "" {
		parts = append(parts, c.Appearance)
	}

	if c.Age != "" {
		parts = append(parts, c.Age)
	}

	return strings.Join(parts, ", ")
}

func (a *StoryParserAgent) generateScenePrompt(scene Scene, characters []Character) string {
	var parts []string

	// Add location
	if scene.Location != "" {
		parts = append(parts, scene.Location)
	}

	// Add mood/atmosphere
	if scene.Mood != "" {
		parts = append(parts, scene.Mood+" atmosphere")
	}

	// Add time of day
	if scene.TimeOfDay != "" {
		parts = append(parts, scene.TimeOfDay+" lighting")
	}

	// Add action
	if scene.Action != "" {
		parts = append(parts, scene.Action)
	}

	// Add description
	if scene.Description != "" {
		parts = append(parts, scene.Description)
	}

	return strings.Join(parts, ", ")
}

// extractJSON extracts JSON from text that might contain other content
func extractJSON(text string) string {
	// Find JSON block
	start := strings.Index(text, "{")
	if start == -1 {
		return ""
	}

	// Count braces to find the end
	depth := 0
	inString := false
	escape := false

	for i := start; i < len(text); i++ {
		c := text[i]

		if escape {
			escape = false
			continue
		}

		if c == '\\' {
			escape = true
			continue
		}

		if c == '"' {
			inString = !inString
			continue
		}

		if !inString {
			if c == '{' {
				depth++
			} else if c == '}' {
				depth--
				if depth == 0 {
					return text[start : i+1]
				}
			}
		}
	}

	return ""
}