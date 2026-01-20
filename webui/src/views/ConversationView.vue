<template>
  <div class="conversation-page">
    <!-- Header -->
    <header class="conv-header d-flex align-items-center justify-content-between">
      <div class="d-flex align-items-center gap-2">
        <!-- Avatar -->
        <div class="conv-avatar">
          <img
            v-if="conversationPhoto"
            :src="toImgSrc(conversationPhoto)"
            alt="avatar"
            class="conv-avatar-img"
          >
          <span v-else class="conv-avatar-fallback">
            {{ conversationAvatarLabel }}
          </span>
        </div>

        <div>
          <h1 class="h5 mb-0">{{ title }}</h1>
          <small class="text-muted">Chat ID: {{ chatId }}</small>
        </div>
      </div>

      <div class="d-flex align-items-center gap-2">
        <!-- Group photo actions (solo gruppi) -->
        <div v-if="isGroup" class="d-flex align-items-center gap-2">
          <input
            ref="groupPhotoInputEl"
            type="file"
            accept="image/*"
            class="d-none"
            :disabled="loading || uploadingGroupPhoto"
            @change="onGroupPhotoSelected"
          >

          <button
            type="button"
            class="btn btn-sm btn-outline-primary"
            :disabled="loading || uploadingGroupPhoto"
            @click="openGroupPhotoPicker"
            title="Change group photo"
          >
            <span v-if="uploadingGroupPhoto">Uploading…</span>
            <span v-else>Change photo</span>
          </button>
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
      </div>
    </header>

    <!-- Error -->
    <ErrorMsg v-if="error" :msg="error" />

    <!-- Corpo della chat -->
    <LoadingSpinner :loading="loading">
      <main class="conv-main" @click="closeActions">
        <!-- Nessun messaggio -->
        <p
          v-if="!loading && !error && messages.length === 0"
          class="text-muted text-center mt-3"
        >
          Nessun messaggio ancora. Inizia la conversazione! 💬
        </p>

        <!-- Lista messaggi -->
        <ul v-else ref="messageListEl" class="message-list">
          <li
            v-for="m in messages"
            :key="m.id"
            class="message-row"
            :class="{ mine: isMine(m) }"
          >
            <div class="bubble" @click.stop>
              <div class="bubble-meta">
                <span class="author">
                  {{ isMine(m) ? 'You' : (m.senderName || m.sender?.name || 'Other') }}
                </span>
                <span class="time">{{ formatTime(m.createdAt) }}</span>
              </div>

              <div class="bubble-content">
                <!-- text -->
                <span v-if="m.text">{{ m.text }}</span>

                <!-- media image -->
                <img
                  v-else-if="m.mediaUrl || m.url || m.photoUrl"
                  class="msg-img"
                  :src="toImgSrc(m.mediaUrl || m.url || m.photoUrl)"
                  alt="media"
                >

                <!-- fallback -->
                <span v-else class="kind-pill">[{{ m.kind || 'message' }}]</span>
              </div>

              <!-- Actions -->
              <div class="bubble-actions">
                <button
                  class="btn btn-sm btn-light"
                  type="button"
                  title="Actions"
                  @click.stop="toggleActions(m.id)"
                >
                  ⋯
                </button>

                <div v-if="showActionsFor === m.id" class="actions-popover">
                  <!-- Reaction -->
                  <div class="d-flex gap-2 align-items-center mb-2">
                    <select v-model="reactionEmoji" class="form-select form-select-sm">
                      <option>👍</option>
                      <option>❤️</option>
                      <option>😂</option>
                      <option>😡</option>
                      <option>🎉</option>
                    </select>
                    <button class="btn btn-sm btn-outline-primary" type="button" @click="onReact(m)">
                      React
                    </button>
                  </div>

                  <!-- Forward -->
                  <div class="d-flex gap-2 align-items-center mb-2">
                    <input
                      v-model.trim="forwardToChatId"
                      class="form-control form-control-sm"
                      placeholder="Forward to chatId"
                    >
                    <button class="btn btn-sm btn-outline-secondary" type="button" @click="onForward(m)">
                      Forward
                    </button>
                  </div>

                  <!-- Delete (solo se mio) -->
                  <button
                    v-if="isMine(m)"
                    class="btn btn-sm btn-outline-danger w-100"
                    type="button"
                    @click="onDelete(m)"
                  >
                    Delete
                  </button>
                </div>
              </div>

            </div>
          </li>
        </ul>
      </main>
    </LoadingSpinner>

    <!-- Input messaggio -->
    <footer class="conv-input">
      <form class="input-row" @submit.prevent="handleSend">
        <input
          v-model="draft"
          type="text"
          class="form-control"
          placeholder="Scrivi un messaggio…"
          :disabled="sending || sendingMedia"
        >

        <input
          type="file"
          accept="image/*"
          class="form-control form-control-sm ms-2"
          :disabled="sending || sendingMedia"
          @change="onFileSelected"
        >

        <button
          type="button"
          class="btn btn-outline-primary ms-2"
          :disabled="sendingMedia || !selectedFile"
          @click="handleSendMedia"
        >
          <span v-if="sendingMedia">Uploading…</span>
          <span v-else>Send photo</span>
        </button>

        <button
          type="submit"
          class="btn btn-success ms-2"
          :disabled="sending || sendingMedia || !draft.trim()"
        >
          <span v-if="sending">Sending…</span>
          <span v-else>Send</span>
        </button>
      </form>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick, watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import ErrorMsg from '../components/ErrorMsg.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

