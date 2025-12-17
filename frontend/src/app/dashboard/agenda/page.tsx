'use client'

import { useState } from 'react'
import { Plus } from 'lucide-react'
import { SessionCalendar } from '@/components/appointments/SessionCalendar'
import { SessionFormModal } from '@/components/appointments/SessionFormModal'
import { SessionDetailModal } from '@/components/appointments/SessionDetailModal'

export default function AgendaPage() {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [isEditModalOpen, setIsEditModalOpen] = useState(false)
  const [isDetailModalOpen, setIsDetailModalOpen] = useState(false)
  const [selectedDate, setSelectedDate] = useState<Date | null>(null)
  const [selectedSessionId, setSelectedSessionId] = useState<string | null>(null)

  const handleDateSelect = (date: Date) => {
    setSelectedDate(date)
    setIsCreateModalOpen(true)
  }

  const handleEventClick = (eventId: string) => {
    setSelectedSessionId(eventId)
    setIsDetailModalOpen(true)
  }

  const handleCloseCreateModal = () => {
    setIsCreateModalOpen(false)
    setSelectedDate(null)
  }

  const handleCloseEditModal = () => {
    setIsEditModalOpen(false)
    setSelectedSessionId(null)
  }

  const handleCloseDetailModal = () => {
    setIsDetailModalOpen(false)
    setSelectedSessionId(null)
  }

  const handleEditFromDetail = (sessionId: string) => {
    setIsDetailModalOpen(false)
    setSelectedSessionId(sessionId)
    setIsEditModalOpen(true)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Agenda</h1>
          <p className="mt-1 text-sm text-gray-500">
            Gerir sessões e agendamentos
          </p>
        </div>
        <button
          onClick={() => setIsCreateModalOpen(true)}
          className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
        >
          <Plus className="h-4 w-4 mr-2" />
          Nova Sessão
        </button>
      </div>

      <SessionCalendar
        onDateSelect={handleDateSelect}
        onEventClick={handleEventClick}
      />

      <SessionFormModal
        isOpen={isCreateModalOpen}
        onClose={handleCloseCreateModal}
        initialDate={selectedDate}
      />

      <SessionFormModal
        isOpen={isEditModalOpen}
        onClose={handleCloseEditModal}
        sessionId={selectedSessionId}
      />

      <SessionDetailModal
        isOpen={isDetailModalOpen}
        onClose={handleCloseDetailModal}
        sessionId={selectedSessionId}
        onEdit={handleEditFromDetail}
      />
    </div>
  )
}
