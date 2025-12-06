<template>
  <div class="conversation-page">
    <!-- Header -->
    <header class="conv-header d-flex align-items-center justify-content-between">
      <div>
        <h1 class="h5 mb-0">
          {{ title }}
        </h1>
        <small class="text-muted">
          Chat ID: {{ chatId }}
        </small>
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
      <main class="conv-main">
        <!-- Nessun messaggio -->
        <p
          v-if="!loading && !error && messages.length === 0"
          class="text-muted text-center mt-3"
        >
          Nessun messaggio ancora. Inizia la conversazione! 💬
        </p>

        <!-- Lista messaggi -->
        <ul
          v-else
          ref="messageListEl"
          class="message-list"
        >
          <li
            v-for="m in messages"
            :key="m.id"
            class="message-row"
            :class="{ mine: isMine(m) }"
          >
            <div class="bubble">
              <div class="bubble-meta">
                <span class="author">
                  {{ isMine(m) ? 'You' : (m.senderName || m.sender?.name || 'Other') }}
                </span>
                <span class="time">
                  {{ formatTime(m.createdAt) }}
                </span>
              </div>

              <div class="bubble-content">
                <span v-if="m.text">
                  {{ m.text }}
                </span>
                <span
                  v-else
                  class="kind-pill"
                >
                  [{{ m.kind || 'message' }}]
                </span>
              </div>
            </div>
          </li>
        </ul>
      </main>
    </LoadingSpinner>

    <!-- Input messaggio -->
    <footer class="conv-input">
      <form
        class="input-row"
        @submit.prevent="handleSend"
      >
        <input
          v-model="draft"
          type="text"
          class="form-control"
          placeholder="Scrivi un messaggio…"
          :disabled="sending"
        >
        <button
          type="submit"
          class="btn btn-success ms-2"
          :disabled="sending || !draft.trim()"
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
import { getConversation, sendMessage } from '../services/api.js'

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

// Heuristica: se il messaggio è mio
function isMine(m) {
  return m.mine === true || m.direction === 'outgoing'
}

// Formatta orario in HH:MM
function formatTime(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit'
  })
}

// Scroll in fondo dopo aggiornamento
function scrollToBottom() {
  const el = messageListEl.value
  if (el) {
    el.scrollTop = el.scrollHeight
  }
}

// Carica dettagli + messaggi conversazione
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
    messages.value = Array.isArray(data)
      ? data
      : (data.messages || [])

    // Se il backend manda un titolo, usiamolo
    if (data.title || data.name || data.chatName) {
      title.value = data.title || data.name || data.chatName
    } else {
      title.value = `Conversation`
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

// Invia messaggio testuale usando la tua API: POST /messages
async function handleSend() {
  const text = draft.value.trim()
  if (!text || !chatId) return

  sending.value = true
  error.value = ''

  try {
    const payload = {
      chatId: Number(chatId), // il prof l’ha definito number
      kind: 'text',
      text
    }

    const created = await sendMessage(payload)

    if (created && created.id) {
      // API ritorna il messaggio creato → push
      messages.value.push(created)
    } else {
      // se non torna nulla, ricarichiamo
      await loadConversation()
    }

    draft.value = ''
    await nextTick()
    scrollToBottom()
  } catch (e) {
    console.error(e)
    error.value = e.message || 'Failed to send message.'
  } finally {
    sending.value = false
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
