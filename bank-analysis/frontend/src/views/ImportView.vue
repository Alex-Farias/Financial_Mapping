<template>
  <div class="import-container">
    <h1>Import Transactions</h1>
    
    <div class="import-options">
      <div class="import-card">
        <div class="import-header">
          <i class="fas fa-file-upload"></i>
          <h2>Upload CSV File</h2>
        </div>
        
        <p>Drag and drop or select a CSV file to import your transactions.</p>
        
        <div 
          class="file-dropzone" 
          :class="{ 'active': isDragging }"
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="handleFileDrop"
        >
          <input 
            type="file" 
            id="file-input" 
            ref="fileInput" 
            accept=".csv" 
            @change="handleFileSelect" 
            hidden
          />
          
          <template v-if="selectedFile">
            <div class="selected-file">
              <i class="fas fa-file-csv"></i>
              <span>{{ selectedFile.name }}</span>
              <button @click.stop="resetFileSelection" class="btn-icon">
                <i class="fas fa-times"></i>
              </button>
            </div>
          </template>
          
          <template v-else>
            <i class="fas fa-cloud-upload-alt"></i>
            <p>Drag & drop your CSV file here or</p>
            <button @click="$refs.fileInput.click()" class="btn-select">
              Select File
            </button>
          </template>
        </div>
        
        <div class="import-actions">
          <button 
            @click="uploadFile" 
            class="btn-primary" 
            :disabled="!selectedFile || isUploading"
          >
            <i class="fas fa-upload"></i>
            {{ isUploading ? 'Uploading...' : 'Upload File' }}
          </button>
        </div>
      </div>
      
      <div class="import-card">
        <div class="import-header">
          <i class="fas fa-folder-open"></i>
          <h2>Scan Download Folder</h2>
        </div>
        
        <p>Automatically scan your downloads folder for CSV files from Nubank.</p>
        
        <div class="folder-input">
          <input 
            type="text" 
            v-model="folderPath" 
            placeholder="Enter your downloads folder path" 
          />
          <button @click="browseFolder" class="btn-browse">Browse</button>
        </div>
        
        <div class="import-actions">
          <button 
            @click="scanFolder" 
            class="btn-primary" 
            :disabled="!folderPath || isScanning"
          >
            <i class="fas fa-search"></i>
            {{ isScanning ? 'Scanning...' : 'Scan Folder' }}
          </button>
        </div>
      </div>
    </div>
    
    <div v-if="importResults.length > 0" class="import-results">
      <h2>Import Results</h2>
      
      <div class="results-list">
        <div v-for="(result, index) in importResults" :key="index" class="result-item">
          <div class="result-icon" :class="result.success ? 'success' : 'error'">
            <i :class="result.success ? 'fas fa-check' : 'fas fa-times'"></i>
          </div>
          <div class="result-content">
            <h3>{{ result.filename }}</h3>
            <p>{{ result.message }}</p>
            <p v-if="result.count" class="count">
              Successfully imported {{ result.count }} transactions
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'ImportView',
  data() {
    return {
      isDragging: false,
      selectedFile: null,
      folderPath: '',
      isUploading: false,
      isScanning: false,
      importResults: []
    }
  },
  methods: {
    handleFileSelect(event) {
      const files = event.target.files
      if (files.length > 0) {
        this.selectedFile = files[0]
        console.log("File selected:", this.selectedFile.name, this.selectedFile.size)
      }
    },
    
    handleFileDrop(event) {
      this.isDragging = false
      const files = event.dataTransfer.files
      
      if (files.length > 0) {
        const file = files[0]
        
        // Check if file is a CSV
        if (file.name.toLowerCase().endsWith('.csv')) {
          this.selectedFile = file
        } else {
          this.addResult({
            filename: file.name,
            success: false,
            message: 'Only CSV files are supported'
          })
        }
      }
    },
    
    resetFileSelection() {
      this.selectedFile = null
      if (this.$refs.fileInput) {
        this.$refs.fileInput.value = ''
      }
    },
    
    browseFolder() {
      // In a real application, this would open a folder browser dialog
      // For this demo, we'll just set a default path
      this.folderPath = '/Users/username/Downloads'
    },
    
    async uploadFile() {
      if (!this.selectedFile) return
      
      this.isUploading = true
      
      try {
        const formData = new FormData();
        formData.append('file', this.selectedFile);  // Make sure it's 'file' not 'csv_file' or something else
        
        const response = await axios.post('/import/upload', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        })
        
        // Add success result
        this.addResult({
          filename: this.selectedFile.name,
          success: true,
          message: response.data.message,
          count: response.data.count
        })
        
        // Reset file selection
        this.resetFileSelection()
        
      } catch (error) {
        const errorMessage = error.response?.data?.message || 'Upload failed'
        
        // Add error result
        this.addResult({
          filename: this.selectedFile.name,
          success: false,
          message: errorMessage
        })
      } finally {
        this.isUploading = false
      }
    },
    
    async scanFolder() {
      if (!this.folderPath) return
      
      this.isScanning = true
      
      try {
        const response = await axios.post('/import/scan', {
          folderPath: this.folderPath
        }, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        })
        
        // Add success result
        this.addResult({
          filename: this.folderPath,
          success: true,
          message: response.data.message,
          count: response.data.count
        })
        
      } catch (error) {
        const errorMessage = error.response?.data?.message || 'Scan failed'
        
        // Add error result
        this.addResult({
          filename: this.folderPath,
          success: false,
          message: errorMessage
        })
      } finally {
        this.isScanning = false
      }
    },
    
    addResult(result) {
      // Add timestamp
      result.timestamp = new Date()
      
      // Add to beginning of array
      this.importResults.unshift(result)
    }
  }
}
</script>