import {
  getConversation,
  sendMessage,
  uploadMedia,
  addReaction,
  deleteMessage,
  forwardMessage,
  setGroupPhoto
} from '../services/api.js'

const route = useRoute()

const chatId = ref(route.params.chatId)

const title = ref('Conversation')
const messages = ref([])
const loading = ref(false)
const sending = ref(false)
const error = ref('')
const draft = ref('')
const messageListEl = ref(null)

/* conversation metadata (title/photo/type) */
const conversation = ref(null)

const isGroup = computed(() => {
  const c = conversation.value || {}
  if (typeof c.isGroup === 'boolean') return c.isGroup
  if (c.type) return String(c.type).toLowerCase() === 'group'
  return false
})

const conversationPhoto = computed(() => {
  const c = conversation.value || {}
  return c.photoUrl || c.photo_url || c.photo || ''
})

const conversationAvatarLabel = computed(() => {
  const t = (title.value || '').trim()
  return t ? t.charAt(0).toUpperCase() : '?'
})

/* group photo upload */
const groupPhotoInputEl = ref(null)
const uploadingGroupPhoto = ref(false)

function openGroupPhotoPicker() {
  if (!groupPhotoInputEl.value) return
  groupPhotoInputEl.value.value = '' // reset per poter scegliere lo stesso file
  groupPhotoInputEl.value.click()
}

async function onGroupPhotoSelected(e) {
  const file = e.target.files?.[0] || null
  if (!file) return

  uploadingGroupPhoto.value = true
  error.value = ''

  try {
    await setGroupPhoto(Number(chatId.value), file)
    // ricarica chat + (opzionale) fai sapere alla lista di refreshare
    await loadConversation()
    window.dispatchEvent(new Event('conversations-updated'))
  } catch (err) {
    console.error(err)
    error.value = err.message || 'Failed to update group photo.'
  } finally {
    uploadingGroupPhoto.value = false
  }
}

/* media */
const selectedFile = ref(null)
const sendingMedia = ref(false)

/* actions */
const showActionsFor = ref(null)
const forwardToChatId = ref('')
const reactionEmoji = ref('👍')

function toggleActions(messageId) {
  showActionsFor.value = (showActionsFor.value === messageId) ? null : messageId
}

function closeActions() {
  showActionsFor.value = null
}

function isMine(m) {
  return m.mine === true || m.direction === 'outgoing'
}

function formatTime(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}

function toImgSrc(url) {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) return url
  if (url.startsWith('/v1/')) return `/api${url}`
  if (url.startsWith('/uploads/')) return `/api/v1${url}`
  return url.startsWith('/') ? url : `/${url}`
}

function scrollToBottom() {
  const el = messageListEl.value
  if (el) el.scrollTop = el.scrollHeight
}

async function loadConversation({ silent = false } = {}) {
  if (!chatId.value) {
    error.value = 'Chat non valida.'
    return
  }

  if (!silent) loading.value = true
  error.value = ''

  try {
    const data = await getConversation(chatId.value)

    conversation.value = data || null
    messages.value = Array.isArray(data) ? data : (data.messages || [])

    if (data?.title || data?.name || data?.chatName) {
      title.value = data.title || data.name || data.chatName
    } else {
      title.value = 'Conversation'
    }

    await nextTick()
    scrollToBottom()
  } catch (e) {
    console.error(e)
    if (!silent) error.value = e.message || 'Failed to load messages.'
  } finally {
    if (!silent) loading.value = false
  }
}

