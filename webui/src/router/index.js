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
    path: '/',
    name: 'Home',
    // صفحه‌ی لیست گفتگوها (Conversations)
    component: () => import('../views/ConversationListView.vue'),
    meta: { requiresAuth: true }
  },
  {
    // صفحه‌ی یک گفتگوی مشخص
    path: '/conversations/:chatId',
    name: 'Conversation',                // ⬅ همینی که در router.push استفاده می‌کنیم
    component: () => import('../views/ConversationView.vue'),
    props: true,
    meta: { requiresAuth: true }
  },
  {
    // fallback: هر آدرس اشتباه → Home
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Guardia: se non ho token → /login
router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !isAuthenticated()) {
    next({ name: 'Login' })
  } else if (to.name === 'Login' && isAuthenticated()) {
    next({ name: 'Home' })
  } else {
    next()
  }
})

export default router
