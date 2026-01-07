<template>
  <div class="conversation-list-root">
    <!-- Header -->
    <div class="d-flex justify-content-between align-items-center mb-3">
      <div>
        <h1 class="h2 mb-0">Conversations</h1>
        <small class="text-muted">
          {{ routeLabel }}
        </small>
      </div>

      <!-- Actions -->
      <div class="d-flex gap-2">
        <button
          type="button"
          class="btn btn-sm btn-outline-secondary"
          :disabled="loading"
          @click="loadConversations"
          title="Refresh"
        >
          ↻
        </button>

        <button
          type="button"
          class="btn btn-sm btn-primary"
          :disabled="loading"
          @click="toggleNewChat"
        >
          <span v-if="newChatMode">Close</span>
          <span v-else>+ New chat</span>
        </button>
      </div>
    </div>

    <!-- New chat overlay/panel -->
    <div v-if="newChatMode" class="newchat-panel card mb-3">
      <div class="card-body">
        <!-- Top row: mode switch -->
        <div class="d-flex align-items-center justify-content-between mb-2">
          <div class="fw-semibold">
            Start a new chat
          </div>

          <div class="btn-group btn-group-sm" role="group" aria-label="New chat mode">
            <button
              type="button"
              class="btn"
              :class="newChatKind === 'direct' ? 'btn-primary' : 'btn-outline-primary'"
              :disabled="creating"
              @click="setNewChatKind('direct')"
            >
              Direct
            </button>
            <button
              type="button"
              class="btn"
              :class="newChatKind === 'group' ? 'btn-primary' : 'btn-outline-primary'"
              :disabled="creating"
              @click="setNewChatKind('group')"
            >
              Group
            </button>
          </div>
        </div>

        <small class="text-muted d-block mb-3">
          <span v-if="newChatKind === 'direct'">
            Search a user and click on them to start a direct chat.
          </span>
          <span v-else>
            Select 2+ users, choose a group name, then create the group.
          </span>
        </small>

        <!-- Group name (only for groups) -->
        <div v-if="newChatKind === 'group'" class="mb-2">
          <input
            v-model.trim="groupName"
            type="text"
            class="form-control"
            placeholder="Group name"
            :disabled="creating"
          >
          <small class="text-muted">
            Tip: choose a clear name (e.g. “Project WASA”).
          </small>
        </div>

        <!-- Search bar -->
        <div class="input-group mb-2">
          <span class="input-group-text">🔎</span>
          <input
            v-model="userSearch"
            type="text"
            class="form-control"
            placeholder="Search users…"
            :disabled="creating || searching"
            @input="debouncedSearch"
            @keyup.enter="handleUserSearch"
          >
          <button
            type="button"
            class="btn btn-outline-secondary"
            :disabled="creating || searching || !userSearch.trim()"
            @click="handleUserSearch"
          >
            <span v-if="searching">Searching…</span>
            <span v-else>Search</span>
          </button>
        </div>

        <small v-if="userSearchError" class="text-danger d-block mb-2">
          {{ userSearchError }}
        </small>

        <!-- Selected chips (group only) -->
        <div v-if="newChatKind === 'group' && selectedUsers.length > 0" class="selected-chips mb-2">
          <span class="me-2 text-muted small">Selected:</span>
          <button
            v-for="u in selectedUsers"
            :key="'chip-' + u.identifier"
            type="button"
            class="chip"
            :disabled="creating"
            @click="removeSelected(u)"
            title="Remove"
          >
            {{ u.username || u.name || 'user' }}
            <span class="chip-x">×</span>
          </button>
        </div>

        <!-- Results list -->
        <div v-if="userResults.length > 0" class="list-group user-results">
          <button
            v-for="u in userResults"
            :key="u.identifier"
            type="button"
            class="list-group-item list-group-item-action d-flex justify-content-between align-items-center"
            :disabled="creating"
            @click="handleUserClick(u)"
          >
            <div class="d-flex flex-column text-start">
              <span class="fw-semibold">{{ u.username || u.name }}</span>
              <small class="text-muted">{{ u.identifier }}</small>
            </div>

            <!-- Right side -->
            <div class="d-flex align-items-center gap-2">
              <!-- Group selection toggle -->
              <template v-if="newChatKind === 'group'">
                <span v-if="isSelected(u)" class="badge text-bg-success">Selected</span>
                <span v-else class="badge text-bg-light">Tap to add</span>
              </template>

              <!-- Direct hint -->
              <template v-else>
                <span class="badge text-bg-light">Tap to chat</span>
              </template>
            </div>
          </button>
        </div>

        <!-- Empty results -->
        <div v-else class="text-muted small mt-2">
          <span v-if="userSearch.trim() && !searching">No users found.</span>
          <span v-else>Type to search users.</span>
        </div>

        <!-- Action row (group only) -->
        <div v-if="newChatKind === 'group'" class="d-flex justify-content-between align-items-center mt-3">
          <small class="text-muted">
            Members selected: {{ selectedUsers.length }}
          </small>

          <button
            type="button"
            class="btn btn-sm btn-success"
            :disabled="creating || selectedUsers.length < 2"
            @click="createGroupFromSelection"
          >
            <span v-if="creating">Creating…</span>
            <span v-else>Create group</span>
          </button>
        </div>

        <small v-if="createError" class="text-danger d-block mt-2">
          {{ createError }}
        </small>
      </div>
    </div>

    <!-- Error -->
    <ErrorMsg v-if="error" :msg="error" />

    <!-- Conversations list -->
    <LoadingSpinner :loading="loading">
      <div v-if="!loading && !error && filteredConversations.length === 0" class="text-muted">
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
              {{ c.title || `Chat ${c.id}` }}
            </div>
            <small class="text-muted">{{ c.isGroup ? '👥' : '👤' }}</small>
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
 * ConversationListView.vue (WhatsApp-like UX)
 *
 * - Direct mode:
 *   Search users -> CLICK a user -> createDirect(otherId) -> open chat
 *
 * - Group mode:
 *   Search users -> CLICK to select -> choose name -> createGroup(name, members) -> open chat
 *
 * IMPORTANT:
 * - We call ONLY services/api.js functions.
 * - axios interceptor handles 401 globally, so we don't duplicate logout/redirect here.
 */

