import React from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Volume2, Subtitles, Film, Download, ArrowLeft, CheckCircle, XCircle, Loader2 } from 'lucide-react'
import { audioApi, AudioFile, Subtitle } from '../services/api'

const AudioConfigPage: React.FC = () => {
  const { id: storyId } = useParams<{ id: string }>()
  const navigate = useNavigate()

  // State
  const [audios, setAudios] = React.useState<AudioFile[]>([])
  const [subtitles, setSubtitles] = React.useState<Subtitle[]>([])
  const [audioTaskId, setAudioTaskId] = React.useState<string | null>(null)
  const [audioStatus, setAudioStatus] = React.useState<string>('pending')
  const [audioProgress, setAudioProgress] = React.useState(0)
  const [synthesisTaskId, setSynthesisTaskId] = React.useState<string | null>(null)
  const [synthesisStatus, setSynthesisStatus] = React.useState<string>('pending')
  const [synthesisProgress, setSynthesisProgress] = React.useState(0)
  const [outputUrl, setOutputUrl] = React.useState<string>('')
  const [loading, setLoading] = React.useState(true)
  const [generatingAudio, setGeneratingAudio] = React.useState(false)
  const [generatingSubtitles, setGeneratingSubtitles] = React.useState(false)
  const [synthesizing, setSynthesizing] = React.useState(false)

  React.useEffect(() => {
    if (storyId) {
      loadData()
    }
  }, [storyId])

  // Poll for audio task status
  React.useEffect(() => {
    if (!audioTaskId || audioStatus === 'completed' || audioStatus === 'failed') return

    const poll = async () => {
      try {
        const res = await audioApi.getStatus(audioTaskId)
        setAudioStatus(res.data.status)
        setAudioProgress(res.data.progress)

        if (res.data.status === 'completed') {
          loadAudios()
        }
      } catch (err) {
        console.error('Failed to poll audio status', err)
      }
    }

    const interval = setInterval(poll, 2000)
    return () => clearInterval(interval)
  }, [audioTaskId, audioStatus])

  // Poll for synthesis task status
  React.useEffect(() => {
    if (!synthesisTaskId || synthesisStatus === 'completed' || synthesisStatus === 'failed') return

    const poll = async () => {
      try {
        const res = await audioApi.getSynthesisStatus(synthesisTaskId)
        setSynthesisStatus(res.data.status)
        setSynthesisProgress(res.data.progress)

        if (res.data.status === 'completed' && res.data.output_url) {
          setOutputUrl(res.data.output_url)
        }
      } catch (err) {
        console.error('Failed to poll synthesis status', err)
      }
    }

    const interval = setInterval(poll, 2000)
    return () => clearInterval(interval)
  }, [synthesisTaskId, synthesisStatus])

  const loadData = async () => {
    setLoading(true)
    try {
      await Promise.all([loadAudios(), loadSubtitles()])
    } finally {
      setLoading(false)
    }
  }

  const loadAudios = async () => {
    if (!storyId) return
    try {
      const res = await audioApi.getAudios(storyId)
      setAudios(res.data.audios || [])
    } catch (err) {
      console.error('Failed to load audios', err)
    }
  }

  const loadSubtitles = async () => {
    if (!storyId) return
    try {
      const res = await audioApi.getSubtitles(storyId)
      setSubtitles(res.data.subtitles || [])
    } catch (err) {
      console.error('Failed to load subtitles', err)
    }
  }

  const handleGenerateAudio = async () => {
    if (!storyId) return
    setGeneratingAudio(true)
    try {
      const res = await audioApi.generate(storyId)
      setAudioTaskId(res.data.task_id)
      setAudioStatus('pending')
      setAudioProgress(0)
    } catch (err: any) {
      alert(err.response?.data?.error || '生成配音失败')
    } finally {
      setGeneratingAudio(false)
    }
  }

  const handleGenerateSubtitles = async () => {
    if (!storyId) return
    setGeneratingSubtitles(true)
    try {
      await audioApi.generateSubtitles(storyId)
      await loadSubtitles()
      alert('字幕生成成功')
    } catch (err: any) {
      alert(err.response?.data?.error || '生成字幕失败')
    } finally {
      setGeneratingSubtitles(false)
    }
  }

  const handleSynthesizeVideo = async () => {
    if (!storyId) return
    setSynthesizing(true)
    try {
      const res = await audioApi.synthesizeVideo(storyId, {
        add_audio: true,
        add_subtitle: true,
      })
      setSynthesisTaskId(res.data.task_id)
      setSynthesisStatus('pending')
      setSynthesisProgress(0)
    } catch (err: any) {
      alert(err.response?.data?.error || '视频合成失败')
    } finally {
      setSynthesizing(false)
    }
  }

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = (seconds % 60).toFixed(1)
    return mins > 0 ? `${mins}分${secs}秒` : `${secs}秒`
  }

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    const ms = Math.floor((seconds % 1) * 100)
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}.${ms.toString().padStart(2, '0')}`
  }

  const totalDuration = audios.reduce((sum, a) => sum + a.duration, 0)

  if (loading) {
    return (
      <div className="p-8 flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
      </div>
    )
  }

  return (
    <div className="p-8">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <button
          onClick={() => navigate(-1)}
          className="p-2 hover:bg-gray-100 rounded-lg"
        >
          <ArrowLeft className="w-5 h-5" />
        </button>
        <h1 className="text-2xl font-bold">配音与字幕配置</h1>
      </div>

      {/* Action Buttons */}
      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="flex flex-wrap gap-3">
          {/* Generate Audio */}
          <button
            onClick={handleGenerateAudio}
            disabled={generatingAudio || (audioTaskId !== null && audioStatus === 'processing')}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 flex items-center gap-2"
          >
            {generatingAudio || audioStatus === 'processing' ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                生成中 ({audioProgress}%)
              </>
            ) : (
              <>
                <Volume2 className="w-4 h-4" />
                生成配音
              </>
            )}
          </button>

          {/* Generate Subtitles */}
          <button
            onClick={handleGenerateSubtitles}
            disabled={generatingSubtitles || audios.length === 0}
            className="px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 disabled:bg-gray-400 flex items-center gap-2"
          >
            {generatingSubtitles ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                生成中...
              </>
            ) : (
              <>
                <Subtitles className="w-4 h-4" />
                生成字幕
              </>
            )}
          </button>

          {/* Synthesize Video */}
          <button
            onClick={handleSynthesizeVideo}
            disabled={synthesizing || (synthesisTaskId !== null && synthesisStatus === 'processing') || audios.length === 0}
            className="px-4 py-2 bg-purple-500 text-white rounded-lg hover:bg-purple-600 disabled:bg-gray-400 flex items-center gap-2"
          >
            {synthesizing || synthesisStatus === 'processing' ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                合成中 ({synthesisProgress}%)
              </>
            ) : (
              <>
                <Film className="w-4 h-4" />
                合成视频
              </>
            )}
          </button>
        </div>

        {/* Task Status */}
        {audioTaskId && audioStatus !== 'pending' && (
          <div className="mt-4 p-3 bg-gray-50 rounded-lg">
            <div className="flex items-center gap-2">
              {audioStatus === 'completed' ? (
                <CheckCircle className="w-5 h-5 text-green-500" />
              ) : audioStatus === 'failed' ? (
                <XCircle className="w-5 h-5 text-red-500" />
              ) : (
                <Loader2 className="w-5 h-5 text-blue-500 animate-spin" />
              )}
              <span className="font-medium">配音任务</span>
              <span className="text-gray-500">- {audioStatus}</span>
              {audioStatus === 'processing' && (
                <span className="text-blue-500 ml-2">{audioProgress}%</span>
              )}
            </div>
          </div>
        )}

        {synthesisTaskId && synthesisStatus !== 'pending' && (
          <div className="mt-4 p-3 bg-gray-50 rounded-lg">
            <div className="flex items-center gap-2">
              {synthesisStatus === 'completed' ? (
                <CheckCircle className="w-5 h-5 text-green-500" />
              ) : synthesisStatus === 'failed' ? (
                <XCircle className="w-5 h-5 text-red-500" />
              ) : (
                <Loader2 className="w-5 h-5 text-purple-500 animate-spin" />
              )}
              <span className="font-medium">视频合成任务</span>
              <span className="text-gray-500">- {synthesisStatus}</span>
              {synthesisStatus === 'processing' && (
                <span className="text-purple-500 ml-2">{synthesisProgress}%</span>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Audio Files List */}
      {audios.length > 0 && (
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold flex items-center gap-2">
              <Volume2 className="w-5 h-5 text-blue-500" />
              音频文件 ({audios.length})
            </h2>
            <span className="text-sm text-gray-500">
              总时长: {formatDuration(totalDuration)}
            </span>
          </div>
          <div className="space-y-2">
            {audios.map((audio, index) => (
              <div key={audio.id} className="flex items-center gap-4 p-3 bg-gray-50 rounded-lg">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center text-blue-600 font-medium">
                  {index + 1}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className={`px-2 py-0.5 text-xs rounded ${
                      audio.audio_type === 'dialogue' ? 'bg-blue-100 text-blue-700' : 'bg-purple-100 text-purple-700'
                    }`}>
                      {audio.audio_type === 'dialogue' ? '对话' : '旁白'}
                    </span>
                    <span className="text-sm text-gray-500">
                      {formatDuration(audio.duration)}
                    </span>
                  </div>
                  <p className="text-sm text-gray-600 truncate mt-1">{audio.text_content}</p>
                </div>
                <audio controls src={audio.audio_url} className="h-8 w-48" />
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Subtitles List */}
      {subtitles.length > 0 && (
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
            <Subtitles className="w-5 h-5 text-green-500" />
            字幕 ({subtitles.length})
          </h2>
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {subtitles.map((sub, index) => (
              <div key={sub.id} className="flex items-start gap-4 p-3 bg-gray-50 rounded-lg">
                <div className="flex-shrink-0 w-8 h-8 bg-green-100 rounded-full flex items-center justify-center text-green-600 font-medium text-sm">
                  {index + 1}
                </div>
                <div className="flex-1">
                  <div className="flex items-center gap-2 text-xs text-gray-500">
                    <span>{formatTime(sub.start_time)}</span>
                    <span>→</span>
                    <span>{formatTime(sub.end_time)}</span>
                    <span className={`px-2 py-0.5 rounded ${
                      sub.subtitle_type === 'dialogue' ? 'bg-blue-100 text-blue-700' : 'bg-purple-100 text-purple-700'
                    }`}>
                      {sub.subtitle_type === 'dialogue' ? '对话' : '旁白'}
                    </span>
                  </div>
                  <p className="text-sm mt-1">{sub.text}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Synthesis Output */}
      {outputUrl && (
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
            <Film className="w-5 h-5 text-purple-500" />
            合成视频
          </h2>
          <video
            src={outputUrl}
            controls
            className="w-full max-w-3xl rounded-lg"
          />
          <div className="mt-3">
            <a
              href={outputUrl}
              download
              className="px-4 py-2 bg-purple-500 text-white rounded-lg hover:bg-purple-600 inline-flex items-center gap-2"
            >
              <Download className="w-4 h-4" />
              下载视频
            </a>
          </div>
        </div>
      )}

      {/* Empty State */}
      {audios.length === 0 && (
        <div className="bg-white rounded-lg shadow p-8 text-center">
          <Volume2 className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <p className="text-gray-500 mb-4">还没有生成配音</p>
          <button
            onClick={handleGenerateAudio}
            disabled={generatingAudio}
            className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400"
          >
            {generatingAudio ? '生成中...' : '开始生成配音'}
          </button>
        </div>
      )}
    </div>
  )
}

export default AudioConfigPage