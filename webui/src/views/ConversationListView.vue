<template>
  <div class="conversation-list-root">
    <!-- Header -->
    <div class="d-flex justify-content-between align-items-center mb-3">
      <div>
        <h1 class="h2 mb-0">Conversations</h1>
        <small class="text-muted">{{ routeLabel }}</small>
      </div>

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

    <!-- New chat panel -->
    <div v-if="newChatMode" class="newchat-panel card mb-3">
      <div class="card-body">
        <div class="d-flex align-items-center justify-content-between mb-2">
          <div class="fw-semibold">Start a new chat</div>

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

        <div v-if="newChatKind === 'group'" class="mb-2">
          <input
            v-model.trim="groupName"
            type="text"
            class="form-control"
            placeholder="Group name"
            :disabled="creating"
          >
          <small class="text-muted">Tip: choose a clear name (e.g. “Project WASA”).</small>
        </div>

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

            <div class="d-flex align-items-center gap-2">
              <template v-if="newChatKind === 'group'">
                <span v-if="isSelected(u)" class="badge text-bg-success">Selected</span>
                <span v-else class="badge text-bg-light">Tap to add</span>
              </template>
              <template v-else>
                <span class="badge text-bg-light">Tap to chat</span>
              </template>
            </div>
          </button>
        </div>

        <div v-else class="text-muted small mt-2">
          <span v-if="userSearch.trim() && !searching">No users found.</span>
          <span v-else>Type to search users.</span>
        </div>

        <div v-if="newChatKind === 'group'" class="d-flex justify-content-between align-items-center mt-3">
          <small class="text-muted">Members selected: {{ selectedUsers.length }}</small>

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

    <ErrorMsg v-if="error" :msg="error" />

    <LoadingSpinner :loading="loading">
      <div v-if="!loading && !error && filteredConversations.length === 0" class="text-muted">
        You do not have any conversations yet.
      </div>

      <ul v-else class="list-group">
        <li
          v-for="c in filteredConversations"
          :key="c.id"
          class="list-group-item list-group-item-action"
          @click="openConversation(c)"
        >
          <div class="d-flex align-items-center justify-content-between">
            <!-- Left: avatar + title -->
            <div class="d-flex align-items-center gap-2">
              <div class="conv-avatar">
                <img
                  v-if="getConversationPhoto(c)"
                  :src="toImgSrc(getConversationPhoto(c))"
                  alt="avatar"
                  class="conv-avatar-img"
                >
                <span v-else class="conv-avatar-fallback">
                  {{ getConversationAvatarLabel(c) }}
                </span>
              </div>

              <div class="d-flex flex-column">
                <div class="fw-semibold">
                  {{ getConversationTitle(c) }}
                </div>

                <small v-if="c.lastMessage" class="text-muted">
                  {{ c.lastMessage.text || '[' + c.lastMessage.kind + ']' }}
                  · {{ c.lastMessage.createdAt }}
                </small>
              </div>
            </div>

            <!-- Right: type icon -->
            <small class="text-muted">{{ c.isGroup ? '👥' : '👤' }}</small>
          </div>
        </li>
      </ul>
    </LoadingSpinner>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'

import ErrorMsg from '../components/ErrorMsg.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

import {
  getMyConversations,
  listUsers,
  createDirect,
  createGroup,
  getDirectChatTitle,
  saveDirectChatTitle
} from '../services/api.js'

const router = useRouter()
const route = useRoute()

const conversations = ref([])
const loading = ref(false)
const error = ref('')

const newChatMode = ref(false)
const newChatKind = ref('direct')
const groupName = ref('')
const creating = ref(false)
const createError = ref('')

const userSearch = ref('')
const searching = ref(false)
const userSearchError = ref('')
const userResults = ref([])

const selectedUsers = ref([])

/* ---------------------------
 * Route label + filtering
 * --------------------------- */
const routeLabel = computed(() => {
  if (route.name === 'DirectConversations') return 'Showing direct chats only'
  if (route.name === 'GroupConversations') return 'Showing group chats only'
  return 'Showing all chats'
})

const filteredConversations = computed(() => {
  if (route.name === 'DirectConversations') {
    return conversations.value.filter(c => c.isGroup === false)
  }
  if (route.name === 'GroupConversations') {
    return conversations.value.filter(c => c.isGroup === true)
  }
  return conversations.value
})

/* ---------------------------
 * Helpers
 * --------------------------- */
function toImgSrc(url) {
  if (!url) return ''

  // absolute
  if (url.startsWith('http://') || url.startsWith('https://')) return url

  // if backend returns "/uploads/..", in dev/prod we must go through /api
  if (url.startsWith('/uploads/')) return `/api/v1${url}`

  // if already returned with "/v1/..", proxy needs /api prefix
  if (url.startsWith('/v1/')) return `/api${url}`

  // fallback
  return url.startsWith('/') ? url : `/${url}`
}

