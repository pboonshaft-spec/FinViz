import React, { useRef, useState } from 'react';

function CSVImportModal({ isOpen, onClose, onFilesSelected, isProcessing }) {
  const inputRef = useRef(null);
  const [isDragging, setIsDragging] = useState(false);

  if (!isOpen) return null;

  const handleDropzoneClick = () => {
    inputRef.current?.click();
  };

  const handleChange = (e) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      onFilesSelected(files);
    }
  };

  const handleDragOver = (e) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (e) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (e) => {
    e.preventDefault();
    setIsDragging(false);
    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      onFilesSelected(files);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content csv-import-modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h3>Import CSV</h3>
          <button className="btn-close" onClick={onClose}>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        <div
          className={`dropzone ${isDragging ? 'active' : ''}`}
          onClick={handleDropzoneClick}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
        >
          <input
            ref={inputRef}
            type="file"
            accept=".csv"
            multiple
            onChange={handleChange}
            style={{ display: 'none' }}
          />
          {isProcessing ? (
            <>
              <div className="dropzone-icon">
                <span className="spinner"></span>
              </div>
              <div className="dropzone-text">Processing...</div>
            </>
          ) : (
            <>
              <div className="dropzone-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                  <polyline points="17 8 12 3 7 8" />
                  <line x1="12" y1="3" x2="12" y2="15" />
                </svg>
              </div>
              <div className="dropzone-text">Drop files here or click to browse</div>
              <div className="dropzone-hint">Supports CSV files with transaction data</div>
            </>
          )}
        </div>

        <div className="modal-footer">
          <p className="hint">
            Your CSV should have columns for date, description, and amount.
          </p>
        </div>
      </div>
    </div>
  );
}

export default CSVImportModal;
