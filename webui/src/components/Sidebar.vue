<template>
  <!-- Custom sidebar -->
  <nav class="app-sidebar d-flex flex-column">
    <!-- Menu -->
    <div class="menu-section">
      <h6 class="text-muted text-uppercase mb-3">Menu</h6>
      <ul class="nav flex-column">
        <li class="nav-item">
          <RouterLink to="/" class="nav-link">🏠 Home</RouterLink>
        </li>
        <li class="nav-item">
          <RouterLink to="/conversations" class="nav-link">💬 Conversations</RouterLink>
        </li>
        <li class="nav-item">
          <RouterLink to="/conversations/direct" class="nav-link">👤 Direct chats</RouterLink>
        </li>
        <li class="nav-item">
          <RouterLink to="/conversations/groups" class="nav-link">👥 Groups</RouterLink>
        </li>
      </ul>
    </div>

    <!-- Profile box (bottom) -->
    <div class="profile-box">
      <button
        type="button"
        class="profile-header btn btn-link p-0 text-start w-100"
        @click="goToAccount"
      >
        <div class="d-flex align-items-center">
          <div class="profile-avatar">
            <img
              v-if="profilePhotoUrl"
              :src="profilePhotoUrl"
              alt="avatar"
              class="avatar-img"
            />
            <span v-else>{{ profileInitial }}</span>
          </div>

          <div class="ms-2">
            <div class="small text-muted">Signed in as</div>
            <div class="fw-semibold">{{ profileUsername || 'unknown' }}</div>
            <div class="small text-primary">Profile & settings</div>
          </div>
        </div>
      </button>

      <button
        type="button"
        class="btn btn-sm btn-outline-danger w-100 mt-3"
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
 * - Shows menu + user profile
 * - Avatar = photo if present, otherwise first letter
 * - Reads data from localStorage
 * - Reacts to "profile-updated" event
 */

import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { logout } from '../services/api.js'

const router = useRouter()

// User data from localStorage
const profileUsername = ref(localStorage.getItem('username') || '')
const profilePhotoUrl = ref(localStorage.getItem('photoUrl') || '')

// Initial letter fallback
const profileInitial = computed(() =>
  profileUsername.value ? profileUsername.value.charAt(0).toUpperCase() : '?'
)

// Keep sidebar in sync after changes (photo/username)
function syncProfileFromStorage() {
  profileUsername.value = localStorage.getItem('username') || ''
  profilePhotoUrl.value = localStorage.getItem('photoUrl') || ''
}

onMounted(() => {
  window.addEventListener('profile-updated', syncProfileFromStorage)
})

onBeforeUnmount(() => {
  window.removeEventListener('profile-updated', syncProfileFromStorage)
})

// Go to account page
function goToAccount() {
  router.push({ name: 'Account' })
}

// Logout
function handleLogout() {
  try {
    logout()
  } catch (_) {}

  localStorage.removeItem('token')
  localStorage.removeItem('username')
  localStorage.removeItem('photoUrl')

  // aggiorna anche la sidebar subito
  syncProfileFromStorage()

  router.push({ name: 'Login' })
}
</script>

<!-- styles are in dashboard.css -->
