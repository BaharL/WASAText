<template>
  <div class="conversation-page">
    <!-- HEADER (sticky) -->
    <header class="conv-header">
      <div class="header-left">
        <button 
          type="button" 
          class="btn btn-sm btn-outline-secondary me-2"
          @click="$router.push('/conversations')"
          title="Back to conversations"
        >
          ← Back
        </button>
        <div class="chat-title">
          <h1 class="h5 mb-0">{{ title }}</h1>
          <small class="text-muted">Chat ID: {{ chatId }}</small>
        </div>
      </div>

      <div class="header-right">
        <div v-if="isGroup" class="group-actions">
          <label class="btn btn-sm btn-outline-primary mb-0" title="Change group photo">
            <span v-if="changingGroupPhoto">Uploading…</span>
            <span v-else>Change photo</span>
            <input
              type="file"
              class="d-none"
              accept="image/*"
              :disabled="changingGroupPhoto"
              @change="onPickGroupPhoto"
            >
          </label>

          <button
            type="button"
            class="btn btn-sm btn-outline-primary"
            @click="openRename = !openRename"
          >
            Rename
          </button>
        </div>

        <button
          type="button"
          class="btn btn-sm btn-outline-secondary"
          title="Temporary toggle (remove later)"
          @click="isGroup = !isGroup"
        >
          Group tools
        </button>

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

    <div v-if="isGroup && openRename" class="rename-bar">
      <input
        v-model.trim="newGroupName"
        class="form-control form-control-sm"
        placeholder="New group name…"
        @keydown.enter="onRenameGroup"
      >
      <button
        class="btn btn-sm btn-success"
        :disabled="!newGroupName || renaming"
        @click="onRenameGroup"
      >
        <span v-if="renaming">Saving…</span>
        <span v-else>Save</span>
      </button>
      <button class="btn btn-sm btn-outline-secondary" @click="openRename=false">
        Cancel
      </button>
    </div>

    <div class="conv-body">
      <div v-if="error" class="alert alert-danger mx-3 mt-3">
        {{ error }}
      </div>

      <main ref="messagesEl" class="conv-main" @click="closeAllPopups">
        <p v-if="!loading && !error && messages.length === 0" class="text-muted mx-3 mt-3">
          No messages yet. Start the conversation!
        </p>

        <div
          v-for="m in messages"
          :key="m.id"
          class="msg-row"
          :class="{ mine: !!m.mine, 'row-open': openMenuId===m.id || openForwardId===m.id }"
        >
          <div class="msg-bubble" :class="{ mine: !!m.mine }">
            <div v-if="m.forwardedFromMessageId" class="pill">Forwarded</div>

            <div
              v-if="m.replyToMessageId"
              class="reply-preview"
              title="Go to replied message"
              @click.stop="scrollToMessage(m.replyToMessageId)"
            >
              <div class="reply-title">
                Reply to <strong>{{ repliedSender(m.replyToMessageId) }}</strong>
              </div>
              <div class="reply-snippet">
                {{ repliedSnippet(m.replyToMessageId) }}
              </div>
            </div>

            <div class="msg-content">
              <div v-if="m.kind === 'text'">
                {{ m.text }}
              </div>

              <div v-else class="media-wrap">
                <img
                  v-if="m.kind === 'image'"
                  class="media-img"
                  :src="normalizeMediaUrl(m.mediaUrl)"
                  alt="image"
                  loading="lazy"
                >
                <video
                  v-else-if="m.kind === 'gif'"
                  class="media-img"
                  :src="normalizeMediaUrl(m.mediaUrl)"
                  autoplay
                  muted
                  loop
                  playsinline
                />
                <a
                  v-else
                  :href="normalizeMediaUrl(m.mediaUrl)"
                  target="_blank"
                  rel="noreferrer"
                >Open media</a>
              </div>
            </div>

            <div v-if="m.reactions && m.reactions.length" class="reactions-row">
              <button
                v-for="r in m.reactions"
                :key="r.emoji"
                class="reaction-chip"
                :class="{ mine: r.mine }"
                :title="r.mine ? 'Remove your reaction' : 'Add this reaction'"
                @click.stop="toggleReaction(m.id, r.emoji, r.mine)"
              >
                <span class="emoji">{{ r.emoji }}</span>
                <span class="count">{{ r.count }}</span>
              </button>
            </div>

            <div class="msg-meta">
              <small class="text-muted">
                {{ m.mine ? 'You' : m.senderName }} · {{ formatTime(m.createdAt) }}
                <span v-if="m.mine" class="ms-1 tick">
                  <span v-if="m.status === 'sent'" title="Sent">✓</span>
                  <span v-else-if="m.status === 'received'" title="Received">✓✓</span>
                  <span v-else-if="m.status === 'read'" title="Read" class="read-tick">✓✓</span>
                </span>
              </small>

              <button 
                class="dots-btn" 
                :aria-label="`Message options for ${m.mine ? 'your' : m.senderName + '\'s'} message`"
                @click.stop="toggleMenu(m.id)"
              >
                …
              </button>

              <div v-if="openMenuId === m.id" class="msg-menu" @click.stop>
                <button class="menu-item" @click="startReply(m)">
                  💬 Reply
                </button>
                <button class="menu-item" @click="openForward(m)">
                  ➡️ Forward
                </button>
                <button class="menu-item" @click="openReact(m, $event)">
                  😊 React
                </button>
                <button v-if="m.mine" class="menu-item danger" @click="onDelete(m)">
                  🗑️ Delete
                </button>
              </div>

              <div v-if="openForwardId === m.id" class="forward-pop" @click.stop>
                <div class="forward-title">Forward to (username)</div>
                <input
                  v-model.trim="forwardQuery"
                  class="form-control form-control-sm"
                  placeholder="Type username…"
                  @input="searchUsers"
                >
                <div class="forward-results">
                  <button
                    v-for="u in forwardResults"
                    :key="u.identifier"
                    class="forward-user"
                    @click="doForwardToUser(m, u)"
                  >
                    {{ u.username || u.name || u.identifier }}
                  </button>

                  <div v-if="forwardQuery && forwardResults.length === 0" class="text-muted small px-1 py-2">
                    No users found
                  </div>
                </div>
              </div>
            </div>

            <div :id="`msg-${m.id}`" />
          </div>
        </div>
      </main>
    </div>

    <footer class="composer" @click.stop>
      <div v-if="replyingTo" class="reply-bar">
        <div class="reply-bar-left">
          Replying to <strong>{{ replyingTo.mine ? 'yourself' : replyingTo.senderName }}</strong>:
          <span class="reply-bar-snippet">
            {{ replyingTo.text || (replyingTo.kind !== 'text' ? '[media]' : '') }}
          </span>
        </div>
        <button 
          class="btn btn-sm btn-outline-secondary" 
          title="Cancel reply"
          @click="cancelReply"
        >
          ×
        </button>
      </div>

      <div class="composer-row">
        <label class="attach-btn" title="Attach media">
          📎
          <input 
            type="file" 
            class="d-none" 
            accept="image/*,video/*" 
            @change="onPickFile"
          >
        </label>

        <input
          ref="messageInput"
          v-model="draft"
          class="form-control"
          placeholder="Write a message…"
          @keydown.enter.exact.prevent="onSend"
        >

        <button 
          class="btn btn-success" 
          :disabled="sending || (!draft && !pickedFile)" 
          @click="onSend"
        >
          <span v-if="sending">Sending…</span>
          <span v-else>Send</span>
        </button>
      </div>

      <div v-if="pickedFile" class="picked-file">
        <span class="small">📎 Selected: {{ pickedFile.name }}</span>
        <button class="btn btn-sm btn-outline-secondary" @click="clearPickedFile">Remove</button>
      </div>
    </footer>

    <!-- Fixed-position reaction popover (always on top) -->
    <div
      v-if="reactOverlay.open"
      class="react-overlay"
      :style="{ left: reactOverlay.x + 'px', top: reactOverlay.y + 'px' }"
      @click.stop
    >
      <button
        v-for="e in emojiList"
        :key="e"
        class="emoji-btn"
        :title="`React with ${e}`"
        @click="onReact(reactOverlay.message, e)"
      >
        {{ e }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  getConversation,
  sendMessage,
  uploadMedia,
  deleteMessage,
  addReaction,
  removeReaction,
  listUsers,
  createDirect,
  forwardMessage,
  setGroupPhoto,
  setGroupName,
  markConversationReceived,
  markConversationRead
} from '../services/api.js'

