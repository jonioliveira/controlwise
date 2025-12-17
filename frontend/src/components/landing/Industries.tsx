'use client'

import { useState } from 'react'
import { HardHat, Stethoscope, Briefcase, ArrowRight } from 'lucide-react'
import type { Dictionary } from '@/lib/i18n/dictionaries'

interface IndustriesProps {
  dict: Dictionary
}

const industryConfig = {
  construction: {
    icon: HardHat,
    gradient: 'from-amber-500 to-orange-600',
    bgGradient: 'from-amber-50 to-orange-50',
  },
  healthcare: {
    icon: Stethoscope,
    gradient: 'from-emerald-500 to-teal-600',
    bgGradient: 'from-emerald-50 to-teal-50',
  },
  freelancers: {
    icon: Briefcase,
    gradient: 'from-violet-500 to-purple-600',
    bgGradient: 'from-violet-50 to-purple-50',
  },
}

export function Industries({ dict }: IndustriesProps) {
  const [activeTab, setActiveTab] = useState<keyof typeof industryConfig>('construction')

  const industries = [
    { key: 'construction' as const, ...dict.industries.construction },
    { key: 'healthcare' as const, ...dict.industries.healthcare },
    { key: 'freelancers' as const, ...dict.industries.freelancers },
  ]

  const activeIndustry = industries.find((i) => i.key === activeTab)!
  const config = industryConfig[activeTab]
  const Icon = config.icon

  return (
    <section className="py-20 lg:py-28 bg-gray-50">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        {/* Section Header */}
        <div className="max-w-3xl mx-auto text-center mb-12">
          <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
            {dict.industries.title}
          </h2>
          <p className="text-lg text-gray-600">
            {dict.industries.subtitle}
          </p>
        </div>

        {/* Tab Buttons */}
        <div className="flex flex-wrap justify-center gap-3 mb-12">
          {industries.map((industry) => {
            const TabIcon = industryConfig[industry.key].icon
            const isActive = activeTab === industry.key
            return (
              <button
                key={industry.key}
                onClick={() => setActiveTab(industry.key)}
                className={`flex items-center gap-2 px-5 py-3 rounded-xl font-medium transition-all ${
                  isActive
                    ? 'bg-white shadow-lg text-gray-900'
                    : 'bg-white/50 text-gray-600 hover:bg-white hover:shadow'
                }`}
              >
                <TabIcon className={`h-5 w-5 ${isActive ? 'text-primary-600' : ''}`} />
                {industry.name}
              </button>
            )
          })}
        </div>

        {/* Content Card */}
        <div className={`max-w-4xl mx-auto bg-gradient-to-br ${config.bgGradient} rounded-3xl p-8 lg:p-12`}>
          <div className="flex flex-col lg:flex-row lg:items-center gap-8">
            {/* Icon */}
            <div className={`flex-shrink-0 w-20 h-20 bg-gradient-to-br ${config.gradient} rounded-2xl flex items-center justify-center shadow-lg`}>
              <Icon className="h-10 w-10 text-white" />
            </div>

            {/* Content */}
            <div className="flex-1">
              <h3 className="text-2xl font-bold text-gray-900 mb-3">
                {activeIndustry.name}
              </h3>
              <p className="text-gray-700 mb-6 leading-relaxed">
                {activeIndustry.description}
              </p>

              {/* Workflow Preview */}
              <div className="bg-white/70 backdrop-blur-sm rounded-xl p-4">
                <p className="text-sm text-gray-500 mb-2">Workflow exemplo:</p>
                <div className="flex flex-wrap items-center gap-2 text-sm font-medium text-gray-700">
                  {activeIndustry.workflow.split(' → ').map((step, index, arr) => (
                    <span key={step} className="flex items-center">
                      <span className="bg-white px-3 py-1.5 rounded-lg shadow-sm">
                        {step}
                      </span>
                      {index < arr.length - 1 && (
                        <ArrowRight className="h-4 w-4 mx-1 text-gray-400" />
                      )}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
