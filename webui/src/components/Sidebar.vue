<template>
  <!-- Our custom sidebar: does NOT conflict with dashboard.css -->
  <nav class="app-sidebar d-flex flex-column">
    <!-- Main menu section - scrollabile -->
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

    <!-- Profile box stays at the bottom - SEMPRE VISIBILE -->
    <div class="profile-box">
      <!-- Clickable profile header → go to Account page -->
      <button
        type="button"
        class="profile-header btn btn-link p-0 text-start w-100"
        @click="goToAccount"
      >
        <div class="d-flex align-items-center">
          <div class="profile-avatar">{{ profileInitial }}</div>
          <div class="ms-2">
            <div class="small text-muted">Signed in as</div>
            <div class="fw-semibold">{{ profileUsername || 'unknown' }}</div>
            <div class="small text-primary">Profile & settings</div>
          </div>
        </div>
      </button>
      <!-- Logout button -->
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
 * -----------
 * - Custom sidebar - stili gestiti in dashboard.css
 * - Displays menu items + a bottom-aligned profile box.
 * - Clicking profile opens AccountView.
 * - Logout clears local storage and returns to Login.
 */
import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { logout } from '../services/api.js'

const router = useRouter()

// Load username from localStorage
const profileUsername = ref(localStorage.getItem('username') || '')

// First letter for avatar circle
const profileInitial = computed(() =>
  profileUsername.value ? profileUsername.value.charAt(0).toUpperCase() : '?'
)

// Navigate to Account page
function goToAccount() {
  router.push({ name: 'Account' })
}

// Logout + redirect to Login
function handleLogout() {
  try {
    logout()
  } catch (_) {}
  router.push({ name: 'Login' })
}
</script>

<!-- 🔹 Tutti gli stili sono ora in dashboard.css -->
