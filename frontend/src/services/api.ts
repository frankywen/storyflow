import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 120000, // Increased timeout for exports and video generation
})

// Story APIs
export const storyApi = {
  create: (data: { title: string; content: string }) =>
    api.post('/stories', data),

  get: (id: string) =>
    api.get(`/stories/${id}`),

  list: (page = 1, pageSize = 20) =>
    api.get('/stories', { params: { page, page_size: pageSize } }),

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

export default api