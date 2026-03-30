<script lang="ts">
  import { onMount } from 'svelte';
  import { studentsApi } from '../../lib/api';
  import type { Student } from '../../lib/types';
  import Modal from '../../lib/components/Modal.svelte';
  import ConfirmDialog from '../../lib/components/ConfirmDialog.svelte';

  let students: Student[] = [];
  let loading = true;
  let error = '';

  let isModalOpen = false;
  let isEditing = false;
  let formData: any = {};

  let isConfirmOpen = false;
  let selectedStudent: Student | null = null;

  async function loadStudents() {
    loading = true;
    error = '';
    try {
      students = await studentsApi.list();
    } catch (err: any) {
      error = err.message || 'Error al cargar alumnos';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadStudents();
  });

  function openCreateModal() {
    isEditing = false;
    formData = {
      student_code: '',
      full_name: '',
      email: '',
      group_name: '',
      semester: 1,
      is_active: true,
      password: ''
    };
    isModalOpen = true;
  }

  function openEditModal(student: Student) {
    isEditing = true;
    formData = { ...student, password: '' };
    isModalOpen = true;
  }

  function confirmDelete(student: Student) {
    selectedStudent = student;
    isConfirmOpen = true;
  }

  async function handleDelete() {
    if (!selectedStudent) return;
    try {
      await studentsApi.delete(selectedStudent.id);
      await loadStudents();
    } catch (err: any) {
      alert(err.message || 'Error al desactivar el alumno');
    } finally {
      isConfirmOpen = false;
      selectedStudent = null;
    }
  }

  async function handleSubmit() {
    try {
      // Limpiar campos vacíos en edición para no sobreescribir con vacío (ej. password)
      const dataToSave = { ...formData };
      dataToSave.username = formData.student_code; // Usar la boleta como username
      
      if (isEditing && !dataToSave.password) {
        delete dataToSave.password;
      }

      // Fallback for password in creation
      if (!isEditing && !dataToSave.password) {
        dataToSave.password = dataToSave.student_code;
      }

      if (isEditing && formData.id) {
        await studentsApi.update(formData.id, dataToSave);
      } else {
        await studentsApi.create(dataToSave);
      }
      isModalOpen = false;
      await loadStudents();
    } catch (err: any) {
      alert(err.message || 'Error al guardar el estudiante');
    }
  }
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Catálogo de Alumnos</h1>
      <p class="page-subtitle">Gestión de estudiantes y cuentas de acceso</p>
    </div>
    <div class="header-actions">
      <button class="btn btn-secondary" disabled>Importar CSV</button>
      <button class="btn btn-primary" on:click={openCreateModal}>
        + Nuevo Alumno
      </button>
    </div>
  </div>

  {#if error}
    <div class="error-alert">{error}</div>
  {/if}

  <div class="card table-container">
    {#if loading}
      <div class="text-center p-8 text-muted">Cargando...</div>
    {:else if students.length === 0}
      <div class="text-center p-8 text-muted">No hay alumnos registrados.</div>
    {:else}
      <table class="data-table">
        <thead>
          <tr>
            <th>Boleta</th>
            <th>Nombre Completo</th>
            <th>Grupo</th>
            <th>Email</th>
            <th>Estado</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          {#each students as student}
            <tr>
              <td class="code">{student.student_code}</td>
              <td class="font-medium">{student.full_name}</td>
              <td>{student.group_name}</td>
              <td>{student.email || '--'}</td>
              <td>
                <span class="badge {student.is_active ? 'badge-success' : 'badge-neutral'}">
                  {student.is_active ? 'Activo' : 'Inactivo'}
                </span>
              </td>
              <td class="actions">
                <button class="btn-icon" on:click={() => openEditModal(student)} title="Editar información">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9"></path><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path></svg>
                </button>
                <button class="btn-icon text-error" on:click={() => confirmDelete(student)} title="Desactivar">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"></path><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>

<Modal bind:open={isModalOpen} title={isEditing ? 'Editar Alumno' : 'Nuevo Alumno'}>
  <form id="student-form" on:submit|preventDefault={handleSubmit}>
    <div class="form-grid">
      <div class="form-group">
        <label for="student_code">Boleta / Matrícula *</label>
        <input class="input" id="student_code" bind:value={formData.student_code} required placeholder="202XXXXXXX" />
      </div>
      <div class="form-group">
        <label for="group_name">Grupo *</label>
        <input class="input" id="group_name" bind:value={formData.group_name} required placeholder="3TV4" />
      </div>
      <div class="form-group full-width">
        <label for="full_name">Nombre Completo *</label>
        <input class="input" id="full_name" bind:value={formData.full_name} required />
      </div>
      <div class="form-group full-width">
        <label for="email">Email Académico</label>
        <input type="email" class="input" id="email" bind:value={formData.email} placeholder="correo@alumno.ipn.mx" />
      </div>
      <div class="form-group">
        <label for="password">{isEditing ? 'Nueva Contraseña (opcional)' : 'Contraseña'}</label>
        <input type="password" class="input" id="password" bind:value={formData.password} placeholder={isEditing ? 'Dejar en blanco para no cambiar' : 'Por defecto es su boleta'} />
      </div>
      <div class="form-group">
        <label for="semester">Semestre</label>
        <input type="number" min="1" max="10" class="input" id="semester" bind:value={formData.semester} />
      </div>
    </div>
  </form>

  <div slot="footer">
    <button class="btn btn-secondary" on:click={() => isModalOpen = false}>Cancelar</button>
    <button class="btn btn-primary" type="submit" form="student-form">Guardar Cambios</button>
  </div>
</Modal>

<ConfirmDialog
  bind:open={isConfirmOpen}
  title="Desactivar Alumno"
  message={`¿Estás seguro de que deseas desactivar a ${selectedStudent?.full_name}? Ya no podrá iniciar sesión.`}
  confirmText="Sí, desactivar"
  danger={true}
  on:confirm={handleDelete}
/>

<style>
  .page-container { display: flex; flex-direction: column; gap: 1.5rem; }
  .page-header { display: flex; justify-content: space-between; align-items: flex-end; }
  .page-title { margin: 0; font-size: 1.875rem; }
  .page-subtitle { margin: 0; color: var(--text-muted); }
  .header-actions { display: flex; gap: 0.75rem; }
  .table-container { padding: 0; overflow-x: auto; }
  .data-table { width: 100%; border-collapse: collapse; text-align: left; }
  .data-table th, .data-table td { padding: 1rem 1.5rem; border-bottom: 1px solid var(--border); }
  .data-table th { background-color: var(--bg-subtle); font-weight: 500; color: var(--text-muted); font-size: 0.875rem; }
  .data-table tr:last-child td { border-bottom: none; }
  .code { font-family: var(--mono); color: var(--text-muted); }
  .font-medium { font-weight: 500; }
  .badge { display: inline-block; padding: 0.25rem 0.5rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 500; }
  .badge-success { background-color: var(--success-bg); color: var(--success); }
  .badge-neutral { background-color: var(--bg-subtle); color: var(--text-muted); }
  .actions { display: flex; gap: 0.5rem; }
  .btn-icon { background: none; border: none; color: var(--text-muted); cursor: pointer; padding: 0.25rem; border-radius: var(--radius-sm); }
  .btn-icon:hover { color: var(--accent); background-color: var(--bg-subtle); }
  .text-error:hover { color: var(--error); }
  .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  .form-group { display: flex; flex-direction: column; gap: 0.25rem; }
  .full-width { grid-column: span 2; }
  .form-group label { font-size: 0.875rem; font-weight: 500; }
  .text-center { text-align: center; }
  .p-8 { padding: 2rem; }
  .text-muted { color: var(--text-muted); }
  .error-alert { background-color: var(--error-bg); color: var(--error); padding: 1rem; border-radius: var(--radius-md); border: 1px solid var(--error-border, rgba(239, 68, 68, 0.3)); }
</style>
