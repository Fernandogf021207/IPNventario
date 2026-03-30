import { writable } from 'svelte/store';
import { authApi } from '../api';
import type { AuthSession } from '../types';

interface AuthState {
  user: AuthSession | null;
  loading: boolean;
  error: string | null;
}

const initialState: AuthState = {
  user: null,
  loading: true,
  error: null
};

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(initialState);

  return {
    subscribe,
    
    async checkSession() {
      update(state => ({ ...state, loading: true, error: null }));
      try {
        const user = await authApi.me();
        set({ user, loading: false, error: null });
        return user;
      } catch (err: any) {
        set({ user: null, loading: false, error: null }); // Expected if no active session
        return null;
      }
    },

    async login(username: string, pass: string) {
      update(state => ({ ...state, loading: true, error: null }));
      try {
        await authApi.login(username, pass);
        const user = await authApi.me();
        set({ user, loading: false, error: null });
        return user;
      } catch (err: any) {
        set({ user: null, loading: false, error: err.message || 'Login failed' });
        throw err;
      }
    },

    async logout() {
      try {
        await authApi.logout();
      } catch (e) {
        // ignore errors on logout
      } finally {
        set({ user: null, loading: false, error: null });
      }
    }
  };
}

export const auth = createAuthStore();
