<script lang="ts">
  import { auth } from '../stores/auth';
  import { push } from 'svelte-spa-router';

  let isSidebarOpen = false;

  function toggleSidebar() {
    isSidebarOpen = !isSidebarOpen;
  }

  function handleLogout() {
    auth.logout().then(() => {
      push('/login');
    });
  }

  $: user = $auth.user;
  $: role = user?.role;
</script>

<style>
  .layout {
    display: flex;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
  }

  .sidebar {
    width: 280px;
    background-color: var(--bg-card);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    transition: transform 0.3s ease;
    z-index: 40;
  }

  .sidebar-header {
    padding: 1.5rem;
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .logo {
    width: 32px;
    height: 32px;
    background-color: var(--accent);
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: bold;
  }

  .title {
    font-weight: 600;
    font-size: 1.25rem;
    color: var(--text-h);
    margin: 0;
  }

  .nav {
    flex: 1;
    overflow-y: auto;
    padding: 1rem 0.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .nav-link {
    display: flex;
    align-items: center;
    padding: 0.75rem 1rem;
    color: var(--text);
    text-decoration: none;
    border-radius: var(--radius-md);
    transition: background-color 0.2s, color 0.2s;
    font-weight: 500;
  }

  .nav-link:hover {
    background-color: var(--bg-subtle);
    color: var(--text-h);
  }

  .nav-divider {
    height: 1px;
    background-color: var(--border);
    margin: 0.5rem 0.5rem;
  }

  .nav-section {
    font-size: 0.7rem;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    padding: 0.25rem 1rem;
  }
  .user-section {
    padding: 1rem;
    border-top: 1px solid var(--border);
  }

  .user-info {
    display: flex;
    flex-direction: column;
    margin-bottom: 1rem;
  }

  .user-name {
    font-weight: 600;
    color: var(--text-h);
    font-size: 0.875rem;
  }

  .user-role {
    font-size: 0.75rem;
    color: var(--text-muted);
    text-transform: capitalize;
  }

  .main-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    position: relative;
    background-color: var(--bg-subtle);
  }

  .topbar {
    height: 64px;
    background-color: var(--bg-card);
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    padding: 0 1.5rem;
    display: none; /* Only show on mobile */
  }

  .content-area {
    flex: 1;
    overflow-y: auto;
    padding: 2rem;
  }

  @media (max-width: 768px) {
    .sidebar {
      position: fixed;
      left: 0;
      top: 0;
      bottom: 0;
      transform: translateX(-100%);
    }

    .sidebar.open {
      transform: translateX(0);
    }

    .topbar {
      display: flex;
    }

    .menu-btn {
      background: none;
      border: none;
      color: var(--text-h);
      cursor: pointer;
      padding: 0.5rem;
    }

    .overlay {
      position: fixed;
      inset: 0;
      background-color: rgba(0,0,0,0.5);
      z-index: 30;
      display: none;
    }

    .overlay.open {
      display: block;
    }
    
    .content-area {
      padding: 1rem;
    }
  }
</style>

<div class="layout">
  <!-- Mobile overlay -->
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="overlay" class:open={isSidebarOpen} on:click={toggleSidebar}></div>

  <aside class="sidebar" class:open={isSidebarOpen}>
    <div class="sidebar-header">
      <div class="logo">L</div>
      <h1 class="title">IPNventario</h1>
    </div>

    <nav class="nav">
      {#if role === 'teacher' || role === 'admin' || role === 'operator'}
        <a href="#/teacher/dashboard" class="nav-link" on:click={() => isSidebarOpen = false}>Dashboard</a>
        <a href="#/teacher/students" class="nav-link" on:click={() => isSidebarOpen = false}>Alumnos</a>
        <a href="#/teacher/assignments" class="nav-link" on:click={() => isSidebarOpen = false}>Prácticas</a>
        <a href="#/teacher/sessions" class="nav-link" on:click={() => isSidebarOpen = false}>Sesiones</a>
        <a href="#/teacher/attendance" class="nav-link" on:click={() => isSidebarOpen = false}>Asistencia</a>
        <div class="nav-divider"></div>
        <span class="nav-section">Recursos</span>
        <a href="#/teacher/inventory" class="nav-link" on:click={() => isSidebarOpen = false}>Inventario</a>
        <a href="#/teacher/requests" class="nav-link" on:click={() => isSidebarOpen = false}>Solicitudes</a>
        <a href="#/teacher/equipment" class="nav-link" on:click={() => isSidebarOpen = false}>Maquinaria</a>
      {:else if role === 'student'}
        <a href="#/student/dashboard" class="nav-link" on:click={() => isSidebarOpen = false}>Dashboard</a>
        <a href="#/student/assignments" class="nav-link" on:click={() => isSidebarOpen = false}>Mis Prácticas</a>
        <a href="#/student/session" class="nav-link" on:click={() => isSidebarOpen = false}>Sesión Actual</a>
        <a href="#/student/attendance" class="nav-link" on:click={() => isSidebarOpen = false}>Mi Asistencia</a>
        <a href="#/student/request-resource" class="nav-link" on:click={() => isSidebarOpen = false}>Solicitar Recurso</a>
      {/if}
    </nav>

    <div class="user-section">
      {#if user}
        <div class="user-info">
          <span class="user-name">{user.full_name}</span>
          <span class="user-role">{user.role}</span>
        </div>
      {/if}
      <button class="btn btn-secondary" style="width: 100%" on:click={handleLogout}>
        Cerrar Sesión
      </button>
    </div>
  </aside>

  <main class="main-content">
    <header class="topbar">
      <button class="menu-btn" on:click={toggleSidebar} aria-label="Toggle Menu">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="3" y1="12" x2="21" y2="12"></line>
          <line x1="3" y1="6" x2="21" y2="6"></line>
          <line x1="3" y1="18" x2="21" y2="18"></line>
        </svg>
      </button>
      <span class="title" style="margin-left: 1rem; font-size: 1.125rem;">IPNventario</span>
    </header>
    
    <div class="content-area">
      <slot></slot>
    </div>
  </main>
</div>
