<template>
  <div class="conversation-page w-100">
    <!-- Header -->
    <header class="conv-header d-flex align-items-center justify-content-between">
      <div class="d-flex align-items-center gap-2">
        <div class="conv-avatar">
          <img v-if="conversationPhoto" :src="toImgSrc(conversationPhoto)" alt="avatar" class="conv-avatar-img">
          <span v-else class="conv-avatar-fallback">{{ conversationAvatarLabel }}</span>
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

        <button type="button" class="btn btn-sm btn-outline-secondary" :disabled="loading" @click="loadConversation">
          <span v-if="loading">Reloading…</span>
          <span v-else>Reload</span>
        </button>
      </div>
    </header>

    <ErrorMsg v-if="error" :msg="error" />

    <LoadingSpinner :loading="loading">
      <main ref="mainEl" class="conv-main" @click="closeActions">
        <p v-if="!loading && !error && messages.length === 0" class="text-muted text-center mt-3">
          Nessun messaggio ancora. Inizia la conversazione! 💬
        </p>

        <ul v-else ref="messageListEl" class="message-list">
          <li
            v-for="m in messages"
            :key="m.id"
            class="message-row"
            :class="{ mine: isMine(m) }"
            :data-msgid="m.id"
          >
            <div class="bubble" @click.stop>
              <div class="bubble-meta">
                <span class="author">{{ isMine(m) ? 'You' : (m.senderName || m.sender?.name || 'Other') }}</span>
                <span class="time">{{ formatTime(m.createdAt) }}</span>
              </div>

              <!-- Forwarded label -->
              <div v-if="m.forwardedFromMessageId" class="pill-forwarded">
                Forwarded
              </div>

              <!-- ✅ Reply preview (WhatsApp style) -->
              <button
                v-if="m.replyToMessageId"
                type="button"
                class="reply-preview"
                @click="scrollToMessage(m.replyToMessageId)"
                title="Go to replied message"
              >
                <div class="reply-line1">
                  Reply to <span class="reply-author">{{ getReplyAuthor(m.replyToMessageId) }}</span>
                </div>
                <div class="reply-line2">
                  {{ getReplySnippet(m.replyToMessageId) }}
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
                  @click.stop="toggleActions(m.id, $event)"
                >
                  ⋯
                </button>
              </div>
            </div>
          </li>
        </ul>

        <!-- ✅ Fixed popover that stays in viewport -->
        <div
          v-if="showActionsFor !== null"
          class="actions-popover-fixed"
          :style="{ top: popoverPos.top + 'px', left: popoverPos.left + 'px' }"
          @click.stop
        >
          <button class="action-item" type="button" @click="onReply(getMessageById(showActionsFor))">
            ↩ Reply
          </button>

          <div class="action-sep"></div>

          <div class="d-flex gap-2 align-items-center mb-2">
            <select v-model="reactionEmoji" class="form-select form-select-sm">
              <option>👍</option>
              <option>❤️</option>
              <option>😂</option>
              <option>😡</option>
              <option>🎉</option>
            </select>
            <button class="btn btn-sm btn-outline-primary" type="button" @click="onReact(getMessageById(showActionsFor))">
              React
            </button>
          </div>

          <div class="d-flex gap-2 align-items-center mb-2">
            <input
              v-model.trim="forwardToChatId"
              class="form-control form-control-sm"
              placeholder="Forward to chatId"
            >
            <button class="btn btn-sm btn-outline-secondary" type="button" @click="onForward(getMessageById(showActionsFor))">
              Forward
            </button>
          </div>

          <button
            v-if="isMine(getMessageById(showActionsFor))"
            class="btn btn-sm btn-outline-danger w-100"
            type="button"
            @click="onDelete(getMessageById(showActionsFor))"
          >
            Delete
          </button>
        </div>

      </main>
    </LoadingSpinner>

    <!-- ✅ Single unified input -->
    <footer class="conv-input">
      <!-- Reply bar -->
      <div v-if="replyingTo" class="reply-bar">
        <div class="reply-bar-text">
          Replying to <b>{{ replyingTo.senderName || (isMine(replyingTo) ? 'You' : 'Other') }}</b>:
          <span class="text-muted">{{ snippet(replyingTo) }}</span>
        </div>
        <button type="button" class="btn btn-sm btn-light" @click="cancelReply" title="Cancel reply">✕</button>
      </div>

      <form class="input-row" @submit.prevent="handleSendUnified">
        <!-- Attach -->
        <input
          ref="fileInputEl"
          type="file"
          accept="image/*"
          class="d-none"
          :disabled="sending"
          @change="onFileSelected"
        >
        <button
          type="button"
          class="btn btn-outline-secondary"
          :disabled="sending"
          @click="openFilePicker"
          title="Attach image"
        >
          📎
        </button>

        <input
          v-model="draft"
          type="text"
          class="form-control ms-2"
          placeholder="Scrivi un messaggio…"
          :disabled="sending"
        >

        <button
          type="submit"
          class="btn btn-success ms-2"
          :disabled="sending || (!draft.trim() && !selectedFile)"
        >
          <span v-if="sending">Sending…</span>
          <span v-else>Send</span>
        </button>
      </form>

      <div v-if="selectedFile" class="attach-preview">
        <span class="badge text-bg-light">
          📷 {{ selectedFile.name }}
        </span>
        <button type="button" class="btn btn-sm btn-link" @click="clearSelectedFile">Remove</button>
      </div>
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
const mainEl = ref(null)

