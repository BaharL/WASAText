<template>
  <!-- Aggiunto pt-4 per distanziarlo dalla navbar -->
  <nav class="sidebar d-flex flex-column p-3 pt-4">
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
      <!-- Cliccando su questo blocco vai alla pagina Account -->
      <button
        type="button"
        class="profile-btn w-100 text-start"
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
            <div class="text-muted small">
              Profile &amp; settings
            </div>
          </div>
        </div>
      </button>

      <!-- Solo logout rimane nel sidebar -->
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
 * - Shows navigation links for logged-in users.
 * - Displays a bigger profile box (avatar + username) at the bottom.
 * - Clicking the profile box opens the Account page.
 * - Visible only when the current route is NOT "Login" (controlled by App.vue).
 */

import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { logout } from '../services/api.js'

const router = useRouter()

// Local profile state (read from localStorage on first load)
const profileUsername = ref(localStorage.getItem('username') || '')

// Initial letter for avatar circle
const profileInitial = computed(() => {
  const u = profileUsername.value
  return u ? u.charAt(0).toUpperCase() : '?'
})

// Navigate to Account page
function goToAccount() {
  router.push({ name: 'Account' })
}

// Logout → clear localStorage (via api.logout) + go to Login page
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
  min-height: calc(100vh - 60px);
  /* meno spazio sotto → il box scende più giù */
  padding-bottom: 0.5rem;
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
  padding-top: 0.75rem;
}

/* Bottone che contiene avatar + testo (senza bordo "button") */
.profile-btn {
  background: transparent;
  border: none;
  padding: 0;
  cursor: pointer;
}

.profile-btn:focus {
  outline: none;
}

/* Avatar tondo */
.profile-avatar {
  width: 60px;
  height: 60px;
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
