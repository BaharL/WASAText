<template>
  <div class="conversation-page">
    <!-- Header -->
    <header class="conv-header d-flex align-items-center justify-content-between">
      <div>
        <h1 class="h5 mb-0">{{ title }}</h1>
        <small class="text-muted">Chat ID: {{ chatId }}</small>
      </div>

      <button
        type="button"
        class="btn btn-sm btn-outline-secondary"
        :disabled="loading"
        @click="loadConversation"
      >
        <span v-if="loading">Reloading…</span>
        <span v-else>Reload</span>
      </button>
    </header>

    <!-- Error -->
    <ErrorMsg v-if="error" :msg="error" />

    <!-- Corpo -->
    <LoadingSpinner :loading="loading">
      <div class="conv-body">
        <!-- Lista messaggi scrollabile (SOLO questa parte scrolla) -->
        <div ref="listEl" class="messages">
          <p v-if="!loading && !error && messages.length === 0" class="text-muted">
            No messages yet.
          </p>

          <div
            v-for="m in messages"
            :key="m.id"
            class="msg-row"
            :class="m.mine ? 'mine' : 'theirs'"
            :id="`msg-${m.id}`"
          >
            <div class="msg-bubble" :class="m.mine ? 'bubble-mine' : 'bubble-theirs'">
              <!-- Forward pill -->
              <div v-if="m.forwardedFromMessageId" class="meta-pill text-muted">
                Forwarded
              </div>

              <!-- Reply preview (click -> scroll to original) -->
              <div
                v-if="m.replyToMessageId"
                class="reply-preview"
                @click="scrollToMessage(m.replyToMessageId)"
                title="Go to replied message"
              >
                <div class="reply-bar"></div>
                <div class="reply-text">
                  Reply to #{{ m.replyToMessageId }}
                </div>
              </div>

              <!-- Sender + time -->
              <div class="msg-meta d-flex justify-content-between align-items-center">
                <small class="text-muted">
                  <strong v-if="!m.mine">{{ m.senderName }}</strong>
                  <span v-else>You</span>
                </small>
                <small class="text-muted">{{ formatTime(m.createdAt) }}</small>
              </div>

              <!-- Content -->
              <div class="msg-content">
                <div v-if="m.kind === 'text'">
                  {{ m.text }}
                </div>

                <div v-else-if="m.kind === 'image' && m.mediaUrl" class="media-wrap">
                  <img class="msg-img" :src="m.mediaUrl" alt="image" />
                </div>

                <div v-else-if="m.kind === 'gif' && m.mediaUrl" class="media-wrap">
                  <img class="msg-img" :src="m.mediaUrl" alt="gif" />
                </div>
              </div>

              <!-- Reactions bar -->
              <div v-if="m.reactions && m.reactions.length" class="reactions">
                <button
                  v-for="r in m.reactions"
                  :key="r.emoji"
                  class="reaction-chip"
                  :class="{ mine: r.mine }"
                  @click="toggleReaction(m, r.emoji, r.mine)"
                  type="button"
                  :title="r.mine ? 'Remove reaction' : 'React'"
                >
                  <span class="emoji">{{ r.emoji }}</span>
                  <span class="count">{{ r.count }}</span>
                </button>
              </div>

              <!-- Actions (…) -->
              <div class="msg-actions">
                <button class="btn btn-sm btn-light action-btn" @click="toggleMenu(m.id)">
                  …
                </button>

                <!-- popover: deve stare dentro bubble e NON uscire -->
                <div v-if="openMenuId === m.id" class="action-menu" @click.stop>
                  <button class="dropdown-item" @click="startReply(m)">Reply</button>
                  <button class="dropdown-item" @click="openForward(m)">Forward</button>

                  <div class="dropdown-divider"></div>

                  <div class="emoji-row">
                    <button class="emoji-btn" @click="react(m, '👍')">👍</button>
                    <button class="emoji-btn" @click="react(m, '❤️')">❤️</button>
                    <button class="emoji-btn" @click="react(m, '😂')">😂</button>
                    <button class="emoji-btn" @click="react(m, '😮')">😮</button>
                    <button class="emoji-btn" @click="react(m, '😢')">😢</button>
                    <button class="emoji-btn" @click="react(m, '🙏')">🙏</button>
                  </div>

                  <div class="dropdown-divider"></div>

                  <button class="dropdown-item text-danger" @click="removeMsg(m)" :disabled="!m.mine">
                    Delete
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div ref="bottomEl"></div>
        </div>

        <!-- Composer fisso sotto -->
        <div class="composer">
          <!-- Reply bar -->
          <div v-if="replyTo" class="replying">
            <div class="replying-left">
              <strong>Reply</strong>
              <small class="text-muted ms-2">to #{{ replyTo.id }}</small>
            </div>
            <button class="btn btn-sm btn-outline-secondary" @click="cancelReply">x</button>
          </div>

          <div class="composer-row">
            <!-- Upload (icona) -->
            <label class="btn btn-outline-secondary btn-sm upload-btn">
              📎
              <input type="file" class="d-none" @change="onPickFile" />
            </label>

            <!-- Text -->
            <input
              v-model="draft"
              class="form-control"
              placeholder="Write a message…"
              @keydown.enter.prevent="sendText"
            />

            <!-- Single send button -->
            <button class="btn btn-success" :disabled="sending" @click="sendText">
              Send
            </button>
          </div>

          <!-- Preview file selezionato -->
          <div v-if="pickedFile" class="file-preview">
            <small class="text-muted">
              Selected: <strong>{{ pickedFile.name }}</strong>
            </small>
            <button class="btn btn-sm btn-outline-secondary" @click="clearPickedFile">Remove</button>
            <button class="btn btn-sm btn-primary" :disabled="sending" @click="sendMedia">
              Send media
            </button>
          </div>
        </div>
      </div>
    </LoadingSpinner>

    <!-- Forward modal -->
    <div v-if="forwardingMsg" class="modal-backdrop" @click="closeForward"></div>
    <div v-if="forwardingMsg" class="forward-modal" @click.stop>
      <div class="d-flex justify-content-between align-items-center mb-2">
        <h6 class="mb-0">Forward message</h6>
        <button class="btn btn-sm btn-outline-secondary" @click="closeForward">x</button>
      </div>

      <input
        v-model="userSearch"
        class="form-control mb-2"
        placeholder="Search user…"
        @input="doSearchUsers"
      />

      <div class="user-results">
        <button
          v-for="u in userResults"
          :key="u.identifier"
          class="user-row"
          @click="forwardToUser(u)"
        >
          <div class="user-name">{{ u.username }}</div>
          <small class="text-muted">{{ u.identifier }}</small>
        </button>

        <div v-if="userSearch && userResults.length === 0" class="text-muted small p-2">
          No users found
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'