const route = useRoute()
const router = useRouter()

const chatId = computed(() => Number(route.params.chatId))

const loading = ref(false)
const sending = ref(false)
const error = ref('')

const title = ref('Conversation')
const messages = ref([])

const draft = ref('')
const pickedFile = ref(null)

const replyingTo = ref(null)

const openMenuId = ref(null)
const openForwardId = ref(null)

const forwardQuery = ref('')
const forwardResults = ref([])

const emojiList = ['👍','❤️','😂','😮','😢','🔥','👏','🙏']

const messagesEl = ref(null)
const messageInput = ref(null)

const isGroup = ref(false)
const openRename = ref(false)
const newGroupName = ref('')
const renaming = ref(false)
const changingGroupPhoto = ref(false)

// Fixed reaction overlay state
const reactOverlay = ref({
  open: false,
  x: 0,
  y: 0,
  message: null
})

function closeAllPopups() {
  openMenuId.value = null
  openForwardId.value = null
  reactOverlay.value = { open: false, x: 0, y: 0, message: null }
}

function toggleMenu(id) {
  const next = (openMenuId.value === id) ? null : id
  openMenuId.value = next
  openForwardId.value = null
  reactOverlay.value = { open: false, x: 0, y: 0, message: null }
}

function openReact(m, ev) {
  // Compute a stable viewport position (fixed overlay)
  const rect = ev?.target?.getBoundingClientRect()
  const baseX = rect ? rect.right : window.innerWidth / 2
  const baseY = rect ? rect.bottom : window.innerHeight / 2

  const width = 230
  const height = 70
  let x = Math.min(baseX - width, window.innerWidth - width - 12)
  let y = Math.min(baseY + 8, window.innerHeight - height - 12)
  if (x < 12) x = 12
  if (y < 12) y = 12

  reactOverlay.value = { open: true, x, y, message: m }
  openMenuId.value = null
  openForwardId.value = null
}

