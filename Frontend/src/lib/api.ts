// Fetch wrapper for /api/* — handles JSON, errors, and session cookies
import type { ApiError, AuthSession } from './types';

export class ApiException extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
    this.name = 'ApiException';
  }
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  isFormData = false
): Promise<T> {
  const headers: Record<string, string> = {};
  if (body && !isFormData) {
    headers['Content-Type'] = 'application/json';
  }

  const res = await fetch(`/api${path}`, {
    method,
    credentials: 'include', // send session cookie
    headers,
    body: body
      ? isFormData
        ? (body as FormData)
        : JSON.stringify(body)
      : undefined,
  });

  if (res.status === 204) return undefined as T;

  let data: T | ApiError;
  try {
    data = await res.json();
  } catch {
    throw new ApiException(`HTTP ${res.status}`, res.status);
  }

  if (!res.ok) {
    const err = data as ApiError;
    throw new ApiException(err.error || err.message || `HTTP ${res.status}`, res.status);
  }

  // Unwrap standardized backend response { success: true, data: ... }
  if (data && typeof data === 'object' && 'success' in data && data.success === true && 'data' in data) {
    return flattenNullables((data as any).data) as T;
  }

  return flattenNullables(data) as T;
}

/**
 * Recursively flattens objects that look like Go's sql.NullString, sql.NullInt64, etc.
 * { String: "val", Valid: true } -> "val"
 * { String: "", Valid: false } -> null
 */
function flattenNullables(obj: any): any {
  if (obj === null || typeof obj !== 'object') return obj;
  
  if (Array.isArray(obj)) {
    return obj.map(flattenNullables);
  }

  const keys = Object.keys(obj);
  // Pattern: { [Type]: Value, Valid: boolean }
  if (keys.length === 2 && 'Valid' in obj && typeof obj.Valid === 'boolean') {
    const valueKey = keys.find(k => k !== 'Valid');
    if (valueKey) {
      return obj.Valid ? obj[valueKey] : null;
    }
  }

  const flattened: any = {};
  for (const [key, value] of Object.entries(obj)) {
    flattened[key] = flattenNullables(value);
  }
  return flattened;
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
  delete: <T>(path: string) => request<T>('DELETE', path),
  postForm: <T>(path: string, form: FormData) => request<T>('POST', path, form, true),
};

// Auth
export const authApi = {
  login: (username: string, password: string) =>
    api.post<{ message: string }>('/auth/login', { username, password }),
  logout: () => api.post<void>('/auth/logout'),
  me: () => api.get<AuthSession>('/auth/me'),
};

// Students
export const studentsApi = {
  list: (group?: string) => api.get<any[]>(`/students${group ? `?group=${encodeURIComponent(group)}` : ''}`),
  get: (id: number) => api.get<any>(`/students/${id}`),
  create: (body: unknown) => api.post<any>('/students', body),
  update: (id: number, body: unknown) => api.put<any>(`/students/${id}`, body),
  delete: (id: number) => api.delete<void>(`/students/${id}`),
};

// Assignments
export const assignmentsApi = {
  list: () => api.get<any[]>('/assignments'),
  get: (id: number) => api.get<any>(`/assignments/${id}`),
  create: (body: unknown) => api.post<any>('/assignments', body),
  update: (id: number, body: unknown) => api.put<any>(`/assignments/${id}`, body),
  publish: (id: number) => api.put<any>(`/assignments/${id}/publish`),
  close: (id: number) => api.put<any>(`/assignments/${id}/close`),
  delete: (id: number) => api.delete<void>(`/assignments/${id}`),
};

// Sessions
export const sessionsApi = {
  list: () => api.get<any[]>('/sessions'),
  get: (id: number) => api.get<any>(`/sessions/${id}`),
  create: (body: unknown) => api.post<any>('/sessions', body),
  update: (id: number, body: unknown) => api.put<any>(`/sessions/${id}`, body),
  open: (id: number) => api.put<any>(`/sessions/${id}/open`),
  close: (id: number) => api.put<any>(`/sessions/${id}/close`),
  cancel: (id: number) => api.put<any>(`/sessions/${id}/cancel`),
};

// Attendance
export const attendanceApi = {
  list: (sessionId: number) => api.get<any[]>(`/sessions/${sessionId}/attendance`),
  checkIn: (sessionId: number, body: unknown) => api.post<any>(`/sessions/${sessionId}/attendance`, body),
  update: (attendanceId: number, body: unknown) => api.put<any>(`/attendance/${attendanceId}`, body),
};

// Inventory
export const inventoryApi = {
  listItems: (params?: Record<string, string>) => {
    const qs = params ? '?' + new URLSearchParams(params).toString() : '';
    return api.get<any[]>(`/items${qs}`);
  },
  getItem: (id: number) => api.get<any>(`/items/${id}`),
  createItem: (body: unknown) => api.post<any>('/items', body),
  updateItem: (id: number, body: unknown) => api.put<any>(`/items/${id}`, body),
  deleteItem: (id: number) => api.delete<void>(`/items/${id}`),
  adjustStock: (id: number, body: unknown) => api.post<any>(`/items/${id}/adjust-stock`, body),
  listCategories: () => api.get<any[]>('/categories'),
  createCategory: (body: unknown) => api.post<any>('/categories', body),
};

// Resources
export const requestsApi = {
  list: () => api.get<any[]>('/requests'),
  create: (body: unknown) => api.post<any>('/requests', body),
  approve: (id: number) => api.put<any>(`/requests/${id}/approve`),
  reject: (id: number) => api.put<any>(`/requests/${id}/reject`),
  return: (id: number) => api.put<any>(`/requests/${id}/return`),
};

// Equipment usage
export const equipmentApi = {
  list: () => api.get<any[]>('/equipment-usage'),
  create: (body: unknown) => api.post<any>('/equipment-usage', body),
  end: (id: number) => api.put<any>(`/equipment-usage/${id}/end`),
  checkAvailable: (itemId: number) => api.get<any>(`/items/${itemId}/available`),
};

// Manuals
export const manualsApi = {
  list: (params?: Record<string, string>) => {
    const qs = params ? '?' + new URLSearchParams(params).toString() : '';
    return api.get<any[]>(`/manuals${qs}`);
  },
  upload: (form: FormData) => api.postForm<any>('/manuals', form),
  download: (id: number) => `/api/manuals/${id}/download`,
  delete: (id: number) => api.delete<void>(`/manuals/${id}`),
};

// Maintenance
export const maintenanceApi = {
  list: (itemId: number) => api.get<any[]>(`/items/${itemId}/maintenance`),
  create: (body: unknown) => api.post<any>('/maintenance', body),
};

// Incidents
export const incidentsApi = {
  list: () => api.get<any[]>('/incidents'),
  create: (body: unknown) => api.post<any>('/incidents', body),
  update: (id: number, body: unknown) => api.put<any>(`/incidents/${id}`, body),
};

// Reports
export const reportsApi = {
  sessionCsv: (sessionId: number) => `/api/sessions/${sessionId}/report/csv`,
};
