'use client'

import { useEffect, useState, useCallback } from 'react'
import type { UserWithOrg } from '@/types'
import { adminApi } from '@/lib/adminApi'
import { useAdminAuthStore } from '@/stores/adminAuthStore'
import { useAuthStore } from '@/stores/authStore'
import { Key, X } from 'lucide-react'

export default function AdminUsersPage() {
  const { token } = useAdminAuthStore()
  const { setAuth } = useAuthStore()
  const [users, setUsers] = useState<UserWithOrg[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)

  // Reset password modal state
  const [resetPasswordUser, setResetPasswordUser] = useState<UserWithOrg | null>(null)
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [resetPasswordError, setResetPasswordError] = useState('')
  const [resetPasswordLoading, setResetPasswordLoading] = useState(false)

  const loadUsers = useCallback(async () => {
    if (!token) return

    setLoading(true)
    try {
      adminApi.setToken(token)
      const response = await adminApi.listUsers({
        page,
        limit: 10,
        search: search || undefined,
      })
      setUsers(response.data || [])
      setTotalPages(response.pagination.total_pages)
      setTotal(response.pagination.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load users')
    } finally {
      setLoading(false)
    }
  }, [token, page, search])

  useEffect(() => {
    loadUsers()
  }, [loadUsers])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setPage(1)
    loadUsers()
  }

  const handleSuspend = async (id: string) => {
    const reason = prompt('Enter reason for suspension:')
    if (!reason) return

    try {
      await adminApi.suspendUser(id, { reason })
      loadUsers()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to suspend user')
    }
  }

  const handleReactivate = async (id: string) => {
    if (!confirm('Are you sure you want to reactivate this user?')) return

    try {
      await adminApi.reactivateUser(id)
      loadUsers()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to reactivate user')
    }
  }

  const handleImpersonate = async (userId: string, userName: string) => {
    const reason = prompt(`Enter reason for impersonating ${userName}:`)
    if (!reason) return

    try {
      // Start impersonation and get the token
      adminApi.setToken(token)
      const response = await adminApi.startImpersonation(userId, { reason })

      // Set the impersonation token in the regular auth store so /dashboard works
      setAuth(response.user, response.token)

      // Redirect to the main application dashboard as the impersonated user
      window.location.href = '/dashboard'
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to start impersonation')
    }
  }

  const openResetPasswordModal = (user: UserWithOrg) => {
    setResetPasswordUser(user)
    setNewPassword('')
    setConfirmPassword('')
    setResetPasswordError('')
  }

  const closeResetPasswordModal = () => {
    setResetPasswordUser(null)
    setNewPassword('')
    setConfirmPassword('')
    setResetPasswordError('')
  }

  const handleResetPassword = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!resetPasswordUser) return

    // Validation
    if (newPassword.length < 8) {
      setResetPasswordError('Password must be at least 8 characters')
      return
    }
    if (!/[A-Z]/.test(newPassword)) {
      setResetPasswordError('Password must contain at least one uppercase letter')
      return
    }
    if (!/[a-z]/.test(newPassword)) {
      setResetPasswordError('Password must contain at least one lowercase letter')
      return
    }
    if (!/[0-9]/.test(newPassword)) {
      setResetPasswordError('Password must contain at least one number')
      return
    }
    if (newPassword !== confirmPassword) {
      setResetPasswordError('Passwords do not match')
      return
    }

    setResetPasswordLoading(true)
    setResetPasswordError('')

    try {
      await adminApi.resetUserPassword(resetPasswordUser.id, newPassword)
      closeResetPasswordModal()
      alert('Password reset successfully')
    } catch (err) {
      setResetPasswordError(err instanceof Error ? err.message : 'Failed to reset password')
    } finally {
      setResetPasswordLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-white">Users</h1>
        <p className="text-gray-400">Manage all users across organizations</p>
      </div>

      {/* Search */}
      <form onSubmit={handleSearch} className="flex gap-4">
        <input
          type="text"
          placeholder="Search users by name or email..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="flex-1 px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <button
          type="submit"
          className="px-4 py-2 bg-gray-700 text-white rounded-lg hover:bg-gray-600"
        >
          Search
        </button>
      </form>

      {error && (
        <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        </div>
      ) : (
        <>
          {/* Users Table */}
          <div className="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
            <table className="w-full">
              <thead className="bg-gray-900">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    User
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Email
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Organization
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Role
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700">
                {users.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-6 py-8 text-center text-gray-400">
                      No users found
                    </td>
                  </tr>
                ) : (
                  users.map((user) => (
                    <tr key={user.id} className="hover:bg-gray-700/50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center">
                            <span className="text-white font-medium text-sm">
                              {user.first_name?.[0]}{user.last_name?.[0]}
                            </span>
                          </div>
                          <div className="ml-3">
                            <p className="text-white">
                              {user.first_name} {user.last_name}
                            </p>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-gray-300">
                        {user.email}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-gray-300">
                        {user.org_name}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="px-2 py-1 text-xs bg-gray-600 text-white rounded capitalize">
                          {user.role}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {user.suspended_at ? (
                          <span className="px-2 py-1 text-xs bg-red-600 text-white rounded">
                            Suspended
                          </span>
                        ) : user.is_active ? (
                          <span className="px-2 py-1 text-xs bg-green-600 text-white rounded">
                            Active
                          </span>
                        ) : (
                          <span className="px-2 py-1 text-xs bg-gray-600 text-white rounded">
                            Inactive
                          </span>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right">
                        <div className="flex justify-end gap-2">
                          <button
                            onClick={() => openResetPasswordModal(user)}
                            className="px-3 py-1 text-sm bg-amber-600 text-white rounded hover:bg-amber-500 flex items-center gap-1"
                            title="Reset password"
                          >
                            <Key className="h-3 w-3" />
                            Reset Password
                          </button>
                          <button
                            onClick={() => handleImpersonate(user.id, `${user.first_name} ${user.last_name}`)}
                            className="px-3 py-1 text-sm bg-purple-600 text-white rounded hover:bg-purple-500"
                            title="Impersonate user"
                          >
                            Impersonate
                          </button>
                          {user.suspended_at ? (
                            <button
                              onClick={() => handleReactivate(user.id)}
                              className="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-500"
                            >
                              Reactivate
                            </button>
                          ) : (
                            <button
                              onClick={() => handleSuspend(user.id)}
                              className="px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-500"
                            >
                              Suspend
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex justify-between items-center">
              <p className="text-sm text-gray-400">
                Showing page {page} of {totalPages} ({total} total)
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage(p => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="px-4 py-2 bg-gray-700 text-white rounded-lg hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Previous
                </button>
                <button
                  onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                  className="px-4 py-2 bg-gray-700 text-white rounded-lg hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}

      {/* Reset Password Modal */}
      {resetPasswordUser && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-lg border border-gray-700 p-6 w-full max-w-md">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold text-white">Reset Password</h2>
              <button
                onClick={closeResetPasswordModal}
                className="text-gray-400 hover:text-white"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <p className="text-gray-400 text-sm mb-4">
              Reset password for <span className="text-white font-medium">{resetPasswordUser.first_name} {resetPasswordUser.last_name}</span> ({resetPasswordUser.email})
            </p>

            {resetPasswordError && (
              <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded mb-4 text-sm">
                {resetPasswordError}
              </div>
            )}

            <form onSubmit={handleResetPassword} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">
                  New Password
                </label>
                <input
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Enter new password"
                  required
                />
                <p className="text-xs text-gray-500 mt-1">
                  Min 8 chars, uppercase, lowercase, and number required
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">
                  Confirm Password
                </label>
                <input
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Confirm new password"
                  required
                />
              </div>

              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={closeResetPasswordModal}
                  className="px-4 py-2 bg-gray-700 text-white rounded-lg hover:bg-gray-600"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={resetPasswordLoading}
                  className="px-4 py-2 bg-amber-600 text-white rounded-lg hover:bg-amber-500 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {resetPasswordLoading ? 'Resetting...' : 'Reset Password'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
