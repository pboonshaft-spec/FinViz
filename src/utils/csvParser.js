import Papa from 'papaparse';

export function parseCSV(file) {
  return new Promise((resolve, reject) => {
    Papa.parse(file, {
      header: true,
      dynamicTyping: false,
      skipEmptyLines: true,
      transformHeader: (header, index) => {
        // Trim whitespace and handle empty headers
        const trimmed = header.trim();
        if (!trimmed) {
          return `Column_${index + 1}`;
        }
        return trimmed;
      },
      complete: (results) => {
        // Check for duplicate headers and rename them
        const headers = results.meta.fields || [];
        const seenHeaders = new Map();
        const uniqueHeaders = headers.map((header, index) => {
          if (seenHeaders.has(header)) {
            const count = seenHeaders.get(header) + 1;
            seenHeaders.set(header, count);
            return `${header}_${count}`;
          } else {
            seenHeaders.set(header, 1);
            return header;
          }
        });

        // Update data with renamed headers if duplicates were found
        if (uniqueHeaders.some((h, i) => h !== headers[i])) {
          results.data = results.data.map(row => {
            const newRow = {};
            headers.forEach((oldHeader, i) => {
              newRow[uniqueHeaders[i]] = row[oldHeader];
            });
            return newRow;
          });
        }

        resolve({
          filename: file.name,
          data: results.data,
          headers: uniqueHeaders
        });
      },
      error: (error) => {
        reject(new Error(`Error parsing ${file.name}: ${error.message}`));
      }
    });
  });
}
