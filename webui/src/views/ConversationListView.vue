<template>
  <div class="conversation-list-root">
    <!-- Header: title + new chat -->
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h1 class="h2 mb-0">Conversations</h1>

      <!-- Toggle "New chat mode" instead of creating an empty group immediately -->
      <button
        type="button"
        class="btn btn-sm btn-primary"
        :disabled="loading"
        @click="toggleNewChat"
      >
        <span v-if="newChatMode">Cancel</span>
        <span v-else>+ New chat</span>
      </button>
    </div>

    <!-- New chat panel: search users + select members + create -->
    <div v-if="newChatMode" class="card mb-3">
      <div class="card-body">
        <div class="d-flex flex-column gap-2">
          <div class="fw-semibold">Create a new chat</div>

          <!-- Optional group name (recommended only when you select 2+ users) -->
          <input
            v-model.trim="newChatName"
            type="text"
            class="form-control"
            placeholder="Group name (optional)"
            :disabled="creatingChat"
          >

          <!-- User search bar -->
          <div class="input-group">
            <input
              v-model="userSearch"
              type="text"
              class="form-control"
              placeholder="Search users…"
              :disabled="searchingUsers || creatingChat"
              @keyup.enter="handleUserSearch"
            >
            <button
              type="button"
              class="btn btn-outline-secondary"
              :disabled="searchingUsers || creatingChat || !userSearch.trim()"
              @click="handleUserSearch"
            >
              <span v-if="searchingUsers">Searching…</span>
              <span v-else>Search</span>
            </button>
          </div>

          <small v-if="userSearchError" class="text-danger">
            {{ userSearchError }}
          </small>

          <!-- Selected users summary + create button -->
          <div class="d-flex align-items-center justify-content-between">
            <small class="text-muted">
              Selected: {{ selectedUsers.length }}
            </small>

            <button
              type="button"
              class="btn btn-sm btn-success"
              :disabled="creatingChat || selectedUsers.length === 0"
              @click="createConversationFromSelection"
            >
              <span v-if="creatingChat">Creating…</span>
              <span v-else>Create</span>
            </button>
          </div>

          <small v-if="createChatError" class="text-danger">
            {{ createChatError }}
          </small>

          <!-- Search results list with checkboxes -->
          <ul
            v-if="userResults.length > 0"
            class="list-group mt-2 user-results"
          >
            <li
              v-for="u in userResults"
              :key="u.identifier"
              class="list-group-item d-flex justify-content-between align-items-center"
            >
              <div class="d-flex flex-column">
                <span class="fw-semibold">{{ u.username || u.name }}</span>
                <small class="text-muted">{{ u.identifier }}</small>
              </div>

              <div class="form-check m-0">
                <input
                  class="form-check-input"
                  type="checkbox"
                  :id="'sel-' + u.identifier"
                  :checked="isSelected(u)"
                  :disabled="creatingChat"
                  @change="toggleSelected(u)"
                >
              </div>
            </li>
          </ul>

          <small class="text-muted">
            Tip: selecting 1 user creates a direct chat (you + them).
            Selecting 2+ users creates a group chat.
          </small>
        </div>
      </div>
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
      <ul v-else class="list-group">
        <li
          v-for="c in filteredConversations"
          :key="c.chatId ?? c.id"
          class="list-group-item list-group-item-action d-flex flex-column"
          @click="openConversation(c)"
        >
          <div class="d-flex align-items-center justify-content-between">
            <div class="fw-semibold">
              {{ getTitle(c) }}
            </div>

            <!-- Unread badge (only if backend provides it) -->
            <span
              v-if="getUnreadCount(c) > 0"
              class="badge bg-danger rounded-pill"
            >
              {{ getUnreadCount(c) }}
            </span>
          </div>

          <!-- Group members summary (only if backend provides them) -->
          <small
            v-if="isGroupChat(c) && getMembersLabel(c)"
            class="text-muted"
          >
            Members: {{ getMembersLabel(c) }}
          </small>

          <!-- Optional preview of last message -->
          <small v-if="c.lastMessage" class="text-muted">
            {{ c.lastMessage.text || '[' + c.lastMessage.kind + ']' }}
            · {{ c.lastMessage.createdAt }}
          </small>

          <!-- Optional label -->
          <small v-if="isGroupChat(c)" class="text-muted">
            Group chat
          </small>
        </li>
      </ul>
    </LoadingSpinner>
  </div>
