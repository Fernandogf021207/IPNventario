<script lang="ts">
  import { auth } from '../../lib/stores/auth';

  $: user = $auth.user;
</script>

<div class="dashboard-container">
  <div class="header">
    <h1 class="title">Hola, {user?.full_name || 'Estudiante'}</h1>
    <p class="subtitle">Bienvenido al portal del Laboratorio de Pesados</p>
  </div>

  <div class="card active-session">
    <div class="session-header">
      <h2 style="margin: 0;">Sesiones Activas</h2>
      <span class="badge" style="background: var(--bg-subtle);">0 sesiones web</span>
    </div>
    
    <div class="session-body">
      <p style="color: var(--text-muted); text-align: center; padding: 2rem 0;">
        No tienes sesiones abiertas en este momento.
      </p>
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
</style>
