'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { adminApi } from '@/lib/adminApi'
import { useAdminAuthStore } from '@/stores/adminAuthStore'
import type { OrganizationModule, ModuleName } from '@/types'

const moduleDescriptions: Record<ModuleName, string> = {
  construction: 'Construction management, budgets, projects and payments',
  appointments: 'Calendar, sessions, patients and therapists',
  notifications: 'WhatsApp notifications and automatic reminders'
}

export default function OrganizationModulesPage() {
  const params = useParams()
  const { token } = useAdminAuthStore()
  const [modules, setModules] = useState<OrganizationModule[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [toggling, setToggling] = useState<ModuleName | null>(null)

  const orgId = params.id as string

  useEffect(() => {
    if (!token || !orgId) return

    const loadModules = async () => {
      try {
        adminApi.setToken(token)
        const data = await adminApi.getOrganizationModules(orgId)
        setModules(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load modules')
      } finally {
        setLoading(false)
      }
    }

    loadModules()
  }, [token, orgId])

  const handleToggle = async (module: OrganizationModule) => {
    if (!token) return

    setToggling(module.module_name)
    setError('') // Clear previous error
    try {
      adminApi.setToken(token)
      if (module.is_enabled) {
        await adminApi.disableOrganizationModule(orgId, module.module_name)
      } else {
        await adminApi.enableOrganizationModule(orgId, module.module_name)
      }

      // Update local state
      setModules(prev => prev.map(m =>
        m.module_name === module.module_name
          ? { ...m, is_enabled: !m.is_enabled }
          : m
      ))
    } catch (err) {
      console.error('Failed to toggle module:', err)
      setError(err instanceof Error ? err.message : 'Failed to toggle module')
    } finally {
      setToggling(null)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <Link
          href={`/admin/organizations/${orgId}`}
          className="text-blue-400 hover:text-blue-300 text-sm"
        >
          &larr; Back to Organization
        </Link>
        <h1 className="text-2xl font-bold text-white mt-2">Manage Modules</h1>
        <p className="text-gray-400">Enable or disable modules for this organization</p>
      </div>

      {error && (
        <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <div className="space-y-4">
        {modules.map((module) => (
          <div
            key={module.module_name}
            className="bg-gray-800 rounded-lg border border-gray-700 p-6"
          >
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center gap-3">
                  <h3 className="text-lg font-semibold text-white">
                    {module.display_name}
                  </h3>
                  <span className={`px-2 py-1 rounded text-xs ${
                    module.is_enabled
                      ? 'bg-green-500/20 text-green-400'
                      : 'bg-gray-500/20 text-gray-400'
                  }`}>
                    {module.is_enabled ? 'Enabled' : 'Disabled'}
                  </span>
                </div>
                <p className="text-sm text-gray-400 mt-1">
                  {module.description || moduleDescriptions[module.module_name]}
                </p>
              </div>

              <button
                onClick={() => handleToggle(module)}
                disabled={toggling === module.module_name}
                className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-800 disabled:cursor-not-allowed disabled:opacity-50 ${
                  module.is_enabled ? 'bg-blue-600' : 'bg-gray-600'
                }`}
              >
                <span
                  className={`pointer-events-none inline-flex h-5 w-5 transform items-center justify-center rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
                    module.is_enabled ? 'translate-x-5' : 'translate-x-0'
                  }`}
                >
                  {toggling === module.module_name && (
                    <div className="h-3 w-3 animate-spin rounded-full border-2 border-gray-400 border-t-transparent"></div>
                  )}
                </span>
              </button>
            </div>
          </div>
        ))}
      </div>

      {modules.length === 0 && (
        <div className="text-center py-12 text-gray-400">
          No modules available
        </div>
      )}
    </div>
  )
}
