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
  try { return JSON.parse(raw) } catch { return {} }
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

async function request(path, { method = 'GET', data, headers, params } = {}) {
  try {
    const response = await axios.request({
      url: path,
      method,
      headers: buildAuthHeaders(headers),
      data,
      params
    })
    return response.data
  } catch (error) {
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

export async function setMyPhoto(file) {
  const formData = new FormData()
  formData.append('file', file)

  return request('/users/photo', {
    method: 'PUT',
    // NON forzare content-type: axios lo mette da solo col boundary
    data: formData
  })
}

export async function getContext() {
  const ctx = await request('/context', { method: 'GET' })

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

export async function getMyConversations() {
  return request('/conversations', { method: 'GET' })
}

export async function getConversation(chatId) {
  return request(`/conversations/${chatId}`, { method: 'GET' })
}

/* ----------------------------------------------------------------------
 * DIRECT CHATS
 * ------------------------------------------------------------------- */

export async function createDirect(otherIdentifier) {
  return request('/direct', {
    method: 'POST',
    data: { members: [otherIdentifier] }
  })
}

/* ----------------------------------------------------------------------
 * MESSAGES
 * ------------------------------------------------------------------- */

export async function sendMessage(payload) {
  return request('/messages', { method: 'POST', data: payload })
}

export async function forwardMessage(messageId, toChatId) {
  return request(`/messages/${messageId}/forward`, {
    method: 'POST',
    data: { toChatId }
  })
}

export async function addReaction(messageId, emoji) {
  return request(`/messages/${messageId}/reactions`, {
    method: 'POST',
    data: { emoji }
  })
}

export async function removeReaction(messageId, emoji) {
  return request(`/messages/${messageId}/reactions/${encodeURIComponent(emoji)}`, {
    method: 'DELETE'
  })
}

export async function deleteMessage(messageId) {
  return request(`/messages/${messageId}`, { method: 'DELETE' })
}

/* ----------------------------------------------------------------------
 * GROUPS
 * ------------------------------------------------------------------- */

export async function createGroup(name, members = []) {
  return request('/groups', {
    method: 'POST',
    data: { name, members }
  })
}

export async function addToGroup(chatId, members) {
  return request(`/groups/${chatId}/members`, {
    method: 'POST',
    data: { members }
  })
}

export async function leaveGroup(chatId) {
  return request(`/groups/${chatId}/members/me`, { method: 'DELETE' })
}

export async function setGroupName(chatId, name) {
  return request(`/groups/${chatId}/name`, {
    method: 'PUT',
    data: { name }
  })
}

export async function setGroupPhoto(chatId, photoUrl) {
  return request(`/groups/${chatId}/photo`, {
    method: 'PUT',
    data: { photoUrl }
  })
}

export async function setGroupPhoto(chatId, file) {
  const form = new FormData()
  form.append('file', file)

  return apiFetch(`/groups/${chatId}/photo`, {
    method: 'PUT',
    body: form,
  })
}


/* ----------------------------------------------------------------------
 * MEDIA
 * ------------------------------------------------------------------- */

export async function uploadMedia(file) {
  const formData = new FormData()
  formData.append('file', file)

  return request('/media', {
    method: 'POST',
    headers: { 'Content-Type': 'multipart/form-data' },
    data: formData
  })
}

