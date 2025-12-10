// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import { isAuthenticated } from '../services/api.js'
import AccountView from '../views/AccountView.vue'

const routes = [
  {
  path: '/account',
  name: 'Account',
  component: AccountView
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
    path: '/:pathMatch(.*)*',
    redirect: { name: 'Home' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Global navigation guard
router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !isAuthenticated()) {
    // Not logged → go to login
    next({ name: 'Login' })
  } else if (to.name === 'Login' && isAuthenticated()) {
    // Already logged → go to Home (all conversations)
    next({ name: 'Home' })
  } else {
    next()
  }
})

export default router
