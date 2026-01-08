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
            >
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
 * - Reads username/photoUrl from localStorage
 * - Reacts to "profile-updated" event
 * - Logout is client-side (clears localStorage) handled by services/api.js
 */

import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { logout as clientLogout } from '../services/api.js'

const router = useRouter()

const profileUsername = ref(localStorage.getItem('username') || '')
const profilePhotoUrl = ref(localStorage.getItem('photoUrl') || '')

const profileInitial = computed(() =>
  profileUsername.value ? profileUsername.value.charAt(0).toUpperCase() : '?'
)

function normalizePhotoUrl(url) {
  if (!url) return ''
  // legacy values saved in storage before backend fix
  if (url.startsWith('/uploads/')) return `/v1${url}`
  return url
}

function syncProfileFromStorage() {
  profileUsername.value = localStorage.getItem('username') || ''
  profilePhotoUrl.value = normalizePhotoUrl(localStorage.getItem('photoUrl') || '')
}

onMounted(() => {
  syncProfileFromStorage()
  window.addEventListener('profile-updated', syncProfileFromStorage)
})

onBeforeUnmount(() => {
  window.removeEventListener('profile-updated', syncProfileFromStorage)
})

function goToAccount() {
  router.push({ name: 'Account' })
}

function handleLogout() {
  // client-side logout (clears token/username/photoUrl + emits profile-updated)
  clientLogout()
  router.replace({ name: 'Login' })
}
</script>

<!-- styles are in dashboard.css -->
