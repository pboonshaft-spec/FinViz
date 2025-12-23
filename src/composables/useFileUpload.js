import { ref } from 'vue';
import { parseCSV } from '../utils/csvParser.js';

export function useFileUpload() {
  const isProcessing = ref(false);
  const error = ref(null);

  async function handleFiles(files) {
    isProcessing.value = true;
    error.value = null;

    try {
      const filesArray = Array.from(files);
      const parsedFiles = [];

      for (const file of filesArray) {
        try {
          const result = await parseCSV(file);
          parsedFiles.push(result);
        } catch (err) {
          console.error(`Error parsing ${file.name}:`, err);
          error.value = err.message;
        }
      }

      return parsedFiles;
    } finally {
      isProcessing.value = false;
    }
  }

  return {
    handleFiles,
    isProcessing,
    error
  };
}
