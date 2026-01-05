// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import AccountView from '../views/AccountView.vue'

// We will NOT rely on "isAuthenticated()" alone anymore.
// Instead, we validate the session by calling a backend endpoint (e.g. /context).
import { getToken, getContext, logout } from '../services/api.js'

const routes = [
  {
    path: '/account',
    name: 'Account',
    component: AccountView,
    // IMPORTANT: Account must also be protected.
    // Without this, anyone can access /account even when logged out.
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
    path: '/:pathMatch
