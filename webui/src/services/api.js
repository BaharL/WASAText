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
    headers['Authorization'] = `Bearer ${token}`
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
  localStorage.removeItem('photoUrl') // ✅ NOME GIUSTO
  localStorage.removeItem('theme')
  localStorage.removeItem('lastChat')

  // opzionale ma utile: aggiorna subito UI
  window.dispatchEvent(new Event('profile-updated'))
}


/* ----------------------------------------------------------------------
 * LOGIN
 * ------------------------------------------------------------------- */

/**
 * Login or create user.
 * POST /session
 *
 * @param {string} name - Username to login with.
 * @returns {Promise<{identifier: string}>}
 */
export async function login(name) {
  const data = await request('/session', {
    method: 'POST',
    data: { name }
  })

  // The backend returns { identifier: "<uuid>" }
  setToken(data.identifier)
  return data
}

/* ----------------------------------------------------------------------
 * USER + PROFILE PHOTO
 * ------------------------------------------------------------------- */

/**
 * Change my username.
 * PUT /users/username
 *
 * @param {string} username - New username.
 * @returns {Promise<any>} - Updated user object.
 */
export async function setMyUserName(username) {
  return request('/users/username', {
    method: 'PUT',
    data: { username }
  })
}

/**
 * Search users by username.
 * GET /users?search=...
 *
 * @param {string} search - Search term.
 * @returns {Promise<{users: Array}>}
 */
export async function listUsers(search) {
  return request('/users', {
    method: 'GET',
    params: { search }
  })
}

/**
 * Upload or change my profile photo.
 * PUT /users/photo
 *
 * @param {File} file - Image file selected from an <input type="file">.
 * @returns {Promise<any>}
 */
export async function setMyPhoto(file) {
  const formData = new FormData()
  formData.append('file', file)

  return request('/users/photo', {
    method: 'PUT',
    // Let the browser set Content-Type (multipart/form-data with boundary)
    headers: {},
    data: formData
  })
}
// GET /context
export async function getContext() {
  return request('/context', { method: 'GET' })
}


/* ----------------------------------------------------------------------
 * CONVERSATIONS
 * ------------------------------------------------------------------- */

/**
 * Get my conversations.
 * GET /conversations
 *
 * @returns {Promise<{conversations: Array}>}
 */
export async function getMyConversations() {
  return request('/conversations', {
    method: 'GET'
  })
}

/**
 * Get messages of a conversation.
 * GET /conversations/{chatId}
 *
 * @param {number} chatId
 * @returns {Promise<{messages: Array}>}
 */
export async function getConversation(chatId) {
  return request(`/conversations/${chatId}`, {
    method: 'GET'
  })
}

/* ----------------------------------------------------------------------
 * MESSAGES
 * ------------------------------------------------------------------- */

/**
 * Send a message.
 * POST /messages
 *
 * Body must match SendMessageRequest schema:
 * {
 *   chatId: number,
 *   kind: "text" | "gif" | "image",
 *   text?: string,
 *   mediaUrl?: string
 * }
 *
 * @param {object} payload
 * @returns {Promise<any>} - Created Message.
 */
export async function sendMessage(payload) {
  return request('/messages', {
    method: 'POST',
    data: payload
  })
}

/**
 * Forward a message to another conversation.
 * POST /messages/{messageId}/forward
 *
 * @param {number} messageId
 * @param {number} toChatId
 * @returns {Promise<any>} - Created forwarded Message.
 */
export async function forwardMessage(messageId, toChatId) {
  return request(`/messages/${messageId}/forward`, {
    method: 'POST',
    data: { toChatId }
  })
}

/**
 * Add a reaction (emoji) to a message.
 * POST /messages/{messageId}/reactions
 *
 * @param {number} messageId
 * @param {string} emoji
 * @returns {Promise<any>} - Created Reaction.
 */
export async function addReaction(messageId, emoji) {
  return request(`/messages/${messageId}/reactions`, {
    method: 'POST',
    data: { emoji }
  })
}

/**
 * Remove a reaction from a message.
 * DELETE /messages/{messageId}/reactions/{emoji}
 *
 * @param {number} messageId
 * @param {string} emoji
 * @returns {Promise<void>}
 */
export async function removeReaction(messageId, emoji) {
  return request(`/messages/${messageId}/reactions/${encodeURIComponent(emoji)}`, {
    method: 'DELETE'
  })
}

/**
 * Delete a sent message.
 * DELETE /messages/{messageId}
 *
 * @param {number} messageId
 * @returns {Promise<void>}
 */
export async function deleteMessage(messageId) {
  return request(`/messages/${messageId}`, {
    method: 'DELETE'
  })
}

/* ----------------------------------------------------------------------
 * GROUPS
 * ------------------------------------------------------------------- */

/**
 * Create a new group.
 * POST /groups
 *
 * @param {string} name - Group name.
 * @param {string[]} [members=[]] - Array of user identifiers (UUIDs).
 * @returns {Promise<{chatId: number}>}
 */
export async function createGroup(name, members = []) {
  return request('/groups', {
    method: 'POST',
    data: { name, members }
  })
}

/**
 * Add members to an existing group.
 * POST /groups/{chatId}/members
 *
 * @param {number} chatId
 * @param {string[]} members - Array of user identifiers (UUIDs).
 * @returns {Promise<void>}
 */
export async function addToGroup(chatId, members) {
  return request(`/groups/${chatId}/members`, {
    method: 'POST',
    data: { members }
  })
}

/**
 * Leave a group (current user).
 * DELETE /groups/{chatId}/members/me
 *
 * @param {number} chatId
 * @returns {Promise<void>}
 */
export async function leaveGroup(chatId) {
  return request(`/groups/${chatId}/members/me`, {
    method: 'DELETE'
  })
}

/**
 * Change group name.
 * PUT /groups/{chatId}/name
 *
 * @param {number} chatId
 * @param {string} name - New group name.
 * @returns {Promise<{message: string}>}
 */
export async function setGroupName(chatId, name) {
  return request(`/groups/${chatId}/name`, {
    method: 'PUT',
    data: { name }
  })
}

/**
 * Set group photo URL.
 * PUT /groups/{chatId}/photo
 *
 * @param {number} chatId
 * @param {string} photoUrl - Public URL of the group photo.
 * @returns {Promise<{message: string}>}
 */
export async function setGroupPhoto(chatId, photoUrl) {
  return request(`/groups/${chatId}/photo`, {
    method: 'PUT',
    data: { photoUrl }
  })
}
