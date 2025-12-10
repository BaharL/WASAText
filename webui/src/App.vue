<template>
  <div id="app">

    <!-- Top navbar is always visible -->
    <AppNavbar />

    <!-- Spacer: pushes sidebar + content below the navbar -->
    <div style="height: 60px;"></div>

    <div class="container-fluid">
      <div class="row">

        <!-- Sidebar (hidden on Login page) -->
        <Sidebar
          v-if="showSidebar"
          class="col-md-2 col-lg-2 d-none d-md-block p-0"
        />

        <!-- Main content area -->
        <main :class="mainClass">
          <RouterView />
        </main>

      </div>
    </div>

  </div>
</template>

<script setup>
/**
 * App.vue is the root layout controller.
 * - Always shows the top navbar.
 * - Shows the sidebar only when NOT on Login page.
 * - Adds a spacing under navbar so the layout aligns correctly.
 */

import { computed } from 'vue'
import { useRoute } from 'vue-router'

import AppNavbar from './components/AppNavbar.vue'
import Sidebar from './components/Sidebar.vue'

const route = useRoute()

// Sidebar only visible if NOT on Login page
const showSidebar = computed(() => route.name !== 'Login')

// Dynamic main column depending on sidebar visibility
const mainClass = computed(() =>
  showSidebar.value
    ? 'col-md-10 col-lg-10 ms-sm-auto px-md-4'
    : 'col-12 px-0'
)
</script>

<style scoped>
/* Layout styles come mostly from dashboard.css and main.css */
</style>
