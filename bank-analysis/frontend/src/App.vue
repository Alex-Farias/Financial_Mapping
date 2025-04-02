<template>
  <div id="app" :class="{ 'dark-mode': isDarkMode }">
    <!-- Navigation Header -->
    <header class="header" v-if="isLoggedIn">
      <div class="logo-container">
        <img src="@/assets/logo.svg" alt="Bank Analysis" class="logo">
        <h1>Bank Analysis</h1>
      </div>
      
      <div class="search-container" v-if="isLoggedIn">
        <input 
          type="text" 
          placeholder="Search transactions..." 
          v-model="searchQuery"
          @keyup.enter="searchTransactions"
        >
        <button @click="searchTransactions" class="search-button">
          <i class="fas fa-search"></i>
        </button>
      </div>
      
      <div class="user-controls">
        <button @click="toggleDarkMode" class="icon-button">
          <i :class="isDarkMode ? 'fas fa-sun' : 'fas fa-moon'"></i>
        </button>
        <button @click="logout" class="icon-button" v-if="isLoggedIn">
          <i class="fas fa-sign-out-alt"></i>
        </button>
      </div>
    </header>

    <!-- Main Content -->
    <div class="main-container">
      <!-- Sidebar (only show when logged in) -->
      <aside class="sidebar" v-if="isLoggedIn">
        <nav>
          <ul>
            <li>
              <router-link to="/dashboard" class="nav-item">
                <i class="fas fa-home"></i>
                <span>Dashboard</span>
              </router-link>
            </li>
            <li>
              <router-link to="/transactions" class="nav-item">
                <i class="fas fa-list"></i>
                <span>Transactions</span>
              </router-link>
            </li>
            <li>
              <router-link to="/import" class="nav-item">
                <i class="fas fa-file-import"></i>
                <span>Import Data</span>
              </router-link>
            </li>
            <li>
              <router-link to="/export" class="nav-item">
                <i class="fas fa-file-export"></i>
                <span>Export Data</span>
              </router-link>
            </li>
          </ul>
        </nav>
      </aside>

      <!-- Main Content -->
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      isDarkMode: localStorage.getItem('darkMode') === 'true',
      searchQuery: '',
      isLoggedIn: false
    }
  },
  created() {
    // Check if user is logged in
    const token = localStorage.getItem('token')
    this.isLoggedIn = !!token
    
    // Listen for auth state changes
    // FIX: Make sure $bus is defined before using it
    if (this.$bus) {
      this.$bus.$on('login', this.handleLogin)
      this.$bus.$on('logout', this.handleLogout)
    }
    
    // Apply dark mode if set
    if (this.isDarkMode) {
      document.body.classList.add('dark-mode')
    }
  },
  // Add destroyed lifecycle hook to remove event listeners
  destroyed() {
    // Clean up event listeners
    if (this.$bus) {
      this.$bus.$off('login', this.handleLogin)
      this.$bus.$off('logout', this.handleLogout)
    }
  },
  methods: {
    toggleDarkMode() {
      this.isDarkMode = !this.isDarkMode
      localStorage.setItem('darkMode', this.isDarkMode)
      document.body.classList.toggle('dark-mode', this.isDarkMode)
    },
    searchTransactions() {
      if (!this.searchQuery.trim()) return
      
      this.$router.push({
        path: '/transactions',
        query: { q: this.searchQuery }
      })
    },
    logout() {
      localStorage.removeItem('token')
      localStorage.removeItem('userId')
      this.isLoggedIn = false
      this.$bus.$emit('logout')
      this.$router.push('/login')
    },
    handleLogin() {
      this.isLoggedIn = true
    },
    handleLogout() {
      this.isLoggedIn = false
    }
  }
}
</script>

<style>
:root {
  /* YouTube Theme Colors */
  --primary-color: #FF0000; /* YouTube Red */
  --secondary-color: #282828; /* YouTube Dark Gray */
  --background-color: #FFFFFF;
  --sidebar-bg: #F9F9F9;
  --text-color: #030303;
  --border-color: #E5E5E5;
  --hover-color: #F2F2F2;
  --input-bg: #F8F8F8;
  --input-border: #CCCCCC;
  --button-hover: #CC0000;
}

/* Dark Mode */
.dark-mode {
  --background-color: #212121;
  --sidebar-bg: #181818;
  --text-color: #FFFFFF;
  --border-color: #383838;
  --hover-color: #383838;
  --input-bg: #121212;
  --input-border: #383838;
}

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: 'Roboto', Arial, sans-serif;
  background-color: var(--background-color);
  color: var(--text-color);
  transition: background-color 0.3s, color 0.3s;
}

#app {
  width: 100%;
  min-height: 100vh;
}

/* Header Styles */
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.8rem 1.5rem;
  background-color: var(--background-color);
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  z-index: 100;
}

.logo-container {
  display: flex;
  align-items: center;
}

.logo {
  height: 30px;
  margin-right: 0.75rem;
}

.logo-container h1 {
  font-size: 1.2rem;
  font-weight: 500;
}

.search-container {
  flex: 1;
  max-width: 600px;
  display: flex;
  margin: 0 2rem;
}

.search-container input {
  width: 100%;
  padding: 0.6rem 1rem;
  border: 1px solid var(--input-border);
  border-radius: 2px 0 0 2px;
  background-color: var(--input-bg);
  color: var(--text-color);
}

.search-button {
  background-color: #F8F8F8;
  border: 1px solid var(--input-border);
  border-left: none;
  padding: 0 1.25rem;
  border-radius: 0 2px 2px 0;
  cursor: pointer;
}

.dark-mode .search-button {
  background-color: #303030;
  color: white;
}

.user-controls {
  display: flex;
  gap: 1rem;
}

.icon-button {
  background: transparent;
  border: none;
  font-size: 1.1rem;
  color: var(--text-color);
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 50%;
}

.icon-button:hover {
  background-color: var(--hover-color);
}

/* Main Container Styles */
.main-container {
  display: flex;
  min-height: calc(100vh - 56px);
}

/* Sidebar Styles */
.sidebar {
  width: 240px;
  background-color: var(--sidebar-bg);
  padding-top: 1rem;
  overflow-y: auto;
  height: calc(100vh - 56px);
  position: sticky;
  top: 56px;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 0.8rem 1.5rem;
  color: var(--text-color);
  text-decoration: none;
  border-radius: 2px;
}

.nav-item:hover, .router-link-active {
  background-color: var(--hover-color);
}

.nav-item i {
  margin-right: 1.5rem;
  font-size: 1.2rem;
  width: 24px;
  text-align: center;
}

.content {
  flex: 1;
  padding: 2rem;
  overflow-y: auto;
}

/* Responsive Styles */
@media (max-width: 768px) {
  .sidebar {
    width: 70px;
  }
  
  .nav-item span {
    display: none;
  }
  
  .nav-item i {
    margin-right: 0;
  }
  
  .content {
    padding: 1rem;
  }
}

@media (max-width: 576px) {
  .search-container {
    margin: 0 0.5rem;
  }
  
  .logo-container h1 {
    display: none;
  }
  
  .header {
    padding: 0.8rem;
  }
  
  .sidebar {
    display: none;
  }
}
</style>