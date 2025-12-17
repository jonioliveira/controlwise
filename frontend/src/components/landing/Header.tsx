'use client'

import Link from 'next/link'
import { useState } from 'react'
import { Layers, Menu, X, Globe } from 'lucide-react'
import type { Dictionary } from '@/lib/i18n/dictionaries'
import type { Locale } from '@/lib/i18n/config'
import { locales, localeNames } from '@/lib/i18n/config'

interface HeaderProps {
  dict: Dictionary
  locale: Locale
}

export function Header({ dict, locale }: HeaderProps) {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [langMenuOpen, setLangMenuOpen] = useState(false)

  const switchLocale = (newLocale: Locale) => {
    document.cookie = `NEXT_LOCALE=${newLocale}; path=/; max-age=31536000`
    window.location.href = `/${newLocale}`
  }

  return (
    <header className="fixed top-0 left-0 right-0 z-50 bg-white/80 backdrop-blur-md border-b border-gray-100">
      <nav className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <Link href={`/${locale}`} className="flex items-center space-x-2">
            <div className="w-9 h-9 bg-primary-600 rounded-lg flex items-center justify-center">
              <Layers className="h-5 w-5 text-white" />
            </div>
            <span className="text-xl font-bold text-gray-900">ControlWise</span>
          </Link>

          {/* Desktop Navigation */}
          <div className="hidden md:flex items-center space-x-6">
            {/* Language Switcher */}
            <div className="relative">
              <button
                onClick={() => setLangMenuOpen(!langMenuOpen)}
                className="flex items-center space-x-1 text-gray-600 hover:text-gray-900 transition-colors"
              >
                <Globe className="h-4 w-4" />
                <span className="text-sm">{localeNames[locale]}</span>
              </button>
              {langMenuOpen && (
                <div className="absolute top-full right-0 mt-2 bg-white rounded-lg shadow-lg border border-gray-100 py-1 min-w-[120px]">
                  {locales.map((l) => (
                    <button
                      key={l}
                      onClick={() => {
                        switchLocale(l)
                        setLangMenuOpen(false)
                      }}
                      className={`w-full px-4 py-2 text-left text-sm hover:bg-gray-50 transition-colors ${
                        l === locale ? 'text-primary-600 font-medium' : 'text-gray-700'
                      }`}
                    >
                      {localeNames[l]}
                    </button>
                  ))}
                </div>
              )}
            </div>

            <Link
              href="/login"
              className="text-gray-600 hover:text-gray-900 font-medium transition-colors"
            >
              {dict.header.login}
            </Link>
            <Link
              href="/register"
              className="bg-primary-600 text-white px-5 py-2.5 rounded-lg font-medium hover:bg-primary-700 transition-colors shadow-sm"
            >
              {dict.header.startFree}
            </Link>
          </div>

          {/* Mobile Menu Button */}
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="md:hidden p-2 text-gray-600 hover:text-gray-900"
          >
            {mobileMenuOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
          </button>
        </div>

        {/* Mobile Menu */}
        {mobileMenuOpen && (
          <div className="md:hidden py-4 border-t border-gray-100">
            <div className="flex flex-col space-y-4">
              {/* Language Switcher Mobile */}
              <div className="flex items-center space-x-2 px-2">
                <Globe className="h-4 w-4 text-gray-500" />
                <div className="flex space-x-2">
                  {locales.map((l) => (
                    <button
                      key={l}
                      onClick={() => switchLocale(l)}
                      className={`text-sm px-2 py-1 rounded ${
                        l === locale
                          ? 'bg-primary-100 text-primary-700'
                          : 'text-gray-600 hover:bg-gray-100'
                      }`}
                    >
                      {l.toUpperCase()}
                    </button>
                  ))}
                </div>
              </div>
              <Link
                href="/login"
                className="text-gray-600 hover:text-gray-900 font-medium px-2"
              >
                {dict.header.login}
              </Link>
              <Link
                href="/register"
                className="bg-primary-600 text-white px-5 py-2.5 rounded-lg font-medium hover:bg-primary-700 transition-colors text-center"
              >
                {dict.header.startFree}
              </Link>
            </div>
          </div>
        )}
      </nav>
    </header>
  )
}