function openForward(m) {
  openForwardId.value = m.id
  reactOverlay.value = { open: false, x: 0, y: 0, message: null }
  forwardQuery.value = ''
  forwardResults.value = []
}

function startReply(m) {
  replyingTo.value = m
  closeAllPopups()
  nextTick(() => {
    messageInput.value?.focus()
  })
}

function cancelReply() {
  replyingTo.value = null
}

function clearPickedFile() {
  pickedFile.value = null
}

function onPickFile(ev) {
  const f = ev.target.files?.[0]
  if (!f) return
  pickedFile.value = f
  // Reset input per permettere di selezionare lo stesso file di nuovo
  ev.target.value = ''
}

function normalizeMediaUrl(u) {
  if (!u) return ''
  if (u.startsWith('http://') || u.startsWith('https://')) return u
  if (u.startsWith('/v1/')) return u
  if (u.startsWith('/uploads/')) return `/v1${u}`
  return u
}

function formatTime(dt) {
  if (!dt) return ''
  
  // Se è già formattato come HH:MM, ritorna così
  const timeMatch = String(dt).match(/(\d{2}):(\d{2})/)
  if (timeMatch) return `${timeMatch[1]}:${timeMatch[2]}`
  
  // Prova a parsare come Date
  try {
    const date = new Date(dt)
    if (!isNaN(date.getTime())) {
      return date.toLocaleTimeString('it-IT', { 
        hour: '2-digit', 
        minute: '2-digit' 
      })
    }
  } catch (e) {
    // Fallback
  }
  
  return String(dt)
}

