'use client'

import { useEffect, useState, useCallback } from 'react'
import Link from 'next/link'
import type { OrganizationWithStats } from '@/types'
import { adminApi } from '@/lib/adminApi'
import { useAdminAuthStore } from '@/stores/adminAuthStore'

export default function AdminOrganizationsPage() {
  const { token } = useAdminAuthStore()
  const [organizations, setOrganizations] = useState<OrganizationWithStats[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)

  const loadOrganizations = useCallback(async () => {
    if (!token) return

    setLoading(true)
    try {
      adminApi.setToken(token)
      const response = await adminApi.listOrganizations({
        page,
        limit: 10,
        search: search || undefined,
      })
      setOrganizations(response.data || [])
      setTotalPages(response.pagination.total_pages)
      setTotal(response.pagination.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load organizations')
    } finally {
      setLoading(false)
    }
  }, [token, page, search])

  useEffect(() => {
    loadOrganizations()
  }, [loadOrganizations])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setPage(1)
    loadOrganizations()
  }

  const handleSuspend = async (id: string) => {
    const reason = prompt('Enter reason for suspension:')
    if (!reason) return

    try {
      await adminApi.suspendOrganization(id, { reason })
      loadOrganizations()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to suspend organization')
    }
  }

  const handleReactivate = async (id: string) => {
    if (!confirm('Are you sure you want to reactivate this organization?')) return

    try {
      await adminApi.reactivateOrganization(id)
      loadOrganizations()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to reactivate organization')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-white">Organizations</h1>
          <p className="text-gray-400">Manage all organizations on the platform</p>
        </div>
        <Link
          href="/admin/organizations/new"
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          Create Organization
        </Link>
      </div>

      {/* Search */}
      <form onSubmit={handleSearch} className="flex gap-4">
        <input
          type="text"
          placeholder="Search organizations..."
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
          {/* Organizations Table */}
          <div className="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
            <table className="w-full">
              <thead className="bg-gray-900">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Organization
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Email
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Users
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Created
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700">
                {organizations.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-6 py-8 text-center text-gray-400">
                      No organizations found
                    </td>
                  </tr>
                ) : (
                  organizations.map((org) => (
                    <tr key={org.id} className="hover:bg-gray-700/50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <Link href={`/admin/organizations/${org.id}`} className="text-white hover:text-blue-400">
                          {org.name}
                        </Link>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-gray-300">
                        {org.email}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-gray-300">
                        {org.active_user_count} / {org.user_count}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {org.suspended_at ? (
                          <span className="px-2 py-1 text-xs bg-red-600 text-white rounded">
                            Suspended
                          </span>
                        ) : org.is_active ? (
                          <span className="px-2 py-1 text-xs bg-green-600 text-white rounded">
                            Active
                          </span>
                        ) : (
                          <span className="px-2 py-1 text-xs bg-gray-600 text-white rounded">
                            Inactive
                          </span>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-gray-400 text-sm">
                        {new Date(org.created_at).toLocaleDateString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right">
                        <div className="flex justify-end gap-2">
                          <Link
                            href={`/admin/organizations/${org.id}`}
                            className="px-3 py-1 text-sm bg-gray-600 text-white rounded hover:bg-gray-500"
                          >
                            View
                          </Link>
                          {org.suspended_at ? (
                            <button
                              onClick={() => handleReactivate(org.id)}
                              className="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-500"
                            >
                              Reactivate
                            </button>
                          ) : (
                            <button
                              onClick={() => handleSuspend(org.id)}
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
    </div>
  )
}