/**
 * Meta from conversation list
 */
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

/* group photo upload */
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

/* unified media */
const fileInputEl = ref(null)
const selectedFile = ref(null)

function openFilePicker() {
  if (!fileInputEl.value) return
  fileInputEl.value.value = ''
  fileInputEl.value.click()
}

function onFileSelected(e) {
  selectedFile.value = e.target.files?.[0] || null
}

function clearSelectedFile() {
  selectedFile.value = null
}

/* actions popover */
const showActionsFor = ref(null)
const forwardToChatId = ref('')
const reactionEmoji = ref('👍')

const popoverPos = ref({ top: 0, left: 0 })

function clamp(n, min, max) {
  return Math.max(min, Math.min(max, n))
}

function toggleActions(messageId, ev) {
  if (showActionsFor.value === messageId) {
    showActionsFor.value = null
    return
  }

  showActionsFor.value = messageId

  // position popover inside viewport
  const btn = ev?.currentTarget
  if (btn) {
    const rect = btn.getBoundingClientRect()
    const w = 260
    const h = 190
    const margin = 8

    const left = clamp(rect.right - w, margin, window.innerWidth - w - margin)
    const top = clamp(rect.bottom + 6, margin, window.innerHeight - h - margin)

    popoverPos.value = { top, left }
  }
}

function closeActions() {
  showActionsFor.value = null
}

function getMessageById(id) {
  if (!id) return null
  return messages.value.find(x => String(x.id) === String(id)) || null
}