async function loadConversation() {
  loading.value = true
  error.value = ''
  closeAllPopups()

  try {
    // Mark as received/read (best effort; do not block UI on failure)
    try { 
      await markConversationReceived(chatId.value) 
    } catch (e) {
      console.warn('Failed to mark as received:', e)
    }
    
    try { 
      await markConversationRead(chatId.value) 
    } catch (e) {
      console.warn('Failed to mark as read:', e)
    }

    const data = await getConversation(chatId.value)

    if (Array.isArray(data)) {
      messages.value = data
      title.value = 'Conversation'
    } else {
      messages.value = data?.messages || data?.Messages || data?.data || []
      title.value = data?.title || data?.name || 'Conversation'
      if (typeof data?.isGroup === 'boolean') {
        isGroup.value = data.isGroup
      }
    }

    await nextTick()
    scrollToBottom()
  } catch (e) {
    console.error('Load conversation error:', e)
    error.value = e.message || 'Cannot load conversation'
  } finally {
    loading.value = false
  }
}

function scrollToBottom() {
  const el = messagesEl.value
  if (!el) return
  requestAnimationFrame(() => {
    el.scrollTop = el.scrollHeight
  })
}

function scrollToMessage(messageId) {
  closeAllPopups()
  nextTick(() => {
    const target = document.getElementById(`msg-${messageId}`)
    if (target) {
      target.scrollIntoView({ behavior: 'smooth', block: 'center' })
      // Highlight effect
      target.parentElement?.classList.add('highlight')
      setTimeout(() => {
        target.parentElement?.classList.remove('highlight')
      }, 2000)
    }
  })
}

function repliedMessage(id) {
  return messages.value.find(x => x.id === id) || null
}

function repliedSender(id) {
  const m = repliedMessage(id)
  return m ? (m.mine ? 'You' : (m.senderName || 'unknown')) : 'unknown'
}

function repliedSnippet(id) {
  const m = repliedMessage(id)
  if (!m) return '(message not found)'
  if (m.kind === 'text') return (m.text || '').slice(0, 80)
  return '[media]'
}

async function onSend() {
  if (sending.value) return
  if (!draft.value.trim() && !pickedFile.value) return

  sending.value = true
  error.value = ''

  try {
    let kind = 'text'
    let mediaUrl = null
    let text = null

    if (pickedFile.value) {
      const up = await uploadMedia(pickedFile.value)
      mediaUrl = up?.mediaUrl || up?.url || up?.path || up
      if (!mediaUrl) throw new Error('Upload failed: no media url returned')

      const t = pickedFile.value.type || ''
      kind = t.startsWith('image/') ? 'image' : 'gif'
      text = draft.value.trim() || null
    } else {
      text = draft.value.trim()
    }

    await sendMessage({
      chatId: chatId.value,
      kind,
      text: text || undefined,
      mediaUrl: mediaUrl || undefined,
      replyToMessageId: replyingTo.value?.id || undefined
    })

    draft.value = ''
    pickedFile.value = null
    replyingTo.value = null

    await loadConversation()
    
    // Refocus input
    nextTick(() => {
      messageInput.value?.focus()
    })
  } catch (e) {
    console.error('Send message error:', e)
    error.value = e.message || 'Cannot send message'
  } finally {
    sending.value = false
  }
}

async function onDelete(m) {
  closeAllPopups()
  if (!m?.id) return
  
  if (!confirm('Are you sure you want to delete this message?')) return

  try {
    await deleteMessage(m.id)
    await loadConversation()
  } catch (e) {
    console.error('Delete error:', e)
    error.value = e.message || 'Cannot delete message'
  }
}

async function onReact(m, emoji) {
  closeAllPopups()
  try {
    await addReaction(m.id, emoji)
    await loadConversation()
  } catch (e) {
    console.error('Add reaction error:', e)
    error.value = e.message || 'Cannot add reaction'
  }
}

async function toggleReaction(messageId, emoji, mine) {
  try {
    if (mine) {
      await removeReaction(messageId, emoji)
    } else {
      await addReaction(messageId, emoji)
    }
    await loadConversation()
  } catch (e) {
    console.error('Toggle reaction error:', e)
    error.value = e.message || 'Cannot toggle reaction'
  }
}

