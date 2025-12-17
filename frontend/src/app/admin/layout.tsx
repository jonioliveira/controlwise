'use client'

import { useEffect } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import Link from 'next/link'
import { useAdminAuthStore } from '@/stores/adminAuthStore'

const navItems = [
  { href: '/admin', label: 'Dashboard', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' },
  { href: '/admin/organizations', label: 'Organizations', icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4' },
  { href: '/admin/users', label: 'Users', icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' },
  { href: '/admin/audit-logs', label: 'Audit Logs', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
]

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const pathname = usePathname()
  const { isAuthenticated, admin, logout, loadAdmin, isImpersonating, impersonatedUser, endImpersonation } = useAdminAuthStore()

  useEffect(() => {
    // Don't check auth on login page
    if (pathname === '/admin/login') return

    // Load admin if we have a token but no admin data
    loadAdmin()
  }, [pathname, loadAdmin])

  useEffect(() => {
    // Redirect to login if not authenticated (except on login page)
    if (pathname !== '/admin/login' && !isAuthenticated) {
      router.push('/admin/login')
    }
  }, [pathname, isAuthenticated, router])

  // Show login page without layout
  if (pathname === '/admin/login') {
    return children
  }

  // Don't show layout until authenticated
  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-900">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    )
  }

  const handleLogout = () => {
    logout()
    router.push('/admin/login')
  }

  const handleEndImpersonation = async () => {
    try {
      await endImpersonation()
      // Redirect back to admin dashboard
      window.location.href = '/admin'
    } catch (error) {
      console.error('Failed to end impersonation:', error)
    }
  }

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Impersonation Banner */}
      {isImpersonating && impersonatedUser && (
        <div className="bg-yellow-600 px-4 py-2 text-center">
          <span className="text-white font-medium">
            Impersonating: {impersonatedUser.first_name} {impersonatedUser.last_name} ({impersonatedUser.email})
          </span>
          <button
            onClick={handleEndImpersonation}
            className="ml-4 px-3 py-1 bg-yellow-800 text-white rounded hover:bg-yellow-900 text-sm"
          >
            End Impersonation
          </button>
        </div>
      )}

      {/* Sidebar */}
      <div className="fixed inset-y-0 left-0 w-64 bg-gray-800">
        <div className="flex flex-col h-full">
          <div className="flex items-center justify-center h-16 px-4 bg-gray-900">
            <span className="text-xl font-bold text-white">ControlWise</span>
            <span className="ml-2 px-2 py-1 text-xs bg-red-600 text-white rounded">Admin</span>
          </div>

          <nav className="flex-1 px-4 py-4 space-y-1">
            {navItems.map((item) => {
              const isActive = pathname === item.href || (item.href !== '/admin' && pathname.startsWith(item.href))
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`flex items-center px-4 py-2 text-sm rounded-lg transition-colors ${
                    isActive
                      ? 'bg-blue-600 text-white'
                      : 'text-gray-300 hover:bg-gray-700 hover:text-white'
                  }`}
                >
                  <svg className="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={item.icon} />
                  </svg>
                  {item.label}
                </Link>
              )
            })}
          </nav>

          <div className="p-4 border-t border-gray-700">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center">
                  <span className="text-white font-medium text-sm">
                    {admin?.first_name?.[0]}{admin?.last_name?.[0]}
                  </span>
                </div>
              </div>
              <div className="ml-3 flex-1 min-w-0">
                <p className="text-sm font-medium text-white truncate">
                  {admin?.first_name} {admin?.last_name}
                </p>
                <p className="text-xs text-gray-400 truncate">{admin?.email}</p>
              </div>
              <button
                onClick={handleLogout}
                className="ml-2 p-2 text-gray-400 hover:text-white"
                title="Logout"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="pl-64">
        <main className="p-8">{children}</main>
      </div>
    </div>
  )
}
