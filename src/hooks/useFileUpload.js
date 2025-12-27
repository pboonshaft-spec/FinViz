import { useState, useCallback } from 'react';
import { parseCSV } from '../utils/csvParser.js';

export function useFileUpload() {
  const [isProcessing, setIsProcessing] = useState(false);
  const [error, setError] = useState(null);

  const handleFiles = useCallback(async (files) => {
    setIsProcessing(true);
    setError(null);

    try {
      const filesArray = Array.from(files);
      const parsedFiles = [];

      for (const file of filesArray) {
        try {
          const result = await parseCSV(file);
          parsedFiles.push(result);
        } catch (err) {
          console.error(`Error parsing ${file.name}:`, err);
          setError(err.message);
        }
      }

      return parsedFiles;
    } finally {
      setIsProcessing(false);
    }
  }, []);

  return {
    handleFiles,
    isProcessing,
    error
  };
}
