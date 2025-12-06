// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import { isAuthenticated } from '../services/api.js'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: LoginView
  },
  {
    // Redirect root "/" to conversations list
    path: '/',
    redirect: { name: 'Conversations' }
  },
  {
    // All conversations (default list)
    path: '/conversations',
    name: 'Conversations',
    component: () => import('../views/ConversationListView.vue'),
    meta: { requiresAuth: true }
  },
  {
    // Only direct (1-to-1) conversations
    path: '/conversations/direct',
    name: 'DirectConversations',
    component: () => import('../views/ConversationListView.vue'),
    meta: { requiresAuth: true }
  },
  {
    // Only group conversations
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
    // Fallback: any unknown route → conversations list
    path: '/:pathMatch(.*)*',
    redirect: { name: 'Conversations' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Global navigation guard
router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !isAuthenticated()) {
    // If route requires auth and user is not logged in → go to login
    next({ name: 'Login' })
  } else if (to.name === 'Login' && isAuthenticated()) {
    // If user is already logged in and tries to go to login → redirect home
    next({ name: 'Conversations' })
  } else {
    next()
  }
})

export default router