import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

import ErrorMsg from '../components/ErrorMsg.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

import {
  getMyConversations,
  listUsers,
  createDirect,
  createGroup
} from '../services/api.js'

const router = useRouter()
const route = useRoute()

// Conversations state
const conversations = ref([])
const loading = ref(false)
const error = ref('')

// New chat state
const newChatMode = ref(false)
const newChatKind = ref('direct') // 'direct' | 'group'
const groupName = ref('')
const creating = ref(false)
const createError = ref('')

// User search state
const userSearch = ref('')
const searching = ref(false)
const userSearchError = ref('')
const userResults = ref([])

// Group selection
const selectedUsers = ref([])

/**
 * Route label helper (nice UI hint)
 */
const routeLabel = computed(() => {
  if (route.name === 'DirectConversations') return 'Showing direct chats only'
  if (route.name === 'GroupConversations') return 'Showing group chats only'
  return 'Showing all chats'
})

/**
 * Filter conversations based on route:
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
  createError.value = ''
  userSearchError.value = ''

  if (!newChatMode.value) {
    resetNewChatState()
  }
}

function setNewChatKind(kind) {
  newChatKind.value = kind
  createError.value = ''
  userSearchError.value = ''

  // When switching mode, keep search term/results but reset selections/name
  selectedUsers.value = []
  groupName.value = ''
}

function resetNewChatState() {
  userSearch.value = ''
  userResults.value = []
  selectedUsers.value = []
  groupName.value = ''
  searching.value = false
  userSearchError.value = ''
  createError.value = ''
}

/**
 * Simple debounce (no external libs).
 * We call search automatically after 350ms of inactivity.
 */
let debounceTimer = null
function debouncedSearch() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    // only auto-search if user typed at least 2 chars
    if (userSearch.value.trim().length >= 2) {
      handleUserSearch()
    } else {
      userResults.value = []
      userSearchError.value = ''
    }
  }, 350)
}

async function handleUserSearch() {
  const term = userSearch.value.trim()
  if (!term) {
    userResults.value = []
    userSearchError.value = ''
    return
  }

  searching.value = true
  userSearchError.value = ''

  try {
    const data = await listUsers(term)
    userResults.value = Array.isArray(data) ? data : (data.users || [])
  } catch (e) {
    console.error(e)
    userSearchError.value = e.message || 'Failed to search users.'
  } finally {
    searching.value = false
  }
}

function isSelected(u) {
  return selectedUsers.value.some(x => x.identifier === u.identifier)
}

function removeSelected(u) {
  selectedUsers.value = selectedUsers.value.filter(x => x.identifier !== u.identifier)
}

/**
 * Handle click on a user result:
 * - direct mode: create direct chat immediately
 * - group mode: toggle selection
 */
async function handleUserClick(u) {
  if (creating.value) return

  if (newChatKind.value === 'group') {
    // Toggle selection
    if (isSelected(u)) {
      removeSelected(u)
    } else {
      selectedUsers.value.push(u)
    }
    return
  }

  // Direct: create chat immediately
  createError.value = ''
  creating.value = true
  try {
    const created = await createDirect(u.identifier)
    await loadConversations()

    // Close panel and open chat
    newChatMode.value = false
    resetNewChatState()

    if (created?.chatId) {
      router.push({ name: 'Conversation', params: { chatId: created.chatId } })
    }
  } catch (e) {
    console.error(e)
    createError.value = e.message || 'Failed to create direct chat.'
  } finally {
    creating.value = false
  }
}

async function createGroupFromSelection() {
  createError.value = ''

  if (selectedUsers.value.length < 2) {
    createError.value = 'Select at least 2 users to create a group.'
    return
  }

  const name = groupName.value.trim()
  if (!name) {
    createError.value = 'Group name is required.'
    return
  }

  creating.value = true
  try {
    const memberIds = selectedUsers.value.map(u => u.identifier)
    const created = await createGroup(name, memberIds)
    await loadConversations()

    // Close panel and open chat
    newChatMode.value = false
    resetNewChatState()

    if (created?.chatId) {
      router.push({ name: 'Conversation', params: { chatId: created.chatId } })
    }
  } catch (e) {
    console.error(e)
    createError.value = e.message || 'Failed to create group.'
  } finally {
    creating.value = false
  }
}

function openConversation(c) {
  if (!c?.id) return
  router.push({ name: 'Conversation', params: { chatId: c.id } })
}

onMounted(() => {
  loadConversations()
})
</script>

<style scoped>
/* Small layout polish */
.conversation-list-root {
  padding-top: 0;
}

/* Make the panel feel like an overlay card */
.newchat-panel {
  border-radius: 14px;
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.08);
}

/* Chips (selected users) */
.selected-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  align-items: center;
}

.chip {
  border: 1px solid rgba(0,0,0,0.15);
  background: #fff;
  padding: 0.25rem 0.55rem;
  border-radius: 999px;
  font-size: 0.85rem;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.chip-x {
  font-weight: 700;
  opacity: 0.6;
}

.user-results .list-group-item {
  text-align: left;
}
</style>
