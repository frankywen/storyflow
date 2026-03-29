export interface Story {
  id: string
  title: string
  content?: string
  summary: string
  genre: string
  status: 'pending' | 'parsing' | 'parsed' | 'generating' | 'completed' | 'error'
  merged_video_url?: string
  created_at: string
  updated_at: string
  characters?: Character[]
  scenes?: Scene[]
  images?: Image[]
}

export interface Character {
  id: string
  story_id: string
  name: string
  description: string
  gender: string
  age: string
  hair_color?: string
  eye_color?: string
  body_type?: string
  clothing?: string
  reference_image_url?: string
  reference_image_id?: string
  seed?: number
  visual_prompt?: string
  prompt_template: string
  created_at: string
}

export interface Scene {
  id: string
  story_id: string
  sequence: number
  title: string
  description: string
  location: string
  time_of_day: string
  mood: string
  character_ids: string[]  // UUID strings
  dialogue: string
  narration: string
  image_prompt: string
  image_url?: string
  video_url?: string
  status: 'pending' | 'generating' | 'completed'
  created_at: string
}

export interface Image {
  id: string
  story_id: string
  scene_id?: string
  prompt: string
  image_url: string
  thumbnail_url?: string
  width: number
  height: number
  seed: number
  model: string
  status: 'pending' | 'completed'
  created_at: string
}

export interface GenerationJob {
  id: string
  story_id: string
  type: 'image' | 'video' | 'batch'
  status: 'pending' | 'running' | 'completed' | 'failed'
  progress: number
  total_items: number
  done_items: number
  error?: string
  result_urls: string[]
  created_at: string
  completed_at?: string
}

export interface ParseResult {
  summary: string
  genre: string
  tone: string
  characters: Character[]
  scenes: Scene[]
  questions?: string[]
}

export interface AudioFile {
  id: string
  scene_id: string
  audio_type: 'dialogue' | 'narration'
  audio_url: string
  duration?: number
  character_id?: string
  created_at: string
}

export interface Subtitle {
  id: string
  scene_id: string
  text: string
  start_time: number
  end_time: number
  character_id?: string
  created_at: string
}