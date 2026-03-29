export interface User {
  id: string
  email: string
  name: string
  avatar_url?: string
  role: 'admin' | 'user'
  status: string
  created_at: string
  last_login_at?: string
}

export interface UserConfig {
  llm_provider?: string
  llm_api_key_masked?: string
  llm_model?: string
  llm_base_url?: string
  image_provider?: string
  image_api_key_masked?: string
  image_base_url?: string
  image_model?: string
  video_provider?: string
  video_api_key_masked?: string
  video_base_url?: string
  video_model?: string
  default_style?: string
}

export interface AuthState {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
}

export interface LoginInput {
  email: string
  password: string
}

export interface RegisterInput {
  email: string
  password: string
  name?: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
}