let searchTimer = null
function searchUsers() {
  clearTimeout(searchTimer)
  const q = forwardQuery.value
  if (!q) {
    forwardResults.value = []
    return
  }
  searchTimer = setTimeout(async () => {
    try {
      const res = await listUsers(q)
      forwardResults.value = Array.isArray(res) ? res : (res?.users || [])
    } catch (e) {
      console.error('Search users error:', e)
      forwardResults.value = []
    }
  }, 300)
}

async function doForwardToUser(m, user) {
  closeAllPopups()
  try {
    const direct = await createDirect(user.identifier)
    const newChatId = direct?.chatId || direct?.id || direct
    if (!newChatId) throw new Error('Cannot create/open direct chat')

    await forwardMessage(m.id, Number(newChatId))
    await router.push(`/conversations/${newChatId}`)
    await loadConversation()
  } catch (e) {
    console.error('Forward error:', e)
    error.value = e.message || 'Cannot forward message'
  }
}

async function onPickGroupPhoto(ev) {
  const f = ev.target.files?.[0]
  if (!f) return
  changingGroupPhoto.value = true
  error.value = ''

  try {
    await setGroupPhoto(chatId.value, f)
    await loadConversation()
  } catch (e) {
    console.error('Change photo error:', e)
    error.value = e.message || 'Cannot change group photo'
  } finally {
    changingGroupPhoto.value = false
    ev.target.value = ''
  }
}

async function onRenameGroup() {
  if (!newGroupName.value) return
  renaming.value = true
  error.value = ''

  try {
    await setGroupName(chatId.value, newGroupName.value)
    openRename.value = false
    newGroupName.value = ''
    await loadConversation()
  } catch (e) {
    console.error('Rename error:', e)
    error.value = e.message || 'Cannot rename group'
  } finally {
    renaming.value = false
  }
}

function onKeyDown(e) {
  if (e.key === 'Escape') {
    closeAllPopups()
  }
}

function onDocClick() {
  closeAllPopups()
}

onMounted(() => {
  document.addEventListener('keydown', onKeyDown)
  document.addEventListener('click', onDocClick)
  loadConversation()
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeyDown)
  document.removeEventListener('click', onDocClick)
  if (searchTimer) clearTimeout(searchTimer)
})

watch(() => chatId.value, async () => {
  await loadConversation()
})
</script>
<style scoped>
.conversation-page{
  height: calc(100vh - 60px);
  display:flex;
  flex-direction:column;
  background:#f6f7f8;
}

.conv-header{
  position: sticky;
  top: 0;
  z-index: 2000;
  background: white;
  border-bottom: 1px solid #e6e6e6;
  padding: 12px 16px;
  display:flex;
  align-items:center;
  justify-content:space-between;
}

.header-right{
  display:flex;
  align-items:center;
  gap:10px;
}

.group-actions{
  display:flex;
  align-items:center;
  gap:8px;
}

.rename-bar{
  padding: 10px 16px;
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
  display:flex;
  align-items:center;
  gap:10px;
}

.conv-body{
  flex: 1;
  min-height: 0;
  display:flex;
  flex-direction:column;
}

.conv-main{
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 16px;
}

.msg-row{
  display:flex;
  margin-bottom: 12px;
  position: relative;
}
.msg-row.mine{
  justify-content:flex-end;
}
.msg-row.row-open{
  z-index: 50;
}

.msg-bubble{
  max-width: 70%;
  background: white;
  border: 1px solid #e7e7e7;
  border-radius: 12px;
  padding: 10px 10px 6px;
  position: relative;
  box-shadow: 0 1px 2px rgba(0,0,0,0.04);
}
.msg-bubble.mine{
  background: #dff6df;
  border-color: #cfeccc;
}

.tick{
  font-weight: 700;
}

.pill{
  display:inline-block;
  font-size: 12px;
  font-weight: 600;
  color:#2b6b2b;
  background:#eef9ee;
  border:1px solid #cfeccc;
  padding:2px 8px;
  border-radius: 999px;
  margin-bottom: 6px;
}

.reply-preview{
  border-left: 3px solid #5b9bd5;
  background: rgba(91,155,213,0.10);
  padding: 6px 8px;
  border-radius: 10px;
  cursor: pointer;
  margin-bottom: 8px;
}
.reply-title{
  font-size: 12px;
  color:#2f5f8c;
}
.reply-snippet{
  font-size: 12px;
  color:#2a2a2a;
  opacity: 0.85;
  margin-top: 2px;
}

