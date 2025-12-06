import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import { isAuthenticated } from '../services/api.js'

/*
 * Application routes
 */
const routes = [
  {
    path: '/login',
    name: 'Login',
    component: LoginView
  },

  {
    path: '/',
    name: 'Home',
    component: () => import('../views/ConversationListView.vue'),
    meta: { requiresAuth: true }
  },

  /*
   * Single conversation page
   * /conversations/:chatId
   *
   * props: true → enables passing chatId to the component as a prop
   */
  {
    path: '/conversations/:chatId',
    name: 'Conversation',
    component: () => import('../views/ConversationView.vue'),
    props: true,
    meta: { requiresAuth: true }
  },

  // Fallback route → redirect unknown paths to Home
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

/*
 * Create Vue Router instance
 */
const router = createRouter({
  history: createWebHistory(),
  routes
})

/*
 * Global navigation guard:
 *
 * - If a route requires authentication and user has no token → redirect to /login
 * - If user is already logged in and tries to access /login → redirect to Home
 */
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
