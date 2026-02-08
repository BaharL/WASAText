<template>
  <div class="conversation-page">
    <!-- HEADER (sticky) -->
    <header class="conv-header">
      <div class="header-left">
        <div class="chat-title">
          <h1 class="h5 mb-0">{{ title }}</h1>
          <small class="text-muted">Chat ID: {{ chatId }}</small>
        </div>
      </div>

      <div class="header-right">
        <!-- Group actions -->
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

        <!-- TEMP toggle (remove later) -->
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

    <!-- RENAME GROUP -->
    <div v-if="isGroup && openRename" class="rename-bar">
      <input
        v-model.trim="newGroupName"
        class="form-control form-control-sm"
        placeholder="New group name…"
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

    <!-- BODY -->
    <div class="conv-body">
      <div v-if="error" class="alert alert-danger mx-3 mt-3">
        {{ error }}
      </div>

      <main ref="messagesEl" class="conv-main" @click="closeAllPopups">
        <p v-if="!loading && !error && messages.length === 0" class="text-muted mx-3 mt-3">
          No messages yet.
        </p>

        <div
          v-for="m in messages"
          :key="m.id"
          class="msg-row"
          :class="{ mine: !!m.mine, 'row-open': isAnyPopupOpen(m.id) }"
        >
          <div class="msg-bubble" :class="{ mine: !!m.mine }">
            <div v-if="m.forwardedFromMessageId" class="pill">
              Forwarded
            </div>

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

            <!-- content -->
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

            <!-- reactions row -->
            <div v-if="m.reactions && m.reactions.length" class="reactions-row">
              <button
                v-for="r in m.reactions"
                :key="r.emoji"
                class="reaction-chip"
                :class="{ mine: r.mine }"
                title="Toggle reaction"
                @click.stop="toggleReaction(m.id, r.emoji, r.mine)"
              >
                <span class="emoji">{{ r.emoji }}</span>
                <span class="count">{{ r.count }}</span>
              </button>
            </div>

            <!-- meta + ticks + dots -->
            <div class="msg-meta">
              <small class="text-muted">
                {{ m.mine ? 'You' : (m.senderName || 'unknown') }} · {{ formatTime(m.createdAt) }}
              </small>

              <div class="meta-right">
                <!-- Delivery ticks for outgoing messages only -->
                <span v-if="m.mine" class="ticks" :title="m.status || ''">
                  <span v-if="tickCount(m) === 1">✓</span>
                  <span v-else-if="tickCount(m) === 2">✓✓</span>
                </span>

                <button class="dots-btn" @click.stop="toggleMenu(m.id)">…</button>
              </div>

              <!-- ACTION MENU -->
              <div v-if="openMenuId === m.id" class="msg-menu" @click.stop>
                <button class="menu-item" @click="startReply(m)">Reply</button>
                <button class="menu-item" @click="openForward(m)">Forward</button>
                <button class="menu-item" @click="openReact(m)">React</button>
                <button v-if="m.mine" class="menu-item danger" @click="onDelete(m)">
                  Delete
                </button>
              </div>

              <!-- REACT POPOVER -->
              <div v-if="openReactId === m.id" class="react-pop" @click.stop>
                <button
                  v-for="e in emojiList"
                  :key="e"
                  class="emoji-btn"
                  @click="onReact(m, e)"
                >
                  {{ e }}
                </button>
              </div>

              <!-- FORWARD POPOVER -->
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

    <!-- COMPOSER -->
    <footer class="composer" @click.stop>
      <div v-if="replyingTo" class="reply-bar">
        <div class="reply-bar-left">
          Replying to <strong>{{ replyingTo.senderName }}</strong>:
          <span class="reply-bar-snippet">
            {{ replyingTo.text || (replyingTo.kind !== 'text' ? '[media]' : '') }}
          </span>
        </div>
        <button class="btn btn-sm btn-outline-secondary" @click="cancelReply">×</button>
      </div>

      <div class="composer-row">
        <label class="attach-btn" title="Attach media">
          📎
          <input type="file" class="d-none" accept="image/*,video/*" @change="onPickFile">
        </label>

        <input
          v-model="draft"
          class="form-control"
          placeholder="Write a message…"
          @keydown.enter.prevent="onSend"
        >

        <button class="btn btn-success" :disabled="sending || (!draft && !pickedFile)" @click="onSend">
          <span v-if="sending">Sending…</span>
          <span v-else>Send</span>
        </button>
      </div>

      <div v-if="pickedFile" class="picked-file">
        <span class="small">Selected: {{ pickedFile.name }}</span>
        <button class="btn btn-sm btn-outline-secondary" @click="clearPickedFile">Remove</button>
      </div>
    </footer>
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
const openReactId = ref(null)
const openForwardId = ref(null)

