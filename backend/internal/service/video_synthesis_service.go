package service

import (
	"context"
	"errors"
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

// VideoSynthesisService handles video synthesis operations
// combining audio, subtitles, and video into final output
type VideoSynthesisService struct {
	audioRepo    *repository.AudioRepository
	storyRepo    *repository.StoryRepository
	subtitleSvc  *SubtitleService
	outputDir    string
	videoBaseURL string
	tempDir      string
}

// NewVideoSynthesisService creates a new video synthesis service
func NewVideoSynthesisService(
	audioRepo *repository.AudioRepository,
	storyRepo *repository.StoryRepository,
	subtitleSvc *SubtitleService,
	outputDir string,
	videoBaseURL string,
) *VideoSynthesisService {
	if outputDir == "" {
		outputDir = "./uploads/synthesis"
	}
	tempDir := filepath.Join(outputDir, "temp")
	os.MkdirAll(outputDir, 0755)
	os.MkdirAll(tempDir, 0755)
	return &VideoSynthesisService{
		audioRepo:    audioRepo,
		storyRepo:    storyRepo,
		subtitleSvc:  subtitleSvc,
		outputDir:    outputDir,
		videoBaseURL: videoBaseURL,
		tempDir:      tempDir,
	}
}

// SynthesizeVideoRequest represents a video synthesis request
type SynthesizeVideoRequest struct {
	StoryID    uuid.UUID `json:"story_id"`
	VideoURL   string    `json:"video_url"`   // Base video URL (optional, uses story's merged video if empty)
	AddAudio   bool      `json:"add_audio"`   // Whether to add audio track
	AddSubtitle bool     `json:"add_subtitle"` // Whether to add subtitles
}

// SynthesizeVideo starts an async video synthesis task
func (s *VideoSynthesisService) SynthesizeVideo(ctx context.Context, req SynthesizeVideoRequest) (*model.VideoSynthesisTask, error) {
	// Validate story exists
	story, err := s.storyRepo.GetWithRelations(ctx, req.StoryID)
	if err != nil {
		return nil, fmt.Errorf("story not found: %w", err)
	}

	// Determine video source
	videoURL := req.VideoURL
	if videoURL == "" {
		// Use story's merged video if available
		videoURL = story.MergedVideoURL
	}
	if videoURL == "" {
		// Check if any scenes have videos
		for _, scene := range story.Scenes {
			if scene.VideoURL != "" {
				videoURL = scene.VideoURL
				break
			}
		}
	}
	if videoURL == "" {
		return nil, errors.New("no video available for synthesis")
	}

	// Create task
	task := &model.VideoSynthesisTask{
		StoryID: req.StoryID,
		Status:  "pending",
		Progress: 0,
	}
	if err := s.audioRepo.CreateVideoSynthesisTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Spawn goroutine for background processing
	go s.processVideoSynthesis(context.Background(), task, videoURL, req.AddAudio, req.AddSubtitle)

	return task, nil
}

// processVideoSynthesis handles the actual video synthesis in background
func (s *VideoSynthesisService) processVideoSynthesis(
	ctx context.Context,
	task *model.VideoSynthesisTask,
	videoURL string,
	addAudio bool,
	addSubtitle bool,
) {
	// Update status to processing
	task.Status = "processing"
	task.Progress = 5
	s.audioRepo.UpdateVideoSynthesisTask(ctx, task)

	// Track temp files for cleanup
	tempFiles := make([]string, 0)
	defer s.cleanupFiles(tempFiles)

	// Step 1: Download base video (Progress: 5-20)
	fmt.Printf("[VideoSynthesis] Task %s: Downloading base video...\n", task.ID.String())
	localVideoPath := filepath.Join(s.tempDir, fmt.Sprintf("base_video_%s.mp4", task.ID.String()))
	if err := s.downloadFile(videoURL, localVideoPath); err != nil {
		s.failTask(ctx, task, fmt.Sprintf("failed to download video: %v", err))
		return
	}
	tempFiles = append(tempFiles, localVideoPath)

	task.Progress = 20
	s.audioRepo.UpdateVideoSynthesisTask(ctx, task)

	// Step 2: Get audio files and concatenate (Progress: 20-50)
	var audioTrackPath string
	var err error
	if addAudio {
		fmt.Printf("[VideoSynthesis] Task %s: Processing audio files...\n", task.ID.String())
		audioTrackPath, err = s.processAudioTrack(ctx, task.StoryID)
		if err != nil {
			s.failTask(ctx, task, fmt.Sprintf("failed to process audio: %v", err))
			return
		}
		if audioTrackPath != "" {
			tempFiles = append(tempFiles, audioTrackPath)
		}
	}

	task.Progress = 50
	s.audioRepo.UpdateVideoSynthesisTask(ctx, task)

	// Step 3: Generate SRT subtitle file (Progress: 50-60)
	var subtitlePath string
	if addSubtitle {
		fmt.Printf("[VideoSynthesis] Task %s: Generating subtitles...\n", task.ID.String())
		subtitlePath, err = s.subtitleSvc.GenerateSRT(ctx, task.StoryID)
		if err != nil {
			// Subtitle generation failure is not critical, continue without subtitles
			fmt.Printf("[VideoSynthesis] Task %s: Warning - subtitle generation failed: %v\n", task.ID.String(), err)
		}
		if subtitlePath != "" {
			tempFiles = append(tempFiles, subtitlePath)
		}
	}

	task.Progress = 60
	s.audioRepo.UpdateVideoSynthesisTask(ctx, task)

	// Step 4: Combine video + audio + subtitles (Progress: 60-90)
	fmt.Printf("[VideoSynthesis] Task %s: Combining media elements...\n", task.ID.String())
	outputFilename := fmt.Sprintf("%s_synthesized_%s.mp4", task.StoryID.String()[:8], time.Now().Format("20060302_150405"))
	outputPath := filepath.Join(s.outputDir, outputFilename)

	if err := s.combineMedia(localVideoPath, audioTrackPath, subtitlePath, outputPath); err != nil {
		s.failTask(ctx, task, fmt.Sprintf("failed to combine media: %v", err))
		return
	}

	task.Progress = 90
	s.audioRepo.UpdateVideoSynthesisTask(ctx, task)

	// Step 5: Complete task (Progress: 90-100)
	outputURL := fmt.Sprintf("%s/%s", s.videoBaseURL, outputFilename)
	task.OutputURL = outputURL
	task.Status = "completed"
	task.Progress = 100
	now := time.Now()
	task.CompletedAt = &now
	s.audioRepo.UpdateVideoSynthesisTask(ctx, task)

	fmt.Printf("[VideoSynthesis] Task %s: Completed successfully - %s\n", task.ID.String(), outputURL)
}

// processAudioTrack concatenates all audio files into a single track
func (s *VideoSynthesisService) processAudioTrack(ctx context.Context, storyID uuid.UUID) (string, error) {
	// Get all audio files for story
	audios, err := s.audioRepo.GetAudiosByStory(ctx, storyID)
	if err != nil {
		return "", fmt.Errorf("failed to get audio files: %w", err)
	}

	if len(audios) == 0 {
		fmt.Printf("[VideoSynthesis] No audio files found for story %s\n", storyID.String())
		return "", nil
	}

	// Download all audio files to temp directory
	audioFiles := make([]string, 0, len(audios))
	for i, audio := range audios {
		filename := fmt.Sprintf("audio_%d_%s.mp3", i, audio.ID.String()[:8])
		localPath := filepath.Join(s.tempDir, filename)

		fmt.Printf("[VideoSynthesis] Downloading audio %d...\n", i)
		if err := s.downloadFile(audio.AudioURL, localPath); err != nil {
			// Clean up downloaded files
			s.cleanupFiles(audioFiles)
			return "", fmt.Errorf("failed to download audio %d: %w", i, err)
		}
		audioFiles = append(audioFiles, localPath)
	}

	// Create concat list file
	concatFile := filepath.Join(s.tempDir, fmt.Sprintf("audio_list_%s.txt", storyID.String()[:8]))
	concatContent := ""
	for _, f := range audioFiles {
		absPath, _ := filepath.Abs(f)
		// Escape single quotes in filename
		escaped := strings.ReplaceAll(absPath, "'", "'\\''")
		concatContent += fmt.Sprintf("file '%s'\n", escaped)
	}
	if err := os.WriteFile(concatFile, []byte(concatContent), 0644); err != nil {
		s.cleanupFiles(audioFiles)
		return "", fmt.Errorf("failed to create concat file: %w", err)
	}

	// Concatenate audio files using ffmpeg
	outputPath := filepath.Join(s.tempDir, fmt.Sprintf("audio_track_%s.mp3", storyID.String()[:8]))
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-c", "copy",
		"-y",
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.cleanupFiles(audioFiles)
		os.Remove(concatFile)
		return "", fmt.Errorf("ffmpeg concat failed: %w, output: %s", err, string(output))
	}

	// Clean up individual audio files and concat list
	s.cleanupFiles(audioFiles)
	os.Remove(concatFile)

	fmt.Printf("[VideoSynthesis] Audio track created: %s\n", outputPath)
	return outputPath, nil
}

