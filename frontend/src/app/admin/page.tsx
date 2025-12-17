'use client'

import { useEffect, useState } from 'react'
import type { PlatformStats, RecentActivity } from '@/types'
import { adminApi } from '@/lib/adminApi'
import { useAdminAuthStore } from '@/stores/adminAuthStore'

export default function AdminDashboardPage() {
  const { token } = useAdminAuthStore()
  const [stats, setStats] = useState<PlatformStats | null>(null)
  const [recentActivity, setRecentActivity] = useState<RecentActivity[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!token) return

    const loadData = async () => {
      try {
        adminApi.setToken(token)
        const [statsData, activityData] = await Promise.all([
          adminApi.getStats(),
          adminApi.getRecentActivity(10),
        ])
        setStats(statsData)
        setRecentActivity(activityData || [])
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load data')
      } finally {
        setLoading(false)
      }
    }

    loadData()
  }, [token])

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

  const statCards = [
    { label: 'Total Organizations', value: stats?.total_organizations || 0, color: 'blue' },
    { label: 'Active Organizations', value: stats?.active_organizations || 0, color: 'green' },
    { label: 'Suspended Organizations', value: stats?.suspended_organizations || 0, color: 'red' },
    { label: 'Total Users', value: stats?.total_users || 0, color: 'purple' },
    { label: 'Active Users', value: stats?.active_users || 0, color: 'teal' },
    { label: 'New Orgs This Month', value: stats?.new_orgs_this_month || 0, color: 'orange' },
  ]

  const colorClasses: Record<string, string> = {
    blue: 'bg-blue-600',
    green: 'bg-green-600',
    red: 'bg-red-600',
    purple: 'bg-purple-600',
    teal: 'bg-teal-600',
    orange: 'bg-orange-600',
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-white">Dashboard</h1>
        <p className="text-gray-400">Platform overview and statistics</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {statCards.map((stat) => (
          <div
            key={stat.label}
            className="bg-gray-800 rounded-lg p-6 border border-gray-700"
          >
            <div className="flex items-center">
              <div className={`w-12 h-12 ${colorClasses[stat.color]} rounded-lg flex items-center justify-center`}>
                <span className="text-2xl font-bold text-white">{stat.value}</span>
              </div>
              <div className="ml-4">
                <p className="text-sm text-gray-400">{stat.label}</p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Modules Usage */}
      {stats?.orgs_by_module && Object.keys(stats.orgs_by_module).length > 0 && (
        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
          <h2 className="text-lg font-semibold text-white mb-4">Organizations by Module</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {Object.entries(stats.orgs_by_module).map(([module, count]) => (
              <div key={module} className="bg-gray-700 rounded-lg p-4">
                <p className="text-sm text-gray-400 capitalize">{module}</p>
                <p className="text-2xl font-bold text-white">{count}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Recent Activity */}
      <div className="bg-gray-800 rounded-lg border border-gray-700">
        <div className="px-6 py-4 border-b border-gray-700">
          <h2 className="text-lg font-semibold text-white">Recent Activity</h2>
        </div>
        <div className="divide-y divide-gray-700">
          {!recentActivity || recentActivity.length === 0 ? (
            <div className="px-6 py-8 text-center text-gray-400">
              No recent activity
            </div>
          ) : (
            recentActivity.map((activity, index) => (
              <div key={index} className="px-6 py-4 flex items-center justify-between">
                <div className="flex items-center">
                  <div className={`w-2 h-2 rounded-full mr-3 ${
                    activity.type === 'org_created' ? 'bg-green-500' : 'bg-blue-500'
                  }`} />
                  <div>
                    <p className="text-sm text-white">{activity.description}</p>
                    <p className="text-xs text-gray-400">
                      {activity.type === 'org_created' ? 'Organization created' : 'User created'}
                    </p>
                  </div>
                </div>
                <span className="text-xs text-gray-500">
                  {new Date(activity.created_at).toLocaleDateString()}
                </span>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