</template>

<script setup>
/**
 * ConversationListView.vue
 *
 * Goals:
 * - Show conversations list (all/direct/groups depending on route)
 * - Provide a safe "New chat" flow:
 *    1) Search users
 *    2) Select members
 *    3) Create chat
 *
 * Important professor requirement:
 * - Do NOT create empty chats (no group with members: [])
 */

import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'

import ErrorMsg from '../components/ErrorMsg.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

import axios from '../services/axios.js'
import { getToken, listUsers } from '../services/api.js'

const router = useRouter()
const route = useRoute()

// Conversations state
const conversations = ref([])
const loading = ref(false)
const error = ref('')

// New chat creation state
const newChatMode = ref(false)
const newChatName = ref('')
const creatingChat = ref(false)
const createChatError = ref('')

// User search state
const userSearch = ref('')
const searchingUsers = ref(false)
const userSearchError = ref('')
const userResults = ref([])

// Selected users to be added to the new chat
const selectedUsers = ref([]) // user objects (identifier, username/name)

/**
 * Filter conversations based on current route:
 * - /conversations/direct  => non-group chats
 * - /conversations/groups  => group chats
 * - /                     => all
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

/**
 * Best-effort detection for group chats (backend field may differ).
 */
function isGroupChat(c) {
  if (typeof c.isGroup === 'boolean') return c.isGroup
  if (typeof c.group === 'boolean') return c.group
  if (typeof c.is_group === 'boolean') return c.is_group

  // Fallback: if members exist and are more than 2, assume group
  const members = c.members || c.participants || c.users
  if (Array.isArray(members) && members.length > 2) return true

  return false
}

/**
 * Unread count (only if backend provides it).
 */
function getUnreadCount(c) {
  return (c.unreadCount ?? c.unread ?? c.unread_messages ?? 0)
}

/**
 * Build a readable members label for group chats.
 */
function getMembersLabel(c) {
  const members = c.members || c.participants || c.users
  if (!Array.isArray(members) || members.length === 0) return ''

  const names = members
    .map(m => (typeof m === 'string' ? '' : (m.username || m.name || '')))
    .filter(Boolean)

  if (names.length > 0) {
    const first = names.slice(0, 3)
    const rest = names.length - first.length
    return rest > 0 ? `${first.join(', ')} +${rest}` : first.join(', ')
  }

  return `${members.length}`
}

/**
 * Build a title:
 * - Prefer explicit backend title/name/chatName
 * - Otherwise try to derive from members
 * - Final fallback: "Chat <id>"
 */
function getTitle(c) {
  if (c.title) return c.title
  if (c.name) return c.name
  if (c.chatName) return c.chatName

  const members = c.members || c.participants || c.users
  if (Array.isArray(members) && members.length > 0) {
    const names = members
      .map(m => (typeof m === 'string' ? '' : (m.username || m.name || '')))
      .filter(Boolean)

    if (names.length > 0) {
      if (!isGroupChat(c)) return names[0]
      const first = names.slice(0, 2)
      const rest = names.length - first.length
      return rest > 0 ? `${first.join(', ')} +${rest}` : first.join(', ')
    }
  }

  return `Chat ${c.id ?? c.chatId}`
}

/**
 * Load conversations:
 * Backend returns: { conversations: [...] }
 */