.media-wrap{
  margin-top: 4px;
}
.media-img{
  max-width: 320px;
  width: 100%;
  border-radius: 10px;
  display:block;
}

.reactions-row{
  display:flex;
  gap:6px;
  flex-wrap: wrap;
  margin-top: 8px;
}
.reaction-chip{
  border: 1px solid #ddd;
  background: white;
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 12px;
  display:flex;
  gap:6px;
  align-items:center;
  cursor:pointer;
}
.reaction-chip.mine{
  border-color:#2b6b2b;
}
.reaction-chip .count{
  font-weight: 700;
}

.msg-meta{
  display:flex;
  align-items:center;
  justify-content:space-between;
  gap: 10px;
  margin-top: 8px;
}

.dots-btn{
  border: 0;
  background: transparent;
  font-size: 18px;
  line-height: 1;
  padding: 0 6px;
  cursor:pointer;
}

.msg-menu{
  position: absolute;
  right: 8px;
  top: 34px;
  z-index: 20;
  background: white;
  border: 1px solid #ddd;
  border-radius: 10px;
  min-width: 140px;
  box-shadow: 0 6px 20px rgba(0,0,0,0.12);
  overflow: hidden;
}
.menu-item{
  width: 100%;
  text-align: left;
  padding: 8px 10px;
  border: 0;
  background: transparent;
  cursor: pointer;
  font-size: 14px;
}
.menu-item:hover{
  background:#f3f3f3;
}
.menu-item.danger{
  color:#b00020;
}

.forward-pop{
  position:absolute;
  right: 8px;
  top: 34px;
  transform: translateY(44px);
  z-index: 20;
  background:white;
  border:1px solid #ddd;
  border-radius: 12px;
  padding: 10px;
  width: 240px;
  box-shadow: 0 6px 20px rgba(0,0,0,0.12);
}

.forward-title{
  font-size: 12px;
  font-weight: 700;
  margin-bottom: 6px;
}
.forward-results{
  margin-top: 8px;
  max-height: 160px;
  overflow:auto;
  display:flex;
  flex-direction:column;
  gap:6px;
}
.forward-user{
  border: 1px solid #e3e3e3;
  background:white;
  border-radius: 10px;
  padding: 6px 8px;
  text-align:left;
  cursor:pointer;
}
.forward-user:hover{
  background:#f3f3f3;
}

.composer{
  position: sticky;
  bottom: 0;
  z-index: 2000;
  background: white;
  border-top: 1px solid #e6e6e6;
  padding: 10px 16px;
}

.reply-bar{
  display:flex;
  align-items:center;
  justify-content:space-between;
  gap: 12px;
  background: #eef5ff;
  border: 1px solid #cfe3ff;
  border-radius: 12px;
  padding: 8px 10px;
  margin-bottom: 8px;
}
.reply-bar-snippet{
  opacity: 0.8;
  margin-left: 6px;
}

.composer-row{
  display:flex;
  align-items:center;
  gap: 10px;
}

.attach-btn{
  border: 1px solid #ddd;
  background: white;
  border-radius: 10px;
  padding: 6px 10px;
  cursor:pointer;
  user-select:none;
}

.picked-file{
  margin-top: 8px;
  display:flex;
  align-items:center;
  justify-content:space-between;
  gap: 10px;
}

/* Fixed overlay for reactions (always on top of everything) */
.react-overlay{
  position: fixed;
  z-index: 999999;
  background: white;
  border: 1px solid #ddd;
  border-radius: 12px;
  padding: 8px;
  display:flex;
  gap:6px;
  flex-wrap: wrap;
  width: 230px;
  box-shadow: 0 10px 30px rgba(0,0,0,0.20);
}

.emoji-btn{
  border: 1px solid #ddd;
  background:white;
  border-radius: 10px;
  padding: 6px 8px;
  cursor:pointer;
  font-size: 16px;
}
.emoji-btn:hover{
  background:#f3f3f3;
}
</style>