const forwardQuery = ref('')
const forwardResults = ref([])

const emojiList = ['👍','❤️','😂','😮','😢','🔥','👏','🙏']

const messagesEl = ref(null)

const isGroup = ref(false)
const openRename = ref(false)
const newGroupName = ref('')
const renaming = ref(false)
const changingGroupPhoto = ref(false)

function closeAllPopups() {
  openMenuId.value = null
  openReactId.value = null
  openForwardId.value = null
}

function isAnyPopupOpen(messageId) {
  return openMenuId.value === messageId || openReactId.value === messageId || openForwardId.value === messageId
}

// Returns 1 tick for "sent", 2 ticks for "received" or "read".
function tickCount(m) {
  const s = String(m?.status || '').toLowerCase()
  if (s === 'sent' || s === '') return 1
  if (s === 'received' || s === 'read') return 2
  // Fallback: unknown statuses still show 1 tick.
  return 1
}

function toggleMenu(id) {
  const next = (openMenuId.value === id) ? null : id
  openMenuId.value = next
  openReactId.value = null
  openForwardId.value = null
}

function openReact(m) {
  openReactId.value = m.id
  openForwardId.value = null
  openMenuId.value = null
}

function openForward(m) {
  openForwardId.value = m.id
  openReactId.value = null
  openMenuId.value = null
  forwardQuery.value = ''
  forwardResults.value = []
}

function startReply(m) {
  replyingTo.value = m
  closeAllPopups()
  nextTick(() => {
    const input = document.querySelector('.composer-row input.form-control')
    input?.focus()
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
  const m = String(dt).match(/(\d{2}):(\d{2})/)
  if (m) return `${m[1]}:${m[2]}`
  return dt
}

async function markIncomingAsReceivedAndRead() {
  // Best-effort: do not block UI if backend does not support these endpoints yet.
  try { await markConversationReceived(chatId.value) } catch {}
  try { await markConversationRead(chatId.value) } catch {}
}

/* LOAD conversation */
async function loadConversation() {
  loading.value = true
  error.value = ''
  closeAllPopups()

  try {
    const data = await getConversation(chatId.value)

    if (Array.isArray(data)) {
      messages.value = data
      title.value = 'Conversation'
    } else {
      messages.value = data?.messages || data?.Messages || data?.data || []
      title.value = data?.title || data?.name || 'Conversation'
      if (typeof data?.isGroup === 'boolean') isGroup.value = data.isGroup
    }

    await nextTick()
    scrollToBottom()

    // Update delivery/read status for incoming messages (best-effort).
    await markIncomingAsReceivedAndRead()
  } catch (e) {
    error.value = e.message || 'Cannot load conversation'
  } finally {
    loading.value = false
  }
}

function scrollToBottom() {
  const el = messagesEl.value
  if (!el) return
  el.scrollTop = el.scrollHeight
}

function scrollToMessage(messageId) {
  closeAllPopups()
  nextTick(() => {
    const target = document.getElementById(`msg-${messageId}`)
    target?.scrollIntoView({ behavior: 'smooth', block: 'center' })
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

/* SEND message */
async function onSend() {
  if (sending.value) return
  if (!draft.value && !pickedFile.value) return

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
      text = draft.value ? draft.value : null
    } else {
      kind = 'text'
      text = draft.value
    }

    const payload = {
      chatId: chatId.value,
      kind,
      text: text || undefined,
      mediaUrl: mediaUrl || undefined,
      replyToMessageId: replyingTo.value?.id || undefined
    }

    await sendMessage(payload)

    draft.value = ''
    pickedFile.value = null
    replyingTo.value = null

    await loadConversation()
  } catch (e) {
    error.value = e.message || 'Cannot send message'
  } finally {
    sending.value = false
  }
}

/* DELETE */
async function onDelete(m) {
  closeAllPopups()
  if (!m?.id) return
  try {
    await deleteMessage(m.id)
    await loadConversation()
  } catch (e) {
    error.value = e.message || 'Cannot delete message'
  }
}

/* REACTIONS */
async function onReact(m, emoji) {
  closeAllPopups()
  try {
    await addReaction(m.id, emoji)
    await loadConversation()
  } catch (e) {
    error.value = e.message || 'Cannot add reaction'
  }
}

async function toggleReaction(messageId, emoji, mine) {
  try {
    if (mine) await removeReaction(messageId, emoji)
    else await addReaction(messageId, emoji)
    await loadConversation()
  } catch (e) {
    error.value = e.message || 'Cannot toggle reaction'
  }
}

/* FORWARD by username */
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
    } catch {
      forwardResults.value = []
    }
  }, 250)
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
    error.value = e.message || 'Cannot forward message'
  }
}

