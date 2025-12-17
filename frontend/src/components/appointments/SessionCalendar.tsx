'use client'

import { useState, useMemo } from 'react'
import FullCalendar from '@fullcalendar/react'
import dayGridPlugin from '@fullcalendar/daygrid'
import timeGridPlugin from '@fullcalendar/timegrid'
import interactionPlugin from '@fullcalendar/interaction'
import { format, startOfMonth, endOfMonth, addMonths, subMonths } from 'date-fns'
import { pt } from 'date-fns/locale'
import { useCalendarEvents } from '@/hooks/useSessions'
import { useTherapists } from '@/hooks/useTherapists'
import { getCalendarEventColor } from '@/schemas/session'
import type { CalendarEvent } from '@/types'
import { ChevronLeft, ChevronRight, Loader2 } from 'lucide-react'

interface SessionCalendarProps {
  onDateSelect?: (date: Date) => void
  onEventClick?: (eventId: string) => void
}

export function SessionCalendar({ onDateSelect, onEventClick }: SessionCalendarProps) {
  const [currentDate, setCurrentDate] = useState(new Date())
  const [selectedTherapist, setSelectedTherapist] = useState<string>('')

  const { data: therapistsData } = useTherapists({ is_active: true })

  const dateRange = useMemo(() => ({
    from_date: format(startOfMonth(subMonths(currentDate, 1)), 'yyyy-MM-dd'),
    to_date: format(endOfMonth(addMonths(currentDate, 1)), 'yyyy-MM-dd'),
    therapist_id: selectedTherapist || undefined,
  }), [currentDate, selectedTherapist])

  const { data: events, isLoading } = useCalendarEvents(dateRange)

  const calendarEvents = useMemo(() => {
    if (!events) return []
    return events.map((event: CalendarEvent) => ({
      id: event.id,
      title: `${event.patient_name}`,
      start: event.start,
      end: event.end,
      backgroundColor: event.color || getCalendarEventColor(event.status),
      borderColor: event.color || getCalendarEventColor(event.status),
      extendedProps: {
        therapist_name: event.therapist_name,
        patient_name: event.patient_name,
        status: event.status,
      },
    }))
  }, [events])

  const handleDateClick = (arg: { date: Date }) => {
    if (onDateSelect) {
      onDateSelect(arg.date)
    }
  }

  const handleEventClick = (arg: { event: { id: string } }) => {
    if (onEventClick) {
      onEventClick(arg.event.id)
    }
  }

  const handleDatesSet = (arg: { start: Date }) => {
    setCurrentDate(arg.start)
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200">
      <div className="p-4 border-b border-gray-200">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <h2 className="text-lg font-semibold text-gray-900">Agenda</h2>
          <div className="flex items-center gap-4">
            <select
              value={selectedTherapist}
              onChange={(e) => setSelectedTherapist(e.target.value)}
              className="block w-full sm:w-48 rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
            >
              <option value="">Todos os terapeutas</option>
              {therapistsData?.therapists?.map((therapist) => (
                <option key={therapist.id} value={therapist.id}>
                  {therapist.name}
                </option>
              ))}
            </select>
            {isLoading && (
              <Loader2 className="h-5 w-5 animate-spin text-primary-600" />
            )}
          </div>
        </div>
      </div>

      <div className="p-4">
        <FullCalendar
          plugins={[dayGridPlugin, timeGridPlugin, interactionPlugin]}
          initialView="timeGridWeek"
          headerToolbar={{
            left: 'prev,next today',
            center: 'title',
            right: 'dayGridMonth,timeGridWeek,timeGridDay',
          }}
          locale="pt"
          firstDay={1}
          slotMinTime="08:00:00"
          slotMaxTime="20:00:00"
          allDaySlot={false}
          weekends={true}
          events={calendarEvents}
          dateClick={handleDateClick}
          eventClick={handleEventClick}
          datesSet={handleDatesSet}
          height="auto"
          slotDuration="00:30:00"
          slotLabelInterval="01:00"
          slotLabelFormat={{
            hour: '2-digit',
            minute: '2-digit',
            hour12: false,
          }}
          eventTimeFormat={{
            hour: '2-digit',
            minute: '2-digit',
            hour12: false,
          }}
          buttonText={{
            today: 'Hoje',
            month: 'Mês',
            week: 'Semana',
            day: 'Dia',
          }}
          eventContent={(eventInfo) => (
            <div className="p-1 overflow-hidden">
              <div className="font-medium text-xs truncate">
                {eventInfo.event.title}
              </div>
              <div className="text-xs opacity-75 truncate">
                {eventInfo.event.extendedProps.therapist_name}
              </div>
            </div>
          )}
        />
      </div>

      {/* Legend */}
      <div className="px-4 pb-4 flex flex-wrap gap-4">
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-yellow-500" />
          <span className="text-xs text-gray-600">Pendente</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-blue-500" />
          <span className="text-xs text-gray-600">Confirmada</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-green-500" />
          <span className="text-xs text-gray-600">Concluída</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-red-500" />
          <span className="text-xs text-gray-600">Cancelada</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-gray-500" />
          <span className="text-xs text-gray-600">Faltou</span>
        </div>
      </div>
    </div>
  )
}