async function handleSend() {
  const text = draft.value.trim()
  if (!text || !chatId.value) return

  sending.value = true
  error.value = ''

  try {
    await sendMessage({
      chatId: Number(chatId.value),
      kind: 'text',
      text
    })

    draft.value = ''
    await loadConversation()
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to send message.'
  } finally {
    sending.value = false
  }
}

function onFileSelected(e) {
  selectedFile.value = e.target.files?.[0] || null
}

async function handleSendMedia() {
  if (!selectedFile.value || !chatId.value) return

  sendingMedia.value = true
  error.value = ''

  try {
    const up = await uploadMedia(selectedFile.value)

    const mediaUrl = up?.url || up?.mediaUrl || up?.photoUrl || up?.path || up?.file || ''
    if (!mediaUrl) throw new Error('Upload ok but no media URL returned.')

    await sendMessage({
      chatId: Number(chatId.value),
      kind: 'media',
      mediaUrl
    })

    selectedFile.value = null
    await loadConversation()
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to send media.'
  } finally {
    sendingMedia.value = false
  }
}

async function onReact(m) {
  try {
    await addReaction(m.id, reactionEmoji.value)
    await loadConversation()
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to react.'
  } finally {
    showActionsFor.value = null
  }
}

async function onForward(m) {
  const toChatId = Number(forwardToChatId.value)
  if (!toChatId) return

  try {
    await forwardMessage(m.id, toChatId)
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to forward.'
  } finally {
    forwardToChatId.value = ''
    showActionsFor.value = null
  }
}

async function onDelete(m) {
  try {
    await deleteMessage(m.id)
    await loadConversation()
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to delete.'
  } finally {
    showActionsFor.value = null
  }
}

/* Polling */
let pollTimer = null

function startPolling() {
  stopPolling()
  pollTimer = setInterval(async () => {
    if (loading.value || sending.value || sendingMedia.value || uploadingGroupPhoto.value) return
    if (showActionsFor.value !== null) return
    await loadConversation({ silent: true })
  }, 2500)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(async () => {
  await loadConversation()
  startPolling()
})

onBeforeUnmount(() => {
  stopPolling()
})

watch(
  () => route.params.chatId,
  async (newId) => {
    chatId.value = newId
    await loadConversation()
    startPolling()
  }
)
</script>

<style scoped>
.conversation-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 56px);
}

.conv-header {
  padding: 0.75rem 0.75rem 0.25rem;
  border-bottom: 1px solid rgba(0, 0, 0, 0.1);
}

.conv-main {
  flex: 1;
  padding: 0.75rem;
  overflow-y: auto;
  background: #f5f7fb;
}

.message-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.message-row {
  display: flex;
  justify-content: flex-start;
}

.message-row.mine {
  justify-content: flex-end;
}

.bubble {
  max-width: 70%;
  padding: 0.5rem 0.65rem;
  border-radius: 14px;
  background: white;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
  font-size: 0.9rem;
  position: relative;
}

.message-row.mine .bubble {
  background: #d1f5e2;
}

.bubble-meta {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.15rem;
  font-size: 0.75rem;
  color: #6c757d;
}

.bubble-content {
  white-space: pre-wrap;
  word-break: break-word;
}

.kind-pill {
  font-size: 0.8rem;
  padding: 0.1rem 0.35rem;
  border-radius: 999px;
  background: #e9ecef;
}

.msg-img {
  max-width: 260px;
  max-height: 260px;
  border-radius: 12px;
  display: block;
}

.bubble-actions {
  position: relative;
  margin-top: 0.35rem;
  display: flex;
  justify-content: flex-end;
}

.actions-popover {
  position: absolute;
  right: 0;
  top: 28px;
  z-index: 10;
  background: white;
  border: 1px solid rgba(0,0,0,0.12);
  border-radius: 12px;
  padding: 0.5rem;
  width: 240px;
  box-shadow: 0 10px 24px rgba(0,0,0,0.12);
}

.conv-input {
  padding: 0.5rem 0.75rem;
  border-top: 1px solid rgba(0, 0, 0, 0.1);
  background: #ffffff;
}

.input-row {
  display: flex;
  align-items: center;
}

/* Header avatar */
.conv-avatar {
  width: 36px;
  height: 36px;
  border-radius: 999px;
  overflow: hidden;
  flex: 0 0 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #e9ecef;
  border: 1px solid rgba(0,0,0,0.06);
}

.conv-avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.conv-avatar-fallback {
  font-weight: 800;
  font-size: 0.85rem;
  color: #495057;
}
</style>
