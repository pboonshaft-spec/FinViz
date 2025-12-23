import Papa from 'papaparse';

export function parseCSV(file) {
  return new Promise((resolve, reject) => {
    Papa.parse(file, {
      header: true,
      dynamicTyping: false,
      skipEmptyLines: true,
      complete: (results) => {
        resolve({
          filename: file.name,
          data: results.data,
          headers: results.meta.fields
        });
      },
      error: (error) => {
        reject(new Error(`Error parsing ${file.name}: ${error.message}`));
      }
    });
  });
}
