<template>
  <div class="conversation-page w-100">
    <!-- Header -->
    <header class="conv-header d-flex align-items-center justify-content-between">
      <div class="d-flex align-items-center gap-2">
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

    <ErrorMsg v-if="error" :msg="error" />

    <LoadingSpinner :loading="loading">
      <main class="conv-main" @click="closeActions">
        <p
          v-if="!loading && !error && messages.length === 0"
          class="text-muted text-center mt-3"
        >
          Nessun messaggio ancora. Inizia la conversazione! 💬
        </p>

        <ul v-else ref="messageListEl" class="message-list">
          <li
            v-for="m in messages"
            :key="m.id"
            class="message-row"
            :class="{ mine: isMine(m) }"
            :data-mid="m.id"
          >
            <div class="bubble" @click.stop>
              <div class="bubble-meta">
                <span class="author">
                  {{ isMine(m) ? 'You' : (m.senderName || m.sender?.name || 'Other') }}
                </span>
                <span class="time">{{ formatTime(m.createdAt) }}</span>
              </div>

              <!-- QUOTE / REPLY BOX -->
              <button
                v-if="getReplyTarget(m)"
                type="button"
                class="quoted"
                @click="jumpToMessage(getReplyTarget(m).id)"
                title="Go to replied message"
              >
                <div class="quoted-author">
                  {{ getReplyTarget(m).mine ? 'You' : (getReplyTarget(m).senderName || 'Other') }}
                </div>
                <div class="quoted-text">
                  <span v-if="getReplyTarget(m).text">{{ getReplyTarget(m).text }}</span>
                  <span v-else>[{{ getReplyTarget(m).kind || 'media' }}]</span>
                </div>
              </button>

              <div class="bubble-content">
                <span v-if="m.text">{{ m.text }}</span>

                <img
                  v-else-if="m.mediaUrl || m.url || m.photoUrl"
                  class="msg-img"
                  :src="toImgSrc(m.mediaUrl || m.url || m.photoUrl)"
                  alt="media"
                >

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

                <!-- Popover: NON esce fuori (flip) -->
                <div
                  v-if="showActionsFor === m.id"
                  class="actions-popover"
                  :class="{ up: shouldPopoverGoUp(m.id) }"
                  @click.stop
                >
                  <!-- Reply -->
                  <button
                    class="btn btn-sm btn-outline-secondary w-100 mb-2"
                    type="button"
                    @click="onReply(m)"
                  >
                    Reply
                  </button>

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

                  <!-- Delete -->
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

    <!-- Composer unico -->
    <footer class="conv-input">
      <!-- Reply preview (WhatsApp-like) -->
      <div v-if="replyTo" class="reply-preview">
        <div class="reply-bar"></div>
        <div class="reply-body">
          <div class="reply-title">Reply to {{ replyTo.sender }}</div>
          <div class="reply-snippet">
            <span v-if="replyTo.text">{{ replyTo.text }}</span>
            <span v-else>[{{ replyTo.kind || 'media' }}]</span>
          </div>
        </div>
        <button type="button" class="reply-close" @click="cancelReply" title="Cancel reply">×</button>
      </div>

      <!-- Media preview -->
      <div v-if="selectedFile" class="media-preview">
        <div class="media-name">
          📎 {{ selectedFile.name }}
        </div>
        <button type="button" class="btn btn-sm btn-outline-secondary" @click="clearSelectedFile">
          Remove
        </button>
      </div>

      <form class="input-row" @submit.prevent="handleSendUnified">
        <input
          v-model="draft"
          type="text"
          class="form-control"
          placeholder="Scrivi un messaggio…"
          :disabled="sending"
        >

        <label class="btn btn-outline-secondary ms-2 mb-0">
          <input
            type="file"
            accept="image/*"
            class="d-none"
            :disabled="sending"
            @change="onFileSelected"
          >
          📷
        </label>

        <button
          type="submit"
          class="btn btn-success ms-2"
          :disabled="sending || (!draft.trim() && !selectedFile)"
        >
          <span v-if="sending">Sending…</span>
          <span v-else>Send</span>
        </button>
      </form>

      <small class="text-muted d-block mt-1">
        Enter = send · Shift+Enter = new line (solo se vuoi lo aggiungiamo)
      </small>
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
  getMyConversations,
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

/* meta */
const conversationMeta = ref(null)
const isGroup = computed(() => {
  const c = conversationMeta.value || {}
  if (typeof c.isGroup === 'boolean') return c.isGroup
  if (c.type) return String(c.type).toLowerCase() === 'group'
  return false
})
const conversationPhoto = computed(() => {
  const c = conversationMeta.value || {}
  return c.photoUrl || c.photo_url || c.photo || ''
})
const conversationAvatarLabel = computed(() => {
  const t = (title.value || '').trim()
  return t ? t.charAt(0).toUpperCase() : '?'
})

