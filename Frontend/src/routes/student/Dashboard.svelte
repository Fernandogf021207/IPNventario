<script lang="ts">
  import { onMount } from 'svelte';
  import { auth } from '../../lib/stores/auth';
  import { sessionsApi } from '../../lib/api';
  import type { LabSession } from '../../lib/types';

  $: user = $auth.user;
  let activeSession: LabSession | null = null;
  let loading = true;

  onMount(async () => {
    try {
      const sessions = await sessionsApi.list();
      activeSession = sessions.find((s: any) => s.status === 'open') || null;
    } catch (err) {
      console.error('Error fetching sessions:', err);
    } finally {
      loading = false;
    }
  });
</script>

<div class="dashboard-container">
  <div class="header">
    <h1 class="title">Hola, {user?.full_name || 'Estudiante'}</h1>
    <p class="subtitle">Bienvenido al portal del Laboratorio de Pesados</p>
  </div>

  <div class="card active-session">
    <div class="session-header">
      <h2 style="margin: 0; color: var(--text-h);">Estado del Laboratorio</h2>
      {#if loading}
        <span class="badge badge-loading">Consultando...</span>
      {:else if activeSession}
        <span class="badge badge-active">🟢 En Curso</span>
      {:else}
        <span class="badge badge-inactive">Sin Práctica Activa</span>
      {/if}
    </div>
    
    <div class="session-body">
      {#if loading}
        <p class="text-muted text-center p-4">Cargando estado...</p>
      {:else if activeSession}
        <div class="active-session-info text-center">
          <h3>{activeSession.assignment_title}</h3>
          <p class="text-muted">La práctica ha comenzado. Registra tu asistencia.</p>
          <div style="margin-top: 1.5rem;">
            <a href="#/student/session" class="btn btn-primary">Ir a la Sesión Actual</a>
          </div>
        </div>
      {:else}
        <p class="text-muted text-center p-4">Tu profesor aún no ha iniciado sesión para tu grupo. Espera instrucciones.</p>
      {/if}
    </div>
  </div>

  <div class="quick-links">
    <h2>Atajos Rápidos</h2>
    <div class="links-grid">
      <a href="#/student/request-resource" class="card link-card">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="link-icon"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        <h3>Solicitar Material</h3>
        <p>Herramientas, consumibles y uso de maquinaria</p>
      </a>
      <a href="#/student/assignments" class="card link-card">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="link-icon"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
        <h3>Mis Prácticas</h3>
        <p>Ver mis calificaciones e instrucciones</p>
      </a>
      <a href="#/student/attendance" class="card link-card">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="link-icon"><rect width="18" height="18" x="3" y="4" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
        <h3>Mi Asistencia</h3>
        <p>Historial de asistencias registradas</p>
      </a>
    </div>
  </div>
</div>

<style>
  .dashboard-container {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }
  
  .header {
    margin-bottom: 0.5rem;
  }
  
  .subtitle {
    color: var(--text-muted);
  }

  .session-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--border);
    padding-bottom: 1rem;
    margin-bottom: 1rem;
  }

  .badge {
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.875rem;
    font-weight: 500;
  }
  
  .badge-loading { background: var(--bg-subtle); color: var(--text-muted); }
  .badge-active { background: var(--success-bg); color: var(--success); }
  .badge-inactive { background: var(--border); color: var(--text-muted); }

  .text-center { text-align: center; }
  .text-muted { color: var(--text-muted); }
  .p-4 { padding: 1rem; }

  .active-session-info h3 { margin: 0 0 0.5rem 0; font-size: 1.5rem; color: var(--text-h); }
  
  .quick-links { margin-top: 1rem; }
  .quick-links h2 { font-size: 1.25rem; margin-bottom: 1rem; color: var(--text-h); }
  
  .links-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1rem; }
  .link-card { display: flex; flex-direction: column; gap: 0.5rem; text-decoration: none; color: inherit; transition: all 200ms ease; }
  .link-card:hover { transform: translateY(-4px); border-color: var(--accent); }
  .link-icon { color: var(--accent); margin-bottom: 0.5rem; }
  .link-card h3 { font-size: 1.125rem; margin: 0; color: var(--text-h); }
  .link-card p { font-size: 0.875rem; color: var(--text-muted); margin: 0; line-height: 1.4; }
</style>
