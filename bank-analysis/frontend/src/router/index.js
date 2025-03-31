import Vue from 'vue'
import VueRouter from 'vue-router'
import LoginView from '@/views/LoginView.vue'
import DashboardView from '@/views/DashboardView.vue'
import TransactionsView from '@/views/TransactionsView.vue'
import ImportView from '@/views/ImportView.vue'
import ExportView from '@/views/ExportView.vue'

Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    meta: {
      requiresAuth: false
    }
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: DashboardView,
    meta: {
      requiresAuth: true
    }
  },
  {
    path: '/transactions',
    name: 'transactions',
    component: TransactionsView,
    meta: {
      requiresAuth: true
    }
  },
  {
    path: '/import',
    name: 'import',
    component: ImportView,
    meta: {
      requiresAuth: true
    }
  },
  {
    path: '/export',
    name: 'export',
    component: ExportView,
    meta: {
      requiresAuth: true
    }
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

// Navigation guard to check authentication
router.beforeEach((to, from, next) => {
  const isLoggedIn = !!localStorage.getItem('token')
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
  
  if (requiresAuth && !isLoggedIn) {
    // Redirect to login if not authenticated
    next('/login')
  } else if (to.path === '/login' && isLoggedIn) {
    // Redirect to dashboard if already logged in
    next('/dashboard')
  } else {
    // Continue navigation
    next()
  }
})

export default router