/* group photo */
const groupPhotoInputEl = ref(null)
const uploadingGroupPhoto = ref(false)
function openGroupPhotoPicker() {
  if (!groupPhotoInputEl.value) return
  groupPhotoInputEl.value.value = ''
  groupPhotoInputEl.value.click()
}
async function onGroupPhotoSelected(e) {
  const file = e.target.files?.[0] || null
  if (!file) return

  uploadingGroupPhoto.value = true
  error.value = ''
  try {
    await setGroupPhoto(Number(chatId.value), file)
    await loadConversation()
    window.dispatchEvent(new Event('conversations-updated'))
  } catch (err) {
    console.error(err)
    error.value = err.message || 'Failed to update group photo.'
  } finally {
    uploadingGroupPhoto.value = false
  }
}

/* reply state */
const replyTo = ref(null)
function onReply(m) {
  replyTo.value = {
    id: m.id,
    sender: isMine(m) ? 'You' : (m.senderName || m.sender?.name || 'Other'),
    text: m.text || '',
    kind: m.kind || '',
    mediaUrl: m.mediaUrl || m.url || m.photoUrl || ''
  }
  showActionsFor.value = null
}
function cancelReply() {
  replyTo.value = null
}

/* media */
const selectedFile = ref(null)
function onFileSelected(e) {
  selectedFile.value = e.target.files?.[0] || null
}
function clearSelectedFile() {
  selectedFile.value = null
}

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

