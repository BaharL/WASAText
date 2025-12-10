<template>
  <div id="app">

    <!-- Top navbar always visible -->
    <AppNavbar />

    <div class="layout-row">
      <!-- Sidebar (hidden on Login page) -->
      <Sidebar
        v-if="showSidebar"
        class="sidebar-column d-none d-md-block"
      />

      <!-- Main content -->
      <main :class="mainClass" class="content-column">
        <RouterView />
      </main>
    </div>

  </div>
</template>

<script setup>
/**
 * App.vue — Root layout controller
 * -------------------------------
 * - Shows AppNavbar on every page.
 * - Shows Sidebar only when NOT on Login.
 * - Adds uniform spacing under the navbar via CSS.
 */

import { computed } from 'vue'
import { useRoute } from 'vue-router'

import AppNavbar from './components/AppNavbar.vue'
import Sidebar from './components/Sidebar.vue'

const route = useRoute()

// Hide sidebar only on Login page
const showSidebar = computed(() => route.name !== 'Login')

// Dynamic main width based on sidebar visibility
const mainClass = computed(() =>
  showSidebar.value
    ? 'with-sidebar'
    : 'no-sidebar'
)
</script>

<style scoped>
/* Whole layout placed under navbar */
.layout-row {
  display: flex;
  padding-top: 60px; /* height of AppNavbar */
}

/* Sidebar column */
.sidebar-column {
  width: 220px;
  border-right: 1px solid #ddd;
}

/* Main content column */
.content-column {
  flex-grow: 1;
  padding: 20px;
}

/* When sidebar is visible */
.with-sidebar {
  padding-left: 20px;
}

/* Full width on Login page */
.no-sidebar {
  width: 100%;
  padding: 20px;
}
</style>