import {
  getConversation,
  sendMessage,
  uploadMedia,
  listUsers,
  createDirect,
  forwardMessage,
  addReaction,
  removeReaction,
  deleteMessage
} from '@/services/api.js'

const route = useRoute()
const router = useRouter()

const chatId = computed(() => Number(route.params.chatId))
const title = ref('Conversation')

const loading = ref(false)
const error = ref('')
const messages = ref([])

const listEl = ref(null)
const bottomEl = ref(null)

const draft = ref('')
const sending = ref(false)

const openMenuId = ref(null)

const replyTo = ref(null)

const pickedFile = ref(null)

const forwardingMsg = ref(null)
const userSearch = ref('')
const userResults = ref([])

function formatTime(iso) {
  if (!iso) return ''
  // iso può essere "YYYY-MM-DD HH:MM:SS"
  return iso.slice(11, 16)
}

function toggleMenu(id) {
  openMenuId.value = openMenuId.value === id ? null : id
}

function closeMenus() {
  openMenuId.value = null
}

function scrollToBottom() {
  nextTick(() => {
    bottomEl.value?.scrollIntoView({ behavior: 'smooth' })
  })
}

function scrollToMessage(id) {
  nextTick(() => {
    const el = document.getElementById(`msg-${id}`)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
      el.classList.add('flash')
      setTimeout(() => el.classList.remove('flash'), 900)
    }
  })
}

