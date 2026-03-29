import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Users, UserCheck, UserX, Trash2, Shield, BarChart3, ChevronRight, Search } from 'lucide-react'
import { adminApi, AdminUser } from '../services/api'
import { useAuth } from '../contexts/AuthContext'

const AdminPage = () => {
  const { user } = useAuth()
  const [stats, setStats] = useState<{
    total_users: number
    active_users: number
    suspended_users: number
    admin_count: number
    total_stories: number
  } | null>(null)
  const [users, setUsers] = useState<AdminUser[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [pageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState('')
  const [roleFilter, setRoleFilter] = useState('')
  const [searchEmail, setSearchEmail] = useState('')
  const [selectedUser, setSelectedUser] = useState<AdminUser | null>(null)
  const [showDetail, setShowDetail] = useState(false)

  // Check if user is admin
  if (user?.role !== 'admin') {
    return (
      <div className="p-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-center">
          <Shield className="w-12 h-12 text-red-500 mx-auto mb-3" />
          <h1 className="text-xl font-bold text-red-700 mb-2">需要管理员权限</h1>
          <p className="text-red-600">您没有权限访问此页面</p>
          <Link to="/" className="mt-4 inline-block px-4 py-2 bg-blue-500 text-white rounded-lg">
            返回首页
          </Link>
        </div>
      </div>
    )
  }

  useEffect(() => {
    loadStats()
    loadUsers()
  }, [page, statusFilter, roleFilter])

  const loadStats = async () => {
    try {
      const res = await adminApi.getStats()
      setStats(res.data)
    } catch (err) {
      console.error('Failed to load stats', err)
    }
  }

  const loadUsers = async () => {
    setLoading(true)
    try {
      const res = await adminApi.listUsers(page, pageSize, statusFilter, roleFilter)
      setUsers(res.data.data || [])
      setTotal(res.data.total)
    } catch (err) {
      console.error('Failed to load users', err)
    } finally {
      setLoading(false)
    }
  }

  const handleSuspend = async (id: string) => {
    if (!confirm('确定要暂停该用户账号吗？')) return
    try {
      await adminApi.suspendUser(id)
      loadUsers()
      loadStats()
    } catch (err) {
      alert('操作失败')
    }
  }

  const handleActivate = async (id: string) => {
    try {
      await adminApi.activateUser(id)
      loadUsers()
      loadStats()
    } catch (err) {
      alert('操作失败')
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除该用户吗？此操作不可恢复！')) return
    try {
      await adminApi.deleteUser(id)
      loadUsers()
      loadStats()
      setShowDetail(false)
      setSelectedUser(null)
    } catch (err) {
      alert('操作失败')
    }
  }

  const handleUpdateRole = async (id: string, role: 'admin' | 'user') => {
    try {
      await adminApi.updateUser(id, { role })
      loadUsers()
      loadStats()
      if (selectedUser?.id === id) {
        setSelectedUser({ ...selectedUser, role })
      }
    } catch (err: any) {
      const msg = err.response?.data?.error || '操作失败'
      alert(msg)
    }
  }

  const filteredUsers = searchEmail
    ? users.filter(u => u.email.toLowerCase().includes(searchEmail.toLowerCase()))
    : users

  const totalPages = Math.ceil(total / pageSize)

  const formatDate = (dateStr: string) => {
    if (!dateStr) return '-'
    return new Date(dateStr).toLocaleString('zh-CN')
  }

  return (
    <div className="p-8">
      <div className="flex items-center gap-3 mb-6">
        <Shield className="w-8 h-8 text-blue-500" />
        <h1 className="text-2xl font-bold">管理员控制台</h1>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-8">
          <div className="bg-white rounded-lg shadow p-4 border">
            <div className="flex items-center gap-2 mb-2">
              <Users className="w-5 h-5 text-gray-500" />
              <span className="text-sm text-gray-500">总用户数</span>
            </div>
            <p className="text-2xl font-bold">{stats.total_users}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-4 border">
            <div className="flex items-center gap-2 mb-2">
              <UserCheck className="w-5 h-5 text-green-500" />
              <span className="text-sm text-gray-500">活跃用户</span>
            </div>
            <p className="text-2xl font-bold text-green-600">{stats.active_users}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-4 border">
            <div className="flex items-center gap-2 mb-2">
              <UserX className="w-5 h-5 text-red-500" />
              <span className="text-sm text-gray-500">暂停用户</span>
            </div>
            <p className="text-2xl font-bold text-red-600">{stats.suspended_users}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-4 border">
            <div className="flex items-center gap-2 mb-2">
              <Shield className="w-5 h-5 text-blue-500" />
              <span className="text-sm text-gray-500">管理员</span>
            </div>
            <p className="text-2xl font-bold text-blue-600">{stats.admin_count}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-4 border">
            <div className="flex items-center gap-2 mb-2">
              <BarChart3 className="w-5 h-5 text-purple-500" />
              <span className="text-sm text-gray-500">总故事数</span>
            </div>
            <p className="text-2xl font-bold text-purple-600">{stats.total_stories}</p>
          </div>
        </div>
      )}

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4 mb-4 border">
        <div className="flex flex-wrap gap-4 items-center">
          <div className="flex items-center gap-2">
            <Search className="w-4 h-4 text-gray-400" />
            <input
              type="text"
              placeholder="搜索邮箱..."
              className="px-3 py-2 border rounded-lg w-64"
              value={searchEmail}
              onChange={(e) => setSearchEmail(e.target.value)}
            />
          </div>
          <select
            className="px-3 py-2 border rounded-lg"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
          >
            <option value="">全部状态</option>
            <option value="active">活跃</option>
            <option value="suspended">暂停</option>
            <option value="deleted">已删除</option>
          </select>
          <select
            className="px-3 py-2 border rounded-lg"
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
          >
            <option value="">全部角色</option>
            <option value="admin">管理员</option>
            <option value="user">普通用户</option>
          </select>
          <button
            onClick={() => loadUsers()}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
          >
            刷新
          </button>
        </div>
      </div>

      {/* Users Table */}
      <div className="bg-white rounded-lg shadow border overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">邮箱</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">姓名</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">角色</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">状态</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">注册时间</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">最后登录</th>
              <th className="px-4 py-3 text-right text-sm font-medium text-gray-500">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {loading ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-gray-500">
                  加载中...
                </td>
              </tr>
            ) : filteredUsers.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-gray-500">
                  暂无数据
                </td>
              </tr>
            ) : (
              filteredUsers.map((u) => (
                <tr key={u.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm">{u.email}</td>
                  <td className="px-4 py-3 text-sm">{u.name || '-'}</td>
                  <td className="px-4 py-3 text-sm">
                    <span className={`px-2 py-1 rounded text-xs ${
                      u.role === 'admin' ? 'bg-blue-100 text-blue-700' : 'bg-gray-100 text-gray-700'
                    }`}>
                      {u.role === 'admin' ? '管理员' : '用户'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <span className={`px-2 py-1 rounded text-xs ${
                      u.status === 'active' ? 'bg-green-100 text-green-700' :
                      u.status === 'suspended' ? 'bg-red-100 text-red-700' :
                      'bg-gray-100 text-gray-500'
                    }`}>
                      {u.status === 'active' ? '活跃' :
                       u.status === 'suspended' ? '暂停' : '已删除'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">{formatDate(u.created_at)}</td>
                  <td className="px-4 py-3 text-sm text-gray-500">{formatDate(u.last_login_at || '')}</td>
                  <td className="px-4 py-3 text-sm text-right">
                    <div className="flex gap-2 justify-end">
                      <button
                        onClick={() => { setSelectedUser(u); setShowDetail(true); }}
                        className="px-2 py-1 text-blue-500 hover:bg-blue-50 rounded"
                      >
                        <ChevronRight className="w-4 h-4" />
                      </button>
                      {u.status === 'active' && u.id !== user?.id && (
                        <button
                          onClick={() => handleSuspend(u.id)}
                          className="px-2 py-1 text-orange-500 hover:bg-orange-50 rounded"
                          title="暂停账号"
                        >
                          <UserX className="w-4 h-4" />
                        </button>
                      )}
                      {u.status === 'suspended' && (
                        <button
                          onClick={() => handleActivate(u.id)}
                          className="px-2 py-1 text-green-500 hover:bg-green-50 rounded"
                          title="激活账号"
                        >
                          <UserCheck className="w-4 h-4" />
                        </button>
                      )}
                      {u.id !== user?.id && (
                        <button
                          onClick={() => handleDelete(u.id)}
                          className="px-2 py-1 text-red-500 hover:bg-red-50 rounded"
                          title="删除账号"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-between items-center px-4 py-3 bg-gray-50 border-t">
            <span className="text-sm text-gray-500">
              共 {total} 条，第 {page} / {totalPages} 页
            </span>
            <div className="flex gap-2">
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 border rounded hover:bg-gray-100 disabled:opacity-50"
              >
                上一页
              </button>
              <button
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-3 py-1 border rounded hover:bg-gray-100 disabled:opacity-50"
              >
                下一页
              </button>
            </div>
          </div>
        )}
      </div>

      {/* User Detail Modal */}
      {showDetail && selectedUser && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-lg max-w-lg w-full mx-4 p-6">
            <div className="flex justify-between items-start mb-4">
              <h2 className="text-xl font-bold">用户详情</h2>
              <button
                onClick={() => { setShowDetail(false); setSelectedUser(null); }}
                className="text-gray-400 hover:text-gray-600"
              >
                ✕
              </button>
            </div>

            <div className="space-y-3">
              <div className="flex justify-between">
                <span className="text-gray-500">邮箱：</span>
                <span className="font-medium">{selectedUser.email}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">姓名：</span>
                <span className="font-medium">{selectedUser.name || '-'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">状态：</span>
                <span className={`px-2 py-1 rounded text-xs ${
                  selectedUser.status === 'active' ? 'bg-green-100 text-green-700' :
                  selectedUser.status === 'suspended' ? 'bg-red-100 text-red-700' :
                  'bg-gray-100 text-gray-500'
                }`}>
                  {selectedUser.status === 'active' ? '活跃' :
                   selectedUser.status === 'suspended' ? '暂停' : '已删除'}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-gray-500">角色：</span>
                {selectedUser.id !== user?.id ? (
                  <select
                    className="px-2 py-1 border rounded"
                    value={selectedUser.role}
                    onChange={(e) => handleUpdateRole(selectedUser.id, e.target.value as 'admin' | 'user')}
                  >
                    <option value="user">普通用户</option>
                    <option value="admin">管理员</option>
                  </select>
                ) : (
                  <span className="px-2 py-1 bg-blue-100 text-blue-700 rounded text-xs">
                    管理员 (自己)
                  </span>
                )}
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">注册时间：</span>
                <span>{formatDate(selectedUser.created_at)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">最后登录：</span>
                <span>{formatDate(selectedUser.last_login_at || '')}</span>
              </div>
            </div>

            {/* Actions */}
            <div className="flex gap-3 mt-6 pt-4 border-t">
              {selectedUser.status === 'active' && selectedUser.id !== user?.id && (
                <button
                  onClick={() => { handleSuspend(selectedUser.id); setShowDetail(false); }}
                  className="px-4 py-2 bg-orange-500 text-white rounded-lg hover:bg-orange-600"
                >
                  暂停账号
                </button>
              )}
              {selectedUser.status === 'suspended' && (
                <button
                  onClick={() => { handleActivate(selectedUser.id); setShowDetail(false); }}
                  className="px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600"
                >
                  激活账号
                </button>
              )}
              {selectedUser.id !== user?.id && (
                <button
                  onClick={() => { handleDelete(selectedUser.id); }}
                  className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600"
                >
                  删除账号
                </button>
              )}
              <button
                onClick={() => { setShowDetail(false); setSelectedUser(null); }}
                className="px-4 py-2 border rounded-lg hover:bg-gray-50"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default AdminPage