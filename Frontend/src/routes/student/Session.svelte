<script lang="ts">
  import { onMount } from 'svelte';
  import { sessionsApi, attendanceApi } from '../../lib/api';
  import type { LabSession } from '../../lib/types';
  import { auth } from '../../lib/stores/auth';

  let activeSession: LabSession | null = null;
  let loading = true;
  let error = '';
  
  let checkInLoading = false;
  let checkInSuccess = false;

  async function loadActiveSession() {
    loading = true;
    error = '';
    try {
      // The backend should return the current open session for the student's group
      // For now, we fetch all sessions and find the first open one for simplicity, 
      // but ideally the backend has a /sessions/active endpoint
      const allSessions = await sessionsApi.list();
      activeSession = allSessions.find(s => s.status === 'open') || null;
    } catch (err: any) {
      error = err.message || 'Error al verificar sesión activa';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadActiveSession();
  });

  async function handleCheckIn() {
    if (!activeSession || !$auth.user?.id) return;
    
    checkInLoading = true;
    error = '';
    try {
      // Usually the backend derives student_id from auth token, but if needed we send it
      await attendanceApi.checkIn(activeSession.id, { student_id: $auth.user.id });
      checkInSuccess = true;
    } catch (err: any) {
      error = err.message || 'Error al registrar asistencia';
    } finally {
      checkInLoading = false;
    }
  }
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Sesión Actual</h1>
      <p class="page-subtitle">Pase de lista y recursos</p>
    </div>
  </div>

  {#if loading}
    <div class="card text-center p-8 text-muted">Buscando sesión activa...</div>
  {:else if !activeSession}
    <div class="card empty-state">
      <div class="empty-icon text-muted" style="background-color: var(--bg-subtle);">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"></path><polyline points="16 2 22 2 22 8"></polyline><line x1="22" y1="2" x2="11" y2="13"></line></svg>
      </div>
      <h2>No hay ninguna sesión en curso</h2>
      <p class="text-muted">Tu profesor aún no ha abierto la sesión para tu grupo. Espera instrucciones.</p>
    </div>
  {:else}
    <div class="card session-card">
      <div class="session-header">
        <div>
          <span class="badge badge-success mb-2">En Curso</span>
          <h2 class="session-title">{activeSession.assignment_title || `Práctica ${activeSession.assignment_id}`}</h2>
          <p class="session-meta">Profesor: {activeSession.created_by_name || 'Desconocido'}</p>
        </div>
        <div class="check-in-container">
          {#if checkInSuccess}
            <div class="success-box">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
              Asistencia Registrada
            </div>
          {:else}
            <button class="btn btn-primary check-in-btn" on:click={handleCheckIn} disabled={checkInLoading}>
              {#if checkInLoading}
                Procesando...
              {:else}
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"></path><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"></path></svg>
                Pasar Lista
              {/if}
            </button>
            <p class="check-in-hint">Registra tu entrada para poder solicitar material.</p>
          {/if}
        </div>
      </div>
      
      {#if error}
        <div class="error-alert mt-4">{error}</div>
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

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem 2rem;
    text-align: center;
  }

  .empty-icon {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 1.5rem;
  }

  .empty-state h2 {
    margin: 0 0 0.5rem 0;
    font-size: 1.25rem;
  }

  .mt-4 { margin-top: 1rem; }
  .mb-2 { margin-bottom: 0.5rem; }

  .session-card {
    padding: 2rem;
  }

  .session-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    flex-wrap: wrap;
    gap: 1.5rem;
  }

  .session-title {
    margin: 0 0 0.25rem 0;
    font-size: 1.5rem;
  }

  .session-meta {
    margin: 0;
    color: var(--text-muted);
  }

  .check-in-container {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
  }

  .check-in-btn {
    padding: 0.75rem 1.5rem;
    font-size: 1rem;
  }

  .check-in-hint {
    margin: 0.5rem 0 0 0;
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .success-box {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background-color: var(--success-bg);
    color: var(--success);
    padding: 0.75rem 1.5rem;
    border-radius: var(--radius-md);
    font-weight: 500;
  }

  .badge {
    display: inline-block;
    padding: 0.25rem 0.5rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 500;
  }

  .badge-success { background-color: var(--success-bg); color: var(--success); }

  .text-center { text-align: center; }
  .p-8 { padding: 2rem; }
  .text-muted { color: var(--text-muted); }

  .error-alert {
    background-color: var(--error-bg);
    color: var(--error);
    padding: 1rem;
    border-radius: var(--radius-md);
    border: 1px solid var(--error-border, rgba(239, 68, 68, 0.3));
  }
</style>
