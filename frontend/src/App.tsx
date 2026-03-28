import React, { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Link, useNavigate } from 'react-router-dom'
import { FileText, Image, Settings, Trash2, RefreshCw, Download, Video, Check, Upload, User, Film } from 'lucide-react'
import { storyApi, imageApi, exportApi, videoApi, characterApi } from './services/api'
import { Story, Character } from './types'

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

      <Link to="/settings" className="p-6 bg-white rounded-lg shadow hover:shadow-md transition-shadow border">
        <Settings className="w-8 h-8 text-purple-500 mb-3" />
        <h2 className="text-xl font-semibold mb-2">设置</h2>
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
        <div className="flex items-center gap-2 px-4 py-2 bg-blue-50 rounded-lg">
          <span className="w-6 h-6 bg-blue-500 text-white rounded-full flex items-center justify-center text-xs">4</span>
          视频生成
        </div>
        <span className="text-gray-400">→</span>
        <div className="flex items-center gap-2 px-4 py-2 bg-green-50 rounded-lg">
          <span className="w-6 h-6 bg-green-500 text-white rounded-full flex items-center justify-center text-xs">5</span>
          导出
        </div>
      </div>
    </div>
  </div>
)

const NewStoryPage = () => {
  const navigate = useNavigate()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [loading, setLoading] = useState(false)

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
  const [stories, setStories] = useState<Story[]>([])
  const [loading, setLoading] = useState(true)

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
  const [uploading, setUploading] = useState(false)

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
  const [story, setStory] = useState<Story | null>(null)
  const [loading, setLoading] = useState(true)
  const [parsing, setParsing] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [generatingVideo, setGeneratingVideo] = useState(false)
  const [videoProgress, setVideoProgress] = useState(0)
  const [mergingVideo, setMergingVideo] = useState(false)
  const [mergedVideo, setMergedVideo] = useState<{ url: string; duration: number } | null>(null)
  const [exporting, setExporting] = useState(false)
  const [selectedStyle, setSelectedStyle] = useState('manga')

  useEffect(() => {
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
      alert('解析失败，请检查CLAUDE_API_KEY是否配置')
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
      alert('生成失败，请检查ComfyUI是否启动')
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
      alert('视频生成失败，请检查VIDEO_PROVIDER是否配置或余额是否充足')
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
        <div>
          <h1 className="text-2xl font-bold">{story.title}</h1>
          <p className="text-gray-500 mt-1">{story.summary}</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleDelete}
            className="p-2 text-red-500 hover:bg-red-50 rounded-lg"
          >
            <Trash2 className="w-5 h-5" />
          </button>
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
      <details className="bg-white rounded-lg shadow p-4">
        <summary className="font-semibold cursor-pointer">原始文本</summary>
        <pre className="mt-4 text-sm text-gray-600 whitespace-pre-wrap">{story.content}</pre>
      </details>
    </div>
  )
}