// combineMedia combines video, audio, and subtitles using ffmpeg
func (s *VideoSynthesisService) combineMedia(videoPath, audioPath, subtitlePath, outputPath string) error {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found, please install ffmpeg")
	}

	// Get absolute paths
	absVideoPath, _ := filepath.Abs(videoPath)
	absOutputPath, _ := filepath.Abs(outputPath)

	args := []string{}

	// Input files
	args = append(args, "-i", absVideoPath)

	if audioPath != "" {
		absAudioPath, _ := filepath.Abs(audioPath)
		args = append(args, "-i", absAudioPath)
	}

	// Build video filter for subtitles
	if subtitlePath != "" {
		absSubtitlePath, _ := filepath.Abs(subtitlePath)
		// Escape special characters for ffmpeg filter
		escapedSubtitlePath := s.escapePathForFFmpeg(absSubtitlePath)
		subtitleFilter := fmt.Sprintf("subtitles='%s':force_style='Fontsize=24,PrimaryColour=&HFFFFFF'", escapedSubtitlePath)
		args = append(args, "-vf", subtitleFilter)
	}

	// Encoding options
	if audioPath != "" {
		args = append(args,
			"-c:v", "libx264",
			"-c:a", "aac",
			"-shortest", // Stop when shortest input ends
		)
	} else {
		args = append(args, "-c:v", "libx264")
		if subtitlePath != "" {
			args = append(args, "-c:a", "copy") // Keep original audio
		} else {
			args = append(args, "-c", "copy") // Just copy everything
		}
	}

	// Output options
	args = append(args,
		"-movflags", "+faststart", // Enable streaming
		"-y", // Overwrite output
		absOutputPath,
	)

	fmt.Printf("[VideoSynthesis] FFmpeg command: ffmpeg %s\n", strings.Join(args, " "))

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}

	fmt.Printf("[VideoSynthesis] Output created: %s\n", outputPath)
	return nil
}

