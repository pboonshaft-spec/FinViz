import React, { useRef, useState } from 'react';

const styles = {
  wrapper: {
    position: 'relative'
  },
  button: {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '8px',
    padding: '10px 20px',
    background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
    color: 'white',
    border: 'none',
    borderRadius: '10px',
    fontSize: '0.9rem',
    fontWeight: '500',
    cursor: 'pointer',
    transition: 'all 0.2s ease',
    boxShadow: '0 4px 12px rgba(99, 102, 241, 0.3)'
  },
  buttonHover: {
    transform: 'translateY(-1px)',
    boxShadow: '0 6px 20px rgba(99, 102, 241, 0.4)'
  },
  dropdown: {
    position: 'absolute',
    top: 'calc(100% + 8px)',
    left: '0',
    background: '#1e1e1e',
    borderRadius: '12px',
    padding: '20px',
    minWidth: '320px',
    boxShadow: '0 8px 32px rgba(0, 0, 0, 0.5)',
    border: '1px solid #2a2a2a',
    zIndex: 100
  },
  dropzone: {
    border: '2px dashed #333',
    borderRadius: '10px',
    padding: '30px 20px',
    textAlign: 'center',
    cursor: 'pointer',
    transition: 'all 0.2s ease',
    background: '#252525'
  },
  dropzoneActive: {
    borderColor: '#6366f1',
    background: 'rgba(99, 102, 241, 0.1)'
  },
  dropzoneIcon: {
    fontSize: '2rem',
    marginBottom: '12px'
  },
  dropzoneText: {
    color: '#fff',
    fontSize: '0.95rem',
    fontWeight: '500',
    marginBottom: '4px'
  },
  dropzoneHint: {
    color: '#666',
    fontSize: '0.8rem'
  },
  processing: {
    color: '#6366f1'
  }
};

function FileUpload({ onFilesSelected, isProcessing }) {
  const inputRef = useRef(null);
  const [isOpen, setIsOpen] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const [isHovered, setIsHovered] = useState(false);

  const handleButtonClick = () => {
    setIsOpen(!isOpen);
  };

  const handleDropzoneClick = () => {
    inputRef.current?.click();
  };

  const handleChange = (e) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      onFilesSelected(files);
      setIsOpen(false);
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
      setIsOpen(false);
    }
  };

  return (
    <div style={styles.wrapper}>
      <button
        style={{
          ...styles.button,
          ...(isHovered ? styles.buttonHover : {})
        }}
        onClick={handleButtonClick}
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
      >
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
          <polyline points="17 8 12 3 7 8" />
          <line x1="12" y1="3" x2="12" y2="15" />
        </svg>
        {isProcessing ? 'Processing...' : 'Upload CSV'}
      </button>

      {isOpen && (
        <div style={styles.dropdown}>
          <div
            style={{
              ...styles.dropzone,
              ...(isDragging ? styles.dropzoneActive : {})
            }}
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
            <div style={styles.dropzoneIcon}>📊</div>
            <div style={styles.dropzoneText}>
              Drop files here or click to browse
            </div>
            <div style={styles.dropzoneHint}>
              Supports CSV files with transactions data
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default FileUpload;
