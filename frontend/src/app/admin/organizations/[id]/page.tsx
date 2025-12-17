'use client'

import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import type { OrganizationWithStats, UserWithOrg } from '@/types'
import { adminApi } from '@/lib/adminApi'
import { useAdminAuthStore } from '@/stores/adminAuthStore'

export default function OrganizationDetailPage() {
  const params = useParams()
  const router = useRouter()
  const { token } = useAdminAuthStore()
  const [organization, setOrganization] = useState<OrganizationWithStats | null>(null)
  const [users, setUsers] = useState<UserWithOrg[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [actionLoading, setActionLoading] = useState(false)
  const [showSuspendModal, setShowSuspendModal] = useState(false)
  const [suspendReason, setSuspendReason] = useState('')

  const orgId = params.id as string

  useEffect(() => {
    if (!token || !orgId) return

    const loadData = async () => {
      try {
        adminApi.setToken(token)
        const [orgData, usersData] = await Promise.all([
          adminApi.getOrganization(orgId),
          adminApi.listOrganizationUsers(orgId),
        ])
        setOrganization(orgData)
        setUsers(usersData.data || [])
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load organization')
      } finally {
        setLoading(false)
      }
    }

    loadData()
  }, [token, orgId])

  const handleSuspend = async () => {
    if (!token || !suspendReason.trim()) return

    setActionLoading(true)
    try {
      adminApi.setToken(token)
      await adminApi.suspendOrganization(orgId, { reason: suspendReason })
      setOrganization(prev => prev ? { ...prev, suspended_at: new Date().toISOString() } : null)
      setShowSuspendModal(false)
      setSuspendReason('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to suspend organization')
    } finally {
      setActionLoading(false)
    }
  }

  const handleReactivate = async () => {
    if (!token) return

    setActionLoading(true)
    try {
      adminApi.setToken(token)
      await adminApi.reactivateOrganization(orgId)
      setOrganization(prev => prev ? { ...prev, suspended_at: undefined } : null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to reactivate organization')
    } finally {
      setActionLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!token) return
    if (!confirm('Are you sure you want to delete this organization? This action cannot be undone.')) return

    setActionLoading(true)
    try {
      adminApi.setToken(token)
      await adminApi.deleteOrganization(orgId)
      router.push('/admin/organizations')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete organization')
      setActionLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded">
        {error}
      </div>
    )
  }

  if (!organization) {
    return (
      <div className="text-gray-400">Organization not found</div>
    )
  }

  const isSuspended = !!organization.suspended_at

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <Link
            href="/admin/organizations"
            className="text-blue-400 hover:text-blue-300 text-sm"
          >
            &larr; Back to Organizations
          </Link>
          <h1 className="text-2xl font-bold text-white mt-2">{organization.name}</h1>
        </div>
        <div className="flex gap-3">
          <Link
            href={`/admin/organizations/${orgId}/modules`}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            Manage Modules
          </Link>
          {isSuspended ? (
            <button
              onClick={handleReactivate}
              disabled={actionLoading}
              className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50"
            >
              Reactivate
            </button>
          ) : (
            <button
              onClick={() => setShowSuspendModal(true)}
              disabled={actionLoading}
              className="px-4 py-2 bg-yellow-600 text-white rounded-md hover:bg-yellow-700 disabled:opacity-50"
            >
              Suspend
            </button>
          )}
          <button
            onClick={handleDelete}
            disabled={actionLoading}
            className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50"
          >
            Delete
          </button>
        </div>
      </div>

      {isSuspended && (
        <div className="bg-yellow-500/10 border border-yellow-500 text-yellow-500 px-4 py-3 rounded">
          This organization is suspended. Reason: {organization.suspend_reason || 'No reason provided'}
        </div>
      )}

      {/* Organization Info */}
      <div className="bg-gray-800 rounded-lg border border-gray-700 p-6">
        <h2 className="text-lg font-semibold text-white mb-4">Organization Details</h2>
        <dl className="grid grid-cols-2 gap-4">
          <div>
            <dt className="text-sm text-gray-400">ID</dt>
            <dd className="text-white font-mono text-sm">{organization.id}</dd>
          </div>
          <div>
            <dt className="text-sm text-gray-400">Created</dt>
            <dd className="text-white">{new Date(organization.created_at).toLocaleDateString()}</dd>
          </div>
          <div>
            <dt className="text-sm text-gray-400">Total Users</dt>
            <dd className="text-white">{organization.user_count || 0}</dd>
          </div>
          <div>
            <dt className="text-sm text-gray-400">Status</dt>
            <dd>
              <span className={`px-2 py-1 rounded text-xs ${
                isSuspended ? 'bg-red-500/20 text-red-400' : 'bg-green-500/20 text-green-400'
              }`}>
                {isSuspended ? 'Suspended' : 'Active'}
              </span>
            </dd>
          </div>
        </dl>
      </div>

      {/* Users */}
      <div className="bg-gray-800 rounded-lg border border-gray-700">
        <div className="px-6 py-4 border-b border-gray-700">
          <h2 className="text-lg font-semibold text-white">Users ({users.length})</h2>
        </div>
        <div className="divide-y divide-gray-700">
          {users.length === 0 ? (
            <div className="px-6 py-8 text-center text-gray-400">
              No users in this organization
            </div>
          ) : (
            users.map((user) => (
              <div key={user.id} className="px-6 py-4 flex items-center justify-between">
                <div>
                  <p className="text-white font-medium">
                    {user.first_name} {user.last_name}
                  </p>
                  <p className="text-sm text-gray-400">{user.email}</p>
                </div>
                <div className="flex items-center gap-3">
                  <span className={`px-2 py-1 rounded text-xs ${
                    user.role === 'owner' ? 'bg-purple-500/20 text-purple-400' :
                    user.role === 'admin' ? 'bg-blue-500/20 text-blue-400' :
                    'bg-gray-500/20 text-gray-400'
                  }`}>
                    {user.role}
                  </span>
                  <Link
                    href={`/admin/users/${user.id}`}
                    className="text-blue-400 hover:text-blue-300 text-sm"
                  >
                    View
                  </Link>
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Suspend Modal */}
      {showSuspendModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-white mb-4">Suspend Organization</h3>
            <p className="text-gray-400 mb-4">
              Please provide a reason for suspending this organization.
            </p>
            <textarea
              value={suspendReason}
              onChange={(e) => setSuspendReason(e.target.value)}
              className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              rows={3}
              placeholder="Reason for suspension..."
            />
            <div className="flex justify-end gap-3 mt-4">
              <button
                onClick={() => setShowSuspendModal(false)}
                className="px-4 py-2 text-gray-300 hover:text-white"
              >
                Cancel
              </button>
              <button
                onClick={handleSuspend}
                disabled={!suspendReason.trim() || actionLoading}
                className="px-4 py-2 bg-yellow-600 text-white rounded-md hover:bg-yellow-700 disabled:opacity-50"
              >
                {actionLoading ? 'Suspending...' : 'Suspend'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