<style scoped>
.import-container {
  max-width: 1200px;
  margin: 0 auto;
}

h1, h2 {
  margin-bottom: 1.5rem;
  color: var(--text-color);
}

.import-options {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(450px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.import-card {
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 1.5rem;
}

.import-header {
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
}

.import-header i {
  font-size: 1.5rem;
  margin-right: 0.75rem;
  color: var(--primary-color);
}

.import-header h2 {
  margin-bottom: 0;
  font-size: 1.25rem;
}

.file-dropzone {
  margin: 1.5rem 0;
  padding: 2rem;
  border: 2px dashed var(--border-color);
  border-radius: 8px;
  background-color: var(--sidebar-bg);
  text-align: center;
  cursor: pointer;
  transition: border-color 0.2s, background-color 0.2s;
}

.file-dropzone.active {
  border-color: var(--primary-color);
  background-color: rgba(255, 0, 0, 0.05);
}

.file-dropzone i {
  font-size: 2.5rem;
  color: var(--text-color);
  margin-bottom: 1rem;
}

.file-dropzone p {
  margin-bottom: 1rem;
}

.selected-file {
  display: flex;
  align-items: center;
  padding: 0.75rem;
  background-color: var(--background-color);
  border-radius: 4px;
}

.selected-file i {
  font-size: 1.25rem;
  margin-right: 0.75rem;
  margin-bottom: 0;
  color: var(--primary-color);
}

.selected-file span {
  flex: 1;
  text-align: left;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.btn-icon {
  background: transparent;
  border: none;
  color: var(--text-color);
  cursor: pointer;
  padding: 0.25rem;
  font-size: 1rem;
}

.btn-select {
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 0.5rem 1rem;
  font-size: 0.9rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-select:hover {
  background-color: var(--hover-color);
}

.folder-input {
  display: flex;
  margin: 1.5rem 0;
}

.folder-input input {
  flex: 1;
  padding: 0.75rem;
  border: 1px solid var(--input-border);
  border-radius: 4px 0 0 4px;
  background-color: var(--input-bg);
  color: var(--text-color);
}

.btn-browse {
  padding: 0.75rem 1rem;
  background-color: var(--secondary-color);
  color: white;
  border: none;
  border-radius: 0 4px 4px 0;
  cursor: pointer;
}

.import-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
}

.btn-primary {
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

.btn-primary:hover:not(:disabled) {
  background-color: var(--button-hover);
}

.btn-primary i {
  margin-right: 0.5rem;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.import-results {
  margin-top: 2rem;
}

.results-list {
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.result-item {
  display: flex;
  padding: 1rem;
  border-bottom: 1px solid var(--border-color);
}

.result-item:last-child {
  border-bottom: none;
}

.result-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 1rem;
}

.result-icon.success {
  background-color: rgba(46, 204, 113, 0.1);
  color: #2ECC71;
}

.result-icon.error {
  background-color: rgba(255, 0, 0, 0.1);
  color: var(--primary-color);
}

.result-content h3 {
  font-size: 1rem;
  margin-bottom: 0.25rem;
}

.result-content p {
  font-size: 0.9rem;
  color: var(--text-color);
  opacity: 0.8;
}

.result-content .count {
  margin-top: 0.25rem;
  font-weight: 500;
  color: #2ECC71;
}

/* Responsive */
@media (max-width: 768px) {
  .import-options {
    grid-template-columns: 1fr;
  }
}
</style>