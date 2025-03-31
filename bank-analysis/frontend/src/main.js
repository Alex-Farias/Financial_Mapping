import Vue from 'vue'
import App from './App.vue'
import router from './router'
import axios from 'axios'

// Setup global event bus
Vue.prototype.$bus = new Vue()

// Configure axios for API requests
axios.defaults.baseURL = process.env.VUE_APP_API_URL || 'http://localhost:8080/api'

// Add axios interceptor for authentication
axios.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Add axios interceptor for handling 401 Unauthorized errors
axios.interceptors.response.use(
  response => response,
  error => {
    if (error.response && error.response.status === 401) {
      // Clear token and redirect to login
      localStorage.removeItem('token')
      router.push('/login')
    }
    return Promise.reject(error)
  }
)

Vue.config.productionTip = false

new Vue({
  router,
  render: h => h(App)
}).$mount('#app')