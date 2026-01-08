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
import { ref, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import ErrorMsg from '../components/ErrorMsg.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

import {
  getConversation,
  sendMessage,
  uploadMedia,
  addReaction,
  deleteMessage,
  forwardMessage
} from '../services/api.js'

const route = useRoute()
// lascio stringa: l’API accetta anche "123" e la converte in numero
const chatId = route.params.chatId

const title = ref('Conversation')
const messages = ref([])
const loading = ref(false)
const sending = ref(false)
const error = ref('')
const draft = ref('')
const messageListEl = ref(null)

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

/* Heuristica: se il messaggio è mio */
function isMine(m) {
  return m.mine === true || m.direction === 'outgoing'
}

/* Formatta orario in HH:MM */
function formatTime(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}

/* Convert media URL to usable img src */
function toImgSrc(url) {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) return url

  // backend might return "/v1/..."
  if (url.startsWith('/v1/')) return `/api${url}`

  // legacy "/uploads/..."
  if (url.startsWith('/uploads/')) return `/api/v1${url}`

  return url.startsWith('/') ? url : `/${url}`
}

/* Scroll in fondo dopo aggiornamento */
function scrollToBottom() {
  const el = messageListEl.value
  if (el) el.scrollTop = el.scrollHeight
}

/* Carica dettagli + messaggi conversazione */
async function loadConversation() {
  if (!chatId) {
    error.value = 'Chat non valida.'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const data = await getConversation(chatId)

    // data.messages (schema del prof) oppure direttamente array
    messages.value = Array.isArray(data) ? data : (data.messages || [])

    // Se il backend manda un titolo, usiamolo
    if (data?.title || data?.name || data?.chatName) {
      title.value = data.title || data.name || data.chatName
    } else {
      title.value = 'Conversation'
    }

    await nextTick()
    scrollToBottom()
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to load messages.'
  } finally {
    loading.value = false
  }
}

/* Invia messaggio testuale */
async function handleSend() {
  const text = draft.value.trim()
  if (!text || !chatId) return

  sending.value = true
  error.value = ''

  try {
    await sendMessage({
      chatId: Number(chatId),
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
  if (!selectedFile.value || !chatId) return

  sendingMedia.value = true
  error.value = ''

  try {
    const up = await uploadMedia(selectedFile.value)

    // try common fields
    const mediaUrl = up?.url || up?.mediaUrl || up?.photoUrl || up?.path || up?.file || ''
    if (!mediaUrl) throw new Error('Upload ok but no media URL returned.')

    // NOTE: backend might expect kind 'media' OR 'photo' and field name might differ
    await sendMessage({
      chatId: Number(chatId),
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

/* Actions */
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

onMounted(() => {
  loadConversation()
})
</script>

<style scoped>
.conversation-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 56px); /* sottrae la navbar top del template */
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

/* Lista messaggi */
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

/* Media */
.msg-img {
  max-width: 260px;
  max-height: 260px;
  border-radius: 12px;
  display: block;
}

/* Actions */
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

/* Input in basso */
.conv-input {
  padding: 0.5rem 0.75rem;
  border-top: 1px solid rgba(0, 0, 0, 0.1);
  background: #ffffff;
}

.input-row {
  display: flex;
  align-items: center;
}
</style>
