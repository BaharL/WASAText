// services/api.js
// High-level API client for WASAText, built on top of axios instance from axios.js.

import axios from './axios.js'

/* ----------------------------------------------------------------------
 * TOKEN + AUTH
 * ------------------------------------------------------------------- */

export function getToken() {
  return localStorage.getItem('token')
}

export function setToken(token) {
  localStorage.setItem('token', token)
}

export function clearToken() {
  localStorage.removeItem('token')
}

export function isAuthenticated() {
  return !!getToken()
}

function buildAuthHeaders(extraHeaders = {}) {
  const token = getToken()
  const headers = { ...(extraHeaders || {}) }
  if (token) headers.Authorization = `Bearer ${token}`
  return headers
}

/* ----------------------------------------------------------------------
 * DIRECT CHAT TITLES (localStorage)
 * ------------------------------------------------------------------- */

const DIRECT_TITLES_KEY = 'directChatNames'

function loadDirectTitleMap() {
  const raw = localStorage.getItem(DIRECT_TITLES_KEY)
  if (!raw) return {}
  try {
    return JSON.parse(raw)
  } catch {
    return {}
  }
}

export function getDirectChatTitle(chatId) {
  const map = loadDirectTitleMap()
  return map[String(chatId)] || ''
}

export function saveDirectChatTitle(chatId, displayName) {
  const map = loadDirectTitleMap()
  map[String(chatId)] = displayName
  localStorage.setItem(DIRECT_TITLES_KEY, JSON.stringify(map))
}

/* ----------------------------------------------------------------------
 * INTERNAL REQUEST WRAPPER
 * ------------------------------------------------------------------- */

async function request(path, { method = 'GET', data, headers, params, signal } = {}) {
  try {
    const response = await axios.request({
      url: path,
      method,
      headers: buildAuthHeaders(headers),
      data,
      params,
      signal // ✅ allow AbortController cancellation
    })
    return response.data
  } catch (error) {
    // ✅ If request was canceled, rethrow it as-is (caller can ignore)
    if (error?.name === 'CanceledError' || error?.code === 'ERR_CANCELED') {
      throw error
    }

    let message = 'Request failed'

    const dataResp = error?.response?.data
    if (dataResp) {
      if (typeof dataResp === 'string') message = dataResp
      else if (dataResp.error) message = dataResp.error
      else if (dataResp.message) message = dataResp.message
    } else if (error?.message) {
      message = error.message
    }

    throw new Error(message)
  }
}

/* ----------------------------------------------------------------------
 * LOGOUT (client-side only)
 * ------------------------------------------------------------------- */

export function logout() {
  // salva token corrente per pulire la sua mappa titoli
  const t = getToken()
  if (t) localStorage.removeItem(`directChatNames:${t}`)

  localStorage.removeItem('token')
  localStorage.removeItem('username')
  localStorage.removeItem('photoUrl')
  localStorage.removeItem('theme')
  localStorage.removeItem('lastChat')

  window.dispatchEvent(new Event('profile-updated'))
}

/* ----------------------------------------------------------------------
 * LOGIN
 * ------------------------------------------------------------------- */

export async function login(name) {
  const data = await request('/session', {
    method: 'POST',
    data: { name }
  })

  setToken(data.identifier)
  return data
}

/* ----------------------------------------------------------------------
 * USER + PROFILE PHOTO
 * ------------------------------------------------------------------- */

export async function setMyUserName(username) {
  return request('/users/username', {
    method: 'PUT',
    data: { username }
  })
}

export async function listUsers(search) {
  return request('/users', {
    method: 'GET',
    params: { search }
  })
}

export async function setMyPhoto(file, { signal } = {}) {
  const formData = new FormData()
  formData.append('file', file)

  return request('/users/photo', {
    method: 'PUT',
    // NON forzare content-type: axios lo mette da solo col boundary
    data: formData,
    signal
  })
}

