import Link from 'next/link'
import { ArrowRight, Sparkles } from 'lucide-react'
import type { Dictionary } from '@/lib/i18n/dictionaries'

interface CTAProps {
  dict: Dictionary
}

export function CTA({ dict }: CTAProps) {
  return (
    <section className="py-20 lg:py-28 bg-gradient-to-br from-primary-600 via-primary-700 to-primary-800 relative overflow-hidden">
      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-10">
        <div className="absolute top-0 left-0 w-96 h-96 bg-white rounded-full -translate-x-1/2 -translate-y-1/2" />
        <div className="absolute bottom-0 right-0 w-96 h-96 bg-white rounded-full translate-x-1/2 translate-y-1/2" />
      </div>

      <div className="container mx-auto px-4 sm:px-6 lg:px-8 relative">
        <div className="max-w-3xl mx-auto text-center">
          {/* Icon */}
          <div className="inline-flex items-center justify-center w-16 h-16 bg-white/10 rounded-2xl mb-8">
            <Sparkles className="h-8 w-8 text-white" />
          </div>

          {/* Headline */}
          <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-white mb-6">
            {dict.cta.title}
          </h2>

          {/* Subtitle */}
          <p className="text-lg sm:text-xl text-primary-100 mb-10">
            {dict.cta.subtitle}
          </p>

          {/* CTA Button */}
          <Link
            href="/register"
            className="group inline-flex items-center bg-white text-primary-700 px-8 py-4 rounded-xl font-semibold hover:bg-primary-50 transition-all shadow-lg hover:shadow-xl"
          >
            {dict.cta.button}
            <ArrowRight className="ml-2 h-5 w-5 group-hover:translate-x-1 transition-transform" />
          </Link>
        </div>
      </div>
    </section>
  )
}
