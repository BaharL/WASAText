<template>
  <div id="app">
    <!-- Top navbar (always visible) -->
    <AppNavbar />

    <div class="container-fluid">
      <div class="row">
        <!-- Sidebar visible only when user is not on Login page -->
        <Sidebar
          v-if="showSidebar"
          class="col-md-2 col-lg-2"
        />

        <!-- Page content -->
        <main :class="showSidebar ? 'col-md-10 col-lg-10' : 'col-12'">
          <RouterView />
        </main>
      </div>
    </div>
  </div>
</template>

<script setup>
/**
 * App.vue controls the global layout:
 * - Always shows the top AppNavbar.
 * - Shows the Sidebar only when current route is not "Login".
 * - Renders the active page via RouterView.
 */

import { computed } from 'vue'
import { useRoute } from 'vue-router'

import AppNavbar from './components/AppNavbar.vue'
import Sidebar from './components/Sidebar.vue'

const route = useRoute()

// True if we should display the sidebar (not on Login page)
const showSidebar = computed(() => route.name !== 'Login')
</script>

<style>
/* You can add global layout tweaks here if needed */
</style>
