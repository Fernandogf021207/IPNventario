<script lang="ts">
  import { onMount } from 'svelte';
  import { attendanceApi, sessionsApi } from '../../lib/api';
  import type { LabSession, Attendance } from '../../lib/types';

  let sessions: LabSession[] = [];
  let selectedSessionId: number | null = null;
  let attendances: Attendance[] = [];
  
  let loadingSessions = true;
  let loadingParams = false;
  let error = '';

  async function loadSessions() {
    loadingSessions = true;
    try {
      // Sort by scheduled_at desc
      sessions = await sessionsApi.list();
      sessions.sort((a, b) => new Date(b.scheduled_at).getTime() - new Date(a.scheduled_at).getTime());
    } catch (err: any) {
      error = err.message || 'Error al cargar sesiones';
    } finally {
      loadingSessions = false;
    }
  }

  onMount(() => {
    loadSessions();
  });

  async function loadAttendance(sessionId: number) {
    loadingParams = true;
    error = '';
    try {
      attendances = await attendanceApi.list(sessionId);
    } catch (err: any) {
      error = err.message || 'Error al cargar lista de asistencia';
      attendances = [];
    } finally {
      loadingParams = false;
    }
  }

  function handleSessionChange() {
    if (selectedSessionId) {
      loadAttendance(selectedSessionId);
    } else {
      attendances = [];
    }
  }

  async function updateStatus(attendanceId: number, newStatus: string) {
    try {
      await attendanceApi.update(attendanceId, { status: newStatus });
      // update local
      const idx = attendances.findIndex(a => a.id === attendanceId);
      if (idx !== -1) {
        attendances[idx].status = newStatus as any;
      }
    } catch (err: any) {
      alert(err.message || 'Error al actualizar asistencia');
    }
  }

  function formatDate(isoString: string) {
    if (!isoString) return '--';
    const date = new Date(isoString);
    return date.toLocaleString('es-MX', {
      day: '2-digit', month: '2-digit', year: 'numeric',
      hour: '2-digit', minute: '2-digit'
    });
  }
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Asistencia</h1>
      <p class="page-subtitle">Pase de lista y reportes por sesión</p>
    </div>
  </div>

  {#if error}
    <div class="error-alert">{error}</div>
  {/if}

  <div class="card selector-card">
    <label for="session-select" style="font-weight: 500; display: block; margin-bottom: 0.5rem;">
      Seleccionar Sesión
    </label>
    {#if loadingSessions}
      <p class="text-muted">Cargando sesiones...</p>
    {:else}
      <select id="session-select" class="input form-select" bind:value={selectedSessionId} on:change={handleSessionChange}>
        <option value={null}>-- Selecciona una sesión --</option>
        {#each sessions as s}
          <option value={s.id}>
            {formatDate(s.scheduled_at)} - {s.group_name} 
            ({s.status === 'open' ? 'En Curso' : s.status === 'closed' ? 'Finalizada' : s.status})
          </option>
        {/each}
      </select>
    {/if}
  </div>

  {#if selectedSessionId}
    <div class="card table-container">
      {#if loadingParams}
        <div class="text-center p-8 text-muted">Cargando asistencia...</div>
      {:else if attendances.length === 0}
        <div class="text-center p-8 text-muted">No hay registros de asistencia para esta sesión. (Revisa si la sesión tiene alumnos asociados en el backend)</div>
      {:else}
        <table class="data-table">
          <thead>
            <tr>
              <th>Boleta</th>
              <th>Nombre</th>
              <th>Hora de Entrada</th>
              <th>Estado</th>
              <th>Modificar</th>
            </tr>
          </thead>
          <tbody>
            {#each attendances as att}
              <tr>
                <td class="code">{att.student_code}</td>
                <td class="font-medium">{att.student_name}</td>
                <td class="text-muted">{att.check_in_at ? formatDate(att.check_in_at) : '--'}</td>
                <td>
                  <span class="badge {att.status === 'present' ? 'badge-success' : att.status === 'late' ? 'badge-warning' : 'badge-error'}">
                    {att.status === 'present' ? 'Presente' : att.status === 'late' ? 'Retardo' : att.status === 'absent' ? 'Falta' : 'Justificado'}
                  </span>
                </td>
                <td>
                  <select 
                    class="input form-select input-sm" 
                    value={att.status} 
                    on:change={(e) => updateStatus(att.id, e.currentTarget.value)}
                  >
                    <option value="present">Presente</option>
                    <option value="late">Retardo</option>
                    <option value="absent">Falta</option>
                    <option value="excused">Justificado</option>
                  </select>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
    </div>
  {/if}
</div>

<style>
  .page-container {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .page-header {
    margin-bottom: 0.5rem;
  }

  .page-title { margin: 0; font-size: 1.875rem; }
  .page-subtitle { margin: 0; color: var(--text-muted); }

  .selector-card {
    padding: 1.5rem;
    max-width: 600px;
  }

  .table-container { padding: 0; overflow-x: auto; }

  .data-table {
    width: 100%;
    border-collapse: collapse;
    text-align: left;
  }
  .data-table th, .data-table td { padding: 1rem 1.5rem; border-bottom: 1px solid var(--border); }
  .data-table th { background-color: var(--bg-subtle); color: var(--text-muted); font-weight: 500; font-size: 0.875rem; }
  .data-table tr:last-child td { border-bottom: none; }

  .code { font-family: var(--mono); color: var(--text-muted); }
  .font-medium { font-weight: 500; }
  .text-muted { color: var(--text-muted); }

  .badge {
    display: inline-block; padding: 0.25rem 0.5rem; border-radius: 9999px;
    font-size: 0.75rem; font-weight: 500;
  }
  .badge-success { background-color: var(--success-bg); color: var(--success); }
  .badge-warning { background-color: var(--warning-bg); color: var(--warning); }
  .badge-error { background-color: var(--error-bg); color: var(--error); }

  .form-select {
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2371717a'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 0.5rem center;
    background-size: 1.5em 1.5em;
    padding-right: 2.5rem;
  }

  .input-sm {
    padding: 0.25rem 2rem 0.25rem 0.5rem;
    font-size: 0.875rem;
  }

  .text-center { text-align: center; }
  .p-8 { padding: 2rem; }

  .error-alert {
    background-color: var(--error-bg);
    color: var(--error);
    padding: 1rem;
    border-radius: var(--radius-md);
  }
</style>
