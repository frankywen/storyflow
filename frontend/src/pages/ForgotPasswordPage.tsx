import { useState } from 'react'
import { Link } from 'react-router-dom'
import { authApi } from '../services/api'
import { FileText, Mail, AlertCircle, ArrowLeft } from 'lucide-react'

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [resetToken, setResetToken] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const res = await authApi.forgotPassword(email)
      setSuccess(true)
      // For development, show the reset token
      if (res.data.reset_token) {
        setResetToken(res.data.reset_token)
      }
    } catch (err: any) {
      const message = err.response?.data?.error || '请求失败，请稍后重试'
      setError(message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="max-w-md w-full">
        <div className="text-center mb-8">
          <div className="flex items-center justify-center gap-2 mb-4">
            <FileText className="w-8 h-8 text-blue-500" />
            <h1 className="text-2xl font-bold">StoryFlow</h1>
          </div>
          <p className="text-gray-600">重置密码</p>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          {success ? (
            <div className="text-center">
              <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <Mail className="w-8 h-8 text-green-500" />
              </div>
              <h2 className="text-xl font-semibold mb-2">邮件已发送</h2>
              <p className="text-gray-600 mb-4">
                如果该邮箱已注册，您将收到密码重置链接。
              </p>

              {/* Development only - show reset token */}
              {resetToken && (
                <div className="mt-4 p-4 bg-yellow-50 border border-yellow-200 rounded-lg text-left">
                  <p className="text-sm font-medium text-yellow-800 mb-2">开发模式 - 重置Token:</p>
                  <p className="text-xs text-yellow-700 break-all font-mono bg-white p-2 rounded">
                    {resetToken}
                  </p>
                  <Link
                    to={`/reset-password?token=${resetToken}`}
                    className="mt-3 inline-block text-sm text-blue-500 hover:text-blue-600"
                  >
                    点击这里重置密码 →
                  </Link>
                </div>
              )}

              <Link
                to="/login"
                className="mt-6 inline-flex items-center gap-2 text-gray-600 hover:text-gray-800"
              >
                <ArrowLeft className="w-4 h-4" />
                返回登录
              </Link>
            </div>
          ) : (
            <>
              <h2 className="text-xl font-semibold mb-4 text-center">忘记密码</h2>
              <p className="text-gray-600 text-sm mb-6 text-center">
                输入您的邮箱地址，我们将发送密码重置链接。
              </p>

              {error && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700">
                  <AlertCircle className="w-5 h-5" />
                  <span>{error}</span>
                </div>
              )}

              <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-2">邮箱</label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                    <input
                      type="email"
                      className="w-full pl-10 pr-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      placeholder="your@email.com"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      required
                    />
                  </div>
                </div>

                <button
                  type="submit"
                  className="w-full py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 transition-colors"
                  disabled={loading || !email}
                >
                  {loading ? '发送中...' : '发送重置链接'}
                </button>
              </form>

              <div className="mt-6 text-center">
                <Link to="/login" className="text-sm text-gray-600 hover:text-gray-800">
                  记得密码？ 返回登录
                </Link>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}