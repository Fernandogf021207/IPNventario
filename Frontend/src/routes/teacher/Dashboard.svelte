<script lang="ts">
  import { onMount } from 'svelte';
  import { auth } from '../../lib/stores/auth';
  import { studentsApi, assignmentsApi, sessionsApi } from '../../lib/api';

  $: user = $auth.user;

  let studentCount: number | string = '--';
  let assignmentCount: number | string = '--';
  let sessionCount: number | string = '--';
  let loading = true;

  onMount(async () => {
    try {
      const [students, assignments, sessions] = await Promise.all([
        studentsApi.list(),
        assignmentsApi.list(),
        sessionsApi.list()
      ]);
      studentCount = students.length;
      assignmentCount = assignments.length;
      sessionCount = sessions.filter((s: any) => s.status === 'scheduled').length;
    } catch (err) {
      console.error('Error fetching stats:', err);
    } finally {
      loading = false;
    }
  });
</script>

<div class="dashboard-container">
  <div class="header">
    <h1 class="title">Hola, {user?.full_name || 'Profesor'}</h1>
    <p class="subtitle">Bienvenido al panel de administración de laboratorio</p>
  </div>

  <div class="stats-grid">
    <div class="card stat-card">
      <div class="stat-icon" style="background-color: var(--accent-bg); color: var(--accent);">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
      </div>
      <div>
        <p class="stat-label">Alumnos Registrados</p>
        <h3 class="stat-value">{studentCount}</h3>
      </div>
    </div>
    
    <div class="card stat-card">
      <div class="stat-icon" style="background-color: var(--success-bg); color: var(--success);">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
      </div>
      <div>
        <p class="stat-label">Prácticas Creadas</p>
        <h3 class="stat-value">{assignmentCount}</h3>
      </div>
    </div>

    <div class="card stat-card">
      <div class="stat-icon" style="background-color: var(--warning-bg); color: var(--warning);">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
      </div>
      <div>
        <p class="stat-label">Sesiones Programadas</p>
        <h3 class="stat-value">{sessionCount}</h3>
      </div>
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

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1.5rem;
  }

  .stat-card {
    display: flex;
    align-items: center;
    gap: 1.25rem;
  }

  .stat-icon {
    width: 3rem;
    height: 3rem;
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .stat-label {
    margin: 0;
    font-size: 0.875rem;
    color: var(--text-muted);
  }

  .stat-value {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 600;
  }
</style>
