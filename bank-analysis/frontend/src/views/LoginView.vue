<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <img src="@/assets/logo.svg" alt="Bank Analysis Logo" class="login-logo">
        <h1>Bank Analysis</h1>
      </div>
      
      <div class="form-group">
        <label for="email">Email</label>
        <input 
          type="email" 
          id="email" 
          v-model="email" 
          placeholder="Enter your email"
          @keyup.enter="login"
        >
      </div>
      
      <div class="form-group">
        <label for="password">Password</label>
        <input 
          type="password" 
          id="password" 
          v-model="password" 
          placeholder="Enter your password"
          @keyup.enter="login"
        >
      </div>
      
      <div class="login-buttons">
        <button @click="login" class="btn-primary" :disabled="isLoading">
          {{ isLoading ? 'Logging in...' : 'Log In' }}
        </button>
        <button @click="register" class="btn-secondary" :disabled="isLoading">
          Register
        </button>
      </div>
      
      <div v-if="errorMessage" class="error-message">
        {{ errorMessage }}
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'LoginView',
  data() {
    return {
      email: '',
      password: '',
      isLoading: false,
      errorMessage: ''
    }
  },
  methods: {
    async login() {
      if (!this.email || !this.password) {
        this.errorMessage = 'Please enter both email and password'
        return
      }
      
      this.isLoading = true
      this.errorMessage = ''
      
      try {
        const response = await axios.post('/api/auth/login', {
          email: this.email,
          password: this.password
        })
        
        // Store token and user info
        localStorage.setItem('token', response.data.token)
        
        // Emit login event
        this.$bus.$emit('login')
        
        // Redirect to dashboard
        this.$router.push('/dashboard')
      } catch (error) {
        this.errorMessage = error.response?.data?.message || 'Login failed'
      } finally {
        this.isLoading = false
      }
    },
    
    async register() {
      if (!this.email || !this.password) {
        this.errorMessage = 'Please enter both email and password'
        return
      }
      
      this.isLoading = true
      this.errorMessage = ''
      
      try {
        const response = await axios.post('/api/auth/register', {
          email: this.email,
          password: this.password
        })
        
        // Store token and user info
        localStorage.setItem('token', response.data.token)
        
        // Emit login event
        this.$bus.$emit('login')
        
        // Redirect to dashboard
        this.$router.push('/dashboard')
      } catch (error) {
        this.errorMessage = error.response?.data?.message || 'Registration failed'
      } finally {
        this.isLoading = false
      }
    }
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: var(--background-color);
}

.login-card {
  background-color: var(--background-color);
  border-radius: 8px;
  padding: 2rem;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.login-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 2rem;
}

.login-logo {
  width: 60px;
  height: 60px;
  margin-bottom: 1rem;
}

.login-header h1 {
  font-size: 1.5rem;
  font-weight: 500;
  color: var(--text-color);
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: var(--text-color);
}

.form-group input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  background-color: var(--input-bg);
  color: var(--text-color);
  font-size: 1rem;
}

.login-buttons {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
}

.btn-primary, .btn-secondary {
  flex: 1;
  padding: 0.75rem;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: var(--button-hover);
}

.btn-secondary {
  background-color: var(--input-bg);
  color: var(--text-color);
  border: 1px solid var(--input-border);
}

.btn-secondary:hover:not(:disabled) {
  background-color: var(--hover-color);
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-message {
  margin-top: 1rem;
  color: var(--primary-color);
  text-align: center;
}
</style>