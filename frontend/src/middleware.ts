import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { locales, defaultLocale, isValidLocale } from '@/lib/i18n/config'

const PUBLIC_FILE = /\.(.*)$/
const LOCALE_COOKIE = 'NEXT_LOCALE'

function getLocaleFromHeaders(request: NextRequest): string {
  const acceptLanguage = request.headers.get('accept-language')
  if (!acceptLanguage) return defaultLocale

  const preferredLocale = acceptLanguage
    .split(',')
    .map((lang) => lang.split(';')[0].trim().substring(0, 2))
    .find((lang) => locales.includes(lang as typeof locales[number]))

  return preferredLocale || defaultLocale
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Skip public files and API routes
  if (
    PUBLIC_FILE.test(pathname) ||
    pathname.startsWith('/api') ||
    pathname.startsWith('/_next')
  ) {
    return NextResponse.next()
  }

  // Check if pathname already has a locale
  const pathnameHasLocale = locales.some(
    (locale) => pathname.startsWith(`/${locale}/`) || pathname === `/${locale}`
  )

  if (pathnameHasLocale) {
    return NextResponse.next()
  }

  // For paths that don't need locale routing (login, register, dashboard, admin)
  const nonLocalePaths = ['/login', '/register', '/dashboard', '/admin']
  if (nonLocalePaths.some((path) => pathname.startsWith(path))) {
    return NextResponse.next()
  }

  // Redirect to locale-prefixed path for landing page
  const cookieLocale = request.cookies.get(LOCALE_COOKIE)?.value
  const locale = cookieLocale && isValidLocale(cookieLocale)
    ? cookieLocale
    : getLocaleFromHeaders(request)

  const url = request.nextUrl.clone()
  url.pathname = `/${locale}${pathname}`

  const response = NextResponse.redirect(url)
  response.cookies.set(LOCALE_COOKIE, locale, {
    path: '/',
    maxAge: 60 * 60 * 24 * 365, // 1 year
  })

  return response
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
}
