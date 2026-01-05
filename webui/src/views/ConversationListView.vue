<template>
  <div class="conversation-list-root">
    <!-- Header: title + new chat -->
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h1 class="h2 mb-0">Conversations</h1>

      <!-- Toggle New chat mode instead of creating an empty group immediately -->
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

    <!-- New chat panel (select users first) -->
    <div v-if="newChatMode" class="card mb-3">
      <div class="card-body">
        <div class="d-flex flex-column gap-2">
          <div class="fw-semibold">Create a new chat</div>

          <!-- Optional group name (used only if more than 1 selected) -->
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

          <!-- Selected users summary -->
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

          <!-- Search results with checkboxes -->
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
            Tip: selecting 1 user creates a direct chat (2 people: you + them).
            Selecting 2+ users creates a group.
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

          <small v-if="c.lastMessage" class="text-muted">
            {{ c.lastMessage.text || '[' + c.lastMessage.kind + ']' }}
            · {{ c.lastMessage.createdAt }}
          </small>

          <!-- Optional: show a tiny label if it is a group -->
          <small v-if="isGroupChat(c)" class="text-muted">
            Group chat
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

import axios from '../services/axios.js'
import { getToken, listUsers } from '../services/api.js'

const router = useRouter()
const route = useRoute()

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

// Selected users for creating a chat
const selectedUsers = ref([]) // array of user objects {identifier, username/name}

/**
 * Compute the conversations to show, based on current route:
 * - /conversations/direct     -> only non-group chats
 * - /conversations/groups     -> only group chats
 * - otherwise                -> all
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

/**
 * Detect whether conversation is a group.
 * Backend may use different field names, so we keep it robust.
 */
function isGroupChat(c) {
  if (typeof c.isGroup === 'boolean') return c.isGroup
  if (typeof c.group === 'boolean') return c.group
  if (typeof c.is_group === 'boolean') return c.is_group
  // Fallback: if it has a "members" array and it's > 2, treat as group
  const members = c.members || c.participants || c.users
  if (Array.isArray(members) && members.length > 2) return true
  return false
}

/**
 * Unread count:
 * We display it only if backend provides something like unreadCount/unread.
 * Otherwise we show nothing (we can't compute unread reliably without server support).
 */
function getUnreadCount(c) {
  return (
    c.unreadCount ??
    c.unread ??
    c.unread_messages ??
    0
  )
}

/**
 * Members label:
 * If backend provides members/participants with usernames, we show them.
 * If it provides only identifiers, we show count.
 */
function getMembersLabel(c) {
  const members = c.members || c.participants || c.users
  if (!Array.isArray(members) || members.length === 0) return ''

  // If members are objects with username/name
  const names = members
    .map(m => (typeof m === 'string' ? '' : (m.username || m.name || '')))
    .filter(Boolean)

  if (names.length > 0) {
    // Show up to 3 names, then +N more
    const first = names.slice(0, 3)
    const rest = names.length - first.length
    return rest > 0 ? `${first.join(', ')} +${rest}` : first.join(', ')
  }

  // Otherwise fallback to count
  return `${members.length}`
}

/**
 * Conversation title:
 * - If backend provides explicit title/name, use it.
 * - Otherwise, build a friendly fallback.
 */
function getTitle(c) {
  // Prefer explicit titles
  if (c.title) return c.title
  if (c.name) return c.name
  if (c.chatName) return c.chatName

  // If backend provides participants with usernames, build a title
  const members = c.members || c.participants || c.users
  if (Array.isArray(members) && members.length > 0) {
    const names = members
      .map(m => (typeof m === 'string' ? '' : (m.username || m.name || '')))
      .filter(Boolean)

    if (names.length > 0) {
      // Direct chat: show the first other name if possible
      if (!isGroupChat(c) && names.length >= 1) return names[0]
      // Group: show first names as summary
      const first = names.slice(0, 2)
      const rest = names.length - first.length
      return rest > 0 ? `${first.join(', ')} +${rest}` : first.join(', ')
    }
  }

  return `Chat ${c.id ?? c.chatId}`
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
    if (token) headers.Authorization = `Bearer ${token}`

    const response = await axios.get('/conversations', { headers })
    const data = response.data

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
 * Toggle New chat panel.
 * We no longer create an empty group immediately (prof requirement).
 */
function toggleNewChat() {
  newChatMode.value = !newChatMode.value
  createChatError.value = ''
  userSearchError.value = ''
  photoResetSearchIfClosing()
}

function photoResetSearchIfClosing() {
  if (!newChatMode.value) {
    userSearch.value = ''
    userResults.value = []
    selectedUsers.value = []
    newChatName.value = ''
  }
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
 * Create conversation from selected users.
 * - 1 selected user => direct chat (you + them)
 * - 2+ selected users => group chat
 *
 * IMPORTANT: We never send an empty members[] list.
 */
async function createConversationFromSelection() {
  createChatError.value = ''

  if (selectedUsers.value.length === 0) {
    createChatError.value = 'Please select at least one user.'
    return
  }

  try {
    creatingChat.value = true
    error.value = ''

    const token = getToken()
    const headers = {}
    if (token) headers.Authorization = `Bearer ${token}`

    const memberIds = selectedUsers.value.map(u => u.identifier)

    // If only 1 user selected, treat as direct chat.
    // If 2+ users selected, treat as group.
    const isGroup = memberIds.length >= 2

    // Build a reasonable name:
    // - direct: other user name
    // - group: input name or default
    const defaultName = isGroup
      ? (newChatName.value.trim() || 'New group')
      : (selectedUsers.value[0].username || selectedUsers.value[0].name || 'Direct chat')

    const response = await axios.post(
      '/groups',
      {
        name: defaultName,
        members: memberIds
      },
      { headers }
    )

    const data = response.data

    await loadConversations()

    // Close panel after creation
    newChatMode.value = false
    photoResetSearchIfClosing()

    if (data && data.chatId) {
      router.push({ name: 'Conversation', params:
