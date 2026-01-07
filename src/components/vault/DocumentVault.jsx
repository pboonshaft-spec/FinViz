import React, { useState, useEffect, useRef } from 'react';
import { useApi } from '../../hooks/useApi';

const CATEGORIES = [
  { id: 'all', label: 'All Documents', icon: '📁' },
  { id: 'reports', label: 'Financial Reports', icon: '📊' },
  { id: 'tax_returns', label: 'Tax Returns', icon: '📋' },
  { id: 'statements', label: 'Statements', icon: '📄' },
  { id: 'estate_docs', label: 'Estate Documents', icon: '📜' },
  { id: 'insurance', label: 'Insurance', icon: '🛡️' },
  { id: 'investments', label: 'Investments', icon: '📈' },
  { id: 'other', label: 'Other', icon: '📎' },
];

const FILE_ICONS = {
  'application/pdf': '📕',
  'image/jpeg': '🖼️',
  'image/png': '🖼️',
  'text/csv': '📊',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': '📗',
  'application/vnd.ms-excel': '📗',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document': '📘',
  'application/msword': '📘',
  'text/plain': '📝',
};

const styles = {
  container: {
    display: 'flex',
    flexDirection: 'column',
    height: '100%',
    background: '#1a1a1a',
    borderRadius: '12px',
    overflow: 'hidden',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '20px 24px',
    borderBottom: '1px solid #2a2a2a',
  },
  title: {
    fontSize: '1.5rem',
    fontWeight: '600',
    color: '#fff',
    margin: 0,
  },
  uploadBtn: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    padding: '10px 20px',
    background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    fontSize: '0.9rem',
    fontWeight: '500',
    cursor: 'pointer',
    transition: 'all 0.2s ease',
  },
  content: {
    display: 'flex',
    flex: 1,
    overflow: 'hidden',
  },
  sidebar: {
    width: '220px',
    borderRight: '1px solid #2a2a2a',
    padding: '16px 0',
    overflow: 'auto',
  },
  categoryItem: {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    padding: '10px 20px',
    color: '#888',
    cursor: 'pointer',
    transition: 'all 0.15s ease',
    fontSize: '0.9rem',
  },
  categoryItemActive: {
    background: 'rgba(99, 102, 241, 0.1)',
    color: '#6366f1',
    borderRight: '3px solid #6366f1',
  },
  categoryCount: {
    marginLeft: 'auto',
    fontSize: '0.75rem',
    background: '#2a2a2a',
    padding: '2px 8px',
    borderRadius: '10px',
  },
  main: {
    flex: 1,
    overflow: 'auto',
    padding: '20px',
  },
  documentGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
    gap: '16px',
  },
  documentCard: {
    background: '#242424',
    borderRadius: '10px',
    padding: '16px',
    border: '1px solid #333',
    transition: 'all 0.2s ease',
    cursor: 'pointer',
  },
  documentCardHover: {
    borderColor: '#6366f1',
    transform: 'translateY(-2px)',
    boxShadow: '0 4px 20px rgba(99, 102, 241, 0.2)',
  },
  docHeader: {
    display: 'flex',
    alignItems: 'flex-start',
    gap: '12px',
    marginBottom: '12px',
  },
  docIcon: {
    fontSize: '2rem',
    lineHeight: 1,
  },
  docInfo: {
    flex: 1,
    minWidth: 0,
  },
  docName: {
    fontSize: '0.95rem',
    fontWeight: '500',
    color: '#fff',
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    marginBottom: '4px',
  },
  docMeta: {
    fontSize: '0.8rem',
    color: '#666',
  },
  docActions: {
    display: 'flex',
    gap: '8px',
    marginTop: '12px',
    paddingTop: '12px',
    borderTop: '1px solid #333',
  },
  actionBtn: {
    flex: 1,
    padding: '6px 12px',
    background: '#333',
    border: 'none',
    borderRadius: '6px',
    color: '#fff',
    fontSize: '0.8rem',
    cursor: 'pointer',
    transition: 'background 0.15s ease',
  },
  deleteBtn: {
    background: 'rgba(239, 68, 68, 0.2)',
    color: '#ef4444',
  },
  emptyState: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    justifyContent: 'center',
    height: '100%',
    color: '#666',
    textAlign: 'center',
    padding: '40px',
  },
  emptyIcon: {
    fontSize: '4rem',
    marginBottom: '16px',
    opacity: 0.5,
  },
  uploadModal: {
    position: 'fixed',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    background: 'rgba(0,0,0,0.7)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    zIndex: 1000,
  },
  modalContent: {
    background: '#1e1e1e',
    borderRadius: '16px',
    padding: '24px',
    width: '100%',
    maxWidth: '480px',
    border: '1px solid #333',
  },
  modalTitle: {
    fontSize: '1.25rem',
    fontWeight: '600',
    color: '#fff',
    marginBottom: '20px',
  },
  dropzone: {
    border: '2px dashed #333',
    borderRadius: '10px',
    padding: '40px 20px',
    textAlign: 'center',
    cursor: 'pointer',
    transition: 'all 0.2s ease',
    marginBottom: '20px',
  },
  dropzoneActive: {
    borderColor: '#6366f1',
    background: 'rgba(99, 102, 241, 0.1)',
  },
  formGroup: {
    marginBottom: '16px',
  },
  label: {
    display: 'block',
    fontSize: '0.85rem',
    color: '#888',
    marginBottom: '6px',
  },
  input: {
    width: '100%',
    padding: '10px 12px',
    background: '#252525',
    border: '1px solid #333',
    borderRadius: '6px',
    color: '#fff',
    fontSize: '0.9rem',
  },
  select: {
    width: '100%',
    padding: '10px 12px',
    background: '#252525',
    border: '1px solid #333',
    borderRadius: '6px',
    color: '#fff',
    fontSize: '0.9rem',
  },
  modalActions: {
    display: 'flex',
    gap: '12px',
    marginTop: '20px',
  },
  cancelBtn: {
    flex: 1,
    padding: '12px',
    background: '#333',
    border: 'none',
    borderRadius: '8px',
    color: '#fff',
    cursor: 'pointer',
  },
  submitBtn: {
    flex: 1,
    padding: '12px',
    background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
    border: 'none',
    borderRadius: '8px',
    color: '#fff',
    cursor: 'pointer',
    fontWeight: '500',
  },
};

function formatFileSize(bytes) {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

function formatDate(dateStr) {
  const date = new Date(dateStr);
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  });
}

function DocumentVault() {
  const { getDocuments, uploadDocument, downloadDocument, deleteDocument, loading } = useApi();
  const [documents, setDocuments] = useState([]);
  const [categories, setCategories] = useState({});
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [hoveredDoc, setHoveredDoc] = useState(null);
  const [isDragging, setIsDragging] = useState(false);
  const [uploadFile, setUploadFile] = useState(null);
  const [uploadMeta, setUploadMeta] = useState({ name: '', category: 'other', description: '', year: '' });
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef(null);

  const fetchDocuments = async () => {
    try {
      const cat = selectedCategory === 'all' ? null : selectedCategory;
      const response = await getDocuments(cat);
      setDocuments(response.documents || []);
      setCategories(response.categories || {});
    } catch (err) {
      console.error('Failed to fetch documents:', err);
    }
  };

  useEffect(() => {
    fetchDocuments();
  }, [selectedCategory]);

  const handleFileSelect = (files) => {
    if (files && files.length > 0) {
      const file = files[0];
      setUploadFile(file);
      setUploadMeta(prev => ({ ...prev, name: file.name }));
    }
  };

  const handleUpload = async () => {
    if (!uploadFile) return;

    setUploading(true);
    try {
      await uploadDocument(uploadFile, uploadMeta);
      setShowUploadModal(false);
      setUploadFile(null);
      setUploadMeta({ name: '', category: 'other', description: '', year: '' });
      fetchDocuments();
    } catch (err) {
      console.error('Upload failed:', err);
    } finally {
      setUploading(false);
    }
  };

  const handleDownload = async (doc) => {
    try {
      await downloadDocument(doc.id, doc.original_name);
    } catch (err) {
      console.error('Download failed:', err);
    }
  };

  const handleDelete = async (doc) => {
    if (!window.confirm(`Delete "${doc.name}"?`)) return;

    try {
      await deleteDocument(doc.id);
      fetchDocuments();
    } catch (err) {
      console.error('Delete failed:', err);
    }
  };

  const filteredDocuments = selectedCategory === 'all'
    ? documents
    : documents.filter(d => d.category === selectedCategory);

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <h2 style={styles.title}>Document Vault</h2>
        <button
          style={styles.uploadBtn}
          onClick={() => setShowUploadModal(true)}
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="17 8 12 3 7 8"/>
            <line x1="12" y1="3" x2="12" y2="15"/>
          </svg>
          Upload Document
        </button>
      </div>

      <div style={styles.content}>
        <div style={styles.sidebar}>
          {CATEGORIES.map(cat => {
            const count = cat.id === 'all'
              ? documents.length
              : (categories[cat.id] || 0);
            return (
              <div
                key={cat.id}
                style={{
                  ...styles.categoryItem,
                  ...(selectedCategory === cat.id ? styles.categoryItemActive : {})
                }}
                onClick={() => setSelectedCategory(cat.id)}
              >
                <span>{cat.icon}</span>
                <span>{cat.label}</span>
                <span style={styles.categoryCount}>{count}</span>
              </div>
            );
          })}
        </div>

        <div style={styles.main}>
          {loading && documents.length === 0 ? (
            <div style={styles.emptyState}>Loading...</div>
          ) : filteredDocuments.length === 0 ? (
            <div style={styles.emptyState}>
              <div style={styles.emptyIcon}>📂</div>
              <div>No documents yet</div>
              <div style={{ fontSize: '0.9rem', marginTop: '8px' }}>
                Upload your first document to get started
              </div>
            </div>
          ) : (
            <div style={styles.documentGrid}>
              {filteredDocuments.map(doc => (
                <div
                  key={doc.id}
                  style={{
                    ...styles.documentCard,
                    ...(hoveredDoc === doc.id ? styles.documentCardHover : {})
                  }}
                  onMouseEnter={() => setHoveredDoc(doc.id)}
                  onMouseLeave={() => setHoveredDoc(null)}
                >
                  <div style={styles.docHeader}>
                    <span style={styles.docIcon}>
                      {FILE_ICONS[doc.mime_type] || '📄'}
                    </span>
                    <div style={styles.docInfo}>
                      <div style={styles.docName} title={doc.name}>
                        {doc.name}
                      </div>
                      <div style={styles.docMeta}>
                        {formatFileSize(doc.size)} • {formatDate(doc.created_at)}
                      </div>
                    </div>
                  </div>
                  <div style={styles.docActions}>
                    <button
                      style={styles.actionBtn}
                      onClick={() => handleDownload(doc)}
                    >
                      Download
                    </button>
                    {doc.can_delete && (
                      <button
                        style={{...styles.actionBtn, ...styles.deleteBtn}}
                        onClick={() => handleDelete(doc)}
                      >
                        Delete
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {showUploadModal && (
        <div style={styles.uploadModal} onClick={() => setShowUploadModal(false)}>
          <div style={styles.modalContent} onClick={e => e.stopPropagation()}>
            <div style={styles.modalTitle}>Upload Document</div>

            <div
              style={{
                ...styles.dropzone,
                ...(isDragging ? styles.dropzoneActive : {})
              }}
              onClick={() => fileInputRef.current?.click()}
              onDragOver={(e) => { e.preventDefault(); setIsDragging(true); }}
              onDragLeave={() => setIsDragging(false)}
              onDrop={(e) => {
                e.preventDefault();
                setIsDragging(false);
                handleFileSelect(e.dataTransfer.files);
              }}
            >
              <input
                ref={fileInputRef}
                type="file"
                style={{ display: 'none' }}
                onChange={(e) => handleFileSelect(e.target.files)}
                accept=".pdf,.jpg,.jpeg,.png,.xlsx,.xls,.docx,.doc,.csv,.txt"
              />
              {uploadFile ? (
                <div>
                  <div style={{ fontSize: '2rem', marginBottom: '8px' }}>
                    {FILE_ICONS[uploadFile.type] || '📄'}
                  </div>
                  <div style={{ color: '#fff' }}>{uploadFile.name}</div>
                  <div style={{ color: '#666', fontSize: '0.85rem' }}>
                    {formatFileSize(uploadFile.size)}
                  </div>
                </div>
              ) : (
                <div>
                  <div style={{ fontSize: '2.5rem', marginBottom: '8px', opacity: 0.5 }}>📤</div>
                  <div style={{ color: '#fff' }}>Drop file here or click to browse</div>
                  <div style={{ color: '#666', fontSize: '0.85rem', marginTop: '4px' }}>
                    PDF, Images, Excel, Word, CSV (max 25MB)
                  </div>
                </div>
              )}
            </div>

            <div style={styles.formGroup}>
              <label style={styles.label}>Document Name</label>
              <input
                type="text"
                style={styles.input}
                value={uploadMeta.name}
                onChange={(e) => setUploadMeta(prev => ({ ...prev, name: e.target.value }))}
                placeholder="Enter document name"
              />
            </div>

            <div style={styles.formGroup}>
              <label style={styles.label}>Category</label>
              <select
                style={styles.select}
                value={uploadMeta.category}
                onChange={(e) => setUploadMeta(prev => ({ ...prev, category: e.target.value }))}
              >
                {CATEGORIES.filter(c => c.id !== 'all').map(cat => (
                  <option key={cat.id} value={cat.id}>{cat.label}</option>
                ))}
              </select>
            </div>

            <div style={styles.formGroup}>
              <label style={styles.label}>Year (optional)</label>
              <input
                type="number"
                style={styles.input}
                value={uploadMeta.year}
                onChange={(e) => setUploadMeta(prev => ({ ...prev, year: e.target.value }))}
                placeholder="e.g., 2024"
                min="1900"
                max="2100"
              />
            </div>

            <div style={styles.formGroup}>
              <label style={styles.label}>Description (optional)</label>
              <input
                type="text"
                style={styles.input}
                value={uploadMeta.description}
                onChange={(e) => setUploadMeta(prev => ({ ...prev, description: e.target.value }))}
                placeholder="Brief description"
              />
            </div>

            <div style={styles.modalActions}>
              <button
                style={styles.cancelBtn}
                onClick={() => {
                  setShowUploadModal(false);
                  setUploadFile(null);
                }}
              >
                Cancel
              </button>
              <button
                style={{
                  ...styles.submitBtn,
                  opacity: !uploadFile || uploading ? 0.6 : 1,
                }}
                onClick={handleUpload}
                disabled={!uploadFile || uploading}
              >
                {uploading ? 'Uploading...' : 'Upload'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default DocumentVault;
