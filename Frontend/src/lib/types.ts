// TypeScript interfaces matching Go models (core/models/models.go)

export interface User {
  id: number;
  username: string;
  full_name: string;
  email: string;
  role: 'admin' | 'teacher' | 'operator' | 'student';
  is_active: boolean;
  created_at: string;
}

export interface AuthSession {
  user_id: number;
  username: string;
  full_name: string;
  role: 'admin' | 'teacher' | 'operator' | 'student';
  student_id?: number;
  group_name?: string;
}

export interface Student {
  id: number;
  student_code: string;
  full_name: string;
  email: string;
  group_name: string;
  semester: number;
  is_active: boolean;
  created_at: string;
}

export interface StudentAccount {
  id: number;
  student_id: number;
  username: string;
  is_active: boolean;
}

export interface Category {
  id: number;
  name: string;
  description: string;
  created_at: string;
}

export type ItemType = 'consumable' | 'tool' | 'machine';
export type MaintenanceStatus = 'ok' | 'scheduled' | 'critical' | 'out_of_service';

export interface Item {
  id: number;
  sku: string;
  name: string;
  description: string;
  item_type: ItemType;
  type?: ItemType; // alias for backward compat
  category_id: number;
  category_name?: string;
  stock: number;
  min_stock: number;
  unit: string;
  location: string;
  maintenance_status: MaintenanceStatus;
  module_data?: string;
  is_active: boolean;
  created_at: string;
  updated_at?: string;
}

export type AssignmentStatus = 'draft' | 'published' | 'closed';

export interface Assignment {
  id: number;
  title: string;
  description: string;
  instructions: string;
  teacher_id: number;
  teacher_name?: string;
  group_name: string;
  semester: number;
  status: AssignmentStatus;
  published_at?: string;
  created_at: string;
}

export interface AssignmentItem {
  id: number;
  assignment_id: number;
  item_id: number;
  item_name?: string;
  quantity: number;
  notes: string;
}

export type SessionStatus = 'planned' | 'open' | 'closed' | 'cancelled';

export interface LabSession {
  id: number;
  assignment_id: number;
  assignment_title?: string;
  title: string;
  group_name: string;
  scheduled_start: string;
  scheduled_end: string;
  opened_at?: string;
  closed_at?: string;
  status: SessionStatus;
  notes: string;
  teacher_id: number;
  teacher_name?: string;
}

export type AttendanceStatus = 'present' | 'absent' | 'late' | 'excused';

export interface Attendance {
  id: number;
  session_id: number;
  student_id: number;
  student_code?: string;
  student_name?: string;
  status: AttendanceStatus;
  check_in_at?: string;
  notes: string;
}

export type RequestStatus = 'pending' | 'approved' | 'rejected' | 'returned';
export type RequestType = 'loan' | 'consumption';

export interface ResourceRequest {
  id: number;
  session_id: number;
  assignment_id?: number;
  student_id: number;
  student_name?: string;
  student_code?: string;
  item_id: number;
  item_name?: string;
  item_sku?: string;
  quantity: number;
  request_type: RequestType;
  type?: RequestType; // alias
  status: RequestStatus;
  notes: string;
  requested_at: string;
  resolved_at?: string;
  resolved_by?: number;
}

export type UsageStatus = 'active' | 'completed' | 'interrupted';

export interface EquipmentUsage {
  id: number;
  session_id: number;
  student_id: number;
  student_name?: string;
  item_id: number;
  item_name?: string;
  supervisor_id: number;
  supervisor_name?: string;
  started_at: string;
  ended_at?: string;
  status: UsageStatus;
  notes: string;
}

export interface Manual {
  id: number;
  title: string;
  description: string;
  item_id?: number;
  item_name?: string;
  assignment_id?: number;
  file_path: string;
  uploaded_by: number;
  uploaded_by_name?: string;
  is_active: boolean;
  created_at: string;
}

export type MaintenanceType = 'preventive' | 'corrective' | 'inspection';

export interface MaintenanceLog {
  id: number;
  item_id: number;
  item_name?: string;
  type: MaintenanceType;
  description: string;
  performed_by: number;
  performed_by_name?: string;
  scheduled_at?: string;
  completed_at?: string;
  maintenance_status: MaintenanceStatus;
  created_at: string;
}

export type IncidentSeverity = 'low' | 'medium' | 'high' | 'critical';
export type IncidentStatus = 'open' | 'in_review' | 'resolved' | 'dismissed';

export interface IncidentReport {
  id: number;
  session_id: number;
  item_id?: number;
  item_name?: string;
  reported_by: number;
  reporter_name?: string;
  severity: IncidentSeverity;
  status: IncidentStatus;
  description: string;
  related_previous_student_id?: number;
  related_previous_student_name?: string;
  created_at: string;
  resolved_at?: string;
}

export interface ApiError {
  error: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
}
