<template>
  <div class="upload-section">
    <div class="upload-area" @click="openFileDialog" @drop.prevent="handleDrop" @dragover.prevent>
      <input
        ref="fileInput"
        type="file"
        accept=".csv"
        multiple
        @change="handleFileChange"
        style="display: none"
      />
      <div class="upload-icon">📊</div>
      <h3>Drop CSV files here or click to upload</h3>
      <p style="margin-top: 10px; color: #666;">
        Supports transactions, expenses, income, and budget data
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useFileUpload } from '../composables/useFileUpload.js';

const emit = defineEmits(['files-parsed']);

const fileInput = ref(null);
const { handleFiles, isProcessing } = useFileUpload();

function openFileDialog() {
  fileInput.value?.click();
}

async function handleFileChange(event) {
  const files = event.target.files;
  if (files && files.length > 0) {
    await processFiles(files);
  }
}

async function handleDrop(event) {
  const files = event.dataTransfer.files;
  if (files && files.length > 0) {
    await processFiles(files);
  }
}

async function processFiles(files) {
  const parsedFiles = await handleFiles(files);
  if (parsedFiles && parsedFiles.length > 0) {
    emit('files-parsed', parsedFiles);
  }
}
</script>

<style scoped>
.upload-section {
  background: white;
  border-radius: 12px;
  padding: 30px;
  margin-bottom: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
}

.upload-area {
  border: 2px dashed #667eea;
  border-radius: 8px;
  padding: 40px;
  text-align: center;
  background: #f8f9ff;
  cursor: pointer;
  transition: all 0.3s ease;
}

.upload-area:hover {
  background: #eef1ff;
  border-color: #764ba2;
}

.upload-icon {
  font-size: 3rem;
  margin-bottom: 15px;
}
</style>