const SettingsPage = () => {
  const [settings, setSettings] = useState({
    llmProvider: 'claude',
    llmApiKey: '',
    llmModel: '',
    imageProvider: 'comfyui',
    imageApiKey: '',
    imageBaseUrl: '',
    defaultStyle: 'manga',
  })
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    // Load settings from localStorage
    const savedSettings = localStorage.getItem('storyflow_settings')
    if (savedSettings) {
      setSettings(JSON.parse(savedSettings))
    }
  }, [])

  const handleSave = () => {
    localStorage.setItem('storyflow_settings', JSON.stringify(settings))
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  const handleChange = (field: string, value: string) => {
    setSettings(prev => ({ ...prev, [field]: value }))
  }

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-6">设置</h1>

      <div className="max-w-2xl space-y-6">
        {/* LLM Provider Settings */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="font-semibold mb-4">大语言模型 (LLM) 配置</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">LLM 提供商</label>
              <select
                className="w-full px-4 py-2 border rounded-lg"
                value={settings.llmProvider}
                onChange={(e) => handleChange('llmProvider', e.target.value)}
              >
                <option value="claude">Claude (Anthropic)</option>
                <option value="volcengine">火山引擎 / 豆包</option>
                <option value="alibaba">阿里云 / 通义千问</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">API Key</label>
              <input
                type="password"
                className="w-full px-4 py-2 border rounded-lg"
                placeholder="sk-..."
                value={settings.llmApiKey}
                onChange={(e) => handleChange('llmApiKey', e.target.value)}
              />
              <p className="text-xs text-gray-500 mt-1">用于故事解析和角色提取</p>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">模型</label>
              <input
                type="text"
                className="w-full px-4 py-2 border rounded-lg"
                placeholder="claude-sonnet-4-20250514"
                value={settings.llmModel}
                onChange={(e) => handleChange('llmModel', e.target.value)}
              />
              <p className="text-xs text-gray-500 mt-1">留空使用默认模型</p>
            </div>
          </div>
        </div>

        {/* Image Generation Settings */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="font-semibold mb-4">图片生成配置</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">图片生成提供商</label>
              <select
                className="w-full px-4 py-2 border rounded-lg"
                value={settings.imageProvider}
                onChange={(e) => handleChange('imageProvider', e.target.value)}
              >
                <option value="comfyui">ComfyUI (本地 GPU)</option>
                <option value="volcengine">火山引擎 / 豆包</option>
                <option value="alibaba">阿里云 / 通义万相</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">API Key</label>
              <input
                type="password"
                className="w-full px-4 py-2 border rounded-lg"
                placeholder="API Key (云端服务需要)"
                value={settings.imageApiKey}
                onChange={(e) => handleChange('imageApiKey', e.target.value)}
              />
              <p className="text-xs text-gray-500 mt-1">ComfyUI 本地部署无需配置</p>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">服务地址</label>
              <input
                type="text"
                className="w-full px-4 py-2 border rounded-lg"
                placeholder="http://localhost:8188"
                value={settings.imageBaseUrl}
                onChange={(e) => handleChange('imageBaseUrl', e.target.value)}
              />
              <p className="text-xs text-gray-500 mt-1">ComfyUI 服务地址或云端 API 地址</p>
            </div>
          </div>
        </div>

        {/* Default Style Settings */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="font-semibold mb-4">默认风格</h2>
          <select
            className="w-full px-4 py-2 border rounded-lg"
            value={settings.defaultStyle}
            onChange={(e) => handleChange('defaultStyle', e.target.value)}
          >
            <option value="manga">日式漫画</option>
            <option value="manhwa">韩式漫画</option>
            <option value="western_comic">美式漫画</option>
            <option value="anime">动漫风格</option>
            <option value="realistic">写实风格</option>
          </select>
        </div>

        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 text-sm text-yellow-800">
          <p className="font-medium mb-1">配置说明</p>
          <ul className="list-disc list-inside space-y-1 text-xs">
            <li>LLM 用于解析小说、提取角色和生成图片提示词</li>
            <li>图片生成支持本地 ComfyUI 和云端服务</li>
            <li>配置修改后需要重启后端服务生效</li>
            <li>火山引擎和阿里云需要申请对应的 API Key</li>
            <li>前端设置保存在浏览器本地存储，后端配置请修改 .env 文件</li>
          </ul>
        </div>

        <div className="flex items-center gap-4">
          <button
            className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 flex items-center gap-2"
            onClick={handleSave}
          >
            {saved ? (
              <>
                <Check className="w-4 h-4" />
                已保存
              </>
            ) : (
              '保存设置'
            )}
          </button>
          {saved && (
            <span className="text-green-600 text-sm">设置已保存到浏览器本地存储</span>
          )}
        </div>
      </div>
    </div>
  )
}

function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-gray-50">
        <nav className="bg-white shadow-sm">
          <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
            <Link to="/" className="text-xl font-bold text-gray-800 flex items-center gap-2">
              <FileText className="w-6 h-6 text-blue-500" />
              StoryFlow
            </Link>
            <div className="flex gap-4 text-sm">
              <Link to="/stories/new" className="text-gray-600 hover:text-gray-800">
                新建
              </Link>
              <Link to="/stories" className="text-gray-600 hover:text-gray-800">
                故事列表
              </Link>
              <Link to="/settings" className="text-gray-600 hover:text-gray-800">
                设置
              </Link>
            </div>
          </div>
        </nav>

        <main className="max-w-7xl mx-auto">
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/stories/new" element={<NewStoryPage />} />
            <Route path="/stories" element={<StoryListPage />} />
            <Route path="/stories/:id" element={<StoryDetailPage />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}

export default App