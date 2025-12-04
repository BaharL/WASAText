<template>
  <div class="pt-3">
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h1 class="h2 mb-0">Conversations</h1>

      <button
        type="button"
        class="btn btn-sm btn-outline-secondary"
        :disabled="loading"
        @click="loadConversations"
      >
        <span v-if="loading">Reloading…</span>
        <span v-else>Reload</span>
      </button>
    </div>

    <!-- Error -->
    <ErrorMsg v-if="error" :msg="error" />

    <!-- Loading / content -->
    <LoadingSpinner :loading="loading">
      <!-- Empty state -->
      <div
        v-if="!loading && !error && conversations.length === 0"
        class="text-muted"
      >
        You do not have any conversations yet.
      </div>

      <!-- List -->
      <ul
        v-else
        class="list-group"
      >
        <li
          v-for="c in conversations"
          :key="c.chatId ?? c.id"
          class="list-group-item list-group-item-action d-flex flex-column"
          @click="openConversation(c)"
        >
          <div class="fw-semibold">
            {{ getTitle(c) }}
          </div>

          <small
            v-if="c.lastMessage"
            class="text-muted"
          >
            {{ c.lastMessage.text || '[' + c.lastMessage.kind + ']' }}
            · {{ c.lastMessage.createdAt }}
          </small>
        </li>
      </ul>
    </LoadingSpinner>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

// axios instance (dentro al componente come volevi tu)
import axios from '../services/axios.js'
// riutilizziamo getToken così l'header Authorization è consistente
import { getToken } from '../services/api.js'

const router = useRouter()
const conversations = ref([])
const loading = ref(false)
const error = ref('')

// Titolo robusto (diversi back-end usano field diversi)
function getTitle(c) {
  return (
    c.title ||
    c.name ||
    c.chatName ||
    `Chat ${c.id ?? c.chatId}`
  )
}

async function loadConversations() {
  loading.value = true
  error.value = ''

  try {
    const token = getToken()
    const headers = {}

    if (token) {
      headers.Authorization = `Bearer ${token}`
    }

    const response = await axios.get('/conversations', { headers })
    const data = response.data

    // Gestisce sia { conversations: [...] } che un array diretto
    conversations.value = Array.isArray(data)
      ? data
      : (data.conversations || [])
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to load conversations.'
  } finally {
    loading.value = false
  }
}

// Apri conversazione → la pagina ConversationView la abbiamo già creata
function openConversation(c) {
  const chatId = c.chatId ?? c.id
  if (!chatId) return

  router.push({
    name: 'Conversation',
    params: { chatId }
  })
}

onMounted(() => {
  loadConversations()
})
</script>

<style scoped>
.list-group-item {
  cursor: pointer;
}
</style>
