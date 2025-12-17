import Link from 'next/link'
import { ArrowRight, Play } from 'lucide-react'
import type { Dictionary } from '@/lib/i18n/dictionaries'

interface HeroProps {
  dict: Dictionary
}

export function Hero({ dict }: HeroProps) {
  return (
    <section className="pt-32 pb-20 lg:pt-40 lg:pb-32 bg-gradient-to-b from-white via-primary-50/30 to-white">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-4xl mx-auto text-center">
          {/* Badge */}
          <div className="inline-flex items-center px-4 py-1.5 bg-primary-100 text-primary-700 rounded-full text-sm font-medium mb-8">
            <span className="w-2 h-2 bg-primary-500 rounded-full mr-2 animate-pulse" />
            Multi-industry platform
          </div>

          {/* Headline */}
          <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold text-gray-900 leading-tight mb-6">
            {dict.hero.title}
            <span className="text-primary-600 block sm:inline"> {dict.hero.titleHighlight}</span>
          </h1>

          {/* Subtitle */}
          <p className="text-lg sm:text-xl text-gray-600 max-w-2xl mx-auto mb-10 leading-relaxed">
            {dict.hero.subtitle}
          </p>

          {/* CTAs */}
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link
              href="/register"
              className="group w-full sm:w-auto bg-primary-600 text-white px-8 py-4 rounded-xl font-semibold hover:bg-primary-700 transition-all shadow-lg shadow-primary-600/25 hover:shadow-xl hover:shadow-primary-600/30 flex items-center justify-center"
            >
              {dict.hero.cta}
              <ArrowRight className="ml-2 h-5 w-5 group-hover:translate-x-1 transition-transform" />
            </Link>
            <Link
              href="/demo"
              className="group w-full sm:w-auto bg-white text-gray-700 px-8 py-4 rounded-xl font-semibold hover:bg-gray-50 transition-all border border-gray-200 flex items-center justify-center"
            >
              <Play className="mr-2 h-5 w-5 text-primary-600" />
              {dict.hero.demo}
            </Link>
          </div>

          {/* Trust Indicators */}
          <div className="mt-16 pt-8 border-t border-gray-100">
            <p className="text-sm text-gray-500 mb-4">
              {dict.cta.subtitle}
            </p>
            <div className="flex justify-center items-center gap-8 opacity-60">
              <div className="text-2xl font-bold text-gray-400">100+</div>
              <div className="w-px h-6 bg-gray-200" />
              <div className="text-2xl font-bold text-gray-400">500+</div>
              <div className="w-px h-6 bg-gray-200" />
              <div className="text-2xl font-bold text-gray-400">10K+</div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