async function loadConversation() {
  error.value = ''
  loading.value = true
  closeMenus()
  try {
    const data = await getConversation(chatId.value)

    // la tua API potrebbe restituire { title, messages } oppure direttamente array
    if (Array.isArray(data)) {
      messages.value = data
    } else {
      title.value = data?.title || title.value
      messages.value = data?.messages || []
    }

    // normalizza media url (alcuni backend danno /v1/uploads...)
    messages.value = messages.value.map(m => ({
      ...m,
      mediaUrl: m.mediaUrl || m.mediaURL || null, // tolleranza
    }))

    scrollToBottom()
  } catch (e) {
    error.value = e.message || 'Cannot load conversation'
  } finally {
    loading.value = false
  }
}

function startReply(m) {
  replyTo.value = m
  closeMenus()
}

function cancelReply() {
  replyTo.value = null
}

async function sendText() {
  if (sending.value) return
  const text = (draft.value || '').trim()
  if (!text) return

  sending.value = true
  error.value = ''
  closeMenus()

  try {
    const payload = {
      chatId: chatId.value,
      kind: 'text',
      text,
      replyToMessageId: replyTo.value ? replyTo.value.id : undefined
    }
    await sendMessage(payload)
    draft.value = ''
    replyTo.value = null
    await loadConversation()
  } catch (e) {
    error.value = e.message || 'Cannot send message'
  } finally {
    sending.value = false
  }
}

function onPickFile(ev) {
  const f = ev.target.files?.[0]
  if (!f) return
  pickedFile.value = f
}

function clearPickedFile() {
  pickedFile.value = null
}

async function sendMedia() {
  if (!pickedFile.value) return
  if (sending.value) return

  sending.value = true
  error.value = ''
  closeMenus()

  try {
    const up = await uploadMedia(pickedFile.value)
    const url = up?.mediaUrl || up?.url || up?.path
    if (!url) throw new Error('Upload succeeded but no mediaUrl returned')

    const kind = pickedFile.value.type.startsWith('image/') ? 'image' : 'gif'

    const payload = {
      chatId: chatId.value,
      kind,
      mediaUrl: url,
      replyToMessageId: replyTo.value ? replyTo.value.id : undefined
    }

    await sendMessage(payload)
    pickedFile.value = null
    replyTo.value = null
    await loadConversation()
  } catch (e) {
    error.value = e.message || 'Cannot send media'
  } finally {
    sending.value = false
  }
}

async function react(m, emoji) {
  closeMenus()
  try {
    await addReaction(m.id, emoji)
    await loadConversation()
  } catch (e) {
    error.value = e.message || 'Cannot react'
  }
}

async function toggleReaction(m, emoji, mine) {
  try {
    if (mine) await removeReaction(m.id, emoji)
    else await addReaction(m.id, emoji)
    await loadConversation()
  } catch (e) {
    error.value = e.message || 'Cannot toggle reaction'
  }
}

async function removeMsg(m) {
  closeMenus()
  if (!m.mine) return
  try {
    await deleteMessage(m.id)
    await loadConversation()
  } catch (e) {
    error.value = e.message || 'Cannot delete message'
  }
}

function openForward(m) {
  closeMenus()
  forwardingMsg.value = m
  userSearch.value = ''
  userResults.value = []
}

function closeForward() {
  forwardingMsg.value = null
  userSearch.value = ''
  userResults.value = []
}

let searchTimer = null
function doSearchUsers() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(async () => {
    const q = userSearch.value.trim()
    if (!q) {
      userResults.value = []
      return
    }
    try {
      userResults.value = await listUsers(q)
    } catch (e) {
      userResults.value = []
    }
  }, 250)
}

async function forwardToUser(u) {
  if (!forwardingMsg.value) return
  try {
    // 1) create/get direct chat
    const direct = await createDirect(u.identifier)
    // direct potrebbe essere { chatId } o solo id
    const toChatId = direct?.chatId ?? direct?.id ?? direct
    if (!toChatId) throw new Error('Cannot create direct chat')

    // 2) forward message
    await forwardMessage(forwardingMsg.value.id, Number(toChatId))

    // 3) go to that chat
    closeForward()
    await router.push(`/conversations/${toChatId}`)
  } catch (e) {
    error.value = e.message || 'Cannot forward'
    closeForward()
  }
}

