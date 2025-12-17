import { Settings, GitMerge, Zap, Activity } from 'lucide-react'
import type { Dictionary } from '@/lib/i18n/dictionaries'

interface WorkflowShowcaseProps {
  dict: Dictionary
}

export function WorkflowShowcase({ dict }: WorkflowShowcaseProps) {
  const steps = [
    { icon: Settings, text: dict.workflow.step1, color: 'bg-blue-500' },
    { icon: GitMerge, text: dict.workflow.step2, color: 'bg-purple-500' },
    { icon: Zap, text: dict.workflow.step3, color: 'bg-amber-500' },
    { icon: Activity, text: dict.workflow.step4, color: 'bg-emerald-500' },
  ]

  return (
    <section className="py-20 lg:py-28 bg-white overflow-hidden">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-6xl mx-auto">
          <div className="grid lg:grid-cols-2 gap-12 lg:gap-16 items-center">
            {/* Content */}
            <div>
              <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
                {dict.workflow.title}
              </h2>
              <p className="text-lg text-gray-600 mb-10">
                {dict.workflow.subtitle}
              </p>

              {/* Steps */}
              <div className="space-y-6">
                {steps.map((step, index) => {
                  const Icon = step.icon
                  return (
                    <div key={index} className="flex items-start gap-4">
                      <div className={`flex-shrink-0 w-10 h-10 ${step.color} rounded-lg flex items-center justify-center shadow-lg`}>
                        <Icon className="h-5 w-5 text-white" />
                      </div>
                      <div className="pt-2">
                        <p className="font-medium text-gray-900">{step.text}</p>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>

            {/* Visual */}
            <div className="relative">
              <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-2xl p-6 shadow-2xl">
                {/* Mock Workflow Editor */}
                <div className="flex items-center gap-2 mb-4">
                  <div className="w-3 h-3 rounded-full bg-red-400" />
                  <div className="w-3 h-3 rounded-full bg-amber-400" />
                  <div className="w-3 h-3 rounded-full bg-emerald-400" />
                </div>

                {/* Workflow Stages */}
                <div className="space-y-3">
                  {['Lead', 'Proposta', 'Negociação', 'Projeto', 'Conclusão'].map(
                    (stage, index) => (
                      <div
                        key={stage}
                        className="flex items-center gap-3 bg-gray-800/50 rounded-lg p-3"
                        style={{
                          transform: `translateX(${index * 8}px)`,
                          opacity: 1 - index * 0.1,
                        }}
                      >
                        <div
                          className={`w-3 h-3 rounded-full ${
                            index === 0
                              ? 'bg-emerald-400'
                              : index === 1
                              ? 'bg-blue-400'
                              : index === 2
                              ? 'bg-amber-400'
                              : index === 3
                              ? 'bg-purple-400'
                              : 'bg-gray-400'
                          }`}
                        />
                        <span className="text-gray-300 text-sm font-medium">
                          {stage}
                        </span>
                        <div className="flex-1" />
                        <span className="text-gray-500 text-xs">
                          Stage {index + 1}
                        </span>
                      </div>
                    )
                  )}
                </div>

                {/* Decorative Lines */}
                <div className="absolute -right-4 top-1/2 transform -translate-y-1/2 w-8 h-px bg-gradient-to-r from-primary-500 to-transparent" />
              </div>

              {/* Floating Elements */}
              <div className="absolute -top-4 -right-4 w-24 h-24 bg-primary-100 rounded-full opacity-60 blur-2xl" />
              <div className="absolute -bottom-4 -left-4 w-32 h-32 bg-purple-100 rounded-full opacity-60 blur-2xl" />
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
