-- Character voices table
CREATE TABLE IF NOT EXISTS character_voices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID UNIQUE NOT NULL,
    voice_id VARCHAR(100) NOT NULL,
    voice_name VARCHAR(100),
    voice_provider VARCHAR(50),
    voice_params JSONB,
    is_custom BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Audio files table
CREATE TABLE IF NOT EXISTS audio_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL,
    scene_id UUID,
    character_id UUID,
    audio_type VARCHAR(20) NOT NULL,
    text_content TEXT NOT NULL,
    audio_url VARCHAR(500) NOT NULL,
    duration FLOAT NOT NULL,
    voice_id VARCHAR(100),
    status VARCHAR(20) DEFAULT 'completed',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audio_story ON audio_files(story_id);
CREATE INDEX IF NOT EXISTS idx_audio_scene ON audio_files(scene_id);

-- Subtitles table
CREATE TABLE IF NOT EXISTS subtitles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL,
    scene_id UUID,
    subtitle_type VARCHAR(20) NOT NULL,
    character_id UUID,
    text TEXT NOT NULL,
    start_time FLOAT NOT NULL,
    end_time FLOAT NOT NULL,
    style_config JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subtitles_story ON subtitles(story_id);
CREATE INDEX IF NOT EXISTS idx_subtitles_scene ON subtitles(scene_id);

-- Audio generation tasks table
CREATE TABLE IF NOT EXISTS audio_generation_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    total_scenes INT DEFAULT 0,
    completed_scenes INT DEFAULT 0,
    progress INT DEFAULT 0,
    failed_scenes JSONB,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audio_gen_story ON audio_generation_tasks(story_id);

-- Video synthesis tasks table
CREATE TABLE IF NOT EXISTS video_synthesis_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    progress INT DEFAULT 0,
    output_url VARCHAR(500),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_video_synthesis_story ON video_synthesis_tasks(story_id);

-- Voice mappings table
CREATE TABLE IF NOT EXISTS voice_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    standard_voice_id VARCHAR(50) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    provider_voice_id VARCHAR(100) NOT NULL,
    voice_name VARCHAR(100),
    UNIQUE(standard_voice_id, provider)
);

-- Initial voice mappings for Edge-TTS
INSERT INTO voice_mappings (standard_voice_id, provider, provider_voice_id, voice_name) VALUES
('male_adult', 'edge-tts', 'zh-CN-YunxiNeural', '男声-成年'),
('female_adult', 'edge-tts', 'zh-CN-XiaoxiaoNeural', '女声-成年'),
('male_child', 'edge-tts', 'zh-CN-YunxiNeural', '男声-儿童'),
('female_child', 'edge-tts', 'zh-CN-XiaoxiaoNeural', '女声-儿童'),
('narrator', 'edge-tts', 'zh-CN-XiaoxiaoNeural', '旁白')
ON CONFLICT DO NOTHING;