onMounted(() => {
  loadConversation()
  window.addEventListener('click', closeMenus)
})

watch(
  () => chatId.value,
  () => loadConversation()
)
</script>

<style scoped>
.conversation-page {
  height: calc(100vh - 60px); /* 60px = navbar circa */
  display: flex;
  flex-direction: column;
}

.conv-header {
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
  background: #fff;
  position: sticky;
  top: 0;
  z-index: 10;
}

.conv-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background: #f5f7fb;
}

.msg-row {
  display: flex;
  margin-bottom: 10px;
}

.msg-row.mine { justify-content: flex-end; }
.msg-row.theirs { justify-content: flex-start; }

.msg-bubble {
  max-width: 520px;
  width: fit-content;
  border-radius: 14px;
  padding: 10px 10px 8px 10px;
  position: relative; /* per popover */
}

.bubble-mine {
  background: #d9fdd3;
}

.bubble-theirs {
  background: #fff;
  border: 1px solid #eee;
}

.msg-meta {
  margin-bottom: 6px;
  gap: 10px;
}

.msg-content {
  white-space: pre-wrap;
  word-break: break-word;
}

.media-wrap {
  margin-top: 6px;
}

.msg-img {
  max-width: 320px;
  border-radius: 10px;
  display: block;
}

.meta-pill {
  font-size: 12px;
  margin-bottom: 6px;
}

.reply-preview {
  display: flex;
  gap: 10px;
  align-items: center;
  background: rgba(0,0,0,0.04);
  border-radius: 10px;
  padding: 6px 8px;
  margin-bottom: 8px;
  cursor: pointer;
}

.reply-bar {
  width: 4px;
  height: 28px;
  border-radius: 10px;
  background: rgba(0,0,0,0.25);
}

.reply-text {
  font-size: 12px;
  color: #444;
}

.reactions {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.reaction-chip {
  border: 1px solid #ddd;
  background: #fff;
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 13px;
  display: inline-flex;
  gap: 6px;
  align-items: center;
}

.reaction-chip.mine {
  border-color: #198754;
}

.reaction-chip .count {
  font-size: 12px;
  color: #555;
}

.msg-actions {
  position: absolute;
  right: 6px;
  bottom: 6px;
}

.action-btn {
  border-radius: 10px;
  padding: 2px 8px;
  font-size: 12px;
}

.action-menu {
  position: absolute;
  right: 0;
  bottom: 28px;
  width: 200px;
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 12px;
  padding: 6px;
  z-index: 999;
  box-shadow: 0 8px 18px rgba(0,0,0,0.12);
}

.emoji-row {
  display: flex;
  gap: 6px;
  padding: 6px;
  justify-content: space-between;
}

.emoji-btn {
  border: 1px solid #eee;
  background: #fafafa;
  border-radius: 10px;
  padding: 4px 6px;
}

.composer {
  background: #fff;
  border-top: 1px solid #eee;
  padding: 10px 12px;
}

.replying {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 10px;
  border: 1px solid #eee;
  border-radius: 10px;
  margin-bottom: 8px;
  background: #fafafa;
}

.composer-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.upload-btn {
  width: 44px;
}

.file-preview {
  margin-top: 8px;
  display: flex;
  gap: 10px;
  align-items: center;
}

/* Forward modal */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.3);
  z-index: 1000;
}

.forward-modal {
  position: fixed;
  top: 120px;
  left: 50%;
  transform: translateX(-50%);
  width: 420px;
  max-width: calc(100vw - 20px);
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 14px;
  padding: 12px;
  z-index: 1001;
  box-shadow: 0 12px 30px rgba(0,0,0,0.2);
}

.user-results {
  max-height: 260px;
  overflow: auto;
  border: 1px solid #eee;
  border-radius: 10px;
}

.user-row {
  width: 100%;
  text-align: left;
  padding: 10px;
  border: 0;
  background: #fff;
  border-bottom: 1px solid #f1f1f1;
}

.user-row:hover {
  background: #f7f7f7;
}

.user-name {
  font-weight: 600;
}

/* flash when jump to message */
.flash {
  outline: 2px solid rgba(25, 135, 84, 0.7);
  border-radius: 14px;
}
</style>
