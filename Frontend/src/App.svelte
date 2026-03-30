<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';
  import Router, { push } from 'svelte-spa-router';
  import { auth } from './lib/stores/auth';
  import Layout from './lib/components/Layout.svelte';
  import Login from './routes/Login.svelte';

  // Placeholder components for Phase 2 implementation
  import TeacherDashboard from './routes/teacher/Dashboard.svelte';
  import TeacherStudents from './routes/teacher/Students.svelte';
  import TeacherAssignments from './routes/teacher/Assignments.svelte';
  import TeacherSessions from './routes/teacher/Sessions.svelte';
  import TeacherAttendance from './routes/teacher/Attendance.svelte';
  import StudentDashboard from './routes/student/Dashboard.svelte';
  import StudentAssignments from './routes/student/Assignments.svelte';
  import StudentSession from './routes/student/Session.svelte';
  import StudentAttendance from './routes/student/Attendance.svelte';

  let isInitialized = false;

  const routes = {
    '/login': Login,
    '/teacher/dashboard': TeacherDashboard,
    '/teacher/students': TeacherStudents,
    '/teacher/assignments': TeacherAssignments,
    '/teacher/sessions': TeacherSessions,
    '/teacher/attendance': TeacherAttendance,
    '/student/dashboard': StudentDashboard,
    '/student/assignments': StudentAssignments,
    '/student/session': StudentSession,
    '/student/attendance': StudentAttendance,
    '*': Login, // Default fallback
  };

  const hash = writable(window.location.hash.replace(/^#/, '') || '/');
  
  window.addEventListener('hashchange', () => {
    hash.set(window.location.hash.replace(/^#/, '') || '/');
  });

  onMount(async () => {
    isInitialized = true;
    await auth.checkSession();
  });

  // Auth Guard
  $: if (isInitialized && !$auth.loading) {
    const isLoginPage = $hash === '/login' || $hash === '';
    const isAuth = !!$auth.user;

    if (!isAuth && !isLoginPage) {
      push('/login');
    } else if (isAuth) {
      if (isLoginPage || $hash === '/') {
        // Redirect based on role
        if ($auth.user?.role === 'teacher' || $auth.user?.role === 'admin' || $auth.user?.role === 'operator') {
          push('/teacher/dashboard');
        } else if ($auth.user?.role === 'student') {
          push('/student/dashboard');
        } else {
          push('/login');
        }
      } else {
        // Simple role check based on path
        if ($hash.startsWith('/teacher') && !['teacher', 'admin', 'operator'].includes($auth.user?.role || '')) {
          push('/login');
        }
      }
    }
  }
</script>

{#if !isInitialized || $auth.loading}
  <div class="loader-container">
    <div class="loader"></div>
    <p>Cargando sesión...</p>
  </div>
{:else}
  {#if $hash === '/login' || $hash === ''}
    <!-- svelte-ignore a11y-missing-attribute -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <Router routes={routes as any} />
  {:else}
    <Layout>
      <Router routes={routes as any} />
    </Layout>
  {/if}
{/if}

<style>
  .loader-container {
    height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background-color: var(--bg);
    color: var(--text-muted);
  }

  .loader {
    width: 40px;
    height: 40px;
    border: 3px solid var(--border);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
