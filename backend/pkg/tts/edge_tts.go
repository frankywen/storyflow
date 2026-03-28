package tts

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// EdgeTTSProvider Edge-TTS实现
type EdgeTTSProvider struct {
	cmdPath      string
	outputDir    string
	audioBaseURL string
	timeout      time.Duration
}

type EdgeTTSConfig struct {
	CmdPath      string
	OutputDir    string
	AudioBaseURL string
	Timeout      time.Duration
}

func NewEdgeTTSProvider(config EdgeTTSConfig) *EdgeTTSProvider {
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.CmdPath == "" {
		config.CmdPath = "edge-tts"
	}
	if config.OutputDir == "" {
		config.OutputDir = "./uploads/audio"
	}
	return &EdgeTTSProvider{
		cmdPath:      config.CmdPath,
		outputDir:    config.OutputDir,
		audioBaseURL: config.AudioBaseURL,
		timeout:      config.Timeout,
	}
}

func (p *EdgeTTSProvider) GetName() string {
	return "edge-tts"
}

func (p *EdgeTTSProvider) GenerateVoice(ctx context.Context, text string, voiceID string, params VoiceParams) (*AudioResult, error) {
	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s.mp3", time.Now().UnixNano(), voiceID)
	outputPath := filepath.Join(p.outputDir, filename)

	// 确保输出目录存在
	if err := os.MkdirAll(p.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// 构建命令
	args := []string{
		"--text", text,
		"--voice", voiceID,
		"--write-media", outputPath,
	}

	// 添加参数
	if params.Speed != 0 {
		args = append(args, "--rate", fmt.Sprintf("%.0f%%", params.Speed*100))
	}
	if params.Volume != 0 {
		args = append(args, "--volume", fmt.Sprintf("%d%%", params.Volume))
	}

	// 执行命令
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.cmdPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("TTS generation timeout")
		}
		return nil, fmt.Errorf("TTS generation failed: %v, output: %s", err, output)
	}

	// 获取音频时长
	duration, err := p.getAudioDuration(outputPath)
	if err != nil {
		duration = 0
	}

	return &AudioResult{
		AudioURL: fmt.Sprintf("%s/%s", p.audioBaseURL, filename),
		Duration: duration,
		VoiceID:  voiceID,
	}, nil
}

func (p *EdgeTTSProvider) GetAvailableVoices(ctx context.Context) ([]Voice, error) {
	// 返回常用的中文音色
	return []Voice{
		{ID: "zh-CN-XiaoxiaoNeural", Name: "晓晓", Gender: "Female", Language: "zh-CN", Description: "女声-成年"},
		{ID: "zh-CN-YunxiNeural", Name: "云希", Gender: "Male", Language: "zh-CN", Description: "男声-成年"},
		{ID: "zh-CN-YunjianNeural", Name: "云健", Gender: "Male", Language: "zh-CN", Description: "男声-成年"},
		{ID: "zh-CN-XiaoyiNeural", Name: "晓伊", Gender: "Female", Language: "zh-CN", Description: "女声-成年"},
		{ID: "zh-CN-YunyangNeural", Name: "云扬", Gender: "Male", Language: "zh-CN", Description: "男声-新闻"},
		{ID: "zh-CN-XiaochenNeural", Name: "晓辰", Gender: "Female", Language: "zh-CN", Description: "女声-成年"},
	}, nil
}

// getAudioDuration 使用ffprobe获取音频时长
func (p *EdgeTTSProvider) getAudioDuration(filePath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-i", filePath,
		"-show_entries", "format=duration",
		"-v", "quiet",
		"-of", "csv=p=0",
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, err
	}

	return duration, nil
}