async function loadConversations() {
  loading.value = true
  error.value = ''

  try {
    const token = getToken()
    const headers = {}
    if (token) headers.Authorization = `Bearer ${token}`

    const response = await axios.get('/conversations', { headers })
    conversations.value = response.data?.conversations || []
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to load conversations.'
  } finally {
    loading.value = false
  }
}

/**
 * Toggle New chat mode and reset states when closing.
 */
function toggleNewChat() {
  newChatMode.value = !newChatMode.value
  createChatError.value = ''
  userSearchError.value = ''

  resetNewChatStateIfClosing()
}

/**
 * Reset UI-only new chat state when user closes the panel.
 */
function resetNewChatStateIfClosing() {
  if (!newChatMode.value) {
    userSearch.value = ''
    userResults.value = []
    selectedUsers.value = []
    newChatName.value = ''
  }
}

/**
 * Search users by keyword.
 * listUsers(term) must call GET /users?search=term (or your backend equivalent).
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
    userResults.value = Array.isArray(data) ? data : (data.users || [])
  } catch (e) {
    console.error(e)
    userSearchError.value = e.message || 'Failed to search users.'
  } finally {
    searchingUsers.value = false
  }
}

/**
 * Selection helpers
 */
function isSelected(u) {
  return selectedUsers.value.some(x => x.identifier === u.identifier)
}

function toggleSelected(u) {
  if (isSelected(u)) {
    selectedUsers.value = selectedUsers.value.filter(x => x.identifier !== u.identifier)
  } else {
    selectedUsers.value.push(u)
  }
}

/**
 * Create chat from selected users:
 * - 1 selected => direct chat
 * - 2+ selected => group chat
 *
 * We do NOT allow empty members[] (prof requirement).
 *
 * NOTE:
 * This code assumes backend endpoint POST /groups with body:
 * { name: string, members: string[] }
 */
async function createConversationFromSelection() {
  createChatError.value = ''

  if (selectedUsers.value.length === 0) {
    createChatError.value = 'Please select at least one user.'
    return
  }

  creatingChat.value = true

  try {
    const token = getToken()
    const headers = {}
    if (token) headers.Authorization = `Bearer ${token}`

    const memberIds = selectedUsers.value.map(u => u.identifier)

    // Group if you selected >= 2 users (you + 2 => at least 3 participants)
    const isGroup = memberIds.length >= 2

    // Build a good default name
    const defaultName = isGroup
      ? (newChatName.value.trim() || 'New group')
      : (selectedUsers.value[0].username || selectedUsers.value[0].name || 'Direct chat')

    const response = await axios.post(
      '/groups',
      { name: defaultName, members: memberIds },
      { headers }
    )

    const created = response.data

    // Refresh conversations list
    await loadConversations()

    // Close panel
    newChatMode.value = false
    resetNewChatStateIfClosing()

    // Navigate to the created conversation if backend returns chatId
    if (created?.chatId) {
      router.push({ name: 'Conversation', params: { chatId: created.chatId } })
    }
  } catch (e) {
    console.error(e)

    // Friendly message for typical backend validations
    const msg = (e.message || '').toLowerCase()
    if (msg.includes('members') || msg.includes('minimum') || msg.includes('at least')) {
      createChatError.value = 'Unable to create the chat. The backend may require more members.'
    } else {
      createChatError.value = e.message || 'Failed to create conversation.'
    }
  } finally {
    creatingChat.value = false
  }
}

/**
 * Open a conversation by navigating to ConversationView.
 */
function openConversation(c) {
  const chatId = c.chatId ?? c.id
  if (!chatId) return

  router.push({ name: 'Conversation', params: { chatId } })
}

onMounted(() => {
  loadConversations()
})
</script>

<style scoped>
.list-group-item {
  cursor: pointer;
}

/* Make search results not look like clickable conversations */
.user-results .list-group-item {
  cursor: default;
}

.conversation-list-root {
  padding-top: 0;
}
</style>