export async function getContext({ signal } = {}) {
  const ctx = await request('/context', { method: 'GET', signal })

  if (ctx?.username !== undefined) {
    localStorage.setItem('username', ctx.username || '')
  }

  if (ctx?.photoUrl !== undefined) {
    let p = ctx.photoUrl || ''
    // legacy fix
    if (p && p.startsWith('/uploads/')) p = `/v1${p}`
    localStorage.setItem('photoUrl', p)
  }

  window.dispatchEvent(new Event('profile-updated'))
  return ctx
}

/* ----------------------------------------------------------------------
 * CONVERSATIONS
 * ------------------------------------------------------------------- */

export async function getMyConversations({ signal } = {}) {
  return request('/conversations', { method: 'GET', signal })
}

export async function getConversation(chatId, { signal } = {}) {
  return request(`/conversations/${chatId}`, { method: 'GET', signal })
}

/* ----------------------------------------------------------------------
 * MESSAGE STATUS (delivered/read)
 * ------------------------------------------------------------------- */

// Marks all incoming messages in a conversation as "received" by the current user.
export async function markConversationReceived(chatId, { signal } = {}) {
  return request(`/conversations/${chatId}/received`, { method: 'POST', signal })
}

// Marks all incoming messages in a conversation as "read" by the current user.
export async function markConversationRead(chatId, { signal } = {}) {
  return request(`/conversations/${chatId}/read`, { method: 'POST', signal })
}

/* ----------------------------------------------------------------------
 * DIRECT CHATS
 * ------------------------------------------------------------------- */

export async function createDirect(otherIdentifier, { signal } = {}) {
  return request('/direct', {
    method: 'POST',
    data: { members: [otherIdentifier] },
    signal
  })
}

/* ----------------------------------------------------------------------
 * MESSAGES
 * ------------------------------------------------------------------- */

export async function sendMessage(payload, { signal } = {}) {
  return request('/messages', { method: 'POST', data: payload, signal })
}

export async function forwardMessage(messageId, toChatId, { signal } = {}) {
  return request(`/messages/${messageId}/forward`, {
    method: 'POST',
    data: { toChatId },
    signal
  })
}

export async function addReaction(messageId, emoji, { signal } = {}) {
  return request(`/messages/${messageId}/reactions`, {
    method: 'POST',
    data: { emoji },
    signal
  })
}

export async function removeReaction(messageId, emoji, { signal } = {}) {
  return request(`/messages/${messageId}/reactions/${encodeURIComponent(emoji)}`, {
    method: 'DELETE',
    signal
  })
}

export async function deleteMessage(messageId, { signal } = {}) {
  return request(`/messages/${messageId}`, { method: 'DELETE', signal })
}

/* ----------------------------------------------------------------------
 * GROUPS
 * ------------------------------------------------------------------- */

export async function createGroup(name, members = [], { signal } = {}) {
  return request('/groups', {
    method: 'POST',
    data: { name, members },
    signal
  })
}

export async function addToGroup(chatId, members, { signal } = {}) {
  return request(`/groups/${chatId}/members`, {
    method: 'POST',
    data: { members },
    signal
  })
}

export async function leaveGroup(chatId, { signal } = {}) {
  return request(`/groups/${chatId}/members/me`, { method: 'DELETE', signal })
}

export async function setGroupName(chatId, name, { signal } = {}) {
  return request(`/groups/${chatId}/name`, {
    method: 'PUT',
    data: { name },
    signal
  })
}

export async function setGroupPhoto(chatId, file, { signal } = {}) {
  const formData = new FormData()
  formData.append('file', file)

  return request(`/groups/${chatId}/photo`, {
    method: 'PUT',
    // NON mettere headers Content-Type: axios lo mette con boundary
    data: formData,
    signal
  })
}

/* ----------------------------------------------------------------------
 * MEDIA
 * ------------------------------------------------------------------- */

export async function uploadMedia(file, { signal } = {}) {
  const formData = new FormData()
  formData.append('file', file)

  return request('/media', {
    method: 'POST',
    data: formData,
    signal
  })
}
