<template>
  <div class="login-page">
    <div class="logo-area">
      <div class="logo-circle">W</div>
      <h1>WASAText</h1>
      <p class="subtitle">Simple • Clean • Fast</p>
    </div>

    <div class="login-card">
      <h2>Accedi</h2>
      <p class="hint">
        Inserisci il tuo <strong>username</strong> per entrare nella chat.
      </p>

      <form @submit.prevent="onSubmit">
        <label for="name">Username</label>
        <input
          id="name"
          v-model.trim="name"
          type="text"
          placeholder="es. bahar"
          :disabled="loading"
        >

        <button type="submit" :disabled="loading || !name">
          <span v-if="loading">Connessione…</span>
          <span v-else>Entra in WASAText</span>
        </button>
      </form>

      <p v-if="error" class="error">
        {{ error }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { login } from '../services/api.js';

const router = useRouter();
const name = ref('');
const loading = ref(false);
const error = ref('');

async function onSubmit() {
  error.value = '';
  loading.value = true;

  try {
    await login(name.value);

    // ⭐ NEW: salviamo anche lo username nel localStorage
    localStorage.setItem('username', name.value);

    await router.push({ name: 'Home' });
  } catch (err) {
    console.error(err);
    error.value = err.message || 'Errore di login.';
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.login-page {
  min-height: calc(100vh - 80px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  background: radial-gradient(circle at top, #3c9fff 0, #141a2a 50%, #0a0c13 100%);
  color: #f5f5f5;
}

.logo-area {
  text-align: center;
  margin-bottom: 1.5rem;
}

.logo-circle {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, #42b883, #33a1ff);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
  font-size: 1.6rem;
  margin: 0 auto 0.35rem;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
}

.logo-area h1 {
  margin: 0;
  font-size: 1.6rem;
  letter-spacing: 0.04em;
}

.subtitle {
  margin: 0.15rem 0 0;
  font-size: 0.9rem;
  opacity: 0.85;
}

.login-card {
  width: 100%;
  max-width: 360px;
  background: rgba(11, 15, 25, 0.92);
  border-radius: 16px;
  padding: 1.6rem 1.7rem;
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.55);
  border: 1px solid rgba(255, 255, 255, 0.06);
  backdrop-filter: blur(6px);
}

.login-card h2 {
  margin-top: 0;
  margin-bottom: 0.4rem;
  font-size: 1.3rem;
}

.hint {
  margin-top: 0;
  margin-bottom: 1.3rem;
  font-size: 0.9rem;
  color: #c6d2ff;
}

form {
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
}

label {
  font-size: 0.85rem;
  color: #d0d7ff;
}

input {
  padding: 0.55rem 0.7rem;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  font-size: 0.95rem;
  background: rgba(7, 10, 18, 0.9);
  color: #f5f5f5;
}

input::placeholder {
  color: #7f88a8;
}

input:focus {
  outline: none;
  border-color: #42b883;
  box-shadow: 0 0 0 1px rgba(66, 184, 131, 0.45);
}

button {
  margin-top: 0.4rem;
  padding: 0.6rem 0.7rem;
  border-radius: 999px;
  border: none;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  background: linear-gradient(135deg, #42b883, #2aa66e);
  color: #0b101b;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.45);
  transition: transform 0.05s ease, box-shadow 0.1s ease, opacity 0.1s ease;
}

button:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 14px 30px rgba(0, 0, 0, 0.55);
}

button:active:not(:disabled) {
  transform: translateY(0);
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.45);
}

button:disabled {
  opacity: 0.7;
  cursor: default;
}

.error {
  margin-top: 0.75rem;
  color: #ff6b6b;
  font-size: 0.85rem;
}
</style>
