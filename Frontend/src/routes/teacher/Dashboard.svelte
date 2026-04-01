<script lang="ts">
  import { onMount } from 'svelte';
  import { auth } from '../../lib/stores/auth';
  import { studentsApi, assignmentsApi, sessionsApi, inventoryApi, requestsApi, equipmentApi } from '../../lib/api';

  $: user = $auth.user;

  let studentCount: number | string = '--';
  let assignmentCount: number | string = '--';
  let sessionCount: number | string = '--';
  let lowStockCount: number | string = '--';
  let pendingRequestsCount: number | string = '--';
  let activeEquipmentCount: number | string = '--';
  let loading = true;

  onMount(async () => {
    try {
      const [students, assignments, sessions, items, requests, equipment] = await Promise.all([
        studentsApi.list(),
        assignmentsApi.list(),
        sessionsApi.list(),
        inventoryApi.listItems(),
        requestsApi.list(),
        equipmentApi.list()
      ]);
      studentCount = students.length;
      assignmentCount = assignments.length;
      sessionCount = sessions.filter((s: any) => s.status === 'scheduled').length;
      
      lowStockCount = items.filter((i: any) => i.item_type !== 'machine' && i.stock <= (i.min_stock || 0)).length;
      pendingRequestsCount = requests.filter((r: any) => r.status === 'pending').length;
      activeEquipmentCount = equipment.filter((e: any) => e.status === 'active').length;
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

    <div class="card stat-card">
      <div class="stat-icon" style="background-color: var(--warning-bg); color: var(--warning);">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
      </div>
      <div>
        <p class="stat-label">Stock Crítico</p>
        <h3 class="stat-value text-warning">{lowStockCount}</h3>
      </div>
    </div>

    <div class="card stat-card">
      <div class="stat-icon" style="background-color: var(--error-bg); color: var(--error);">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="12" y1="18" x2="12" y2="12"/><line x1="9" y1="15" x2="15" y2="15"/></svg>
      </div>
      <div>
        <p class="stat-label">Solicitudes Mtrl.</p>
        <h3 class="stat-value text-error">{pendingRequestsCount}</h3>
      </div>
    </div>

    <div class="card stat-card">
      <div class="stat-icon" style="background-color: var(--success-bg); color: var(--success);">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
      </div>
      <div>
        <p class="stat-label">Maquinaria Activa</p>
        <h3 class="stat-value text-success">{activeEquipmentCount}</h3>
      </div>
    </div>
  </div>

  <div class="quick-links">
    <h2>Operación del Laboratorio (Atajos)</h2>
    <div class="links-grid">
      <a href="#/teacher/inventory" class="card link-card">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="link-icon"><rect width="20" height="14" x="2" y="7" rx="2" ry="2"/><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/></svg>
        <h3>Inventario</h3>
        <p>Gestionar stock, herramientas y herramientas consumibles</p>
      </a>
      <a href="#/teacher/requests" class="card link-card">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="link-icon"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        <h3>Solicitudes</h3>
        <p>Aprobar y devolver herramientas para prácticas</p>
      </a>
      <a href="#/teacher/equipment" class="card link-card">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="link-icon"><path d="M12 2v20"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
        <h3>Maquinaria</h3>
        <p>Registro de uso de equipo fijo pesado y trazabilidad</p>
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
  
  .text-warning { color: var(--warning); }
  .text-error { color: var(--error); }
  .text-success { color: var(--success); }

  .quick-links { margin-top: 2rem; }
  .quick-links h2 { font-size: 1.25rem; margin-bottom: 1rem; color: var(--text-h); }
  
  .links-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1.5rem; }
  .link-card { display: flex; flex-direction: column; gap: 0.5rem; text-decoration: none; color: inherit; transition: all 200ms ease; }
  .link-card:hover { transform: translateY(-4px); border-color: var(--accent); }
  .link-icon { color: var(--accent); margin-bottom: 0.5rem; }
  .link-card h3 { font-size: 1.125rem; margin: 0; color: var(--text-h); }
  .link-card p { font-size: 0.875rem; color: var(--text-muted); margin: 0; line-height: 1.4; }
</style>
