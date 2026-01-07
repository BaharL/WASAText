<template>
  <div class="conversation-list-root">
    <!-- Header: title + new chat -->
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h1 class="h2 mb-0">Conversations</h1>

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

    <!-- New chat panel -->
    <div v-if="newChatMode" class="card mb-3">
      <div class="card-body">
        <div class="d-flex flex-column gap-2">
          <div class="fw-semibold">Start a new chat</div>

          <!-- Group name only matters if 2+ users -->
          <input
            v-model.trim="newChatName"
            type="text"
            class="form-control"
            placeholder="Group name (only for groups)"
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

          <!-- Search results -->
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
            Tip: select 1 user for a direct chat, 2+ users for a group.
          </small>
        </div>
      </div>
    </div>

    <!-- Error -->
    <ErrorMsg v-if="error" :msg="error" />

    <!-- Loading / content -->
    <LoadingSpinner :loading="loading">
      <div
        v-if="!loading && !error && filteredConversations.length === 0"
        class="text-muted"
      >
        You do not have any conversations yet.
      </div>

      <ul v-else class="list-group">
        <li
          v-for="c in filteredConversations"
          :key="c.id"
          class="list-group-item list-group-item-action d-flex flex-column"
          @click="openConversation(c)"
        >
          <div class="d-flex align-items-center justify-content-between">
            <div class="fw-semibold">
              {{ getTitle(c) }}
            </div>
          </div>

          <small v-if="c.lastMessage" class="text-muted">
            {{ c.lastMessage.text || '[' + c.lastMessage.kind + ']' }}
            · {{ c.lastMessage.createdAt }}
          </small>

          <small v-if="c.isGroup" class="text-muted">Group chat</small>
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
 * - Provide a WhatsApp-like "New chat" flow:
 *   1) Search users
 *   2) Select members
 *   3) Create chat:
 *      - 1 member => POST /direct (creates or returns existing)
 *      - 2+ members => POST /groups
 *
 * Important:
 * - Do NOT call axios directly here for auth logic.
 *   Axios interceptor (services/axios.js) already handles 401 globally.
 */

import { ref, onMounted, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'

import ErrorMsg from '../components/ErrorMsg.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

import {
  getMyConversations,
  listUsers,
  createGroup,
  createDirect
} from '../services/api.js'

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

// Selected users for the new chat
const selectedUsers = ref([]) // [{ identifier, username/name }]

/**
 * Filter conversations based on current route:
 * - /conversations/direct => isGroup false
 * - /conversations/groups => isGroup true
 * - / => all
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

function getTitle(c) {
  return c.title || `Chat ${c.id}`
}

/**
 * Load conversations.
 * Backend returns: { conversations: [...] }
 */
async function loadConversations() {
  loading.value = true
  error.value = ''

  try {
    const data = await getMyConversations()
    conversations.value = data?.conversations || []
  } catch (e) {
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
    // depending on backend it might return { users: [] } or [].
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
 * Create a conversation from selected users:
 * - 1 selected => direct chat => createDirect(otherId)
 * - 2+ selected => group => createGroup(name, members)
 */
async function createConversationFromSelection() {
  createChatError.value = ''

  if (selectedUsers.value.length === 0) {
    createChatError.value = 'Please select at least one user.'
    return
  }

  creatingChat.value = true

  try {
    const memberIds = selectedUsers.value.map(u => u.identifier)

    let created = null

    if (memberIds.length >= 2) {
      const name = newChatName.value.trim() || 'New group'
      created = await createGroup(name, memberIds)
    } else {
      created = await createDirect(memberIds[0])
    }

    await loadConversations()

    // Reset UI
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
    createChatError.value = e.message || 'Failed to create conversation.'
  } finally {
    creatingChat.value = false
  }
}

function openConversation(c) {
  if (!c?.id) return
  router.push({ name: 'Conversation', params: { chatId: c.id } })
}

onMounted(() => {
  loadConversations()
})

/**
 * Optional: if user switches between Home/Direct/Groups,
 * we keep the same conversations list already loaded, no reload needed.
 * But if you want a refresh each time, uncomment:
 */
// watch(() => route.name, () => loadConversations())
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