// escapePathForFFmpeg escapes path for ffmpeg filter syntax
func (s *VideoSynthesisService) escapePathForFFmpeg(path string) string {
	// Escape backslashes and colons for Windows paths, and special chars
	escaped := strings.ReplaceAll(path, "\\", "/")
	escaped = strings.ReplaceAll(escaped, ":", "\\:")
	return escaped
}

// downloadFile downloads a file from URL to local path
func (s *VideoSynthesisService) downloadFile(url, localPath string) error {
	// Handle relative URLs (local paths)
	if strings.HasPrefix(url, "/") || strings.HasPrefix(url, "./") {
		// It's a local path, copy the file
		srcPath := url
		if strings.HasPrefix(url, "/api/") {
			// Convert API URL to actual file path
			// Assuming files are stored in uploads directory
			srcPath = filepath.Join("./uploads", strings.TrimPrefix(url, "/api/v1/videos/view?file="))
		}
		return s.copyFile(srcPath, localPath)
	}

	// Download from remote URL
	client := &http.Client{
		Timeout: 120 * time.Second, // 2 minutes timeout for downloads
	}

	resp, err := client.Get(url)
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

// copyFile copies a file from source to destination
func (s *VideoSynthesisService) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// failTask marks a task as failed with error message
func (s *VideoSynthesisService) failTask(ctx context.Context, task *model.VideoSynthesisTask, errorMsg string) {
	task.Status = "failed"
	task.ErrorMessage = errorMsg
	s.audioRepo.UpdateVideoSynthesisTask(ctx, task)
	fmt.Printf("[VideoSynthesis] Task %s: Failed - %s\n", task.ID.String(), errorMsg)
}

// cleanupFiles removes temporary files
func (s *VideoSynthesisService) cleanupFiles(files []string) {
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			fmt.Printf("[VideoSynthesis] Warning: failed to remove temp file %s: %v\n", f, err)
		}
	}
}

// GetTaskStatus retrieves the status of a synthesis task
func (s *VideoSynthesisService) GetTaskStatus(ctx context.Context, taskID uuid.UUID) (*model.VideoSynthesisTask, error) {
	return s.audioRepo.GetVideoSynthesisTask(ctx, taskID)
}

// GetTaskByStory retrieves the latest synthesis task for a story
func (s *VideoSynthesisService) GetTaskByStory(ctx context.Context, storyID uuid.UUID) (*model.VideoSynthesisTask, error) {
	return s.audioRepo.GetVideoSynthesisTaskByStory(ctx, storyID)
}

// CancelTask cancels a pending or processing synthesis task
func (s *VideoSynthesisService) CancelTask(ctx context.Context, taskID uuid.UUID) error {
	task, err := s.audioRepo.GetVideoSynthesisTask(ctx, taskID)
	if err != nil {
		return err
	}

	if task.Status == "completed" {
		return errors.New("cannot cancel completed task")
	}

	task.Status = "cancelled"
	task.ErrorMessage = "Task cancelled by user"
	return s.audioRepo.UpdateVideoSynthesisTask(ctx, task)
}

// GetVideoDuration returns the duration of a video file using ffprobe
func (s *VideoSynthesisService) GetVideoDuration(path string) (float64, error) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return 0, fmt.Errorf("ffprobe not found")
	}

	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	var duration float64
	fmt.Sscanf(string(output), "%f", &duration)
	return duration, nil
}