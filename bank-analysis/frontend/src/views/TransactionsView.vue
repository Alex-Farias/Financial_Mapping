<template>
  <div class="transactions-container">
    <h1>Transactions</h1>
    
    <div class="filter-container">
      <div class="search-box">
        <input 
          type="text" 
          v-model="searchQuery" 
          placeholder="Search transactions..." 
          @keyup.enter="searchTransactions"
        >
        <button @click="searchTransactions" class="search-button">
          <i class="fas fa-search"></i>
        </button>
      </div>
      
      <div class="filter-box">
        <div class="filter-group">
          <label>Date Range:</label>
          <div class="date-inputs">
            <input 
              type="date" 
              v-model="startDate" 
              placeholder="Start Date"
            >
            <span>to</span>
            <input 
              type="date" 
              v-model="endDate" 
              placeholder="End Date"
            >
          </div>
        </div>
        
        <div class="filter-actions">
          <button @click="applyFilters" class="btn-filter">
            <i class="fas fa-filter"></i>
            Apply Filters
          </button>
          <button @click="resetFilters" class="btn-reset">
            <i class="fas fa-times"></i>
            Reset
          </button>
        </div>
      </div>
    </div>
    
    <div class="transactions-content">
      <div v-if="isLoading" class="loading">
        <i class="fas fa-spinner fa-spin"></i>
        Loading transactions...
      </div>
      
      <div v-else-if="!transactions.length" class="no-data">
        <i class="fas fa-search"></i>
        <p>No transactions found</p>
        <p class="sub-text">Try changing your search filters or import new data</p>
      </div>
      
      <div v-else>
        <!-- Selected transactions actions -->
        <div v-if="selectedTransactions.length > 0" class="selected-actions">
          <span>{{ selectedTransactions.length }} items selected</span>
          
          <div class="action-buttons">
            <div class="category-select">
              <select v-model="selectedCategory">
                <option value="" disabled>Set Category</option>
                <option v-for="category in categories" :key="category" :value="category">
                  {{ category }}
                </option>
              </select>
              <button 
                @click="updateCategories" 
                class="btn-update" 
                :disabled="!selectedCategory"
              >
                Apply
              </button>
            </div>
            
            <button @click="clearSelection" class="btn-cancel">
              Cancel
            </button>
          </div>
        </div>
        
        <!-- Transactions table -->
        <div class="transactions-table-container">
          <table class="transactions-table">
            <thead>
              <tr>
                <th class="checkbox-col">
                  <input 
                    type="checkbox" 
                    :checked="isAllSelected" 
                    @change="toggleSelectAll"
                  >
                </th>
                <th>Date</th>
                <th>Description</th>
                <th>Category</th>
                <th>Amount</th>
                <th>Type</th>
              </tr>
            </thead>
            <tbody>
              <tr 
                v-for="transaction in transactions" 
                :key="transaction.id"
                :class="{ 'selected': isSelected(transaction.id) }"
              >
                <td>
                  <input 
                    type="checkbox" 
                    :checked="isSelected(transaction.id)" 
                    @change="toggleSelect(transaction.id)"
                  >
                </td>
                <td>{{ formatDate(transaction.date) }}</td>
                <td class="description-col">{{ transaction.description }}</td>
                <td>
                  <div class="category-cell">
                    <span>{{ transaction.category }}</span>
                    <button @click="editCategory(transaction)" class="btn-icon">
                      <i class="fas fa-pencil-alt"></i>
                    </button>
                  </div>
                </td>
                <td :class="transaction.type === 'credit' ? 'positive' : 'negative'">
                  {{ formatCurrency(transaction.amount, transaction.type) }}
                </td>
                <td>{{ transaction.type === 'credit' ? 'Income' : 'Expense' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        
        <!-- Pagination -->
        <div class="pagination">
          <button 
            @click="prevPage" 
            class="btn-page" 
            :disabled="currentPage === 1"
          >
            <i class="fas fa-chevron-left"></i>
            Previous
          </button>
          
          <span class="page-info">
            Page {{ currentPage }} of {{ totalPages }}
          </span>
          
          <button 
            @click="nextPage" 
            class="btn-page" 
            :disabled="currentPage === totalPages || totalPages === 0"
          >
            Next
            <i class="fas fa-chevron-right"></i>
          </button>
        </div>
      </div>
    </div>
    
    <!-- Edit Category Modal -->
    <div v-if="showCategoryModal" class="modal-overlay" @click="closeCategoryModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>Edit Category</h3>
          <button @click="closeCategoryModal" class="btn-icon">
            <i class="fas fa-times"></i>
          </button>
        </div>
        
        <div class="modal-body">
          <p class="transaction-desc">{{ editingTransaction.description }}</p>
          
          <div class="category-input">
            <label for="category">Category</label>
            <select id="category" v-model="editingTransaction.category">
              <option v-for="category in categories" :key="category" :value="category">
                {{ category }}
              </option>
            </select>
          </div>
        </div>
        
        <div class="modal-footer">
          <button @click="closeCategoryModal" class="btn-cancel">Cancel</button>
          <button @click="saveCategoryChanges" class="btn-save">Save Changes</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'TransactionsView',
  data() {
    return {
      isLoading: true,
      transactions: [],
      totalTransactions: 0,
      currentPage: 1,
      pageSize: 20,
      searchQuery: '',
      startDate: '',
      endDate: '',
      
      // Selection
      selectedTransactions: [],
      selectedCategory: '',
      
      // Category editing
      showCategoryModal: false,
      editingTransaction: null,
      originalTransaction: null,
      
      // Categories (would come from the backend in a real app)
      categories: [
        'Uncategorized',
        'Housing',
        'Transportation',
        'Food',
        'Utilities',
        'Insurance',
        'Healthcare',
        'Debt',
        'Subscriptions',
        'Entertainment',
        'Shopping',
        'Personal Care',
        'Education',
        'Gifts & Donations',
        'Travel',
        'Income',
        'Other'
      ]
    }
  },
  computed: {
    totalPages() {
      return Math.ceil(this.totalTransactions / this.pageSize)
    },
    isAllSelected() {
      return this.transactions.length > 0 && this.selectedTransactions.length === this.transactions.length
    }
  },
  created() {
    // Get search query from route if present
    if (this.$route.query.q) {
      this.searchQuery = this.$route.query.q
      this.searchTransactions()
    } else {
      this.fetchTransactions()
    }
  },
  methods: {
    async fetchTransactions() {
      this.isLoading = true
      
      try {
        const offset = (this.currentPage - 1) * this.pageSize
        
        // Build query params
        let params = {
          limit: this.pageSize,
          offset: offset
        }
        
        if (this.startDate) {
          params.startDate = this.startDate
        }
        
        if (this.endDate) {
          params.endDate = this.endDate
        }
        
        const response = await axios.get('/api/transactions', {
          params,
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        
        this.transactions = response.data.transactions
        this.totalTransactions = response.data.total
        
        // Clear selection when fetching new transactions
        this.selectedTransactions = []
        
      } catch (error) {
        console.error('Error fetching transactions:', error)
      } finally {
        this.isLoading = false
      }
    },
    
    async searchTransactions() {
      if (!this.searchQuery.trim()) {
        this.fetchTransactions()
        return
      }
      
      this.isLoading = true
      
      try {
        const response = await axios.get('/api/transactions/search', {
          params: { q: this.searchQuery },
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        
        this.transactions = response.data.transactions
        this.totalTransactions = response.data.total
        this.currentPage = 1
        
        // Clear selection
        this.selectedTransactions = []
        
      } catch (error) {
        console.error('Error searching transactions:', error)
      } finally {
        this.isLoading = false
      }
    },
    
    applyFilters() {
      this.currentPage = 1
      this.fetchTransactions()
    },
    
    resetFilters() {
      this.searchQuery = ''
      this.startDate = ''
      this.endDate = ''
      this.currentPage = 1
      this.fetchTransactions()
    },
    
    prevPage() {
      if (this.currentPage > 1) {
        this.currentPage--
        this.fetchTransactions()
      }
    },
    
    nextPage() {
      if (this.currentPage < this.totalPages) {
        this.currentPage++
        this.fetchTransactions()
      }
    },
    
    // Selection methods
    toggleSelectAll() {
      if (this.isAllSelected) {
        this.selectedTransactions = []
      } else {
        this.selectedTransactions = this.transactions.map(t => t.id)
      }
    },
    
    toggleSelect(id) {
      const index = this.selectedTransactions.indexOf(id)
      if (index === -1) {
        this.selectedTransactions.push(id)
      } else {
        this.selectedTransactions.splice(index, 1)
      }
    },
    
    isSelected(id) {
      return this.selectedTransactions.includes(id)
    },
    
    clearSelection() {
      this.selectedTransactions = []
      this.selectedCategory = ''
    },
    
    // Category editing
    editCategory(transaction) {
      this.editingTransaction = { ...transaction }
      this.originalTransaction = transaction
      this.showCategoryModal = true
    },
    
    closeCategoryModal() {
      this.showCategoryModal = false
      this.editingTransaction = null
      this.originalTransaction = null
    },
    
    async saveCategoryChanges() {
      if (!this.editingTransaction || this.editingTransaction.category === this.originalTransaction.category) {
        this.closeCategoryModal()
        return
      }
      
      try {
        await axios.put('/api/categories', {
          transactionIds: [this.editingTransaction.id],
          category: this.editingTransaction.category
        }, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        
        // Update the transaction in the local list
        const index = this.transactions.findIndex(t => t.id === this.editingTransaction.id)
        if (index !== -1) {
          this.transactions[index].category = this.editingTransaction.category
        }
        
        this.closeCategoryModal()
        
      } catch (error) {
        console.error('Error updating category:', error)
        // Show error message
      }
    },
    
    async updateCategories() {
      if (!this.selectedCategory || this.selectedTransactions.length === 0) return
      
      try {
        await axios.put('/api/categories', {
          transactionIds: this.selectedTransactions,
          category: this.selectedCategory
        }, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        
        // Update the transactions in the local list
        this.transactions.forEach(transaction => {
          if (this.selectedTransactions.includes(transaction.id)) {
            transaction.category = this.selectedCategory
          }
        })
        
        // Clear selection
        this.clearSelection()
        
      } catch (error) {
        console.error('Error updating categories:', error)
        // Show error message
      }
    },
    
    // Formatting helpers
    formatCurrency(amount, type) {
      // If type is provided, adjust sign based on type
      const value = type === 'debit' ? -Math.abs(amount) : amount
      
      return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD'
      }).format(value)
    },
    
    formatDate(dateString) {
      const date = new Date(dateString)
      return new Intl.DateTimeFormat('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      }).format(date)
    }
  }
}
</script>

<style scoped>
.transactions-container {
  max-width: 1200px;
  margin: 0 auto;
}

h1 {
  margin-bottom: 1.5rem;
  color: var(--text-color);
}

/* Filter container */
.filter-container {
  margin-bottom: 1.5rem;
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.search-box {
  display: flex;
  border-bottom: 1px solid var(--border-color);
}

.search-box input {
  flex: 1;
  padding: 0.75rem 1rem;
  border: none;
  background-color: var(--background-color);
  color: var(--text-color);
  font-size: 1rem;
}

.search-button {
  padding: 0 1.25rem;
  background-color: var(--background-color);
  border: none;
  cursor: pointer;
  color: var(--text-color);
}

.search-button:hover {
  background-color: var(--hover-color);
}

.filter-box {
  padding: 1rem;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.filter-group {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
}

.filter-group label {
  font-weight: 500;
}

.date-inputs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.date-inputs input {
  padding: 0.5rem;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  background-color: var(--input-bg);
  color: var(--text-color);
}

.filter-actions {
  display: flex;
  gap: 0.75rem;
}

.btn-filter, .btn-reset {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  font-size: 0.9rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-filter {
  background-color: var(--primary-color);
  color: white;
  border: none;
}

.btn-filter:hover {
  background-color: var(--button-hover);
}

.btn-reset {
  background-color: transparent;
  border: 1px solid var(--input-border);
  color: var(--text-color);
}

.btn-reset:hover {
  background-color: var(--hover-color);
}

/* Transactions content */
.transactions-content {
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.selected-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background-color: var(--sidebar-bg);
  border-bottom: 1px solid var(--border-color);
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.category-select {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.category-select select {
  padding: 0.5rem;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  background-color: var(--input-bg);
  color: var(--text-color);
}

.btn-update, .btn-cancel {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  font-size: 0.9rem;
  cursor: pointer;
}

.btn-update {
  background-color: var(--primary-color);
  color: white;
  border: none;
}

.btn-update:hover:not(:disabled) {
  background-color: var(--button-hover);
}

.btn-update:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-cancel {
  background-color: transparent;
  border: 1px solid var(--input-border);
  color: var(--text-color);
}

.btn-cancel:hover {
  background-color: var(--hover-color);
}

/* Transactions table */
.transactions-table-container {
  overflow-x: auto;
}

.transactions-table {
  width: 100%;
  border-collapse: collapse;
}

.transactions-table th,
.transactions-table td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
}

.transactions-table th {
  background-color: var(--sidebar-bg);
  font-weight: 500;
}

.checkbox-col {
  width: 40px;
}

.description-col {
  max-width: 300px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.category-cell {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.btn-icon {
  background: transparent;
  border: none;
  color: var(--text-color);
  opacity: 0.5;
  cursor: pointer;
  padding: 0.25rem;
}

.btn-icon:hover {
  opacity: 1;
}

.transactions-table tr.selected {
  background-color: rgba(255, 0, 0, 0.05);
}

/* Pagination */
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-top: 1px solid var(--border-color);
}

.btn-page {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background-color: var(--input-bg);
  border: 1px solid var(--input-border);
  border-radius: 4px;
  color: var(--text-color);
  cursor: pointer;
}

.btn-page:hover:not(:disabled) {
  background-color: var(--hover-color);
}

.btn-page:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  color: var(--text-color);
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background-color: var(--background-color);
  border-radius: 8px;
  max-width: 400px;
  width: 90%;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
}

.modal-body {
  padding: 1.5rem 1rem;
}

.transaction-desc {
  margin-bottom: 1.25rem;
  font-size: 0.95rem;
}

.category-input {
  margin-bottom: 1rem;
}

.category-input label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.category-input select {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  background-color: var(--input-bg);
  color: var(--text-color);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem;
  border-top: 1px solid var(--border-color);
}

.btn-save {
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 4px;
  padding: 0.5rem 1rem;
  cursor: pointer;
}

.btn-save:hover {
  background-color: var(--button-hover);
}

/* Utility classes */
.positive {
  color: #2ECC71;
}

.negative {
  color: var(--primary-color);
}

.loading, .no-data {
  padding: 3rem;
  text-align: center;
  color: var(--text-color);
}

.loading i, .no-data i {
  font-size: 2rem;
  margin-bottom: 1rem;
  opacity: 0.7;
}

.no-data .sub-text {
  margin-top: 0.5rem;
  opacity: 0.7;
}

/* Responsive */
@media (max-width: 768px) {
  .filter-box {
    flex-direction: column;
    align-items: stretch;
  }
  
  .filter-group, .date-inputs {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .filter-actions {
    flex-wrap: wrap;
  }
  
  .btn-filter, .btn-reset {
    flex: 1;
  }
  
  .selected-actions {
    flex-direction: column;
    gap: 0.75rem;
  }
  
  .action-buttons {
    width: 100%;
    flex-wrap: wrap;
  }
  
  .category-select {
    flex: 1;
  }
}
</style>