import React, { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { authApi } from '../services/api'
import { AuthState, LoginInput, RegisterInput } from '../types/auth'

interface AuthContextType extends AuthState {
  login: (input: LoginInput) => Promise<void>
  register: (input: RegisterInput) => Promise<void>
  logout: () => Promise<void>
  refreshAuth: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

const TOKEN_KEY = 'storyflow_access_token'
const REFRESH_TOKEN_KEY = 'storyflow_refresh_token'
const USER_KEY = 'storyflow_user'

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState>({
    user: null,
    isAuthenticated: false,
    isLoading: true,
  })
  const navigate = useNavigate()
  const location = useLocation()

  // Load user from storage on mount
  useEffect(() => {
    const loadAuth = async () => {
      const token = localStorage.getItem(TOKEN_KEY)
      const savedUser = localStorage.getItem(USER_KEY)

      if (token && savedUser) {
        try {
          // Verify token is still valid
          const res = await authApi.getMe()
          const userData = (res.data as any).user || res.data
          setState({
            user: userData,
            isAuthenticated: true,
            isLoading: false,
          })
          localStorage.setItem(USER_KEY, JSON.stringify(userData))
        } catch (err) {
          // Token expired, try refresh
          const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
          if (refreshToken) {
            try {
              const res = await authApi.refresh(refreshToken)
              const tokens = (res.data as any).tokens || res.data
              localStorage.setItem(TOKEN_KEY, tokens.access_token)
              localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token)

              const userRes = await authApi.getMe()
              const userData = (userRes.data as any).user || userRes.data
              setState({
                user: userData,
                isAuthenticated: true,
                isLoading: false,
              })
              localStorage.setItem(USER_KEY, JSON.stringify(userData))
            } catch (refreshErr) {
              // Refresh failed, clear storage
              localStorage.removeItem(TOKEN_KEY)
              localStorage.removeItem(REFRESH_TOKEN_KEY)
              localStorage.removeItem(USER_KEY)
              setState({
                user: null,
                isAuthenticated: false,
                isLoading: false,
              })
            }
          } else {
            localStorage.removeItem(TOKEN_KEY)
            localStorage.removeItem(USER_KEY)
            setState({
              user: null,
              isAuthenticated: false,
              isLoading: false,
            })
          }
        }
      } else {
        setState({
          user: null,
          isAuthenticated: false,
          isLoading: false,
        })
      }
    }

    loadAuth()
  }, [])

  const login = useCallback(async (input: LoginInput) => {
    try {
      const res = await authApi.login(input)
      const data = res.data as any
      // Handle both response formats: { tokens: { access_token, refresh_token } } or { access_token, refresh_token }
      const tokens = data.tokens || data
      const { access_token, refresh_token } = tokens

      localStorage.setItem(TOKEN_KEY, access_token)
      localStorage.setItem(REFRESH_TOKEN_KEY, refresh_token)

      // Get user info
      const userRes = await authApi.getMe()
      const userData = (userRes.data as any).user || userRes.data
      localStorage.setItem(USER_KEY, JSON.stringify(userData))

      setState({
        user: userData,
        isAuthenticated: true,
        isLoading: false,
      })

      // Redirect to intended page or home
      const from = location.state?.from?.pathname || '/'
      navigate(from, { replace: true })
    } catch (err: any) {
      const message = err.response?.data?.error || '登录失败'
      throw new Error(message)
    }
  }, [navigate, location])

  const register = useCallback(async (input: RegisterInput) => {
    try {
      const res = await authApi.register(input)
      const data = res.data as any
      // Handle both response formats
      const tokens = data.tokens || data
      const { access_token, refresh_token } = tokens

      localStorage.setItem(TOKEN_KEY, access_token)
      localStorage.setItem(REFRESH_TOKEN_KEY, refresh_token)

      // Get user info
      const userRes = await authApi.getMe()
      const userData = (userRes.data as any).user || userRes.data
      localStorage.setItem(USER_KEY, JSON.stringify(userData))

      setState({
        user: userData,
        isAuthenticated: true,
        isLoading: false,
      })

      navigate('/', { replace: true })
    } catch (err: any) {
      const message = err.response?.data?.error || '注册失败'
      throw new Error(message)
    }
  }, [navigate])

  const logout = useCallback(async () => {
    try {
      await authApi.logout()
    } catch (err) {
      // Ignore logout API errors
    }

    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    localStorage.removeItem(USER_KEY)

    setState({
      user: null,
      isAuthenticated: false,
      isLoading: false,
    })

    navigate('/login', { replace: true })
  }, [navigate])

  const refreshAuth = useCallback(async () => {
    const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
    if (!refreshToken) {
      throw new Error('No refresh token')
    }

    try {
      const res = await authApi.refresh(refreshToken)
      const data = res.data as any
      const tokens = data.tokens || data
      localStorage.setItem(TOKEN_KEY, tokens.access_token)
      localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token)

      const userRes = await authApi.getMe()
      const userData = (userRes.data as any).user || userRes.data
      localStorage.setItem(USER_KEY, JSON.stringify(userData))

      setState({
        user: userData,
        isAuthenticated: true,
        isLoading: false,
      })
    } catch (err) {
      logout()
      throw err
    }
  }, [logout])

  return (
    <AuthContext.Provider
      value={{
        ...state,
        login,
        register,
        logout,
        refreshAuth,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

export function getAccessToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}