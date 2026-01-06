// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'

import LoginView from '../views/LoginView.vue'
import AccountView from '../views/AccountView.vue'
import { isAuthenticated } from '../services/api.js'

/**
 * Routes overview:
 * - /login is public
 * - /, /conversations/* are protected (requireAuth)
 * - /account should also be protected (otherwise it can crash if token is stale)
 *
 * IMPORTANT:
 * - We keep a fallback route to avoid blank pages on unknown URLs.
 * - We preserve the "next" query parameter so after login the user can go back
 *   to the originally requested page.
 */
const routes = [
  {
    path: '/login',
    name: 'Login',
    component: LoginView
  },

  {
    path: '/account',
    name: 'Account',
    component: AccountView,
    meta: { requiresAuth: true } // ✅ protect account too
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
    // Fallback: any unknown route -> Home (or Login if not authenticated)
    path: '/:pathMatch(.*)*',
    redirect: { name: 'Home' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

/**
 * Global navigation guard:
 * - If user is NOT authenticated and tries to access a protected route,
 *   redirect to /login and store original route in `next`.
 * - If user IS authenticated and tries to go to /login, redirect to Home.
 */
router.beforeEach((to, from, next) => {
  const requiresAuth = !!to.meta.requiresAuth
  const authed = isAuthenticated()

  if (requiresAuth && !authed) {
    // Preserve the original destination to return after successful login.
    next({
      name: 'Login',
      query: { next: to.fullPath }
    })
    return
  }

  if (to.name === 'Login' && authed) {
    // Already logged in -> go to Home
    next({ name: 'Home' })
    return
  }

  next()
})

export default router
