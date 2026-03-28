package ai

import (
	"context"
	"fmt"
	"time"
)

// ComfyUIImageGenerator implements ImageGenerator for ComfyUI
type ComfyUIImageGenerator struct {
	client     *ComfyUIClient
	workflowDir string
}

// NewComfyUIImageGenerator creates a new ComfyUI image generator
func NewComfyUIImageGenerator(cfg ImageGeneratorConfig) *ComfyUIImageGenerator {
	return &ComfyUIImageGenerator{
		client: NewComfyUIClient(cfg.BaseURL),
	}
}

// GetName returns the provider name
func (g *ComfyUIImageGenerator) GetName() string {
	return "comfyui"
}

// Generate generates a single image from prompt
func (g *ComfyUIImageGenerator) Generate(ctx context.Context, req *ImageRequest) (*ImageResult, error) {
	// Set defaults
	if req.Width == 0 {
		req.Width = 1024
	}
	if req.Height == 0 {
		req.Height = 1024
	}
	if req.Steps == 0 {
		req.Steps = 20
	}
	if req.Seed == 0 {
		req.Seed = time.Now().UnixNano()
	}
	if req.NegativePrompt == "" {
		req.NegativePrompt = "low quality, blurry, distorted, ugly, bad anatomy, worst quality, watermark, text"
	}

	// Build workflow
	workflow := g.buildWorkflow(req)

	// Queue to ComfyUI
	clientID := fmt.Sprintf("storyflow-%d", time.Now().UnixNano())
	resp, err := g.client.QueuePrompt(ctx, workflow, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to queue prompt: %w", err)
	}

	// Wait for completion
	timeout := 3 * time.Minute
	history, err := g.client.WaitForCompletion(ctx, resp.PromptID, timeout)
	if err != nil {
		return nil, fmt.Errorf("generation timeout: %w", err)
	}

	// Extract first image
	for _, output := range history.Outputs {
		for _, img := range output.Images {
			imageURL := fmt.Sprintf("/api/v1/images/view?filename=%s&subfolder=%s&type=%s",
				img.Filename, img.Subfolder, img.Type)

			return &ImageResult{
				ID:       resp.PromptID,
				ImageURL: imageURL,
				Seed:     req.Seed,
				Width:    req.Width,
				Height:   req.Height,
				Model:    "sdxl",
			}, nil
		}
	}

	return nil, fmt.Errorf("no images generated")
}

// GenerateBatch generates multiple images
func (g *ComfyUIImageGenerator) GenerateBatch(ctx context.Context, req *ImageRequest, count int) ([]*ImageResult, error) {
	results := make([]*ImageResult, count)
	for i := 0; i < count; i++ {
		// Use different seeds for each image
		reqCopy := *req
		if req.Seed == 0 {
			reqCopy.Seed = time.Now().UnixNano() + int64(i)
		} else {
			reqCopy.Seed = req.Seed + int64(i)
		}

		result, err := g.Generate(ctx, &reqCopy)
		if err != nil {
			return results, err
		}
		results[i] = result
	}
	return results, nil
}

// buildWorkflow builds a ComfyUI workflow for image generation
func (g *ComfyUIImageGenerator) buildWorkflow(req *ImageRequest) map[string]interface{} {
	stylePrompt := req.Prompt
	if req.Style != "" {
		stylePrompt = fmt.Sprintf("%s style, %s", req.Style, req.Prompt)
	}

	return map[string]interface{}{
		"1": map[string]interface{}{
			"class_type": "KSampler",
			"inputs": map[string]interface{}{
				"seed":          req.Seed,
				"steps":         req.Steps,
				"cfg":           7,
				"sampler_name":  "euler",
				"scheduler":     "normal",
				"denoise":       1,
				"model":         []interface{}{"4", 0},
				"positive":      []interface{}{"6", 0},
				"negative":      []interface{}{"7", 0},
				"latent_image":  []interface{}{"5", 0},
			},
		},
		"2": map[string]interface{}{
			"class_type": "VAEDecode",
			"inputs": map[string]interface{}{
				"samples": []interface{}{"1", 0},
				"vae":     []interface{}{"4", 2},
			},
		},
		"3": map[string]interface{}{
			"class_type": "SaveImage",
			"inputs": map[string]interface{}{
				"filename_prefix": "storyflow",
				"images":          []interface{}{"2", 0},
			},
		},
		"4": map[string]interface{}{
			"class_type": "CheckpointLoaderSimple",
			"inputs": map[string]interface{}{
				"ckpt_name": "sdxl_base_1.0.safetensors",
			},
		},
		"5": map[string]interface{}{
			"class_type": "EmptyLatentImage",
			"inputs": map[string]interface{}{
				"width":      req.Width,
				"height":     req.Height,
				"batch_size": 1,
			},
		},
		"6": map[string]interface{}{
			"class_type": "CLIPTextEncode",
			"inputs": map[string]interface{}{
				"text": stylePrompt,
				"clip": []interface{}{"4", 1},
			},
		},
		"7": map[string]interface{}{
			"class_type": "CLIPTextEncode",
			"inputs": map[string]interface{}{
				"text": req.NegativePrompt,
				"clip": []interface{}{"4", 1},
			},
		},
	}
}