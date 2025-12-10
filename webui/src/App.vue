<template>
  <div id="app">
    <!-- Top navbar sempre visibile -->
    <AppNavbar />

    <!-- Corpo dell'app: push sotto la navbar -->
    <div class="app-body container-fluid">
      <div class="row">
        <!-- Sidebar: nascosta sulla pagina Login -->
        <Sidebar
          v-if="showSidebar"
          class="col-md-2 col-lg-2 d-none d-md-block p-0"
        />

        <!-- Contenuto principale -->
        <main :class="mainClass">
          <RouterView />
        </main>
      </div>
    </div>
  </div>
</template>

<script setup>
/**
 * App.vue è il layout root:
 * - Navbar sempre visibile
 * - Sidebar nascosta solo su route "Login"
 * - Spazio sotto la navbar per non sovrapporre il contenuto
 */

import { computed } from 'vue'
import { useRoute } from 'vue-router'

import AppNavbar from './components/AppNavbar.vue'
import Sidebar from './components/Sidebar.vue'

const route = useRoute()

// Sidebar visibile solo se NON siamo su Login
const showSidebar = computed(() => route.name !== 'Login')

// Colonna principale: larghezza diversa se la sidebar c'è o no
const mainClass = computed(() =>
  showSidebar.value
    ? 'col-md-10 col-lg-10 ms-sm-auto px-md-4'
    : 'col-12 px-0'
)
</script>

<style scoped>
/* Spinge tutto il layout sotto la navbar (alta ~60px) */
.app-body {
  padding-top: 1.2px;
}

/* Il resto del layout è gestito da dashboard.css / main.css */
</style>
