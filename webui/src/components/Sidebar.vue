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
 * - Custom sidebar that avoids conflicts with dashboard.css.
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

<style scoped>
/* 🔹 FIXED positioning come il template originale */
.app-sidebar {
  position: fixed;
  top: 60px; /* Sotto la navbar */
  bottom: 0;
  left: 0;
  width: 220px;
  border-right: 1px solid #ddd;
  background: #fff;
  display: flex;
  flex-direction: column;
  z-index: 50;
}

/* 🔹 Sezione menu - scrollabile */
.menu-section {
  flex: 1; /* Occupa tutto lo spazio disponibile */
  overflow-y: auto; /* Scrollabile se il menu è lungo */
  padding: 1.5rem 1rem;
}

/* 🔹 Nasconde la sidebar su mobile */
@media (max-width: 767.98px) {
  .app-sidebar {
    display: none;
  }
}

/* Menu links */
.nav-link {
  color: #333;
  padding: 6px 0;
}

.nav-link.router-link-active {
  font-weight: bold;
  color: #2470dc;
}

.nav-link:hover {
  color: #2470dc;
}

/* 🔹 Profile section - SEMPRE IN BASSO, NON SCROLLA */
.profile-box {
  border-top: 1px solid #e0e0e0;
  padding: 1rem;
  background: #fff;
  /* NON ha flex o margin-top auto perché la struttura flex lo mette automaticamente in fondo */
}

.profile-header {
  text-decoration: none;
  color: inherit;
}

.profile-header:hover .profile-avatar {
  opacity: 0.85;
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
  font-size: 1.1rem;
}
</style>
