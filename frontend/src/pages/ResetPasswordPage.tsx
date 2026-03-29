import { useState, useEffect } from 'react'
import { Link, useSearchParams, useNavigate } from 'react-router-dom'
import { authApi } from '../services/api'
import { FileText, Lock, AlertCircle, CheckCircle } from 'lucide-react'

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [token, setToken] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    const tokenParam = searchParams.get('token')
    if (tokenParam) {
      setToken(tokenParam)
    }
  }, [searchParams])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (password.length < 6) {
      setError('密码长度至少6位')
      return
    }

    if (password !== confirmPassword) {
      setError('两次输入的密码不一致')
      return
    }

    setLoading(true)

    try {
      await authApi.resetPassword(token, password)
      setSuccess(true)
      setTimeout(() => {
        navigate('/login')
      }, 3000)
    } catch (err: any) {
      const message = err.response?.data?.error || '重置失败，请重试'
      setError(message)
    } finally {
      setLoading(false)
    }
  }

  const passwordStrength = () => {
    if (password.length === 0) return { level: 0, text: '', color: '' }
    if (password.length < 6) return { level: 1, text: '弱', color: 'text-red-500' }
    if (password.length < 10) return { level: 2, text: '中等', color: 'text-yellow-500' }
    return { level: 3, text: '强', color: 'text-green-500' }
  }

  const strength = passwordStrength()

  if (!token) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="max-w-md w-full">
          <div className="bg-white rounded-lg shadow p-6 text-center">
            <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-xl font-semibold mb-2">无效链接</h2>
            <p className="text-gray-600 mb-4">密码重置链接无效或已过期。</p>
            <Link to="/forgot-password" className="text-blue-500 hover:text-blue-600">
              重新获取重置链接
            </Link>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="max-w-md w-full">
        <div className="text-center mb-8">
          <div className="flex items-center justify-center gap-2 mb-4">
            <FileText className="w-8 h-8 text-blue-500" />
            <h1 className="text-2xl font-bold">StoryFlow</h1>
          </div>
          <p className="text-gray-600">设置新密码</p>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          {success ? (
            <div className="text-center">
              <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <CheckCircle className="w-8 h-8 text-green-500" />
              </div>
              <h2 className="text-xl font-semibold mb-2">密码已重置</h2>
              <p className="text-gray-600 mb-4">
                您的密码已成功重置，正在跳转到登录页面...
              </p>
              <Link to="/login" className="text-blue-500 hover:text-blue-600">
                立即登录
              </Link>
            </div>
          ) : (
            <>
              <h2 className="text-xl font-semibold mb-6 text-center">设置新密码</h2>

              {error && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700">
                  <AlertCircle className="w-5 h-5" />
                  <span>{error}</span>
                </div>
              )}

              <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-2">新密码</label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                    <input
                      type="password"
                      className="w-full pl-10 pr-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      placeholder="至少6位"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      required
                    />
                  </div>
                  {password && (
                    <div className={`mt-1 text-sm ${strength.color}`}>
                      密码强度: {strength.text}
                    </div>
                  )}
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">确认密码</label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                    <input
                      type="password"
                      className="w-full pl-10 pr-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      placeholder="再次输入密码"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      required
                    />
                    {confirmPassword && password === confirmPassword && (
                      <CheckCircle className="absolute right-3 top-1/2 -translate-y-1/2 w-5 h-5 text-green-500" />
                    )}
                  </div>
                </div>

                <button
                  type="submit"
                  className="w-full py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 transition-colors"
                  disabled={loading || !password || !confirmPassword}
                >
                  {loading ? '重置中...' : '重置密码'}
                </button>
              </form>

              <div className="mt-6 text-center text-sm text-gray-600">
                记得密码？{' '}
                <Link to="/login" className="text-blue-500 hover:text-blue-600 font-medium">
                  登录
                </Link>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}