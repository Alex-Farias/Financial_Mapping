<template>
  <div class="export-container">
    <h1>Export Data</h1>
    
    <div class="export-card">
      <div class="export-header">
        <i class="fas fa-file-export"></i>
        <h2>Export Transactions</h2>
      </div>
      
      <p>Export your transaction data to a CSV file for use in spreadsheets or other applications.</p>
      
      <div class="export-options">
        <div class="option-group">
          <h3>Date Range</h3>
          
          <div class="date-range">
            <div class="date-input">
              <label for="start-date">Start Date</label>
              <input 
                type="date" 
                id="start-date" 
                v-model="startDate"
              >
            </div>
            
            <div class="date-input">
              <label for="end-date">End Date</label>
              <input 
                type="date" 
                id="end-date" 
                v-model="endDate"
              >
            </div>
          </div>
        </div>
        
        <div class="option-group">
          <h3>Export Options</h3>
          
          <div class="option-checkboxes">
            <div class="option-checkbox">
              <input type="checkbox" id="include-checking" v-model="includeChecking">
              <label for="include-checking">Include Checking Account</label>
            </div>
            
            <div class="option-checkbox">
              <input type="checkbox" id="include-credit" v-model="includeCredit">
              <label for="include-credit">Include Credit Card</label>
            </div>
          </div>
        </div>
        
        <div class="export-actions">
          <button @click="exportCSV" class="btn-export" :disabled="isExporting">
            <i class="fas fa-download"></i>
            {{ isExporting ? 'Exporting...' : 'Export to CSV' }}
          </button>
        </div>
      </div>
    </div>
    
    <div class="export-history">
      <h2>Export History</h2>
      
      <div v-if="exportHistory.length === 0" class="no-history">
        <i class="fas fa-history"></i>
        <p>No export history available</p>
      </div>
      
      <div v-else class="history-list">
        <div v-for="(item, index) in exportHistory" :key="index" class="history-item">
          <div class="history-icon">
            <i class="fas fa-file-csv"></i>
          </div>
          <div class="history-content">
            <h3>{{ item.filename }}</h3>
            <p class="history-date">{{ formatDate(item.date) }}</p>
            <p class="history-details">
              {{ item.count }} transactions
              <span v-if="item.startDate && item.endDate">
                ({{ formatDateShort(item.startDate) }} to {{ formatDateShort(item.endDate) }})
              </span>
            </p>
          </div>
          <div class="history-actions">
            <button @click="redownload(item)" class="btn-redownload">
              <i class="fas fa-download"></i>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'ExportView',
  data() {
    return {
      startDate: '',
      endDate: '',
      includeChecking: true,
      includeCredit: true,
      isExporting: false,
      
      // In a real app, this would come from the backend
      exportHistory: [
        {
          filename: 'bank_transactions_2023-09-01.csv',
          date: new Date('2023-09-01T15:30:00'),
          count: 245,
          startDate: '2023-01-01',
          endDate: '2023-08-31'
        },
        {
          filename: 'bank_transactions_2023-08-15.csv',
          date: new Date('2023-08-15T10:15:00'),
          count: 210,
          startDate: '2023-01-01',
          endDate: '2023-07-31'
        }
      ]
    }
  },
  methods: {
    async exportCSV() {
      this.isExporting = true
      
      try {
        // Build query parameters
        let params = {}
        
        if (this.startDate) {
          params.start = this.startDate
        }
        
        if (this.endDate) {
          params.end = this.endDate
        }
        
        // In a real app, we would also add filters for account types
        if (!this.includeChecking || !this.includeCredit) {
          const sources = []
          if (this.includeChecking) sources.push('checking')
          if (this.includeCredit) sources.push('credit_card')
          
          if (sources.length > 0) {
            params.sources = sources.join(',')
          }
        }
        
        // Get the export URL with authentication token
        const token = localStorage.getItem('token')
        const url = '/api/export/csv' + this.buildQueryString(params)
        
        // Create a temporary link element and trigger a download
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', '')
        
        // Add authentication header (this works for simple cases, but in production
        // you'd want to use a more secure approach)
        link.setAttribute('data-auth', `Bearer ${token}`)
        
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        
        // Add to export history (in a real app, this would come from the backend)
        const currentDate = new Date()
        const filename = `bank_transactions_${this.formatDateForFilename(currentDate)}.csv`
        
        this.exportHistory.unshift({
          filename,
          date: currentDate,
          count: 0, // In a real app, this would come from the backend
          startDate: this.startDate || null,
          endDate: this.endDate || null
        })
        
      } catch (error) {
        console.error('Error exporting CSV:', error)
        // Show error message
      } finally {
        this.isExporting = false
      }
    },
    
    redownload(item) {
      // In a real app, this would re-trigger the download with the same parameters
      const url = '/api/export/csv'
      const token = localStorage.getItem('token')
      
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', '')
      link.setAttribute('data-auth', `Bearer ${token}`)
      
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    },
    
    buildQueryString(params) {
      if (Object.keys(params).length === 0) return ''
      
      const queryParts = []
      
      for (const [key, value] of Object.entries(params)) {
        if (value !== undefined && value !== null && value !== '') {
          queryParts.push(`${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
        }
      }
      
      return queryParts.length > 0 ? '?' + queryParts.join('&') : ''
    },
    
    formatDate(date) {
      return new Intl.DateTimeFormat('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: 'numeric',
        minute: 'numeric'
      }).format(date)
    },
    
    formatDateShort(dateString) {
      const date = new Date(dateString)
      return new Intl.DateTimeFormat('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      }).format(date)
    },
    
    formatDateForFilename(date) {
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      return `${year}-${month}-${day}`
    }
  }
}
</script>

<style scoped>
.export-container {
  max-width: 800px;
  margin: 0 auto;
}

h1, h2 {
  margin-bottom: 1.5rem;
  color: var(--text-color);
}

/* Export Card */
.export-card {
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 2rem;
}

.export-header {
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
}

.export-header i {
  font-size: 1.5rem;
  margin-right: 0.75rem;
  color: var(--primary-color);
}

.export-header h2 {
  margin-bottom: 0;
  font-size: 1.25rem;
}

.export-options {
  margin-top: 1.5rem;
}

.option-group {
  margin-bottom: 1.5rem;
}

.option-group h3 {
  font-size: 1rem;
  margin-bottom: 0.75rem;
  font-weight: 500;
}

.date-range {
  display: flex;
  gap: 1rem;
}

.date-input {
  flex: 1;
}

.date-input label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
}

.date-input input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  background-color: var(--input-bg);
  color: var(--text-color);
}

.option-checkboxes {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
}

.option-checkbox {
  display: flex;
  align-items: center;
}

.option-checkbox input {
  margin-right: 0.5rem;
}

.export-actions {
  margin-top: 1.5rem;
  display: flex;
  justify-content: flex-end;
}

.btn-export {
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 4px;
  padding: 0.75rem 1.5rem;
  font-size: 1rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: background-color 0.2s;
}

.btn-export i {
  margin-right: 0.5rem;
}

.btn-export:hover:not(:disabled) {
  background-color: var(--button-hover);
}

.btn-export:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Export History */
.export-history {
  margin-top: 2rem;
}

.no-history {
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 3rem 1.5rem;
  text-align: center;
}

.no-history i {
  font-size: 2rem;
  color: var(--text-color);
  opacity: 0.6;
  margin-bottom: 1rem;
}

.history-list {
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.history-item {
  display: flex;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border-color);
}

.history-item:last-child {
  border-bottom: none;
}

.history-icon {
  width: 40px;
  height: 40px;
  background-color: var(--hover-color);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 1rem;
}

.history-icon i {
  color: var(--primary-color);
}

.history-content {
  flex: 1;
}

.history-content h3 {
  font-size: 1rem;
  margin-bottom: 0.25rem;
}

.history-date {
  font-size: 0.85rem;
  color: var(--text-color);
  opacity: 0.7;
  margin-bottom: 0.25rem;
}

.history-details {
  font-size: 0.9rem;
}

.history-actions {
  margin-left: 1rem;
}

.btn-redownload {
  width: 40px;
  height: 40px;
  background-color: var(--hover-color);
  border: none;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--text-color);
  transition: background-color 0.2s;
}

.btn-redownload:hover {
  background-color: var(--input-border);
}

/* Responsive */
@media (max-width: 600px) {
  .date-range {
    flex-direction: column;
  }
  
  .history-item {
    flex-wrap: wrap;
  }
  
  .history-content {
    width: 100%;
    margin: 0.5rem 0;
  }
  
  .history-actions {
    margin-left: auto;
  }
}
</style>