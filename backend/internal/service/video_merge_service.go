package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"storyflow/internal/model"
	"storyflow/internal/repository"
)

// VideoMergeService handles video merging operations
type VideoMergeService struct {
	repo        *repository.StoryRepository
	outputDir   string
	tempDir     string
}

// NewVideoMergeService creates a new video merge service
func NewVideoMergeService(repo *repository.StoryRepository, outputDir string) *VideoMergeService {
	if outputDir == "" {
		outputDir = "./uploads/videos"
	}
	tempDir := filepath.Join(outputDir, "temp")
	os.MkdirAll(outputDir, 0755)
	os.MkdirAll(tempDir, 0755)
	return &VideoMergeService{
		repo:      repo,
		outputDir: outputDir,
		tempDir:   tempDir,
	}
}

// MergeStoryVideos merges all scene videos into one final video
func (s *VideoMergeService) MergeStoryVideos(ctx context.Context, storyID uuid.UUID, options MergeOptions) (*MergeResult, error) {
	// Get story with scenes
	story, err := s.repo.GetWithRelations(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get story: %w", err)
	}

	// Filter scenes with videos, sorted by sequence
	var scenesWithVideos []model.Scene
	for _, scene := range story.Scenes {
		if scene.VideoURL != "" {
			scenesWithVideos = append(scenesWithVideos, scene)
		}
	}

	if len(scenesWithVideos) == 0 {
		return nil, fmt.Errorf("no videos to merge")
	}

	// Download all videos to temp directory
	videoFiles := make([]string, 0, len(scenesWithVideos))
	for _, scene := range scenesWithVideos {
		filename := fmt.Sprintf("scene_%d_%d.mp4", scene.Sequence, time.Now().Unix())
		localPath := filepath.Join(s.tempDir, filename)

		fmt.Printf("[VideoMerge] Downloading scene %d video...\n", scene.Sequence)
		if err := s.downloadFile(scene.VideoURL, localPath); err != nil {
			// Clean up downloaded files
			s.cleanupFiles(videoFiles)
			return nil, fmt.Errorf("failed to download scene %d video: %w", scene.Sequence, err)
		}
		videoFiles = append(videoFiles, localPath)
		fmt.Printf("[VideoMerge] Downloaded: %s\n", localPath)
	}

	// Generate output filename
	outputFilename := fmt.Sprintf("%s_merged_%s.mp4", storyID.String()[:8], time.Now().Format("20060302_150405"))
	outputPath := filepath.Join(s.outputDir, outputFilename)

	// Merge videos using ffmpeg
	fmt.Printf("[VideoMerge] Merging %d videos...\n", len(videoFiles))
	if err := s.mergeVideosWithFFmpeg(videoFiles, outputPath, options); err != nil {
		s.cleanupFiles(videoFiles)
		return nil, fmt.Errorf("failed to merge videos: %w", err)
	}

	// Clean up temp files
	s.cleanupFiles(videoFiles)

	// Build output URL
	outputURL := "/api/v1/videos/view?file=" + outputFilename

	// Update story with merged video URL
	story.MergedVideoURL = outputURL
	if err := s.repo.Update(ctx, story); err != nil {
		fmt.Printf("[VideoMerge] Warning: failed to update story with merged video URL: %v\n", err)
	}

	// Return result
	result := &MergeResult{
		OutputPath:    outputPath,
		OutputURL:     outputURL,
		TotalScenes:   len(scenesWithVideos),
		TotalDuration: s.getVideoDuration(outputPath),
	}

	fmt.Printf("[VideoMerge] Merge completed: %s (duration: %.1fs)\n", outputPath, result.TotalDuration)
	return result, nil
}

// MergeOptions defines options for video merging
type MergeOptions struct {
	Transition   string  `json:"transition"`   // none, fade, crossfade
	TransitionDuration float64 `json:"transition_duration"` // seconds
	AddAudio     bool    `json:"add_audio"`    // whether to add background audio
	AudioPath    string  `json:"audio_path"`   // path to background audio
	OutputFPS    int     `json:"output_fps"`   // output video fps
	OutputWidth  int     `json:"output_width"` // output width
	OutputHeight int     `json:"output_height"` // output height
}

// MergeResult contains the result of video merging
type MergeResult struct {
	OutputPath    string  `json:"output_path"`
	OutputURL     string  `json:"output_url"`
	TotalScenes   int     `json:"total_scenes"`
	TotalDuration float64 `json:"total_duration"`
}

// downloadFile downloads a file from URL to local path
func (s *VideoMergeService) downloadFile(url, localPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// mergeVideosWithFFmpeg merges videos using ffmpeg
func (s *VideoMergeService) mergeVideosWithFFmpeg(videoFiles []string, outputPath string, options MergeOptions) error {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found, please install ffmpeg")
	}

	// Get absolute paths for all video files
	absVideoFiles := make([]string, len(videoFiles))
	for i, f := range videoFiles {
		absPath, err := filepath.Abs(f)
		if err != nil {
			absPath = f
		}
		absVideoFiles[i] = absPath
	}

	// Create concat file list with absolute paths
	concatFile := filepath.Join(s.tempDir, "concat_list.txt")
	concatContent := ""
	for _, f := range absVideoFiles {
		// Escape single quotes in filename
		escaped := strings.ReplaceAll(f, "'", "'\\''")
		concatContent += fmt.Sprintf("file '%s'\n", escaped)
	}
	if err := os.WriteFile(concatFile, []byte(concatContent), 0644); err != nil {
		return fmt.Errorf("failed to create concat file: %w", err)
	}
	defer os.Remove(concatFile)

	// Get absolute paths for concat file and output
	absConcatFile, _ := filepath.Abs(concatFile)
	absOutputPath, _ := filepath.Abs(outputPath)

	// Build ffmpeg command
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", absConcatFile,
	}

	// Add video filters if transitions are needed
	if options.Transition == "fade" || options.Transition == "crossfade" {
		// Build complex filter for crossfade transitions
		filterComplex := s.buildTransitionFilter(videoFiles, options.TransitionDuration)
		args = append(args, "-filter_complex", filterComplex)
	} else if options.OutputWidth > 0 && options.OutputHeight > 0 {
		// Scale to specified resolution
		args = append(args, "-vf", fmt.Sprintf("scale=%d:%d", options.OutputWidth, options.OutputHeight))
	}

	// Add audio if specified
	if options.AddAudio && options.AudioPath != "" {
		args = append(args, "-i", options.AudioPath, "-c:v", "copy", "-c:a", "aac", "-shortest")
	} else {
		args = append(args, "-c", "copy")
	}

	// Output options
	args = append(args,
		"-movflags", "+faststart", // Enable streaming
		"-y", // Overwrite output
		absOutputPath,
	)

	fmt.Printf("[VideoMerge] FFmpeg command: ffmpeg %s\n", strings.Join(args, " "))

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}

	return nil
}

// buildTransitionFilter builds ffmpeg filter for transitions
func (s *VideoMergeService) buildTransitionFilter(videoFiles []string, duration float64) string {
	// For simplicity, use xfade filter for crossfade transitions
	// This requires all inputs to have the same resolution
	// For now, we'll use a simpler approach without transitions
	return ""
}

// getVideoDuration gets the duration of a video file using ffprobe
func (s *VideoMergeService) getVideoDuration(path string) float64 {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return 0
	}

	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	var duration float64
	fmt.Sscanf(string(output), "%f", &duration)
	return duration
}

// cleanupFiles removes temporary files
func (s *VideoMergeService) cleanupFiles(files []string) {
	for _, f := range files {
		os.Remove(f)
	}
}