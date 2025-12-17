import { notFound } from 'next/navigation'
import { isValidLocale } from '@/lib/i18n/config'
import { getDictionary } from '@/lib/i18n/dictionaries'
import {
  Header,
  Hero,
  Features,
  Industries,
  WorkflowShowcase,
  CTA,
  Footer,
} from '@/components/landing'

interface LandingPageProps {
  params: Promise<{ locale: string }>
}

export default async function LandingPage({ params }: LandingPageProps) {
  const { locale } = await params

  if (!isValidLocale(locale)) {
    notFound()
  }

  const dict = await getDictionary(locale)

  return (
    <div className="min-h-screen bg-white">
      {/* Skip Link for Accessibility */}
      <a href="#main-content" className="skip-link">
        Skip to content
      </a>

      <Header dict={dict} locale={locale} />

      <main id="main-content">
        <Hero dict={dict} />
        <Features dict={dict} />
        <Industries dict={dict} />
        <WorkflowShowcase dict={dict} />
        <CTA dict={dict} />
      </main>

      <Footer dict={dict} locale={locale} />
    </div>
  )
}
