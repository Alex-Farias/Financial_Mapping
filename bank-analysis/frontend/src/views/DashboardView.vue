<template>
  <div class="dashboard">
    <h1>Financial Dashboard</h1>
    
    <div class="dashboard-stats">
      <div class="stat-card">
        <div class="stat-icon">
          <i class="fas fa-arrow-down"></i>
        </div>
        <div class="stat-content">
          <h3>Total Income</h3>
          <p>{{ formatCurrency(totalIncome) }}</p>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">
          <i class="fas fa-arrow-up"></i>
        </div>
        <div class="stat-content">
          <h3>Total Expenses</h3>
          <p>{{ formatCurrency(totalExpenses) }}</p>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">
          <i class="fas fa-wallet"></i>
        </div>
        <div class="stat-content">
          <h3>Net Cashflow</h3>
          <p :class="netCashflow >= 0 ? 'positive' : 'negative'">
            {{ formatCurrency(netCashflow) }}
          </p>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">
          <i class="fas fa-receipt"></i>
        </div>
        <div class="stat-content">
          <h3>Transactions</h3>
          <p>{{ totalTransactions }}</p>
        </div>
      </div>
    </div>
    
    <div class="chart-container">
      <div class="chart-wrapper">
        <h2>Monthly Overview</h2>
        <div class="chart-card">
          <bar-chart 
            :chart-data="monthlyChartData" 
            :options="chartOptions"
          />
        </div>
      </div>
      
      <div class="chart-wrapper">
        <h2>Spending by Category</h2>
        <div class="chart-card">
          <doughnut-chart 
            :chart-data="categoryChartData" 
            :options="doughnutOptions"
          />
        </div>
      </div>
    </div>
    
    <div class="recent-transactions">
      <h2>Recent Transactions</h2>
      <div class="transaction-list">
        <div v-if="isLoading" class="loading">
          <i class="fas fa-spinner fa-spin"></i>
          Loading...
        </div>
        
        <div v-else-if="!recentTransactions.length" class="no-data">
          No transactions found
        </div>
        
        <div v-else class="transaction-table-container">
          <table class="transaction-table">
            <thead>
              <tr>
                <th>Date</th>
                <th>Description</th>
                <th>Category</th>
                <th>Amount</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(transaction, index) in recentTransactions" :key="index">
                <td>{{ formatDate(transaction.date) }}</td>
                <td>{{ transaction.description }}</td>
                <td>{{ transaction.category }}</td>
                <td :class="transaction.type === 'credit' ? 'positive' : 'negative'">
                  {{ formatCurrency(transaction.amount, transaction.type) }}
                </td>
              </tr>
            </tbody>
          </table>
          
          <div class="view-all">
            <router-link to="/transactions" class="view-all-link">
              View all transactions
              <i class="fas fa-arrow-right"></i>
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import { BarChart, DoughnutChart } from 'vue-chartjs'

