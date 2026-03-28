import axios, { AxiosError } from 'axios'
import { getAccessToken } from '../contexts/AuthContext'
import { LoginInput, RegisterInput, User, UserConfig } from '../types/auth'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 120000, // Increased timeout for exports and video generation
})

// Request interceptor to add auth header
api.interceptors.request.use(
  (config) => {
    const token = getAccessToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor to handle 401 errors
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    if (error.response?.status === 401) {
      // Clear auth data and redirect to login
      localStorage.removeItem('storyflow_access_token')
      localStorage.removeItem('storyflow_refresh_token')
      localStorage.removeItem('storyflow_user')

      // Only redirect if not already on login/register page
      if (!window.location.pathname.startsWith('/login') && !window.location.pathname.startsWith('/register')) {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

// Auth APIs
export const authApi = {
  login: (data: LoginInput) =>
    api.post<{ access_token: string; refresh_token: string; expires_in: number }>('/auth/login', data),

  register: (data: RegisterInput) =>
    api.post<{ access_token: string; refresh_token: string; expires_in: number }>('/auth/register', data),

  refresh: (refreshToken: string) =>
    api.post<{ access_token: string; refresh_token: string; expires_in: number }>('/auth/refresh', { refresh_token: refreshToken }),

  logout: () =>
    api.post('/auth/logout'),

  getMe: () =>
    api.get<User>('/auth/me'),

  forgotPassword: (email: string) =>
    api.post<{ message: string; reset_token?: string; reset_url?: string }>('/auth/forgot-password', { email }),

  resetPassword: (token: string, password: string) =>
    api.post<{ message: string }>('/auth/reset-password', { token, password }),
}

// User Config APIs
export const configApi = {
  get: () =>
    api.get<{ config: UserConfig }>('/user/config'),

  update: (data: Partial<UserConfig>) =>
    api.put<UserConfig>('/user/config', data),

  updateLLM: (data: {
    provider?: string
    api_key?: string
    model?: string
    base_url?: string
  }) =>
    api.put('/user/config/llm', data),

  updateImage: (data: {
    provider?: string
    api_key?: string
    base_url?: string
    model?: string
  }) =>
    api.put('/user/config/image', data),

  updateVideo: (data: {
    provider?: string
    api_key?: string
    base_url?: string
    model?: string
  }) =>
    api.put('/user/config/video', data),

  validateAPIKey: (data: {
    type: 'llm' | 'image' | 'video'
    provider: string
    api_key?: string
    base_url?: string
  }) =>
    api.post<{ valid: boolean; message: string }>('/user/config/validate', data),
}

// Story APIs
export const storyApi = {
  create: (data: { title: string; content: string }) =>
    api.post('/stories', data),

  get: (id: string) =>
    api.get(`/stories/${id}`),

  list: (page = 1, pageSize = 20) =>
    api.get('/stories', { params: { page, page_size: pageSize } }),

  update: (id: string, data: { title?: string; content?: string }) =>
    api.put(`/stories/${id}`, data),

  parse: (id: string, style = 'manga') =>
    api.post(`/stories/${id}/parse`, { style }),

  delete: (id: string) =>
    api.delete(`/stories/${id}`),

  getCharacters: (id: string) =>
    api.get(`/stories/${id}/characters`),

  getScenes: (id: string) =>
    api.get(`/stories/${id}/scenes`),
}

// Image APIs
export const imageApi = {
  generate: (data: {
    prompt: string
    negative_prompt?: string
    width?: number
    height?: number
    style?: string
  }) => api.post('/images/generate', data),

  batchGenerate: (storyId: string, style = 'manga', useConsistency = true) =>
    api.post('/images/batch', { story_id: storyId, style, use_consistency: useConsistency }),

  getJobStatus: (jobId: string) =>
    api.get(`/images/jobs/${jobId}`),
}

// Export APIs
export const exportApi = {
  exportStory: async (storyId: string, format: 'png' | 'pdf' = 'png') => {
    const response = await api.post('/export', {
      story_id: storyId,
      format,
      include_title: true,
    }, {
      responseType: 'blob',
    })
    return response
  },

  downloadExport: (storyId: string, format: 'png' | 'pdf' = 'png') => {
    // Create a download link
    api.post('/export', {
      story_id: storyId,
      format,
      include_title: true,
    }, {
      responseType: 'blob',
    }).then((response) => {
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `storyflow_export.${format}`)
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
    })
  },
}

// Video APIs
export const videoApi = {
  generate: (storyId: string, sceneId?: string, options?: {
    duration?: number
    prompt?: string
    motion_level?: string
  }) =>
    api.post('/videos/generate', {
      story_id: storyId,
      scene_id: sceneId,
      ...options,
    }),

  batchGenerate: (storyId: string, options?: {
    duration?: number
    prompt?: string
    motion_level?: string
  }) =>
    api.post('/videos/batch', {
      story_id: storyId,
      ...options,
    }),

  getStatus: (taskId: string) =>
    api.get(`/videos/status/${taskId}`),

  merge: (storyId: string, options?: {
    transition?: string
    transition_duration?: number
    add_audio?: boolean
  }) =>
    api.post('/videos/merge', {
      story_id: storyId,
      ...options,
    }),

  getViewUrl: (filename: string) =>
    `/api/v1/videos/view?file=${filename}`,
}

// Admin APIs
export const adminApi = {
  getStats: () =>
    api.get<{
      total_users: number
      active_users: number
      suspended_users: number
      admin_count: number
      total_stories: number
    }>('/admin/stats'),

  listUsers: (page = 1, pageSize = 20, status?: string, role?: string) =>
    api.get<{
      data: AdminUser[]
      total: number
      page: number
      page_size: number
    }>('/admin/users', { params: { page, page_size: pageSize, status, role } }),

  getUser: (id: string) =>
    api.get<{
      user: AdminUser
      config: UserConfig
      story_count: number
    }>(`/admin/users/${id}`),

  updateUser: (id: string, data: {
    name?: string
    role?: 'admin' | 'user'
    status?: 'active' | 'suspended' | 'deleted'
  }) =>
    api.put<{ message: string; user: AdminUser }>(`/admin/users/${id}`, data),

  suspendUser: (id: string) =>
    api.post<{ message: string }>(`/admin/users/${id}/suspend`),

  activateUser: (id: string) =>
    api.post<{ message: string }>(`/admin/users/${id}/activate`),

  deleteUser: (id: string) =>
    api.delete<{ message: string }>(`/admin/users/${id}`),
}

export interface AdminUser {
  id: string
  email: string
  name: string
  avatar_url?: string
  role: 'admin' | 'user'
  status: 'active' | 'suspended' | 'deleted'
  created_at: string
  updated_at: string
  last_login_at?: string
}

// Character APIs
export const characterApi = {
  uploadReference: (characterId: string, file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post(`/characters/${characterId}/reference`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },

  deleteReference: (characterId: string) =>
    api.delete(`/characters/${characterId}/reference`),

  regenerateReference: (characterId: string, style = 'manga') =>
    api.post(`/characters/${characterId}/regenerate`, { style }),

  generateAllReferences: (storyId: string, style = 'manga') =>
    api.post(`/stories/${storyId}/generate-references`, { style }),
}

// Audio APIs
export const audioApi = {
  generate: (storyId: string) =>
    api.post<{ success: boolean; task_id: string; message: string }>('/audio/generate', { story_id: storyId }),

  getStatus: (taskId: string) =>
    api.get<{
      task_id: string
      status: string
      progress: number
      total_scenes: number
      completed_scenes: number
      failed_scenes: Record<string, string>
    }>(`/audio/status/${taskId}`),

  getAudios: (storyId: string) =>
    api.get<{ audios: AudioFile[] }>(`/audio/story/${storyId}`),

  generateSubtitles: (storyId: string) =>
    api.post<{ success: boolean; message: string }>(`/audio/subtitles/${storyId}`),

  getSubtitles: (storyId: string) =>
    api.get<{ subtitles: Subtitle[] }>(`/audio/subtitles/${storyId}`),

  synthesizeVideo: (storyId: string, options?: {
    video_url?: string
    add_audio?: boolean
    add_subtitle?: boolean
  }) =>
    api.post<{ success: boolean; task_id: string; message: string }>('/audio/synthesis', {
      story_id: storyId,
      ...options,
    }),

  getSynthesisStatus: (taskId: string) =>
    api.get<{
      task_id: string
      status: string
      progress: number
      output_url: string
      error_message: string
    }>(`/videos/synthesis/${taskId}`),
}

export interface AudioFile {
  id: string
  story_id: string
  scene_id: string
  character_id?: string
  audio_type: 'dialogue' | 'narration'
  text_content: string
  audio_url: string
  duration: number
  voice_id: string
  status: string
  created_at: string
}

export interface Subtitle {
  id: string
  story_id: string
  scene_id: string
  subtitle_type: 'dialogue' | 'narration'
  character_id?: string
  text: string
  start_time: number
  end_time: number
  style_config?: Record<string, any>
  created_at: string
}

export default api