function getConversationTitle(c) {
  const base = c.title || c.name || c.chatName || `Chat ${c.id}`

  if (c.isGroup === false) {
    const saved = getDirectChatTitle(c.id)
    if (saved) return saved
  }
  return base
}

function getConversationPhoto(c) {
  return c.photoUrl || c.photo_url || c.photo || ''
}

/**
 * Build "A", "A2", ... when multiple conversations share the same title.
 */
const titleCounts = computed(() => {
  const map = new Map()
  for (const c of filteredConversations.value) {
    const t = getConversationTitle(c).trim()
    map.set(t, (map.get(t) || 0) + 1)
  }
  return map
})

const titleIndexById = computed(() => {
  const seen = new Map()
  const idx = new Map()
  for (const c of filteredConversations.value) {
    const t = getConversationTitle(c).trim()
    const n = (seen.get(t) || 0) + 1
    seen.set(t, n)
    idx.set(c.id, n)
  }
  return idx
})

function getConversationAvatarLabel(c) {
  const title = getConversationTitle(c).trim()
  const first = title ? title.charAt(0).toUpperCase() : '?'

  const count = titleCounts.value.get(title) || 0
  if (count <= 1) return first

  const n = titleIndexById.value.get(c.id) || 1
  return `${first}${n}`
}

/* ---------------------------
 * Load conversations
 * --------------------------- */
async function loadConversations({ silent = false } = {}) {
  if (!silent) {
    loading.value = true
  }
  error.value = ''

  try {
    const data = await getMyConversations()
    conversations.value = data?.conversations || []
  } catch (e) {
    console.error(e)
    if (!silent) error.value = e.message || 'Failed to load conversations.'
  } finally {
    if (!silent) loading.value = false
  }
}

/* ---------------------------
 * ✅ AUTREFRESH (Polling)
 * --------------------------- */
let pollTimer = null

function startPolling() {
  stopPolling()
  pollTimer = setInterval(async () => {
    // se stai creando una chat o cercando utenti, non stressiamo la rete
    if (loading.value || creating.value || searching.value) return
    // mentre sei nel pannello "new chat" evitiamo refresh (così non “salta” la UI)
    if (newChatMode.value) return

    await loadConversations({ silent: true })
  }, 8000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

/* ---------------------------
 * New chat panel logic
 * --------------------------- */
function toggleNewChat() {
  newChatMode.value = !newChatMode.value
  createError.value = ''
  userSearchError.value = ''
  if (!newChatMode.value) resetNewChatState()
}

function setNewChatKind(kind) {
  newChatKind.value = kind
  createError.value = ''
  userSearchError.value = ''
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

let debounceTimer = null
function debouncedSearch() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    if (userSearch.value.trim().length >= 2) handleUserSearch()
    else {
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

async function handleUserClick(u) {
  if (creating.value) return

  if (newChatKind.value === 'group') {
    if (isSelected(u)) removeSelected(u)
    else selectedUsers.value.push(u)
    return
  }

  // DIRECT
  createError.value = ''
  creating.value = true
  try {
    const created = await createDirect(u.identifier)

    if (created?.chatId) {
      const display = u.username || u.name || 'Direct'
      saveDirectChatTitle(created.chatId, display)
    }

    await loadConversations()

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

onMounted(async () => {
  await loadConversations()
  startPolling()
})

onBeforeUnmount(() => {
  stopPolling()
})
</script>

<style scoped>
.conversation-list-root { padding-top: 0; }
.newchat-panel { border-radius: 14px; box-shadow: 0 10px 24px rgba(0,0,0,0.08); }
.selected-chips { display:flex; flex-wrap:wrap; gap:0.35rem; align-items:center; }
.chip { border:1px solid rgba(0,0,0,0.15); background:#fff; padding:0.25rem 0.55rem; border-radius:999px; font-size:0.85rem; display:inline-flex; align-items:center; gap:0.35rem; }
.chip-x { font-weight:700; opacity:0.6; }
.user-results .list-group-item { text-align:left; }

.conv-avatar {
  width: 36px; height: 36px; border-radius: 999px; overflow: hidden;
  flex: 0 0 36px; display:flex; align-items:center; justify-content:center;
  background:#e9ecef; border:1px solid rgba(0,0,0,0.06);
}
.conv-avatar-img { width: 100%; height: 100%; object-fit: cover; }
.conv-avatar-fallback { font-weight:800; font-size:0.85rem; color:#495057; }
</style>
