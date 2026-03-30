<script lang="ts">
  import { auth } from '../lib/stores/auth';
  import { push } from 'svelte-spa-router';

  let username = '';
  let password = '';
  let error = '';
  let loading = false;

  async function handleSubmit() {
    error = '';
    loading = true;
    try {
      await auth.login(username, password);
      // Let the router handle redirection based on auth state, or push manually:
      push('/');
    } catch (err: any) {
      error = err.message || 'Credenciales inválidas';
    } finally {
      loading = false;
    }
  }
</script>

<style>
  .login-container {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1.5rem;
    background-color: var(--bg-subtle);
  }

  .login-card {
    width: 100%;
    max-width: 24rem;
    padding: 2.5rem;
  }

  .logo-area {
    text-align: center;
    margin-bottom: 2rem;
  }

  .logo-box {
    width: 48px;
    height: 48px;
    background-color: var(--accent);
    color: white;
    border-radius: var(--radius-md);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 1.5rem;
    font-weight: bold;
    margin-bottom: 1rem;
  }

  .title {
    font-size: 1.5rem;
    margin-bottom: 0.5rem;
  }

  .subtitle {
    color: var(--text-muted);
    font-size: 0.875rem;
  }

  .form-group {
    margin-bottom: 1.25rem;
  }

  label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
    color: var(--text-h);
    font-size: 0.875rem;
  }

  .error-message {
    color: var(--error);
    background-color: var(--error-bg);
    padding: 0.75rem;
    border-radius: var(--radius-md);
    margin-bottom: 1.5rem;
    font-size: 0.875rem;
    text-align: center;
  }

  .submit-btn {
    width: 100%;
    margin-top: 1rem;
    padding: 0.75rem;
    font-size: 1rem;
  }
</style>

<div class="login-container">
  <div class="card login-card">
    <div class="logo-area">
      <div class="logo-box">L</div>
      <h1 class="title">Bienvenido</h1>
      <p class="subtitle">IPNventario Laboratorio de Pesados</p>
    </div>

    {#if error}
      <div class="error-message">
        {error}
      </div>
    {/if}

    <form on:submit|preventDefault={handleSubmit}>
      <div class="form-group">
        <label for="username">Usuario / Boleta</label>
        <input 
          id="username" 
          type="text" 
          class="input" 
          bind:value={username} 
          required 
          autocomplete="username"
          placeholder="Ingresa tu usuario"
        />
      </div>

      <div class="form-group">
        <label for="password">Contraseña</label>
        <input 
          id="password" 
          type="password" 
          class="input" 
          bind:value={password} 
          required 
          autocomplete="current-password"
          placeholder="••••••••"
        />
      </div>

      <button type="submit" class="btn btn-primary submit-btn" disabled={loading}>
        {#if loading}
          Iniciando sesión...
        {:else}
          Iniciar Sesión
        {/if}
      </button>
    </form>
  </div>
</div>
