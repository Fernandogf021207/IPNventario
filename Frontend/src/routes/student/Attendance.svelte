<script lang="ts">
  import { onMount } from 'svelte';
  import { attendanceApi, sessionsApi } from '../../lib/api';
  import type { Attendance, LabSession } from '../../lib/types';
  import { auth } from '../../lib/stores/auth';

  // We fetch sessions and then attendance for those sessions to see if the student attended
  let error = '';
  let loading = true;
  
  type SessionWithAttendance = {
    session: LabSession;
    attendance: Attendance | null;
  };
  
  let history: SessionWithAttendance[] = [];

  async function loadHistory() {
    loading = true;
    error = '';
    try {
      if (!$auth.user?.id) return;
      
      const allSessions = await sessionsApi.list();
      const pastSessions = allSessions.filter(s => s.status !== 'scheduled');
      
      const historyData: SessionWithAttendance[] = [];
      
      // In a real scenario, there'd be a /students/me/attendance endpoint.
      // We simulate by calling attendance list per session and finding the current student
      await Promise.all(pastSessions.map(async (session) => {
        try {
          const attendances = await attendanceApi.list(session.id);
          const studentAtt = attendances.find(a => a.student_id === $auth.user?.id) || null;
          historyData.push({ session, attendance: studentAtt });
        } catch {
          // ignore error for a single session
        }
      }));
      
      // Sort by scheduled_at desc
      historyData.sort((a, b) => new Date(b.session.scheduled_at).getTime() - new Date(a.session.scheduled_at).getTime());
      history = historyData;

    } catch (err: any) {
      error = err.message || 'Error al cargar historial de asistencia';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadHistory();
  });

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
      <h1 class="page-title">Historial de Asistencia</h1>
      <p class="page-subtitle">Tus registros de entrada a sesiones de laboratorio</p>
    </div>
  </div>

  {#if error}
    <div class="error-alert">{error}</div>
  {/if}

  <div class="card table-container">
    {#if loading}
      <div class="text-center p-8 text-muted">Cargando historial...</div>
    {:else if history.length === 0}
      <div class="text-center p-8 text-muted">No se encontraron sesiones pasadas o activas.</div>
    {:else}
      <table class="data-table">
        <thead>
          <tr>
            <th>Fecha</th>
            <th>Práctica</th>
            <th>Hora de Entrada</th>
            <th>Estado</th>
          </tr>
        </thead>
        <tbody>
          {#each history as item}
            <tr>
              <td class="code">{formatDate(item.session.scheduled_at)}</td>
              <td class="font-medium">{item.session.assignment_title || `Práctica ${item.session.assignment_id}`}</td>
              <td class="text-muted">
                {item.attendance?.check_in_at ? formatDate(item.attendance.check_in_at) : '--'}
              </td>
              <td>
                {#if item.attendance}
                  <span class="badge {item.attendance.status === 'present' ? 'badge-success' : item.attendance.status === 'late' ? 'badge-warning' : 'badge-error'}">
                    {item.attendance.status === 'present' ? 'Presente' : item.attendance.status === 'late' ? 'Retardo' : item.attendance.status === 'absent' ? 'Falta' : 'Justificado'}
                  </span>
                {:else}
                  <span class="badge badge-error">Falta (No Registrada)</span>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
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

  .text-center { text-align: center; }
  .p-8 { padding: 2rem; }

  .error-alert {
    background-color: var(--error-bg);
    color: var(--error);
    padding: 1rem;
    border-radius: var(--radius-md);
  }
</style>