/* popover positioning: flip up if near bottom */
function shouldPopoverGoUp(messageId) {
  const container = document.querySelector('.conv-main')
  if (!container) return false
  const row = container.querySelector(`[data-mid="${messageId}"]`)
  if (!row) return false

  const rect = row.getBoundingClientRect()
  const viewH = window.innerHeight
  // se la parte bassa è troppo vicina al fondo, apri verso l'alto
  return rect.bottom > (viewH - 260)
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

/* Reply link: find replied msg in current list */
function getReplyId(m) {
  return m.replyToMessageId ?? m.reply_to_message_id ?? null
}
function getReplyTarget(m) {
  const id = getReplyId(m)
  if (!id) return null
  return messages.value.find(x => x.id === id) || null
}

/* Jump to replied message + highlight */
const highlightId = ref(null)
function jumpToMessage(messageId) {
  const container = document.querySelector('.conv-main')
  if (!container) return
  const el = container.querySelector(`[data-mid="${messageId}"]`)
  if (!el) return

  el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  highlightId.value = messageId
  setTimeout(() => {
    if (highlightId.value === messageId) highlightId.value = null
  }, 900)
}

/* Load conversation */
async function loadConversation({ silent = false } = {}) {
  if (!chatId.value) {
    error.value = 'Chat non valida.'
    return
  }
  if (!silent) loading.value = true
  error.value = ''

  try {
    const data = await getConversation(chatId.value)
    messages.value = Array.isArray(data) ? data : (data.messages || [])

    try {
      const list = await getMyConversations()
      const arr = Array.isArray(list) ? list : (list.conversations || list.items || [])
      const found = arr.find(c => String(c.id ?? c.chatId ?? c.conversationId) === String(chatId.value))
      conversationMeta.value = found || null

      const metaTitle = found?.name || found?.title || found?.chatName || ''
      if (metaTitle) title.value = metaTitle
      else if (data?.title || data?.name || data?.chatName) title.value = data.title || data.name || data.chatName
      else title.value = 'Conversation'
    } catch {
      conversationMeta.value = null
      title.value = data?.title || data?.name || data?.chatName || 'Conversation'
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

/* ✅ Composer unico: invia text oppure image con un solo tasto */
async function handleSendUnified() {
  if (!chatId.value) return
  if (!draft.value.trim() && !selectedFile.value) return

  sending.value = true
  error.value = ''

  try {
    let kind = 'text'
    let text = draft.value.trim()
    let mediaUrl = null

    if (selectedFile.value) {
      // 1) upload file
      const up = await uploadMedia(selectedFile.value)
      mediaUrl = up?.url || up?.mediaUrl || up?.photoUrl || up?.path || up?.file || ''
      if (!mediaUrl) throw new Error('Upload ok but no media URL returned.')
      kind = 'image'
      // opzionale: se vuoi permettere anche testo + foto, NON svuotare text.
      // per ora: se text esiste, lo mando comunque (dipende dal backend)
      if (!text) text = null
    }

    // 2) send message
    await sendMessage({
      chatId: Number(chatId.value),
      kind,
      text: kind === 'text' ? text : (text || null),
      mediaUrl: mediaUrl || null,
      replyToMessageId: replyTo.value?.id ?? null
    })

    // reset
    draft.value = ''
    selectedFile.value = null
    replyTo.value = null

    await loadConversation()
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to send.'
  } finally {
    sending.value = false
  }
}

/* reactions / forward / delete */
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
  const ok = confirm('Delete this message?')
  if (!ok) return

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
    if (loading.value || sending.value || uploadingGroupPhoto.value) return
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
onBeforeUnmount(() => stopPolling())

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
  width: 100%;
  height: calc(100vh - 56px);
  min-width: 0;
}

.conv-header {
  padding: 0.75rem 0.75rem 0.25rem;
  border-bottom: 1px solid rgba(0, 0, 0, 0.1);
  background: #fff;
}

.conv-main {
  flex: 1;
  min-height: 0;
  padding: 0.75rem;
  overflow-y: auto;
  background: #f5f7fb;
  position: relative;
}

.message-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.message-row {
  display: flex;
  justify-content: flex-start;
}
.message-row.mine {
  justify-content: flex-end;
}

.bubble {
  max-width: 72%;
  padding: 0.55rem 0.65rem;
  border-radius: 14px;
  background: white;
  box-shadow: 0 2px 7px rgba(0, 0, 0, 0.07);
  font-size: 0.92rem;
  position: relative;
}
.message-row.mine .bubble {
  background: #d8f6e6;
}

.bubble-meta {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.2rem;
  font-size: 0.75rem;
  color: #6c757d;
}

.bubble-content {
  white-space: pre-wrap;
  word-break: break-word;
}

.kind-pill {
  font-size: 0.8rem;
  padding: 0.12rem 0.4rem;
  border-radius: 999px;
  background: #e9ecef;
}

.msg-img {
  max-width: 280px;
  max-height: 280px;
  border-radius: 12px;
  display: block;
}

/* quoted */
.quoted {
  width: 100%;
  text-align: left;
  border: none;
  background: rgba(0,0,0,0.04);
  border-radius: 12px;
  padding: 0.45rem 0.55rem;
  margin-bottom: 0.4rem;
  cursor: pointer;
}
.quoted:hover { background: rgba(0,0,0,0.06); }
.quoted-author {
  font-weight: 700;
  font-size: 0.78rem;
  color: #0d6efd;
  margin-bottom: 0.1rem;
}
.quoted-text {
  font-size: 0.82rem;
  color: #495057;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* actions */
.bubble-actions {
  position: relative;
  margin-top: 0.35rem;
  display: flex;
  justify-content: flex-end;
}

/* anchored popover */
.actions-popover {
  position: absolute;
  right: 0;
  top: 28px;
  z-index: 50;
  background: white;
  border: 1px solid rgba(0,0,0,0.12);
  border-radius: 12px;
  padding: 0.55rem;
  width: 240px;
  box-shadow: 0 10px 24px rgba(0,0,0,0.14);
}
.actions-popover.up {
  top: auto;
  bottom: 28px;
}

/* input area */
.conv-input {
  padding: 0.55rem 0.75rem;
  border-top: 1px solid rgba(0, 0, 0, 0.1);
  background: #ffffff;
}

.input-row {
  display: flex;
  align-items: center;
}

/* Reply preview */
.reply-preview {
  display: flex;
  align-items: stretch;
  gap: 0.6rem;
  padding: 0.5rem 0.55rem;
  border-radius: 12px;
  background: rgba(13,110,253,0.07);
  border: 1px solid rgba(13,110,253,0.15);
  margin-bottom: 0.45rem;
}
.reply-bar {
  width: 4px;
  border-radius: 999px;
  background: #0d6efd;
  flex: 0 0 4px;
}
.reply-body { flex: 1; min-width: 0; }
.reply-title {
  font-size: 0.78rem;
  font-weight: 800;
  color: #0d6efd;
}
.reply-snippet {
  font-size: 0.82rem;
  color: #495057;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.reply-close {
  border: none;
  background: transparent;
  font-size: 1.2rem;
  line-height: 1;
  opacity: 0.6;
}
.reply-close:hover { opacity: 1; }

/* media preview */
.media-preview {
  display:flex;
  align-items:center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.45rem 0.55rem;
  border-radius: 12px;
  background: rgba(0,0,0,0.04);
  border: 1px solid rgba(0,0,0,0.08);
  margin-bottom: 0.45rem;
}
.media-name {
  font-size: 0.85rem;
  color: #495057;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* header avatar */
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
