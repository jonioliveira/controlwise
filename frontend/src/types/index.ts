export type Role = 'owner' | 'admin' | 'manager' | 'employee' | 'client' | 'accountant'

export type WorkSheetStatus = 'draft' | 'under_review' | 'approved'
export type BudgetStatus = 'draft' | 'sent' | 'approved' | 'rejected' | 'expired'
export type ProjectStatus = 'in_progress' | 'on_hold' | 'completed' | 'cancelled'
export type TaskStatus = 'todo' | 'in_progress' | 'completed' | 'cancelled'
export type PaymentStatus = 'pending' | 'paid' | 'overdue' | 'cancelled'
export type Priority = 'low' | 'medium' | 'high' | 'urgent'

export interface Organization {
  id: string
  name: string
  email: string
  phone: string
  address: string
  tax_id: string
  logo?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface User {
  id: string
  organization_id: string
  email: string
  first_name: string
  last_name: string
  phone?: string
  avatar?: string
  role: Role
  is_active: boolean
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface Client {
  id: string
  organization_id: string
  name: string
  email: string
  phone: string
  address?: string
  tax_id?: string
  notes?: string
  user_id?: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface WorkSheet {
  id: string
  organization_id: string
  client_id: string
  title: string
  description: string
  status: WorkSheetStatus
  created_by: string
  reviewed_by?: string
  reviewed_at?: string
  created_at: string
  updated_at: string
}

export interface WorkSheetItem {
  id: string
  worksheet_id: string
  description: string
  quantity: number
  unit: string
  notes?: string
  order: number
  created_at: string
  updated_at: string
}

export interface Budget {
  id: string
  organization_id: string
  worksheet_id: string
  budget_number: string
  status: BudgetStatus
  subtotal: number
  tax: number
  total: number
  valid_until: string
  notes?: string
  created_by: string
  sent_at?: string
  approved_by?: string
  approved_at?: string
  rejected_at?: string
  rejection_notes?: string
  created_at: string
  updated_at: string
}

export interface BudgetItem {
  id: string
  budget_id: string
  worksheet_item_id?: string
  description: string
  quantity: number
  unit: string
  unit_price: number
  tax: number
  total: number
  order: number
  created_at: string
  updated_at: string
}

export interface Project {
  id: string
  organization_id: string
  budget_id: string
  project_number: string
  title: string
  description?: string
  status: ProjectStatus
  progress: number
  start_date: string
  expected_end_date: string
  actual_end_date?: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface Task {
  id: string
  project_id: string
  title: string
  description?: string
  assigned_to?: string
  status: TaskStatus
  priority: Priority
  due_date?: string
  completed_at?: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface Payment {
  id: string
  organization_id: string
  project_id: string
  amount: number
  status: PaymentStatus
  due_date: string
  paid_at?: string
  method?: string
  reference?: string
  notes?: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface Photo {
  id: string
  organization_id: string
  entity_type: 'worksheet' | 'budget' | 'project' | 'task'
  entity_id: string
  file_name: string
  file_size: number
  mime_type: string
  url: string
  thumbnail_url?: string
  caption?: string
  uploaded_by: string
  created_at: string
}

export interface Notification {
  id: string
  user_id: string
  type: string
  title: string
  message: string
  entity_type?: string
  entity_id?: string
  is_read: boolean
  read_at?: string
  created_at: string
}

export interface ApiResponse<T> {
  data: T
  message?: string
}

export interface ApiError {
  error: string
  message: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  organization_name: string
  email: string
  password: string
  first_name: string
  last_name: string
  phone: string
}

export interface AuthResponse {
  token: string
  user: User
}

// ============ Module Types ============

export type ModuleName = 'construction' | 'appointments' | 'notifications'

export interface AvailableModule {
  name: ModuleName
  display_name: string
  description?: string
  icon?: string
  dependencies: string[]
  is_active: boolean
  created_at: string
}

export interface OrganizationModule {
  id: string
  organization_id: string
  module_name: ModuleName
  is_enabled: boolean
  config: Record<string, unknown>
  enabled_at?: string
  enabled_by?: string
  display_name: string
  description?: string
  icon?: string
  dependencies: string[]
  created_at: string
  updated_at: string
}

// ============ Appointments Types ============

export type SessionStatus = 'pending' | 'confirmed' | 'cancelled' | 'completed' | 'no_show'
export type SessionType = 'regular' | 'evaluation' | 'follow_up'

// Patient is a healthcare extension of Client - name/email/phone come from linked client
export interface Patient {
  id: string
  organization_id: string
  client_id: string // Required - link to clients table
  date_of_birth?: string
  notes?: string // Medical notes
  emergency_contact?: string
  emergency_phone?: string
  is_active: boolean
  created_by?: string
  created_at: string
  updated_at: string
  // Joined fields from client
  client_name: string
  client_email: string
  client_phone: string
}

export interface WorkingHours {
  start: string
  end: string
}

export interface Therapist {
  id: string
  organization_id: string
  user_id?: string
  name: string
  email?: string
  phone?: string
  specialty?: string
  working_hours: Record<string, WorkingHours>
  session_duration_minutes: number
  default_price_cents: number
  timezone: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Session {
  id: string
  organization_id: string
  therapist_id: string
  patient_id: string
  scheduled_at: string
  duration_minutes: number
  price_cents: number
  status: SessionStatus
  session_type: SessionType
  notes?: string
  cancel_reason?: string
  cancelled_at?: string
  cancelled_by?: string
  completed_at?: string
  created_by?: string
  created_at: string
  updated_at: string
}

export interface SessionWithDetails extends Session {
  therapist_name: string
  patient_name: string
  patient_phone: string
  patient_email?: string
}

export interface CalendarEvent {
  id: string
  title: string
  start: string
  end: string
  status: SessionStatus
  therapist_id: string
  therapist_name: string
  patient_id: string
  patient_name: string
  color?: string
}

// ============ Notification Config Types ============

export interface NotificationConfig {
  id: string
  organization_id: string
  whatsapp_enabled: boolean
  twilio_configured: boolean
  twilio_whatsapp_number?: string
  reminder_24h_enabled: boolean
  reminder_2h_enabled: boolean
  reminder_24h_template?: string
  reminder_2h_template?: string
  confirmation_response_template?: string
  created_at: string
  updated_at: string
}

// ============ System Admin Types ============

export interface SystemAdmin {
  id: string
  email: string
  first_name: string
  last_name: string
  is_active: boolean
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface SystemAdminAuthResponse {
  token: string
  admin: SystemAdmin
}

export interface PlatformStats {
  total_organizations: number
  active_organizations: number
  suspended_organizations: number
  total_users: number
  active_users: number
  new_orgs_this_month: number
  new_users_this_month: number
  orgs_by_module: Record<string, number>
}

export interface OrganizationWithStats extends Organization {
  user_count: number
  active_user_count: number
  enabled_modules: string[]
  suspended_at?: string
  suspended_by?: string
  suspend_reason?: string
}

export interface UserWithOrg extends User {
  org_name: string
  suspended_at?: string
  suspended_by?: string
  suspend_reason?: string
}

export interface AuditLogEntry {
  id: string
  admin_id: string
  action: string
  entity_type: string
  entity_id?: string
  details: Record<string, unknown>
  ip_address: string
  user_agent: string
  created_at: string
}

export interface ImpersonationSession {
  id: string
  admin_id: string
  impersonated_user_id: string
  started_at: string
  ended_at?: string
  reason: string
  ip_address: string
}

export interface ImpersonationSessionWithDetails extends ImpersonationSession {
  user_email: string
  user_name: string
  org_name: string
  admin_name: string
}

export interface ImpersonationToken {
  token: string
  session_id: string
  user: User
  expires_at: string
}

export interface RecentActivity {
  type: string
  description: string
  created_at: string
  entity_id?: string
  entity_type?: string
}

// Admin request types
export interface AdminLoginRequest {
  email: string
  password: string
}

export interface AdminChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface AdminCreateOrganizationRequest {
  name: string
  email: string
  phone?: string
  address?: string
  tax_id?: string
  admin_email: string
  admin_password: string
  admin_first_name: string
  admin_last_name: string
  admin_phone?: string
}

export interface AdminUpdateOrganizationRequest {
  name?: string
  email?: string
  phone?: string
  address?: string
  tax_id?: string
}

export interface AdminSuspendRequest {
  reason: string
}

export interface AdminStartImpersonationRequest {
  reason: string
}

// Paginated response type
export interface PaginatedResponse<T> {
  data: T[]
  pagination: {
    page: number
    limit: number
    total: number
    total_pages: number
  }
}

// ============ Workflow Types ============

export type WorkflowModule = 'appointments' | 'construction'
export type WorkflowEntityType = 'session' | 'budget' | 'project'
export type StateType = 'initial' | 'intermediate' | 'final'
export type TriggerType = 'on_enter' | 'on_exit' | 'time_before' | 'time_after' | 'recurring'
export type ActionType = 'send_whatsapp' | 'send_email' | 'update_field' | 'create_task'
export type MessageChannel = 'whatsapp' | 'email'
export type SessionPaymentStatus = 'unpaid' | 'partial' | 'paid'
export type SessionPaymentMethod = 'cash' | 'transfer' | 'insurance' | 'card'

export interface Workflow {
  id: string
  organization_id: string
  name: string
  description?: string
  module: WorkflowModule
  entity_type: WorkflowEntityType
  is_active: boolean
  is_default: boolean
  created_at: string
  updated_at: string
  states?: WorkflowState[]
  transitions?: WorkflowTransition[]
  triggers?: WorkflowTrigger[]
}

export interface WorkflowWithStats extends Workflow {
  state_count: number
  trigger_count: number
  action_count: number
}

export interface WorkflowState {
  id: string
  workflow_id: string
  name: string
  display_name: string
  description?: string
  state_type: StateType
  color?: string
  position: number
  created_at: string
  triggers?: WorkflowTrigger[]
}

export interface WorkflowTransition {
  id: string
  workflow_id: string
  from_state_id: string
  to_state_id: string
  name: string
  requires_confirmation: boolean
  created_at: string
  from_state_name?: string
  to_state_name?: string
  triggers?: WorkflowTrigger[]
}

export interface WorkflowTrigger {
  id: string
  workflow_id: string
  state_id?: string
  transition_id?: string
  trigger_type: TriggerType
  time_offset_minutes?: number
  time_field?: string
  recurring_cron?: string
  conditions?: Record<string, unknown>
  is_active: boolean
  created_at: string
  actions?: WorkflowAction[]
}

export interface WorkflowAction {
  id: string
  trigger_id: string
  action_type: ActionType
  action_order: number
  template_id?: string
  action_config?: Record<string, unknown>
  is_active: boolean
  created_at: string
  template?: MessageTemplate
}

export interface MessageTemplate {
  id: string
  organization_id: string
  name: string
  channel: MessageChannel
  subject?: string
  body: string
  variables?: TemplateVariable[]
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface TemplateVariable {
  name: string
  description: string
}

export interface SessionPayment {
  id: string
  session_id: string
  amount_cents: number
  payment_status: SessionPaymentStatus
  payment_method?: SessionPaymentMethod
  insurance_provider?: string
  insurance_amount_cents?: number
  due_date?: string
  paid_at?: string
  notes?: string
  created_at: string
  updated_at: string
}

export interface SessionPaymentWithDetails extends SessionPayment {
  patient_name: string
  therapist_name: string
  scheduled_at: string
}

export interface SessionPaymentStats {
  total_sessions: number
  paid_count: number
  unpaid_count: number
  partial_count: number
  total_paid_cents: number
  total_unpaid_cents: number
}

export interface UpdateSessionPaymentRequest {
  amount_cents: number
  payment_status: SessionPaymentStatus
  payment_method?: SessionPaymentMethod
  insurance_provider?: string
  insurance_amount_cents?: number
  due_date?: string
  notes?: string
}

export interface MarkAsPaidRequest {
  payment_method?: SessionPaymentMethod
}

// Workflow Execution Types
export type WorkflowEventType = 'state_change' | 'trigger_fired' | 'action_executed' | 'action_failed'
export type ScheduledJobStatus = 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled'

export interface WorkflowExecutionLog {
  id: string
  organization_id: string
  workflow_id: string
  entity_type: string
  entity_id: string
  trigger_id?: string
  action_id?: string
  event_type: WorkflowEventType
  from_state?: string
  to_state?: string
  details?: Record<string, unknown>
  created_at: string
}

export interface ScheduledJob {
  id: string
  organization_id: string
  trigger_id: string
  entity_type: string
  entity_id: string
  scheduled_for: string
  status: ScheduledJobStatus
  attempts: number
  last_error?: string
  created_at: string
  processed_at?: string
}
