'use client'

import Link from 'next/link'
import { Layers, Globe } from 'lucide-react'
import type { Dictionary } from '@/lib/i18n/dictionaries'
import type { Locale } from '@/lib/i18n/config'
import { locales, localeNames } from '@/lib/i18n/config'

interface FooterProps {
  dict: Dictionary
  locale: Locale
}

export function Footer({ dict, locale }: FooterProps) {
  const switchLocale = (newLocale: Locale) => {
    document.cookie = `NEXT_LOCALE=${newLocale}; path=/; max-age=31536000`
    window.location.href = `/${newLocale}`
  }

  const currentYear = new Date().getFullYear()

  return (
    <footer className="bg-gray-900 text-gray-400">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        {/* Main Footer */}
        <div className="py-12 lg:py-16 grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-8">
          {/* Brand Column */}
          <div className="col-span-2 md:col-span-4 lg:col-span-2">
            <Link href={`/${locale}`} className="flex items-center space-x-2 mb-4">
              <div className="w-9 h-9 bg-primary-600 rounded-lg flex items-center justify-center">
                <Layers className="h-5 w-5 text-white" />
              </div>
              <span className="text-xl font-bold text-white">ControleWise</span>
            </Link>
            <p className="text-gray-500 max-w-xs mb-6">
              {dict.metadata.description}
            </p>

            {/* Language Switcher */}
            <div className="flex items-center gap-2">
              <Globe className="h-4 w-4 text-gray-500" />
              <div className="flex gap-1">
                {locales.map((l) => (
                  <button
                    key={l}
                    onClick={() => switchLocale(l)}
                    className={`text-sm px-2 py-1 rounded transition-colors ${
                      l === locale
                        ? 'bg-gray-800 text-white'
                        : 'text-gray-500 hover:text-gray-300'
                    }`}
                  >
                    {localeNames[l]}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Product Links */}
          <div>
            <h4 className="text-white font-semibold mb-4">{dict.footer.product}</h4>
            <ul className="space-y-3">
              <li>
                <Link href="#features" className="hover:text-white transition-colors">
                  {dict.footer.features}
                </Link>
              </li>
              <li>
                <Link href="/pricing" className="hover:text-white transition-colors">
                  {dict.footer.pricing}
                </Link>
              </li>
              <li>
                <Link href="/demo" className="hover:text-white transition-colors">
                  {dict.footer.demo}
                </Link>
              </li>
            </ul>
          </div>

          {/* Company Links */}
          <div>
            <h4 className="text-white font-semibold mb-4">{dict.footer.company}</h4>
            <ul className="space-y-3">
              <li>
                <Link href="/about" className="hover:text-white transition-colors">
                  {dict.footer.about}
                </Link>
              </li>
              <li>
                <Link href="/contact" className="hover:text-white transition-colors">
                  {dict.footer.contact}
                </Link>
              </li>
              <li>
                <Link href="/blog" className="hover:text-white transition-colors">
                  {dict.footer.blog}
                </Link>
              </li>
            </ul>
          </div>

          {/* Legal Links */}
          <div>
            <h4 className="text-white font-semibold mb-4">{dict.footer.legal}</h4>
            <ul className="space-y-3">
              <li>
                <Link href="/privacy" className="hover:text-white transition-colors">
                  {dict.footer.privacy}
                </Link>
              </li>
              <li>
                <Link href="/terms" className="hover:text-white transition-colors">
                  {dict.footer.terms}
                </Link>
              </li>
            </ul>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="py-6 border-t border-gray-800 text-center text-sm text-gray-500">
          <p>&copy; {currentYear} ControleWise. {dict.footer.rights}</p>
        </div>
      </div>
    </footer>
  )
}