function isMine(m) {
  if (!m) return false
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

/* ✅ Reply helpers */
const replyingTo = ref(null)

function onReply(m) {
  if (!m) return
  replyingTo.value = m
  closeActions()
}

function cancelReply() {
  replyingTo.value = null
}

function snippet(m) {
  if (!m) return ''
  if (m.text) return String(m.text).slice(0, 80)
  if (m.mediaUrl) return '[image]'
  return `[${m.kind || 'message'}]`
}

function getReplySnippet(replyId) {
  const m = getMessageById(replyId)
  return m ? snippet(m) : '[message]'
}

function getReplyAuthor(replyId) {
  const m = getMessageById(replyId)
  if (!m) return 'unknown'
  return isMine(m) ? 'You' : (m.senderName || 'Other')
}

async function scrollToMessage(messageId) {
  closeActions()
  await nextTick()

  const el = mainEl.value
  if (!el) return

  // find li by data-msgid
  const target = el.querySelector(`[data-msgid="${messageId}"]`)
  if (!target) return

  target.scrollIntoView({ behavior: 'smooth', block: 'center' })

  // small highlight
  target.classList.add('flash')
  setTimeout(() => target.classList.remove('flash'), 700)
}

/* load conversation */
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

/* ✅ unified send */
async function handleSendUnified() {
  if (!chatId.value) return
  if (!draft.value.trim() && !selectedFile.value) return

  sending.value = true
  error.value = ''

  try {
    // if file attached -> upload then send image message
    if (selectedFile.value) {
      const up = await uploadMedia(selectedFile.value)
      const mediaUrl = up?.url || up?.mediaUrl || up?.photoUrl || up?.path || up?.file || ''
      if (!mediaUrl) throw new Error('Upload ok but no media URL returned.')

      await sendMessage({
        chatId: Number(chatId.value),
        kind: 'image',
        mediaUrl,
        replyToMessageId: replyingTo.value?.id ?? undefined
      })

      selectedFile.value = null
      draft.value = ''
      replyingTo.value = null
      await loadConversation()
      return
    }

    // text
    await sendMessage({
      chatId: Number(chatId.value),
      kind: 'text',
      text: draft.value.trim(),
      replyToMessageId: replyingTo.value?.id ?? undefined
    })

    draft.value = ''
    replyingTo.value = null
    await loadConversation()
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to send message.'
  } finally {
    sending.value = false
  }
}

/* actions handlers */
async function onReact(m) {
  if (!m) return
  try {
    await addReaction(m.id, reactionEmoji.value)
    await loadConversation()
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to react.'
  } finally {
    closeActions()
  }
}

async function onForward(m) {
  if (!m) return
  const toChatId = Number(forwardToChatId.value)
  if (!toChatId) return

  try {
    await forwardMessage(m.id, toChatId)
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to forward.'
  } finally {
    forwardToChatId.value = ''
    closeActions()
  }
}

async function onDelete(m) {
  if (!m) return
  try {
    await deleteMessage(m.id)
    await loadConversation()
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to delete.'
  } finally {
    closeActions()
  }
}

/* polling */
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

function onWindowResizeOrScroll() {
  // close popover on resize/scroll to avoid weird positions
  closeActions()
}

onMounted(async () => {
  await loadConversation()
  startPolling()
  window.addEventListener('resize', onWindowResizeOrScroll)
  window.addEventListener('scroll', onWindowResizeOrScroll, true)
})

onBeforeUnmount(() => {
  stopPolling()
  window.removeEventListener('resize', onWindowResizeOrScroll)
  window.removeEventListener('scroll', onWindowResizeOrScroll, true)
})

watch(
  () => route.params.chatId,
  async (newId) => {
    chatId.value = newId
    replyingTo.value = null
    selectedFile.value = null
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
  margin-top: 0.35rem;
  display: flex;
  justify-content: flex-end;
}

/* ✅ fixed popover */
.actions-popover-fixed {
  position: fixed;
  z-index: 9999;
  background: white;
  border: 1px solid rgba(0,0,0,0.12);
  border-radius: 12px;
  padding: 0.5rem;
  width: 260px;
  box-shadow: 0 10px 24px rgba(0,0,0,0.12);
}

.action-item {
  width: 100%;
  text-align: left;
  border: 0;
  background: transparent;
  padding: 0.35rem 0.25rem;
  border-radius: 8px;
  font-weight: 600;
}
.action-item:hover {
  background: rgba(0,0,0,0.05);
}
.action-sep {
  height: 1px;
  background: rgba(0,0,0,0.08);
  margin: 0.35rem 0;
}

/* ✅ reply preview (WhatsApp-ish) */
.reply-preview {
  width: 100%;
  border: 0;
  background: rgba(0,0,0,0.04);
  border-left: 4px solid rgba(13,110,253,0.8);
  border-radius: 10px;
  padding: 0.35rem 0.5rem;
  margin-bottom: 0.35rem;
  text-align: left;
  cursor: pointer;
}
.reply-line1 { font-size: 0.75rem; color: #495057; }
.reply-author { font-weight: 800; }
.reply-line2 { font-size: 0.8rem; color: #6c757d; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* forwarded pill */
.pill-forwarded {
  font-size: 0.72rem;
  font-weight: 700;
  color: #6c757d;
  margin-bottom: 0.25rem;
}

/* input */
.conv-input {
  padding: 0.5rem 0.75rem;
  border-top: 1px solid rgba(0, 0, 0, 0.1);
  background: #ffffff;
}

.input-row {
  display: flex;
  align-items: center;
}

.attach-preview {
  margin-top: 0.35rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

/* reply bar */
.reply-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.35rem 0.5rem;
  margin-bottom: 0.4rem;
  background: rgba(13,110,253,0.08);
  border: 1px solid rgba(13,110,253,0.18);
  border-radius: 12px;
}
.reply-bar-text {
  font-size: 0.85rem;
  color: #212529;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  padding-right: 0.5rem;
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
.conv-avatar-img { width: 100%; height: 100%; object-fit: cover; }
.conv-avatar-fallback { font-weight: 800; font-size: 0.85rem; color: #495057; }

/* highlight when jump to reply */
.message-row.flash .bubble,
.flash .bubble {
  outline: 2px solid rgba(13,110,253,0.35);
  box-shadow: 0 0 0 6px rgba(13,110,253,0.12);
}
</style>
