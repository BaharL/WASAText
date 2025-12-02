<template>
  <div class="pt-3">
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h1 class="h2 mb-0">Conversations</h1>
      <button
        type="button"
        class="btn btn-sm btn-outline-secondary"
        @click="loadConversations"
        :disabled="loading"
      >
        Reload
      </button>
    </div>

    <ErrorMsg v-if="error" :msg="error" />

    <LoadingSpinner :loading="loading">
      <div v-if="!loading && !error && conversations.length === 0">
        <p class="text-muted mb-0">You do not have any conversations yet.</p>
      </div>

      <ul
        v-else
        class="list-group"
      >
        <li
          v-for="conv in conversations"
          :key="conv.id"
          class="list-group-item d-flex flex-column"
        >
          <div class="fw-semibold">
            {{ conv.title || ('Chat ' + conv.id) }}
          </div>

          <small
            v-if="conv.lastMessage"
            class="text-muted"
          >
            {{ conv.lastMessage.text || '[' + conv.lastMessage.kind + ']' }}
            · {{ conv.lastMessage.createdAt }}
          </small>
        </li>
      </ul>
    </LoadingSpinner>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getMyConversations } from '../services/api.js'

const conversations = ref([])
const loading = ref(false)
const error = ref('')

async function loadConversations() {
  loading.value = true
  error.value = ''

  try {
    const data = await getMyConversations()
    conversations.value = data.conversations ?? []
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to load conversations.'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadConversations()
})
</script>

<style scoped>
.list-group-item {
  cursor: default;
}
</style>
