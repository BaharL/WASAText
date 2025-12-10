<template>
  <nav class="sidebar d-flex flex-column p-3">
    <!-- Main menu links -->
    <div>
      <h6 class="text-muted text-uppercase mb-2">Menu</h6>

      <ul class="nav flex-column">
        <li class="nav-item">
          <RouterLink to="/" class="nav-link">
            🏠 Home
          </RouterLink>
        </li>

        <li class="nav-item">
          <RouterLink to="/conversations" class="nav-link">
            💬 Conversations
          </RouterLink>
        </li>

        <li class="nav-item">
          <RouterLink to="/conversations/direct" class="nav-link">
            👤 Direct chats
          </RouterLink>
        </li>

        <li class="nav-item">
          <RouterLink to="/conversations/groups" class="nav-link">
            👥 Groups
          </RouterLink>
        </li>
      </ul>
    </div>

    <!-- Profile box at the bottom -->
    <div class="profile-box mt-auto">
      <!-- Clickable header → va alla pagina Account -->
      <button
        type="button"
        class="profile-header btn btn-link p-0 text-start w-100"
        @click="goToAccount"
      >
        <div class="d-flex align-items-center">
          <div class="profile-avatar">
            {{ profileInitial }}
          </div>
          <div class="ms-2">
            <div class="small text-muted">Signed in as</div>
            <div class="fw-semibold">
              {{ profileUsername || 'unknown' }}
            </div>
            <div class="small text-primary">Profile &amp; settings</div>
          </div>
        </div>
      </button>

      <button
        type="button"
        class="btn btn-sm btn-outline-danger w-100 mt-2"
        @click="handleLogout"
      >
        Logout
      </button>
    </div>
  </nav>
</template>

<script setup>
/**
 * Sidebar.vue
 * ----------
 * - Mostra il menu a sinistra per gli utenti loggati.
 * - Box profilo in fondo: avatar + nome + link a pagina Account.
 * - Logout pulisce il token e torna alla pagina Login.
 */

import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { logout } from '../services/api.js'

const router = useRouter()

// Local profile state (letto da localStorage al primo load)
const profileUsername = ref(localStorage.getItem('username') || '')

// Iniziale per l’avatar tondo
const profileInitial = computed(() => {
  const u = profileUsername.value
  return u ? u.charAt(0).toUpperCase() : '?'
})

// Vai alla pagina Account
function goToAccount() {
  router.push({ name: 'Account' })
}

// Logout → clear localStorage (via api.logout) + torna a Login
function handleLogout() {
  try {
    logout()
  } catch (_) {
    // ignore
  }
  router.push({ name: 'Login' })
}
</script>

<style scoped>
.sidebar {
  border-right: 1px solid #ddd;
  min-height: calc(100vh - 60px); /* altezza viewport meno navbar */
  padding-top:5rem;
  padding-bottom: 0;              
}

.nav-link {
  color: #333;
  padding: 6px 0;
}

.nav-link.router-link-active {
  font-weight: bold;
}

/* Profile box at bottom */
.profile-box {
  border-top: 1px solid #e0e0e0;
  padding-top: 4rem;
}

/* Bottone “header” del profilo senza look da bottone */
.profile-header {
  text-decoration: none;
}

.profile-header:hover .profile-avatar {
  opacity: 0.9;
}

.profile-avatar {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: #0d6efd;
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1rem;
}
</style>
