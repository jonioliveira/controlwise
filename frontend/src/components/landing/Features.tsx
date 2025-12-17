import {
  Users,
  GitBranch,
  FileText,
  FolderKanban,
  CreditCard,
  Building2,
} from 'lucide-react'
import type { Dictionary } from '@/lib/i18n/dictionaries'

interface FeaturesProps {
  dict: Dictionary
}

const featureIcons = {
  clients: Users,
  workflows: GitBranch,
  budgets: FileText,
  projects: FolderKanban,
  payments: CreditCard,
  multiTenant: Building2,
}

export function Features({ dict }: FeaturesProps) {
  const features = [
    { key: 'clients', ...dict.features.clients },
    { key: 'workflows', ...dict.features.workflows },
    { key: 'budgets', ...dict.features.budgets },
    { key: 'projects', ...dict.features.projects },
    { key: 'payments', ...dict.features.payments },
    { key: 'multiTenant', ...dict.features.multiTenant },
  ] as const

  return (
    <section className="py-20 lg:py-28 bg-white">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        {/* Section Header */}
        <div className="max-w-3xl mx-auto text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
            {dict.features.title}
          </h2>
          <p className="text-lg text-gray-600">
            {dict.features.subtitle}
          </p>
        </div>

        {/* Features Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 lg:gap-8">
          {features.map((feature) => {
            const Icon = featureIcons[feature.key]
            return (
              <div
                key={feature.key}
                className="group p-6 lg:p-8 bg-gray-50 rounded-2xl hover:bg-white hover:shadow-xl hover:shadow-gray-100 transition-all duration-300 border border-transparent hover:border-gray-100"
              >
                <div className="w-12 h-12 bg-primary-100 rounded-xl flex items-center justify-center mb-5 group-hover:bg-primary-600 transition-colors">
                  <Icon className="h-6 w-6 text-primary-600 group-hover:text-white transition-colors" />
                </div>
                <h3 className="text-xl font-semibold text-gray-900 mb-3">
                  {feature.title}
                </h3>
                <p className="text-gray-600 leading-relaxed">
                  {feature.description}
                </p>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
