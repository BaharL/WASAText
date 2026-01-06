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
 * Responsibilities:
 * - Load and show conversations (all/direct/groups depending on route)
 * - Provide a safe "New chat" flow:
 *   1) Search users
 *   2) Select members
 *   3) Create chat (direct or group)
 *
 * Notes:
 * - The backend currently returns only:
 *   { id, title, isGroup, lastMessage }
 *   => No members list, no unread count.
 * - Therefore:
 *   - "Members:" label will remain empty unless backend provides membersPreview
 *   - Unread badge will stay 0 unless backend provides unreadCount
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

function isGroupChat(c) {
  return !!c.isGroup
}

function getUnreadCount(c) {
  // Backend does NOT provide unreadCount yet => always 0
  return (c.unreadCount ?? 0)
}

function getMembersLabel(c) {
  // Backend does NOT provide members yet => always empty
  // This is here only for future improvements.
  const members = c.members || c.participants || c.users
  if (!Array.isArray(members) || members.length === 0) return ''
  return `${members.length}`
}

function getTitle(c) {
  // Backend already returns "title"
  // For direct chats, this may still be "Chat <id>" until backend provides the other user's name.
  return c.title || c.name || `Chat ${c.id ?? c.chatId}`
}

/**
 * Clear local session (used when we detect 401 / stale token).
 */
function clearLocalSession() {
  localStorage.removeItem('token')
  localStorage.removeItem('username')
  localStorage.removeItem('photoUrl')
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
    // If backend returns 401 => session expired => force login
    const status = e?.response?.status
    if (status === 401) {
      clearLocalSession()
      router.replace({ name: 'Login' })
      return
    }

    console.error(e)
    error.value = e.message || 'Failed to load conversations.'
  } finally {
    loading.value = false
  }
}

function toggleNewChat() {
  newChatMode.value = !newChatMode.value
  createChatError.value = ''
  userSearchError.value = ''

  if (!newChatMode.value) {
    userSearch.value = ''
    userResults.value = []
    selectedUsers.value = []
    newChatName.value = ''
  }
}

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
 * Create a conversation from selected users.
 *
 * IMPORTANT:
 * - If you selected 1 user => this should create a DIRECT chat.
 * - If you selected 2+ users => this should create a GROUP chat.
 *
 * The exact endpoint depends on your backend:
 * - group creation might be POST /groups
 * - direct creation might be POST /conversations or POST /direct
 *
 * TODO: adjust endpoints to match your backend.
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

    // Decide direct vs group
    const shouldCreateGroup = memberIds.length >= 2

    let created = null

    if (shouldCreateGroup) {
      // GROUP: you + at least 2 other users
      const name = newChatName.value.trim() || 'New group'

      created = (await axios.post(
        '/groups',
        { name, members: memberIds },
        { headers }
      )).data
    } else {
      // DIRECT: you + 1 other user
      // TODO: replace '/direct' with the real endpoint in your backend.
      created = (await axios.post(
        '/direct',
        { member: memberIds[0] },
        { headers }
      )).data
    }

    await loadConversations()

    newChatMode.value = false
    userSearch.value = ''
    userResults.value = []
    selectedUsers.value = []
    newChatName.value = ''

    if (created?.chatId) {
      router.push({ name: 'Conversation', params: { chatId: created.chatId } })
    }
  } catch (e) {
    console.error(e)
    const status = e?.response?.status

    if (status === 401) {
      clearLocalSession()
      router.replace({ name: 'Login' })
      return
    }

    createChatError.value = e?.response?.data?.message || e.message || 'Failed to create conversation.'
  } finally {
    creatingChat.value = false
  }
}

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
