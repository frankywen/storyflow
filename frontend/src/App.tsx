import React from 'react'
import { BrowserRouter, Routes, Route, Link, useNavigate } from 'react-router-dom'
import { FileText, Image, Settings, Trash2, RefreshCw, Download, Video, Upload, User, Film, LogOut, Shield, Edit3, Save, X, Volume2, Subtitles } from 'lucide-react'

// Auth
import { AuthProvider, useAuth } from './contexts/AuthContext'
import ProtectedRoute from './components/auth/ProtectedRoute'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import ForgotPasswordPage from './pages/ForgotPasswordPage'
import ResetPasswordPage from './pages/ResetPasswordPage'
import ConfigPage from './pages/ConfigPage'
import AdminPage from './pages/AdminPage'
import AudioConfigPage from './pages/AudioConfigPage'

// API
import { storyApi, imageApi, exportApi, videoApi, characterApi, audioApi } from './services/api'

// Types
import { Story, Character, AudioFile, Subtitle } from './types'

// Pages
const HomePage = () => (
  <div className="p-8">
    <h1 className="text-3xl font-bold mb-2">StoryFlow</h1>
    <p className="text-gray-600 mb-8">AI驱动的小说→漫画/视频自动化生成工具</p>

    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
      <Link to="/stories/new" className="p-6 bg-white rounded-lg shadow hover:shadow-md transition-shadow border">
        <FileText className="w-8 h-8 text-blue-500 mb-3" />
        <h2 className="text-xl font-semibold mb-2">创建故事</h2>
        <p className="text-gray-500 text-sm">输入小说文本，AI自动解析角色和场景</p>
      </Link>

      <Link to="/stories" className="p-6 bg-white rounded-lg shadow hover:shadow-md transition-shadow border">
        <Image className="w-8 h-8 text-green-500 mb-3" />
        <h2 className="text-xl font-semibold mb-2">我的故事</h2>
        <p className="text-gray-500 text-sm">查看已创建的故事和生成进度</p>
      </Link>

      <Link to="/config" className="p-6 bg-white rounded-lg shadow hover:shadow-md transition-shadow border">
        <Settings className="w-8 h-8 text-purple-500 mb-3" />
        <h2 className="text-xl font-semibold mb-2">配置</h2>
        <p className="text-gray-500 text-sm">配置AI模型和生成参数</p>
      </Link>
    </div>

    <div className="mt-12">
      <h2 className="text-xl font-semibold mb-4">使用流程</h2>
      <div className="flex items-center gap-4 text-sm flex-wrap">
        <div className="flex items-center gap-2 px-4 py-2 bg-blue-50 rounded-lg">
          <span className="w-6 h-6 bg-blue-500 text-white rounded-full flex items-center justify-center text-xs">1</span>
          输入故事
        </div>
        <span className="text-gray-400">→</span>
        <div className="flex items-center gap-2 px-4 py-2 bg-blue-50 rounded-lg">
          <span className="w-6 h-6 bg-blue-500 text-white rounded-full flex items-center justify-center text-xs">2</span>
          AI解析
        </div>
        <span className="text-gray-400">→</span>
        <div className="flex items-center gap-2 px-4 py-2 bg-blue-50 rounded-lg">
          <span className="w-6 h-6 bg-blue-500 text-white rounded-full flex items-center justify-center text-xs">3</span>
          生成图片
        </div>
        <span className="text-gray-400">→</span>
        <div className="flex items-center gap-2 px-4 py-2 bg-teal-50 rounded-lg">
          <span className="w-6 h-6 bg-teal-500 text-white rounded-full flex items-center justify-center text-xs">4</span>
          配音字幕
        </div>
        <span className="text-gray-400">→</span>
        <div className="flex items-center gap-2 px-4 py-2 bg-blue-50 rounded-lg">
          <span className="w-6 h-6 bg-blue-500 text-white rounded-full flex items-center justify-center text-xs">5</span>
          视频生成
        </div>
        <span className="text-gray-400">→</span>
        <div className="flex items-center gap-2 px-4 py-2 bg-green-50 rounded-lg">
          <span className="w-6 h-6 bg-green-500 text-white rounded-full flex items-center justify-center text-xs">6</span>
          导出
        </div>
      </div>
    </div>
  </div>
)

