<template>
  <div id="app">
    <!-- Top navbar is always visible -->
    <AppNavbar />

    <div class="container-fluid">
      <div class="row">
        <!-- Sidebar is hidden on Login page -->
        <Sidebar
          v-if="showSidebar"
          class="col-md-2 col-lg-2 d-none d-md-block p-0"
        />

        <!-- Main content area; full width when sidebar is hidden -->
        <main :class="mainClass">
          <RouterView />
        </main>
      </div>
    </div>
  </div>
</template>

<script setup>
/**
 * App.vue is the layout controller of the whole app.
 * - Always shows the top navbar.
 * - Shows the sidebar only if the current route is not "Login".
 * - RouterView renders pages inside the main content area.
 */
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import AppNavbar from './components/AppNavbar.vue'
import Sidebar from './components/Sidebar.vue'

const route = useRoute()

// We hide sidebar on login page
const showSidebar = computed(() => route.name !== 'Login')

// Dynamic main column class when sidebar is present/absent
const mainClass = computed(() =>
  showSidebar.value
    ? 'col-md-10 col-lg-10 ms-sm-auto px-md-4'
    : 'col-12 px-0'
)
</script>

<style scoped>
/* Layout styles are mostly handled by dashboard.css / main.css */
</style>