export default {
  name: 'DashboardView',
  components: {
    BarChart,
    DoughnutChart
  },
  data() {
    return {
      isLoading: true,
      totalIncome: 0,
      totalExpenses: 0,
      netCashflow: 0,
      totalTransactions: 0,
      recentTransactions: [],
      monthlyData: [],
      categoryData: {},
      
      // Chart data
      monthlyChartData: {
        labels: [],
        datasets: [
          {
            label: 'Income',
            backgroundColor: '#36A2EB',
            data: []
          },
          {
            label: 'Expenses',
            backgroundColor: '#FF6384',
            data: []
          }
        ]
      },
      categoryChartData: {
        labels: [],
        datasets: [
          {
            backgroundColor: [
              '#FF6384', '#36A2EB', '#FFCE56', '#4BC0C0', '#9966FF',
              '#FF9F40', '#C9CBCF', '#7FB3D5', '#F39C12', '#2ECC71'
            ],
            data: []
          }
        ]
      },
      chartOptions: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          y: {
            beginAtZero: true
          }
        }
      },
      doughnutOptions: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: 'right'
          }
        }
      }
    }
  },
  created() {
    this.fetchDashboardData()
  },
  methods: {
    async fetchDashboardData() {
      try {
        this.isLoading = true
        
        // Get token from localStorage
        const token = localStorage.getItem('token')
        if (!token) {
          console.error('No token found in localStorage')
          this.$router.push('/login')
          return
        }
        
        // Fetch monthly analysis data
        const analysisResponse = await axios.get('/analysis/monthly', {
          headers: { 'Authorization': `Bearer ${token}` }
        })
        
        this.monthlyData = analysisResponse.data
        
        // Fetch recent transactions
        const transactionsResponse = await axios.get('/transactions?limit=5', {
          headers: { 'Authorization': `Bearer ${token}` }
        })
        
        this.recentTransactions = transactionsResponse.data.transactions
        this.totalTransactions = transactionsResponse.data.total
        
        // Calculate totals from monthly data
        this.calculateTotals()
        
        // Prepare chart data
        this.prepareChartData()
        this.prepareCategoryData()
        
      } catch (error) {
        console.error('Error fetching dashboard data:', error)
        
        // If unauthorized, redirect to login
        if (error.response && error.response.status === 401) {
          localStorage.removeItem('token')
          this.$router.push('/login')
        }
      } finally {
        this.isLoading = false
      }
    },
    
    calculateTotals() {
      this.totalIncome = this.monthlyData.reduce((sum, month) => sum + month.totalIncome, 0)
      this.totalExpenses = this.monthlyData.reduce((sum, month) => sum + month.totalExpenses, 0)
      this.netCashflow = this.totalIncome - this.totalExpenses
    },
    
    prepareChartData() {
      // Sort monthly data by date
      const sortedData = [...this.monthlyData].sort((a, b) => {
        return new Date(a.year, this.getMonthIndex(a.month)) - new Date(b.year, this.getMonthIndex(b.month))
      })
      
      // Get labels and values
      const labels = sortedData.map(month => `${month.month.substr(0, 3)} ${month.year}`)
      const incomeData = sortedData.map(month => month.totalIncome)
      const expenseData = sortedData.map(month => month.totalExpenses)
      
      // Update chart data
      this.monthlyChartData = {
        labels,
        datasets: [
          {
            label: 'Income',
            backgroundColor: '#36A2EB',
            data: incomeData
          },
          {
            label: 'Expenses',
            backgroundColor: '#FF6384',
            data: expenseData
          }
        ]
      }
    },
    
    prepareCategoryData() {
      // Aggregate spending by category across all months
      const categories = {}
      
      this.monthlyData.forEach(month => {
        Object.entries(month.categoryBreakdown).forEach(([category, amount]) => {
          // Only include expenses (negative amounts)
          if (amount < 0) {
            const absAmount = Math.abs(amount)
            categories[category] = (categories[category] || 0) + absAmount
          }
        })
      })
      
      // Sort categories by amount (descending)
      const sortedCategories = Object.entries(categories)
        .sort((a, b) => b[1] - a[1])
        .slice(0, 10) // Top 10 categories
      
      // Update chart data
      this.categoryChartData = {
        labels: sortedCategories.map(([category]) => category),
        datasets: [
          {
            backgroundColor: [
              '#FF6384', '#36A2EB', '#FFCE56', '#4BC0C0', '#9966FF',
              '#FF9F40', '#C9CBCF', '#7FB3D5', '#F39C12', '#2ECC71'
            ],
            data: sortedCategories.map(([, amount]) => amount)
          }
        ]
      }
    },
    
    getMonthIndex(monthName) {
      const months = {
        'January': 0, 'February': 1, 'March': 2, 'April': 3, 'May': 4, 'June': 5,
        'July': 6, 'August': 7, 'September': 8, 'October': 9, 'November': 10, 'December': 11
      }
      return months[monthName] || 0
    },
    
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
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
}

h1 {
  margin-bottom: 1.5rem;
  color: var(--text-color);
}

h2 {
  margin-bottom: 1rem;
  color: var(--text-color);
}

/* Dashboard Stats */
.dashboard-stats {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 1.25rem;
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 48px;
  height: 48px;
  background-color: var(--hover-color);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  margin-right: 1rem;
}

.stat-content h3 {
  font-size: 0.9rem;
  font-weight: 500;
  margin-bottom: 0.5rem;
  color: var(--text-color);
}

.stat-content p {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-color);
}

/* Chart Containers */
.chart-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(450px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.chart-wrapper {
  width: 100%;
}

.chart-card {
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 1rem;
  height: 300px;
}

/* Transaction List */
.recent-transactions {
  margin-bottom: 2rem;
}

.transaction-table-container {
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.transaction-table {
  width: 100%;
  border-collapse: collapse;
}

.transaction-table th,
.transaction-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
}

.transaction-table th {
  background-color: var(--sidebar-bg);
  font-weight: 500;
}

.transaction-table tr:last-child td {
  border-bottom: none;
}

.view-all {
  padding: 1rem;
  text-align: center;
  border-top: 1px solid var(--border-color);
}

.view-all-link {
  color: var(--primary-color);
  text-decoration: none;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
}

.view-all-link i {
  margin-left: 0.5rem;
}

/* Utility classes */
.positive {
  color: #2ECC71;
}

.negative {
  color: var(--primary-color);
}

.loading, .no-data {
  padding: 2rem;
  text-align: center;
  color: var(--text-color);
}

.loading i {
  margin-right: 0.5rem;
}

/* Responsive */
@media (max-width: 768px) {
  .chart-container {
    grid-template-columns: 1fr;
  }
}
</style>