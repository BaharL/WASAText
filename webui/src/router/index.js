// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import AccountView from '../views/AccountView.vue'

// We validate authentication by calling a real backend endpoint (/context),
// not just by checking if a token exists in localStorage.
import { getToken, getContext, logout } from '../services/api.js'

const routes = [
  {
    path: '/account',
    name: 'Account',
    component: AccountView,
    // IMPORTANT: Account page must be protected as well
    meta: { requiresAuth: true }
  },
  {
    path: '/login',
    name: 'Login',
    component: LoginView
  },
  {
    // HOME: all conversations (direct + groups together)
    path: '/',
    name: 'Home',
    component: () => import('../views/ConversationListView.vue'),
    meta: { requiresAuth: true }
  },
  {
    // Redirect /conversations -> Home (so both URLs work)
    path: '/conversations',
    redirect: { name: 'Home' }
  },
  {
    // Only direct / personal chats
    path: '/conversations/direct',
    name: 'DirectConversations',
    component: () => import('../views/ConversationListView.vue'),
    meta: { requiresAuth: true }
  },
  {
    // Only group chats
    path: '/conversations/groups',
    name: 'GroupConversations',
    component: () => import('../views/ConversationListView.vue'),
    meta: { requiresAuth: true }
  },
  {
    // Single conversation page
    path: '/conversations/:chatId',
    name: 'Conversation',
    component: () => import('../views/ConversationView.vue'),
    props: true,
    meta: { requiresAuth: true }
  },
  {
    // Fallback: any unknown route → Home
    // Home is protected, so unauthenticated users will be redirected to Login
    path: '/:pathMatch(.*)*',
    redirect: { name: 'Home' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

/**
 * Returns true only if:
 * 1) a token exists AND
 * 2) the backend confirms the session is valid (GET /context).
 *
 * If the token is stale or invalid (e.g. user not found in DB),
 * logout() is executed and the session is considered invalid.
 */
async function hasValidSession() {
  const token = getToken()
  if (!token) return false

  try {
    await getContext()
    return true
  } catch (err) {
    logout()
    return false
  }
}

/**
 * Global navigation guard:
 * - Protected routes require a valid session.
 * - Login route redirects to Home if the user is already authenticated.
 *
 * This prevents reopening the last page when the token is stale.
 */
router.beforeEach(async (to) => {
  const needsAuth = !!to.meta.requiresAuth

  if (needsAuth) {
    const ok = await hasValidSession()
    if (!ok) {
      return { name: 'Login' }
    }
  }

  if (to.name === 'Login') {
    const ok = await hasValidSession()
    if (ok) {
      return { name: 'Home' }
    }
  }

  return true
})

export default router
