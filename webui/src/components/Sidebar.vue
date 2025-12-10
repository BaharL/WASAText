<template>
  <!-- AGGIUNGI pt-3 QUI -->
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
      <div class="d-flex align-items-center mb-2">
        <div class="profile-avatar">
          {{ profileInitial }}
        </div>
        <div class="ms-2">
          <div class="small text-muted">Signed in as</div>
          <div class="fw-semibold">
            {{ profileUsername || 'unknown' }}
          </div>
        </div>
      </div>

      <button
        type="button"
        class="btn btn-sm btn-outline-secondary w-100 mb-1"
        @click="handleChangeUsername"
      >
        Change username
      </button>

      <button
        type="button"
        class="btn btn-sm btn-outline-danger w-100"
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
 * - Displays a bigger profile box (avatar + username + actions) at the bottom.
 * - Visible only when the current route is NOT "Login" (controlled by App.vue).
 */

import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { setMyUserName, logout } from '../services/api.js'

const router = useRouter()

// Local profile state (read from localStorage on first load)
const profileUsername = ref(localStorage.getItem('username') || '')

// Initial letter for avatar circle
const profileInitial = computed(() => {
  const u = profileUsername.value
  return u ? u.charAt(0).toUpperCase() : '?'
})

// Change username → calls PUT /users/username via api.js
async function handleChangeUsername() {
  const current = profileUsername.value || ''
  const next = window.prompt('New username (3–16 chars):', current)
  if (!next) return

  const trimmed = next.trim()
  if (!trimmed) return

  try {
    await setMyUserName(trimmed)
    profileUsername.value = trimmed
    localStorage.setItem('username', trimmed)
  } catch (e) {
    console.error(e)
    window.alert(e.message || 'Failed to change username.')
  }
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

.profile-avatar {
  width: 40px;
  height: 40px;
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