/* GROUP: change photo */
async function onPickGroupPhoto(ev) {
  const f = ev.target.files?.[0]
  if (!f) return
  changingGroupPhoto.value = true
  error.value = ''

  try {
    await setGroupPhoto(chatId.value, f)
    await loadConversation()
  } catch (e) {
    error.value = e.message || 'Cannot change group photo'
  } finally {
    changingGroupPhoto.value = false
    ev.target.value = ''
  }
}

/* GROUP: rename */
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
    error.value = e.message || 'Cannot rename group'
  } finally {
    renaming.value = false
  }
}

/* CLOSE menus */
function onKeyDown(e) {
  if (e.key === 'Escape') closeAllPopups()
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
  z-index: 1;
}
.msg-row.mine{
  justify-content:flex-end;
}

/* When a popup is open, bring this row on top */
.msg-row.row-open{
  z-index: 99999;
}

.msg-bubble{
  max-width: 70%;
  background: white;
  border: 1px solid #e7e7e7;
  border-radius: 12px;
  padding: 10px 10px 6px;
  position: relative;
  overflow: visible; /* Important: allow popovers to overflow */
  box-shadow: 0 1px 2px rgba(0,0,0,0.04);
}
.msg-bubble.mine{
  background: #dff6df;
  border-color: #cfeccc;
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

.media-wrap{ margin-top: 4px; }
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

.msg-meta{
  display:flex;
  align-items:center;
  justify-content:space-between;
  gap: 10px;
  margin-top: 8px;
}

.meta-right{
  display:flex;
  align-items:center;
  gap: 8px;
}

.ticks{
  font-size: 12px;
  line-height: 1;
  user-select:none;
}

.dots-btn{
  border: 0;
  background: transparent;
  font-size: 18px;
  line-height: 1;
  padding: 0 6px;
  cursor:pointer;
}

/* Menus / popovers must be above everything */
.msg-menu,
.react-pop,
.forward-pop{
  position: absolute;
  right: 8px;
  top: 34px;
  z-index: 100000;
  background: white;
  border: 1px solid #ddd;
  box-shadow: 0 6px 20px rgba(0,0,0,0.12);
}

.msg-menu{
  border-radius: 10px;
  min-width: 140px;
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

.menu-item:hover{ background:#f3f3f3; }
.menu-item.danger{ color:#b00020; }

.react-pop{
  transform: translateY(44px);
  border-radius: 12px;
  padding: 8px;
  display:flex;
  gap:6px;
  flex-wrap: wrap;
  width: 210px;
}

.emoji-btn{
  border: 1px solid #ddd;
  background:white;
  border-radius: 10px;
  padding: 6px 8px;
  cursor:pointer;
  font-size: 16px;
}

.forward-pop{
  transform: translateY(44px);
  border-radius: 12px;
  padding: 10px;
  width: 240px;
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
</style>