const NewStoryPage = () => {
  const navigate = useNavigate()
  const [title, setTitle] = React.useState('')
  const [content, setContent] = React.useState('')
  const [loading, setLoading] = React.useState(false)

  const handleSubmit = async () => {
    if (!title || !content) return

    setLoading(true)
    try {
      const res = await storyApi.create({ title, content })
      navigate(`/stories/${res.data.id}`)
    } catch (err) {
      alert('创建失败，请检查后端服务是否启动')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-6">新建故事</h1>
      <div className="max-w-4xl">
        <div className="mb-4">
          <label className="block text-sm font-medium mb-2">故事标题</label>
          <input
            type="text"
            className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            placeholder="例如：穿越到异世界的我成了勇者"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
        </div>
        <div className="mb-4">
          <label className="block text-sm font-medium mb-2">故事内容</label>
          <textarea
            className="w-full px-4 py-2 border rounded-lg h-96 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            placeholder="粘贴小说文本内容...&#10;&#10;AI会自动分析故事中的角色、场景和情节，生成结构化的分镜脚本。"
            value={content}
            onChange={(e) => setContent(e.target.value)}
          />
        </div>
        <div className="flex gap-4">
          <button
            className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400"
            onClick={handleSubmit}
            disabled={loading || !title || !content}
          >
            {loading ? '创建中...' : '创建故事'}
          </button>
          <button
            className="px-6 py-2 border rounded-lg hover:bg-gray-50"
            onClick={() => navigate('/')}
          >
            取消
          </button>
        </div>
      </div>
    </div>
  )
}

const StoryListPage = () => {
  const [stories, setStories] = React.useState<Story[]>([])
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    loadStories()
  }, [])

  const loadStories = async () => {
    try {
      const res = await storyApi.list()
      setStories(res.data.data || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">我的故事</h1>
        <Link to="/stories/new" className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600">
          新建故事
        </Link>
      </div>

      {loading ? (
        <p className="text-gray-500">加载中...</p>
      ) : stories.length === 0 ? (
        <p className="text-gray-500">还没有故事，点击上方按钮创建一个吧</p>
      ) : (
        <div className="space-y-4">
          {stories.map((story) => (
            <Link
              key={story.id}
              to={`/stories/${story.id}`}
              className="block p-4 bg-white rounded-lg shadow hover:shadow-md transition-shadow border"
            >
              <h3 className="font-semibold">{story.title}</h3>
              <p className="text-sm text-gray-500 mt-1">
                {story.summary || '尚未解析'}
              </p>
              <div className="flex items-center gap-4 mt-2 text-xs text-gray-400">
                <span>{story.genre}</span>
                <span>•</span>
                <span>{story.status}</span>
                <span>•</span>
                <span>{new Date(story.created_at).toLocaleDateString()}</span>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

// Character Card Component
const CharacterCard = ({
  character,
  onRegenerate,
  onUpload,
}: {
  character: Character
  onRegenerate: (id: string) => void
  onUpload: (id: string, file: File) => void
}) => {
  const [uploading, setUploading] = React.useState(false)

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      setUploading(true)
      await onUpload(character.id, file)
      setUploading(false)
    }
  }

  return (
    <div className="bg-white rounded-lg shadow p-4 border">
      {/* Reference Image */}
      {character.reference_image_url ? (
        <div className="mb-3">
          <img
            src={character.reference_image_url}
            alt={character.name}
            className="w-full h-32 object-cover rounded-lg"
          />
        </div>
      ) : (
        <div className="mb-3 h-32 bg-gray-100 rounded-lg flex items-center justify-center">
          <User className="w-12 h-12 text-gray-300" />
        </div>
      )}

      {/* Character Info */}
      <h3 className="font-semibold">{character.name}</h3>
      <p className="text-sm text-gray-500 mt-1 line-clamp-2">{character.description}</p>

      {/* Visual Tags */}
      <div className="flex gap-2 mt-2 text-xs flex-wrap">
        <span className="px-2 py-1 bg-gray-100 rounded">{character.gender}</span>
        <span className="px-2 py-1 bg-gray-100 rounded">{character.age}</span>
        {character.hair_color && (
          <span className="px-2 py-1 bg-gray-100 rounded">{character.hair_color}</span>
        )}
      </div>

      {/* Action Buttons */}
      <div className="flex gap-2 mt-3">
        <button
          onClick={() => onRegenerate(character.id)}
          className="flex-1 px-3 py-1.5 text-sm bg-blue-500 text-white rounded hover:bg-blue-600 flex items-center justify-center gap-1"
        >
          <RefreshCw className="w-3 h-3" />
          {character.reference_image_url ? '重新生成' : '生成参考图'}
        </button>
        <label className="flex-1 px-3 py-1.5 text-sm bg-gray-200 rounded cursor-pointer hover:bg-gray-300 flex items-center justify-center gap-1">
          <Upload className="w-3 h-3" />
          {uploading ? '上传中...' : '上传'}
          <input
            type="file"
            accept="image/*"
            className="hidden"
            onChange={handleFileChange}
            disabled={uploading}
          />
        </label>
      </div>
    </div>
  )
}

const StoryDetailPage = () => {
  const id = window.location.pathname.split('/').pop() || ''
  const navigate = useNavigate()
  const [story, setStory] = React.useState<Story | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [parsing, setParsing] = React.useState(false)
  const [generating, setGenerating] = React.useState(false)
  const [generatingVideo, setGeneratingVideo] = React.useState(false)
  const [videoProgress, setVideoProgress] = React.useState(0)
  const [mergingVideo, setMergingVideo] = React.useState(false)
  const [mergedVideo, setMergedVideo] = React.useState<{ url: string; duration: number } | null>(null)
  const [exporting, setExporting] = React.useState(false)
  const [selectedStyle, setSelectedStyle] = React.useState('manga')
  const [isEditing, setIsEditing] = React.useState(false)
  const [editTitle, setEditTitle] = React.useState('')
  const [editContent, setEditContent] = React.useState('')
  const [saving, setSaving] = React.useState(false)

  // 场景测试状态
  const [sceneAudios, setSceneAudios] = React.useState<Record<string, AudioFile[]>>({})
  const [sceneSubtitles, setSceneSubtitles] = React.useState<Record<string, Subtitle[]>>({})
  const [generatingScene, setGeneratingScene] = React.useState<string | null>(null)
  const [generatingType, setGeneratingType] = React.useState<'audio' | 'subtitle' | 'video' | null>(null)

  React.useEffect(() => {
    loadStory()
  }, [id])

  const loadStory = async () => {
    try {
      const res = await storyApi.get(id)
      setStory(res.data)
      // Check if there's a merged video
      if (res.data.merged_video_url) {
        setMergedVideo({
          url: res.data.merged_video_url,
          duration: 0, // Duration not stored in DB
        })
      }
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleParse = async () => {
    setParsing(true)
    try {
      const res = await storyApi.parse(id, selectedStyle)
      setStory(res.data)
    } catch (err) {
      alert('解析失败，请检查API配置')
    } finally {
      setParsing(false)
    }
  }

  const handleGenerateImages = async () => {
    setGenerating(true)
    try {
      await imageApi.batchGenerate(id, selectedStyle)
      alert('图片生成任务已启动，请稍后刷新页面查看结果')
    } catch (err) {
      alert('生成失败，请检查图片服务配置')
    } finally {
      setGenerating(false)
    }
  }

  const handleDelete = async () => {
    if (!confirm('确定要删除这个故事吗？')) return
    try {
      await storyApi.delete(id)
      navigate('/stories')
    } catch (err) {
      alert('删除失败')
    }
  }

  const handleStartEdit = () => {
    if (story) {
      setEditTitle(story.title)
      setEditContent(story.content || '')
      setIsEditing(true)
    }
  }

  const handleCancelEdit = () => {
    setIsEditing(false)
    setEditTitle('')
    setEditContent('')
  }

  const handleSaveEdit = async () => {
    if (!editTitle.trim() || !editContent.trim()) {
      alert('标题和内容不能为空')
      return
    }
    setSaving(true)
    try {
      const res = await storyApi.update(id, {
        title: editTitle,
        content: editContent,
      })
      setStory(res.data)
      setIsEditing(false)
    } catch (err) {
      alert('保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleExport = async (format: 'png' | 'pdf') => {
    setExporting(true)
    try {
      await exportApi.downloadExport(id, format)
    } catch (err) {
      alert('导出失败，请检查故事是否已生成图片')
    } finally {
      setExporting(false)
    }
  }

  const handleGenerateVideo = async () => {
    setGeneratingVideo(true)
    setVideoProgress(0)
    try {
      const res = await videoApi.batchGenerate(id, {
        duration: 5,
        motion_level: 'medium',
      })

      // Check if all videos already exist
      if (res.data.message === '所有场景已有视频' || res.data.total_to_generate === 0) {
        alert('所有场景已生成视频')
        setGeneratingVideo(false)
        return
      }

      const skipped = res.data.already_has_video || 0
      const total = res.data.total_to_generate || 0
      alert(`视频生成任务已启动\n跳过 ${skipped} 个已完成场景\n待生成 ${total} 个场景`)

      // Batch generation processes multiple scenes, poll story to check for videos
      pollForVideos()
    } catch (err) {
      alert('视频生成失败，请检查视频服务配置或余额是否充足')
      setGeneratingVideo(false)
    }
  }

  const pollForVideos = () => {
    let attempts = 0
    const maxAttempts = 60 // 60 * 5s = 5 minutes max

    const poll = async () => {
      if (attempts >= maxAttempts) {
        setGeneratingVideo(false)
        alert('视频生成超时，请稍后刷新页面查看')
        return
      }

      attempts++
      setVideoProgress(Math.min(100, attempts * 2)) // Rough progress indicator

      try {
        const res = await storyApi.get(id)
        const story = res.data

        // Check if any scenes have videos now
        const scenesWithVideos = story.scenes?.filter((s: any) => s.video_url).length || 0
        const totalScenes = story.scenes?.filter((s: any) => s.image_url).length || 0

        if (scenesWithVideos > 0) {
          setStory(story)
          setVideoProgress(Math.round((scenesWithVideos / totalScenes) * 100))
        }

        // If all scenes have videos, stop polling
        if (scenesWithVideos === totalScenes && totalScenes > 0) {
          setGeneratingVideo(false)
          return
        }

        // Continue polling
        setTimeout(poll, 5000)
      } catch (err) {
        console.error('Failed to poll for videos', err)
        setTimeout(poll, 5000)
      }
    }
    poll()
  }

  const handleMergeVideos = async () => {
    setMergingVideo(true)
    try {
      const res = await videoApi.merge(id)
      setMergedVideo({
        url: res.data.output_url,
        duration: res.data.total_duration,
      })
      alert(res.data.message)
    } catch (err: any) {
      const errorMsg = err.response?.data?.error || '视频合成失败'
      alert(errorMsg)
    } finally {
      setMergingVideo(false)
    }
  }

  // Character Reference handlers
  const handleGenerateAllReferences = async () => {
    try {
      const res = await characterApi.generateAllReferences(id, selectedStyle)
      alert(`已生成 ${res.data.total} 个角色参考图`)
      loadStory() // Refresh
    } catch (err) {
      alert('批量生成参考图失败')
    }
  }

  const handleRegenerateReference = async (characterId: string) => {
    try {
      await characterApi.regenerateReference(characterId, selectedStyle)
      loadStory() // Refresh
    } catch (err) {
      alert('生成参考图失败')
    }
  }

  const handleUploadReference = async (characterId: string, file: File) => {
    try {
      await characterApi.uploadReference(characterId, file)
      loadStory() // Refresh
    } catch (err) {
      alert('上传参考图失败')
    }
  }

  // 场景测试处理函数
  const handleGenerateSceneAudio = async (sceneId: string) => {
    setGeneratingScene(sceneId)
    setGeneratingType('audio')
    try {
      const res = await audioApi.generateSceneAudio(sceneId)
      setSceneAudios(prev => ({ ...prev, [sceneId]: res.data.audios }))
      alert(`成功生成 ${res.data.audios.length} 个音频文件`)
    } catch (err: any) {
      alert(err.response?.data?.error || '生成配音失败')
    } finally {
      setGeneratingScene(null)
      setGeneratingType(null)
    }
  }

  const handleGenerateSceneSubtitle = async (sceneId: string) => {
    setGeneratingScene(sceneId)
    setGeneratingType('subtitle')
    try {
      const res = await audioApi.generateSceneSubtitles(sceneId)
      setSceneSubtitles(prev => ({ ...prev, [sceneId]: res.data.subtitles }))
      alert(`成功生成 ${res.data.subtitles.length} 个字幕`)
    } catch (err: any) {
      alert(err.response?.data?.error || '生成字幕失败')
    } finally {
      setGeneratingScene(null)
      setGeneratingType(null)
    }
  }

  const handleGenerateSceneVideo = async (sceneId: string) => {
    setGeneratingScene(sceneId)
    setGeneratingType('video')
    try {
      const res = await videoApi.generate(id, sceneId, { duration: 5, motion_level: 'medium' })
      alert(`视频生成任务已启动，任务ID: ${res.data.task_id}`)
      pollForVideos()
    } catch (err: any) {
      alert(err.response?.data?.error || '生成视频失败')
    } finally {
      setGeneratingScene(null)
      setGeneratingType(null)
    }
  }

  // Check if any scenes have images
  const hasImages = story?.scenes?.some(s => s.image_url)

  // Video generation progress stats
  const scenesWithImages = story?.scenes?.filter(s => s.image_url).length || 0
  const scenesWithVideos = story?.scenes?.filter(s => s.video_url).length || 0
  const videosProgress = scenesWithImages > 0 ? Math.round((scenesWithVideos / scenesWithImages) * 100) : 0

  if (loading) return <div className="p-8">加载中...</div>
  if (!story) return <div className="p-8">故事不存在</div>

  return (
    <div className="p-8">
      <div className="flex justify-between items-start mb-6">
        <div className="flex-1">
          {isEditing ? (
            <input
              type="text"
              className="text-2xl font-bold w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
              value={editTitle}
              onChange={(e) => setEditTitle(e.target.value)}
              placeholder="故事标题"
            />
          ) : (
            <h1 className="text-2xl font-bold">{story.title}</h1>
          )}
          {!isEditing && <p className="text-gray-500 mt-1">{story.summary}</p>}
        </div>
        <div className="flex gap-2 ml-4">
          {isEditing ? (
            <>
              <button
                onClick={handleSaveEdit}
                disabled={saving}
                className="flex items-center gap-1 px-3 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400"
              >
                <Save className="w-4 h-4" />
                {saving ? '保存中...' : '保存'}
              </button>
              <button
                onClick={handleCancelEdit}
                disabled={saving}
                className="flex items-center gap-1 px-3 py-2 border rounded-lg hover:bg-gray-50 disabled:opacity-50"
              >
                <X className="w-4 h-4" />
                取消
              </button>
            </>
          ) : (
            <>
              <button
                onClick={handleStartEdit}
                className="p-2 text-blue-500 hover:bg-blue-50 rounded-lg"
                title="编辑故事"
              >
                <Edit3 className="w-5 h-5" />
              </button>
              <button
                onClick={handleDelete}
                className="p-2 text-red-500 hover:bg-red-50 rounded-lg"
                title="删除故事"
              >
                <Trash2 className="w-5 h-5" />
              </button>
            </>
          )}
        </div>
      </div>

      {/* Status and Actions */}
      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="flex items-center justify-between flex-wrap gap-4">
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-500">状态:</span>
            <span className={`px-3 py-1 rounded-full text-sm ${
              story.status === 'completed' ? 'bg-green-100 text-green-700' :
              story.status === 'parsing' ? 'bg-yellow-100 text-yellow-700' :
              story.status === 'parsed' ? 'bg-blue-100 text-blue-700' :
              'bg-gray-100 text-gray-700'
            }`}>
              {story.status}
            </span>
          </div>

          <div className="flex items-center gap-3 flex-wrap">
            <select
              className="px-3 py-2 border rounded-lg"
              value={selectedStyle}
              onChange={(e) => setSelectedStyle(e.target.value)}
            >
              <option value="manga">日式漫画</option>
              <option value="manhwa">韩式漫画</option>
              <option value="western_comic">美式漫画</option>
              <option value="anime">动漫风格</option>
              <option value="realistic">写实风格</option>
            </select>

            <button
              className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 flex items-center gap-2"
              onClick={handleParse}
              disabled={parsing}
            >
              <RefreshCw className={`w-4 h-4 ${parsing ? 'animate-spin' : ''}`} />
              {parsing ? '解析中...' : 'AI解析故事'}
            </button>

            {story.status === 'parsed' && (
              <button
                className="px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 disabled:bg-gray-400 flex items-center gap-2"
                onClick={handleGenerateImages}
                disabled={generating}
              >
                <Image className="w-4 h-4" />
                {generating ? '生成中...' : '生成图片'}
              </button>
            )}

            {/* Video Generation Button */}
            {hasImages && (
              <div className="flex items-center gap-3">
                <button
                  className="px-4 py-2 bg-purple-500 text-white rounded-lg hover:bg-purple-600 disabled:bg-gray-400 flex items-center gap-2"
                  onClick={handleGenerateVideo}
                  disabled={generatingVideo}
                >
                  <Video className="w-4 h-4" />
                  {generatingVideo ? `生成视频中 (${videoProgress}%)` : '生成视频'}
                </button>
                {scenesWithImages > 0 && (
                  <span className="text-sm text-gray-600">
                    {scenesWithVideos}/{scenesWithImages} 已完成 ({videosProgress}%)
                  </span>
                )}
              </div>
            )}

            {/* Audio Config Button */}
            {story.status === 'parsed' && (
              <Link
                to={`/stories/${id}/audio`}
                className="px-4 py-2 bg-teal-500 text-white rounded-lg hover:bg-teal-600 flex items-center gap-2"
              >
                <Volume2 className="w-4 h-4" />
                配音与字幕
              </Link>
            )}

            {/* Merge Videos Button */}
            {scenesWithVideos > 1 && (
              <button
                className="px-4 py-2 bg-indigo-500 text-white rounded-lg hover:bg-indigo-600 disabled:bg-gray-400 flex items-center gap-2"
                onClick={handleMergeVideos}
                disabled={mergingVideo}
              >
                <Film className="w-4 h-4" />
                {mergingVideo ? '合成中...' : '合成完整视频'}
              </button>
            )}

            {/* Export Buttons */}
            {hasImages && (
              <div className="flex gap-2">
                <button
                  className="px-4 py-2 bg-orange-500 text-white rounded-lg hover:bg-orange-600 disabled:bg-gray-400 flex items-center gap-2"
                  onClick={() => handleExport('png')}
                  disabled={exporting}
                >
                  <Download className="w-4 h-4" />
                  PNG
                </button>
                <button
                  className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 disabled:bg-gray-400 flex items-center gap-2"
                  onClick={() => handleExport('pdf')}
                  disabled={exporting}
                >
                  <Download className="w-4 h-4" />
                  PDF
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Characters */}
      {story.characters && story.characters.length > 0 && (
        <div className="mb-6">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-lg font-semibold">角色 ({story.characters.length})</h2>
            {story.status === 'parsed' && (
              <button
                className="px-4 py-2 bg-purple-500 text-white rounded-lg hover:bg-purple-600 flex items-center gap-2 text-sm"
                onClick={handleGenerateAllReferences}
              >
                <User className="w-4 h-4" />
                批量生成参考图
              </button>
            )}
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {story.characters.map((char) => (
              <CharacterCard
                key={char.id}
                character={char}
                onRegenerate={handleRegenerateReference}
                onUpload={handleUploadReference}
              />
            ))}
          </div>
        </div>
      )}

      {/* Scenes */}
      {story.scenes && story.scenes.length > 0 && (
        <div className="mb-6">
          <h2 className="text-lg font-semibold mb-3">场景 ({story.scenes.length})</h2>
          <div className="space-y-4">
            {story.scenes.map((scene) => (
              <div key={scene.id} className="bg-white rounded-lg shadow p-4 border">
                <div className="flex gap-4">
                  <div className="flex-shrink-0 w-16 h-16 bg-gray-100 rounded-lg flex items-center justify-center text-xl font-bold text-gray-400">
                    {scene.sequence}
                  </div>
                  <div className="flex-1">
                    <h3 className="font-semibold">{scene.title || `场景 ${scene.sequence}`}</h3>
                    <p className="text-sm text-gray-600 mt-1">{scene.description}</p>
                    <div className="flex flex-wrap gap-2 mt-2 text-xs">
                      <span className="px-2 py-1 bg-blue-50 text-blue-700 rounded">{scene.location}</span>
                      <span className="px-2 py-1 bg-purple-50 text-purple-700 rounded">{scene.time_of_day}</span>
                      <span className="px-2 py-1 bg-orange-50 text-orange-700 rounded">{scene.mood}</span>
                    </div>
                    {scene.dialogue && (
                      <p className="text-sm text-gray-500 mt-2 italic">"{scene.dialogue}"</p>
                    )}
                    {scene.image_url && (
                      <div className="mt-3">
                        <img src={scene.image_url} alt="" className="rounded-lg max-w-md" />
                        {scene.video_url && (
                          <div className="mt-3">
                            <video
                              src={scene.video_url}
                              controls
                              className="rounded-lg max-w-md w-full"
                              poster={scene.image_url}
                            >
                              您的浏览器不支持视频播放
                            </video>
                          </div>
                        )}
                      </div>
                    )}

                    {/* 场景测试面板 */}
                    {scene.image_url && (
                      <div className="mt-4 pt-4 border-t border-gray-100">
                        <h4 className="text-sm font-medium text-gray-700 mb-2">场景测试</h4>
                        <div className="flex gap-2 flex-wrap mb-3">
                          {!scene.video_url && (
                            <button
                              onClick={() => handleGenerateSceneVideo(scene.id)}
                              disabled={generatingScene === scene.id && generatingType === 'video'}
                              className="px-3 py-1.5 text-sm bg-purple-500 text-white rounded hover:bg-purple-600 disabled:bg-gray-400 flex items-center gap-1"
                            >
                              <Video className="w-3 h-3" />
                              {generatingScene === scene.id && generatingType === 'video' ? '生成中...' : '生成视频'}
                            </button>
                          )}
                          <button
                            onClick={() => handleGenerateSceneAudio(scene.id)}
                            disabled={generatingScene === scene.id}
                            className="px-3 py-1.5 text-sm bg-blue-500 text-white rounded hover:bg-blue-600 disabled:bg-gray-400 flex items-center gap-1"
                          >
                            <Volume2 className="w-3 h-3" />
                            {generatingScene === scene.id && generatingType === 'audio' ? '生成中...' : '生成配音'}
                          </button>
                          <button
                            onClick={() => handleGenerateSceneSubtitle(scene.id)}
                            disabled={generatingScene === scene.id}
                            className="px-3 py-1.5 text-sm bg-green-500 text-white rounded hover:bg-green-600 disabled:bg-gray-400 flex items-center gap-1"
                          >
                            <Subtitles className="w-3 h-3" />
                            {generatingScene === scene.id && generatingType === 'subtitle' ? '生成中...' : '生成字幕'}
                          </button>
                        </div>

                        {/* 测试结果预览 */}
                        {sceneAudios[scene.id] && sceneAudios[scene.id].length > 0 && (
                          <div className="bg-gray-50 rounded p-2 mt-2">
                            <p className="text-xs text-gray-600 mb-2">配音预览:</p>
                            {sceneAudios[scene.id].map((audio, idx) => (
                              <div key={audio.id || idx} className="flex items-center gap-2 mb-1">
                                <span className="text-xs text-gray-500">
                                  {audio.audio_type === 'dialogue' ? '对话' : '旁白'}
                                </span>
                                <audio controls src={audio.audio_url} className="h-6 w-40" />
                              </div>
                            ))}
                          </div>
                        )}

                        {sceneSubtitles[scene.id] && sceneSubtitles[scene.id].length > 0 && (
                          <div className="bg-gray-50 rounded p-2 mt-2">
                            <p className="text-xs text-gray-600 mb-2">字幕预览:</p>
                            {sceneSubtitles[scene.id].map((subtitle, idx) => (
                              <div key={subtitle.id || idx} className="mb-1">
                                <span className="text-xs text-gray-500">
                                  [{subtitle.start_time.toFixed(1)}s - {subtitle.end_time.toFixed(1)}s]
                                </span>
                                <span className="text-xs text-gray-700 ml-2">{subtitle.text}</span>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Merged Video */}
      {mergedVideo && (
        <div className="mb-6 bg-white rounded-lg shadow p-4">
          <h2 className="text-lg font-semibold mb-3 flex items-center gap-2">
            <Film className="w-5 h-5" />
            完整视频 ({mergedVideo.duration.toFixed(1)}秒)
          </h2>
          <video
            src={mergedVideo.url}
            controls
            className="w-full max-w-3xl rounded-lg"
          />
          <div className="mt-3 flex gap-2">
            <a
              href={mergedVideo.url}
              download
              className="px-4 py-2 bg-indigo-500 text-white rounded-lg hover:bg-indigo-600 flex items-center gap-2"
            >
              <Download className="w-4 h-4" />
              下载视频
            </a>
          </div>
        </div>
      )}

      {/* Raw Content */}
      {isEditing ? (
        <div className="bg-white rounded-lg shadow p-4">
          <label className="block font-semibold mb-2">原始文本</label>
          <textarea
            className="w-full h-96 px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 font-mono text-sm"
            value={editContent}
            onChange={(e) => setEditContent(e.target.value)}
            placeholder="故事内容..."
          />
          <p className="mt-2 text-sm text-gray-500">
            修改内容后保存，故事状态将重置为草稿，需要重新解析。
          </p>
        </div>
      ) : (
        <details className="bg-white rounded-lg shadow p-4" open>
          <summary className="font-semibold cursor-pointer">原始文本</summary>
          <pre className="mt-4 text-sm text-gray-600 whitespace-pre-wrap">{story.content || ''}</pre>
        </details>
      )}
    </div>
  )
}

// Navigation bar with auth
const NavBar = () => {
  const { user, isAuthenticated, logout } = useAuth()

  const handleLogout = async () => {
    await logout()
  }

  return (
    <nav className="bg-white shadow-sm">
      <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
        <Link to="/" className="text-xl font-bold text-gray-800 flex items-center gap-2">
          <FileText className="w-6 h-6 text-blue-500" />
          StoryFlow
        </Link>
        <div className="flex items-center gap-4 text-sm">
          {isAuthenticated ? (
            <>
              <Link to="/stories/new" className="text-gray-600 hover:text-gray-800">
                新建
              </Link>
              <Link to="/stories" className="text-gray-600 hover:text-gray-800">
                故事列表
              </Link>
              <Link to="/config" className="text-gray-600 hover:text-gray-800">
                配置
              </Link>
              {user?.role === 'admin' && (
                <Link to="/admin" className="text-blue-600 hover:text-blue-800 flex items-center gap-1">
                  <Shield className="w-4 h-4" />
                  管理后台
                </Link>
              )}
              <span className="text-gray-400">|</span>
              <span className="text-gray-600">{user?.email}</span>
              <button
                onClick={handleLogout}
                className="flex items-center gap-1 text-gray-600 hover:text-gray-800"
              >
                <LogOut className="w-4 h-4" />
                退出
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="px-4 py-1.5 text-blue-500 hover:text-blue-600">
                登录
              </Link>
              <Link to="/register" className="px-4 py-1.5 bg-blue-500 text-white rounded-lg hover:bg-blue-600">
                注册
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  )
}

function AppContent() {
  return (
    <div className="min-h-screen bg-gray-50">
      <NavBar />
      <main className="max-w-7xl mx-auto">
        <Routes>
          {/* Public routes */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/forgot-password" element={<ForgotPasswordPage />} />
          <Route path="/reset-password" element={<ResetPasswordPage />} />

          {/* Protected routes */}
          <Route path="/" element={
            <ProtectedRoute>
              <HomePage />
            </ProtectedRoute>
          } />
          <Route path="/stories/new" element={
            <ProtectedRoute>
              <NewStoryPage />
            </ProtectedRoute>
          } />
          <Route path="/stories" element={
            <ProtectedRoute>
              <StoryListPage />
            </ProtectedRoute>
          } />
          <Route path="/stories/:id" element={
            <ProtectedRoute>
              <StoryDetailPage />
            </ProtectedRoute>
          } />
          <Route path="/stories/:id/audio" element={
            <ProtectedRoute>
              <AudioConfigPage />
            </ProtectedRoute>
          } />
          <Route path="/config" element={
            <ProtectedRoute>
              <ConfigPage />
            </ProtectedRoute>
          } />
          <Route path="/admin" element={
            <ProtectedRoute>
              <AdminPage />
            </ProtectedRoute>
          } />
        </Routes>
      </main>
    </div>
  )
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <AppContent />
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App