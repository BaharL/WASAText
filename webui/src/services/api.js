// services/api.js
// High-level API client for WASAText, built on top of axios instance from axios.js.
// It manages:
// - auth token (identifier returned by /session)
// - Authorization header
// - error normalization
// - a typed function for each backend endpoint.

import axios from './axios.js'

/**
 * Get the stored auth token (user identifier).
 */
export function getToken() {
  return localStorage.getItem('token')
}

/**
 * Store auth token (user identifier) in localStorage.
 * @param {string} token
 */
export function setToken(token) {
  localStorage.setItem('token', token)
}

/**
 * Remove auth token from localStorage.
 */
export function clearToken() {
  localStorage.removeItem('token')
}

/**
 * Check whether the user is authenticated (token present).
 * @returns {boolean}
 */
export function isAuthenticated() {
  return !!getToken()
}

/**
 * Build headers including Authorization if a token is present.
 * @param {object} extraHeaders
 * @returns {object}
 */
function buildAuthHeaders(extraHeaders = {}) {
  const token = getToken()

  const headers = {
    ...(extraHeaders || {})
  }

  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  return headers
}

/**
 * Generic HTTP request wrapper around axios.
 *
 * @param {string} path - Relative API path (e.g. "/session", "/users").
 * @param {object} options
 * @param {string} [options.method="GET"]
 * @param {object} [options.data] - Request body for POST/PUT.
 * @param {object} [options.headers] - Extra headers.
 * @param {object} [options.params] - Query parameters.
 * @returns {Promise<any>} - Parsed response data.
 */
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

    if (error.response && error.response.data) {
      const data = error.response.data

      if (typeof data === 'string') {
        message = data
      } else if (data.error) {
        message = data.error
      } else if (data.message) {
        message = data.message
      }
    } else if (error.message) {
      message = error.message
    }

    throw new Error(message)
  }
}

/**
 * Clear token and effectively log out on the client side.
 */
export function logout() {
  localStorage.removeItem('token')
  localStorage.removeItem('username')
  localStorage.removeItem('photoUrl')
  localStorage.removeItem('theme')
  localStorage.removeItem('lastChat')

  // optional but useful: refresh UI immediately
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
    headers: {},
    data: formData
  })
}

export async function getContext() {
  const ctx = await request('/context', { method: 'GET' })

  // sync username
  if (ctx?.username !== undefined) {
    localStorage.setItem('username', ctx.username || '')
  }

  // sync photoUrl (ensure it is the new correct prefix)
  if (ctx?.photoUrl !== undefined) {
    let p = ctx.photoUrl || ''
    if (p && p.startsWith('/uploads/')) p = `/v1${p}` // legacy fix
    localStorage.setItem('photoUrl', p)
  }

  window.dispatchEvent(new Event('profile-updated'))
  return ctx
}



/* ----------------------------------------------------------------------
 * CONVERSATIONS
 * ------------------------------------------------------------------- */

export async function getMyConversations() {
  return request('/conversations', {
    method: 'GET'
  })
}

export async function getConversation(chatId) {
  return request(`/conversations/${chatId}`, {
    method: 'GET'
  })
}

/* ----------------------------------------------------------------------
 * DIRECT CHATS
 * ------------------------------------------------------------------- */

/**
 * Create (or get) a direct chat with exactly one other user.
 * POST /direct
 *
 * Backend expects:
 *   { "members": ["<otherIdentifier>"] }
 *
 * Returns:
 *   { "chatId": number }
 *
 * @param {string} otherIdentifier
 * @returns {Promise<{chatId: number}>}
 */
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
  return request('/messages', {
    method: 'POST',
    data: payload
  })
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
  return request(`/messages/${messageId}`, {
    method: 'DELETE'
  })
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
  return request(`/groups/${chatId}/members/me`, {
    method: 'DELETE'
  })
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
