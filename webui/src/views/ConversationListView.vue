<template>
  <div class="pt-3">
    <!-- Header: title + buttons -->
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h1 class="h2 mb-0">Conversations</h1>

      <div class="btn-group">
        <!-- Reload conversations -->
        <button
          type="button"
          class="btn btn-sm btn-outline-secondary"
          :disabled="loading"
          @click="loadConversations"
        >
          <span v-if="loading">Reloading…</span>
          <span v-else>Reload</span>
        </button>

        <!-- Create a new group conversation -->
        <button
          type="button"
          class="btn btn-sm btn-primary"
          :disabled="loading"
          @click="createNewConversation"
        >
          + New chat
        </button>
      </div>
    </div>

    <!-- User search bar (search users by name) -->
    <div class="mb-3">
      <div class="input-group">
        <input
          v-model="userSearch"
          type="text"
          class="form-control"
          placeholder="Search users…"
          :disabled="searchingUsers"
        >
        <button
          type="button"
          class="btn btn-outline-secondary"
          :disabled="searchingUsers || !userSearch.trim()"
          @click="handleUserSearch"
        >
          <span v-if="searchingUsers">Searching…</span>
          <span v-else>Search</span>
        </button>
      </div>

      <!-- Search error -->
      <small
        v-if="userSearchError"
        class="text-danger"
      >
        {{ userSearchError }}
      </small>

      <!-- Search results list -->
      <ul
        v-if="userResults.length > 0"
        class="list-group mt-2 user-results"
      >
        <li
          v-for="u in userResults"
          :key="u.identifier"
          class="list-group-item d-flex justify-content-between align-items-center"
        >
          <span>
            {{ u.name }}
            <small class="text-muted">({{ u.identifier }})</small>
          </span>
          <button
            type="button"
            class="btn btn-sm btn-outline-primary"
            @click="startDirectChat(u)"
          >
            Chat
          </button>
        </li>
      </ul>
    </div>

    <!-- Error -->
    <ErrorMsg v-if="error" :msg="error" />

    <!-- Loading / content -->
    <LoadingSpinner :loading="loading">
      <!-- Empty state -->
      <div
        v-if="!loading && !error && filteredConversations.length === 0"
        class="text-muted"
      >
        You do not have any conversations yet.
      </div>

      <!-- Conversations list -->
      <ul
        v-else
        class="list-group"
      >
        <li
          v-for="c in filteredConversations"
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
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'

import ErrorMsg from '../components/ErrorMsg.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

// axios instance (inside the component, as you preferred)
import axios from '../services/axios.js'
// reuse getToken + listUsers for consistent Authorization header
import { getToken, listUsers } from '../services/api.js'

const router = useRouter()
const route = useRoute()

const conversations = ref([])
const loading = ref(false)
const error = ref('')

// User search state
const userSearch = ref('')
const searchingUsers = ref(false)
const userSearchError = ref('')
const userResults = ref([])

/**
 * Compute the conversations to show, based on current route:
 * - /conversations            -> all
 * - /conversations/direct     -> only non-group chats
 * - /conversations/groups     -> only group chats
 *
 * It expects each conversation to have "isGroup" from backend.
 */
const filteredConversations = computed(() => {
  if (route.name === 'DirectConversations') {
    return conversations.value.filter(c => c.isGroup === false)
  }

  if (route.name === 'GroupConversations') {
    return conversations.value.filter(c => c.isGroup === true)
  }

  return conversations.value
})

// Robust title (backend can use different field names)
function getTitle(c) {
  return (
    c.title ||
    c.name ||
    c.chatName ||
    `Chat ${c.id ?? c.chatId}`
  )
}

/**
 * Load all my conversations from backend.
 */
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

    // Handle both { conversations: [...] } and a plain array
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

/**
 * Create a new group conversation (no members yet) and open it.
 * Uses the /groups endpoint.
 */
async function createNewConversation() {
  try {
    loading.value = true
    error.value = ''

    const token = getToken()
    const headers = {}

    if (token) {
      headers.Authorization = `Bearer ${token}`
    }

    // For now: fixed name "New chat"
    const response = await axios.post(
      '/groups',
      { name: 'New chat', members: [] },
      { headers }
    )

    const data = response.data // expected: { chatId: ... }

    // Reload conversations
    await loadConversations()

    // If chatId exists, go to conversation page
    if (data && data.chatId) {
      router.push({
        name: 'Conversation',
        params: { chatId: data.chatId }
      })
    }
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to create conversation.'
  } finally {
    loading.value = false
  }
}

/**
 * Open a conversation in the ConversationView page.
 */
function openConversation(c) {
  const chatId = c.chatId ?? c.id
  if (!chatId) return

  router.push({
    name: 'Conversation',
    params: { chatId }
  })
}

/**
 * Search users by name using GET /users?search=...
 */
async function handleUserSearch() {
  const term = userSearch.value.trim()
  if (!term) {
    userResults.value = []
    userSearchError.value = ''
    return
  }

  searchingUsers.value = true
  userSearchError.value = ''

  try {
    const data = await listUsers(term)

    // Handle both { users: [...] } and a plain array
    userResults.value = Array.isArray(data)
      ? data
      : (data.users || [])
  } catch (e) {
    console.error(e)
    userSearchError.value = e.message || 'Failed to search users.'
  } finally {
    searchingUsers.value = false
  }
}

/**
 * Start a "direct" chat with a selected user.
 *
 * For now we create a group with that single user + me.
 * Backend: POST /groups with members = [ user.identifier ].
 * Depending on backend rules, this chat may be treated as "direct"
 * (2 members) or as "group"; our isGroup flag controls where it appears.
 */
async function startDirectChat(user) {
  try {
    loading.value = true
    error.value = ''

    const token = getToken()
    const headers = {}

    if (token) {
      headers.Authorization = `Bearer ${token}`
    }

    // Group name based on user name
    const response = await axios.post(
      '/groups',
      {
        name: user.name || 'Chat',
        members: [user.identifier]
      },
      { headers }
    )

    const data = response.data

    // Reload conversations so the new one appears
    await loadConversations()

    // Open the newly created chat if backend returns chatId
    if (data && data.chatId) {
      router.push({
        name: 'Conversation',
        params: { chatId: data.chatId }
      })
    }
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to start direct chat.'
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
  cursor: pointer;
}

/* Optional: style for user search results list */
.user-results .list-group-item {
  cursor: default;
}
</style>
