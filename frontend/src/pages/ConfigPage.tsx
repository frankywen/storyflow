import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { configApi } from '../services/api'
import { UserConfig } from '../types/auth'
import { Settings, Key, Server, Palette, Check, AlertCircle, Eye, EyeOff, Loader2 } from 'lucide-react'

export default function ConfigPage() {
  const { user } = useAuth()
  const [config, setConfig] = useState<UserConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [error, setError] = useState('')

  // Validation states
  const [validatingLLM, setValidatingLLM] = useState(false)
  const [validatingImage, setValidatingImage] = useState(false)
  const [validatingVideo, setValidatingVideo] = useState(false)
  const [validationResults, setValidationResults] = useState<{
    llm?: { valid: boolean; message: string }
    image?: { valid: boolean; message: string }
    video?: { valid: boolean; message: string }
  }>({})

  // Form states
  const [llmProvider, setLlmProvider] = useState('claude')
  const [llmApiKey, setLlmApiKey] = useState('')
  const [llmModel, setLlmModel] = useState('')
  const [llmBaseUrl, setLlmBaseUrl] = useState('')
  const [imageProvider, setImageProvider] = useState('comfyui')
  const [imageApiKey, setImageApiKey] = useState('')
  const [imageBaseUrl, setImageBaseUrl] = useState('http://localhost:8188')
  const [imageModel, setImageModel] = useState('')
  const [videoProvider, setVideoProvider] = useState('')
  const [videoApiKey, setVideoApiKey] = useState('')
  const [videoBaseUrl, setVideoBaseUrl] = useState('')
  const [videoModel, setVideoModel] = useState('')
  const [defaultStyle, setDefaultStyle] = useState('manga')

  // Password visibility states
  const [showLlmKey, setShowLlmKey] = useState(false)
  const [showImageKey, setShowImageKey] = useState(false)
  const [showVideoKey, setShowVideoKey] = useState(false)

  useEffect(() => {
    loadConfig()
  }, [])

  const loadConfig = async () => {
    try {
      const res = await configApi.get()
      const configData = (res.data as any).config || res.data
      setConfig(configData)

      // Set form values from config
      if (configData) {
        setLlmProvider(configData.llm_provider || 'claude')
        setLlmModel(configData.llm_model || '')
        setLlmBaseUrl(configData.llm_base_url || '')
        setImageProvider(configData.image_provider || 'comfyui')
        setImageBaseUrl(configData.image_base_url || 'http://localhost:8188')
        setImageModel(configData.image_model || '')
        setVideoProvider(configData.video_provider || '')
        setVideoBaseUrl(configData.video_base_url || '')
        setVideoModel(configData.video_model || '')
        setDefaultStyle(configData.default_style || 'manga')
      }
    } catch (err) {
      console.error('Failed to load config:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleValidateLLM = async () => {
    setValidatingLLM(true)
    try {
      const res = await configApi.validateAPIKey({
        type: 'llm',
        provider: llmProvider,
        api_key: llmApiKey || undefined,
        base_url: llmBaseUrl || undefined,
      })
      setValidationResults(prev => ({ ...prev, llm: { valid: res.data.valid, message: res.data.message } }))
    } catch (err: any) {
      setValidationResults(prev => ({ ...prev, llm: { valid: false, message: err.response?.data?.error || '验证失败' } }))
    } finally {
      setValidatingLLM(false)
    }
  }

  const handleValidateImage = async () => {
    setValidatingImage(true)
    try {
      const res = await configApi.validateAPIKey({
        type: 'image',
        provider: imageProvider,
        api_key: imageApiKey || undefined,
        base_url: imageBaseUrl || undefined,
      })
      setValidationResults(prev => ({ ...prev, image: { valid: res.data.valid, message: res.data.message } }))
    } catch (err: any) {
      setValidationResults(prev => ({ ...prev, image: { valid: false, message: err.response?.data?.error || '验证失败' } }))
    } finally {
      setValidatingImage(false)
    }
  }

  const handleValidateVideo = async () => {
    if (!videoProvider) return
    setValidatingVideo(true)
    try {
      const res = await configApi.validateAPIKey({
        type: 'video',
        provider: videoProvider,
        api_key: videoApiKey || undefined,
        base_url: videoBaseUrl || undefined,
      })
      setValidationResults(prev => ({ ...prev, video: { valid: res.data.valid, message: res.data.message } }))
    } catch (err: any) {
      setValidationResults(prev => ({ ...prev, video: { valid: false, message: err.response?.data?.error || '验证失败' } }))
    } finally {
      setValidatingVideo(false)
    }
  }

  const handleSaveLLM = async () => {
    setSaving(true)
    setError('')
    try {
      await configApi.updateLLM({
        provider: llmProvider,
        api_key: llmApiKey || undefined,
        model: llmModel || undefined,
        base_url: llmBaseUrl || undefined,
      })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
      loadConfig()
    } catch (err: any) {
      setError(err.response?.data?.error || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleSaveImage = async () => {
    setSaving(true)
    setError('')
    try {
      await configApi.updateImage({
        provider: imageProvider,
        api_key: imageApiKey || undefined,
        base_url: imageBaseUrl || undefined,
        model: imageModel || undefined,
      })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
      loadConfig()
    } catch (err: any) {
      setError(err.response?.data?.error || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleSaveVideo = async () => {
    setSaving(true)
    setError('')
    try {
      await configApi.updateVideo({
        provider: videoProvider,
        api_key: videoApiKey || undefined,
        base_url: videoBaseUrl || undefined,
        model: videoModel || undefined,
      })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
      loadConfig()
    } catch (err: any) {
      setError(err.response?.data?.error || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleSaveStyle = async () => {
    setSaving(true)
    setError('')
    try {
      await configApi.update({ default_style: defaultStyle })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
      loadConfig()
    } catch (err: any) {
      setError(err.response?.data?.error || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="p-8">
        <p className="text-gray-500">加载中...</p>
      </div>
    )
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Settings className="w-6 h-6" />
          配置
        </h1>
        {saved && (
          <div className="flex items-center gap-2 text-green-600">
            <Check className="w-5 h-5" />
            已保存
          </div>
        )}
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700">
          <AlertCircle className="w-5 h-5" />
          <span>{error}</span>
        </div>
      )}

      <div className="mb-4 p-4 bg-blue-50 border border-blue-200 rounded-lg text-sm text-blue-700">
        <p className="font-medium mb-1">用户: {user?.email}</p>
        <p>API Key 配置会加密存储，显示时只保留后4位用于确认。</p>
      </div>

      <div className="max-w-2xl space-y-6">
        {/* LLM Provider Settings */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="font-semibold mb-4 flex items-center gap-2">
            <Key className="w-5 h-5" />
            大语言模型 (LLM) 配置
          </h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">LLM 提供商</label>
              <select
                className="w-full px-4 py-2 border rounded-lg"
                value={llmProvider}
                onChange={(e) => setLlmProvider(e.target.value)}
              >
                <option value="claude">Claude (Anthropic)</option>
                <option value="openai">OpenAI / ChatGPT</option>
                <option value="volcengine">火山引擎 / 豆包</option>
                <option value="alibaba">阿里云 / 通义千问</option>
                <option value="deepseek">DeepSeek</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">API Key</label>
              <div className="relative">
                <input
                  type={showLlmKey ? 'text' : 'password'}
                  className="w-full px-4 py-2 pr-10 border rounded-lg"
                  placeholder={config?.llm_api_key_masked || 'sk-...'}
                  value={llmApiKey}
                  onChange={(e) => setLlmApiKey(e.target.value)}
                />
                <button
                  type="button"
                  className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-gray-400 hover:text-gray-600"
                  onClick={() => setShowLlmKey(!showLlmKey)}
                >
                  {showLlmKey ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
              {config?.llm_api_key_masked && !llmApiKey && (
                <p className="text-xs text-gray-500 mt-1">当前: {config.llm_api_key_masked}（如需修改请输入新的Key）</p>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">模型</label>
              <input
                type="text"
                className="w-full px-4 py-2 border rounded-lg"
                placeholder="claude-sonnet-4-20250514"
                value={llmModel}
                onChange={(e) => setLlmModel(e.target.value)}
              />
              <p className="text-xs text-gray-500 mt-1">留空使用默认模型</p>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">自定义 API 地址</label>
              <div className="relative">
                <Server className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="text"
                  className="w-full pl-10 pr-4 py-2 border rounded-lg"
                  placeholder="https://api.anthropic.com"
                  value={llmBaseUrl}
                  onChange={(e) => setLlmBaseUrl(e.target.value)}
                />
              </div>
              <p className="text-xs text-gray-500 mt-1">可选，用于自定义API端点</p>
            </div>

            {/* Validation Result */}
            {validationResults.llm && (
              <div className={`p-3 rounded-lg flex items-center gap-2 ${validationResults.llm.valid ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
                {validationResults.llm.valid ? <Check className="w-5 h-5" /> : <AlertCircle className="w-5 h-5" />}
                <span>{validationResults.llm.message}</span>
              </div>
            )}

            <div className="flex gap-2">
              <button
                className="px-4 py-2 border border-blue-500 text-blue-500 rounded-lg hover:bg-blue-50 disabled:bg-gray-100 disabled:text-gray-400 disabled:border-gray-300 flex items-center gap-2"
                onClick={handleValidateLLM}
                disabled={validatingLLM || (!llmApiKey && !config?.llm_api_key_masked)}
              >
                {validatingLLM ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
                {validatingLLM ? '验证中...' : '验证 API Key'}
              </button>
              <button
                className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 flex items-center gap-2"
                onClick={handleSaveLLM}
                disabled={saving}
              >
                {saving ? '保存中...' : '保存 LLM 配置'}
              </button>
            </div>
          </div>
        </div>

        {/* Image Generation Settings */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="font-semibold mb-4 flex items-center gap-2">
            <Key className="w-5 h-5" />
            图片生成配置
          </h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">图片生成提供商</label>
              <select
                className="w-full px-4 py-2 border rounded-lg"
                value={imageProvider}
                onChange={(e) => setImageProvider(e.target.value)}
              >
                <option value="comfyui">ComfyUI (本地 GPU)</option>
                <option value="siliconflow">SiliconFlow</option>
                <option value="volcengine">火山引擎 / 豆包</option>
                <option value="alibaba">阿里云 / 通义万相</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">API Key</label>
              <div className="relative">
                <input
                  type={showImageKey ? 'text' : 'password'}
                  className="w-full px-4 py-2 pr-10 border rounded-lg"
                  placeholder={config?.image_api_key_masked || 'API Key (云端服务需要)'}
                  value={imageApiKey}
                  onChange={(e) => setImageApiKey(e.target.value)}
                />
                <button
                  type="button"
                  className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-gray-400 hover:text-gray-600"
                  onClick={() => setShowImageKey(!showImageKey)}
                >
                  {showImageKey ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
              {config?.image_api_key_masked && !imageApiKey && (
                <p className="text-xs text-gray-500 mt-1">当前: {config.image_api_key_masked}</p>
              )}
              <p className="text-xs text-gray-500 mt-1">ComfyUI 本地部署无需配置</p>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">服务地址</label>
              <div className="relative">
                <Server className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="text"
                  className="w-full pl-10 pr-4 py-2 border rounded-lg"
                  placeholder="http://localhost:8188"
                  value={imageBaseUrl}
                  onChange={(e) => setImageBaseUrl(e.target.value)}
                />
              </div>
              <p className="text-xs text-gray-500 mt-1">ComfyUI 服务地址或云端 API 地址</p>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">模型</label>
              <input
                type="text"
                className="w-full px-4 py-2 border rounded-lg"
                placeholder="flux-dev"
                value={imageModel}
                onChange={(e) => setImageModel(e.target.value)}
              />
            </div>

            {/* Validation Result */}
            {validationResults.image && (
              <div className={`p-3 rounded-lg flex items-center gap-2 ${validationResults.image.valid ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
                {validationResults.image.valid ? <Check className="w-5 h-5" /> : <AlertCircle className="w-5 h-5" />}
                <span>{validationResults.image.message}</span>
              </div>
            )}

            <div className="flex gap-2">
              <button
                className="px-4 py-2 border border-blue-500 text-blue-500 rounded-lg hover:bg-blue-50 disabled:bg-gray-100 disabled:text-gray-400 disabled:border-gray-300 flex items-center gap-2"
                onClick={handleValidateImage}
                disabled={validatingImage}
              >
                {validatingImage ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
                {validatingImage ? '验证中...' : '验证配置'}
              </button>
              <button
                className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 flex items-center gap-2"
                onClick={handleSaveImage}
                disabled={saving}
              >
                {saving ? '保存中...' : '保存图片配置'}
              </button>
            </div>
          </div>
        </div>

        {/* Video Generation Settings */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="font-semibold mb-4 flex items-center gap-2">
            <Key className="w-5 h-5" />
            视频生成配置（可选）
          </h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">视频生成提供商</label>
              <select
                className="w-full px-4 py-2 border rounded-lg"
                value={videoProvider}
                onChange={(e) => setVideoProvider(e.target.value)}
              >
                <option value="">不配置视频生成</option>
                <option value="runway">Runway</option>
                <option value="pika">Pika</option>
                <option value="siliconflow">SiliconFlow</option>
                <option value="volcengine">火山引擎</option>
              </select>
            </div>
            {videoProvider && (
              <>
                <div>
                  <label className="block text-sm font-medium mb-2">API Key</label>
                  <div className="relative">
                    <input
                      type={showVideoKey ? 'text' : 'password'}
                      className="w-full px-4 py-2 pr-10 border rounded-lg"
                      placeholder={config?.video_api_key_masked || 'API Key'}
                      value={videoApiKey}
                      onChange={(e) => setVideoApiKey(e.target.value)}
                    />
                    <button
                      type="button"
                      className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-gray-400 hover:text-gray-600"
                      onClick={() => setShowVideoKey(!showVideoKey)}
                    >
                      {showVideoKey ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                    </button>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-2">服务地址</label>
                  <div className="relative">
                    <Server className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                    <input
                      type="text"
                      className="w-full pl-10 pr-4 py-2 border rounded-lg"
                      placeholder="https://api.runwayml.com"
                      value={videoBaseUrl}
                      onChange={(e) => setVideoBaseUrl(e.target.value)}
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-2">模型</label>
                  <input
                    type="text"
                    className="w-full px-4 py-2 border rounded-lg"
                    placeholder="gen3a_turbo"
                    value={videoModel}
                    onChange={(e) => setVideoModel(e.target.value)}
                  />
                </div>

                {/* Validation Result */}
                {validationResults.video && (
                  <div className={`p-3 rounded-lg flex items-center gap-2 ${validationResults.video.valid ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
                    {validationResults.video.valid ? <Check className="w-5 h-5" /> : <AlertCircle className="w-5 h-5" />}
                    <span>{validationResults.video.message}</span>
                  </div>
                )}

                <div className="flex gap-2">
                  <button
                    className="px-4 py-2 border border-blue-500 text-blue-500 rounded-lg hover:bg-blue-50 disabled:bg-gray-100 disabled:text-gray-400 disabled:border-gray-300 flex items-center gap-2"
                    onClick={handleValidateVideo}
                    disabled={validatingVideo || !videoApiKey}
                  >
                    {validatingVideo ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
                    {validatingVideo ? '验证中...' : '验证 API Key'}
                  </button>
                </div>
              </>
            )}
            <button
              className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 flex items-center gap-2"
              onClick={handleSaveVideo}
              disabled={saving}
            >
              {saving ? '保存中...' : '保存视频配置'}
            </button>
          </div>
        </div>

        {/* Default Style Settings */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="font-semibold mb-4 flex items-center gap-2">
            <Palette className="w-5 h-5" />
            默认风格
          </h2>
          <div className="space-y-4">
            <select
              className="w-full px-4 py-2 border rounded-lg"
              value={defaultStyle}
              onChange={(e) => setDefaultStyle(e.target.value)}
            >
              <option value="manga">日式漫画</option>
              <option value="manhwa">韩式漫画</option>
              <option value="western_comic">美式漫画</option>
              <option value="anime">动漫风格</option>
              <option value="realistic">写实风格</option>
            </select>
            <button
              className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 flex items-center gap-2"
              onClick={handleSaveStyle}
              disabled={saving}
            >
              {saving ? '保存中...' : '保存默认风格'}
            </button>
          </div>
        </div>

        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 text-sm text-yellow-800">
          <p className="font-medium mb-1">配置说明</p>
          <ul className="list-disc list-inside space-y-1 text-xs">
            <li>LLM 用于解析小说、提取角色和生成图片提示词</li>
            <li>图片生成支持本地 ComfyUI 和云端服务</li>
            <li>API Key 会加密存储，安全性较高</li>
            <li>配置修改后立即生效，无需重启服务</li>
            <li>验证功能会实际调用API测试，可能产生少量费用</li>
          </ul>
        </div>
      </div>
    </div>
  )
}