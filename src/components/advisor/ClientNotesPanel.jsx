import React, { useState, useEffect } from 'react';
import { useClientContext } from '../../contexts/ClientContext';
import { useAuth } from '../../contexts/AuthContext';
import './AdvisorStyles.css';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8085';

const CATEGORIES = [
  { value: 'general', label: 'General', color: '#6b7280' },
  { value: 'meeting', label: 'Meeting', color: '#3b82f6' },
  { value: 'goal', label: 'Goal', color: '#10b981' },
  { value: 'concern', label: 'Concern', color: '#f59e0b' },
  { value: 'action_item', label: 'Action Item', color: '#ef4444' },
  { value: 'personal', label: 'Personal', color: '#8b5cf6' },
];

export default function ClientNotesPanel() {
  const { activeClient } = useClientContext();
  const { token } = useAuth();
  const [notes, setNotes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filterCategory, setFilterCategory] = useState('');

  // New note form
  const [newNote, setNewNote] = useState('');
  const [newCategory, setNewCategory] = useState('general');
  const [saving, setSaving] = useState(false);

  // Edit state
  const [editingId, setEditingId] = useState(null);
  const [editText, setEditText] = useState('');
  const [editCategory, setEditCategory] = useState('');

  const fetchNotes = async () => {
    if (!activeClient) return;

    try {
      setLoading(true);
      const params = new URLSearchParams();
      if (filterCategory) params.append('category', filterCategory);

      const response = await fetch(
        `${API_BASE_URL}/api/advisor/clients/${activeClient.id}/notes?${params.toString()}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) throw new Error('Failed to fetch notes');
      const data = await response.json();
      setNotes(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNotes();
  }, [activeClient?.id, filterCategory]);

  const handleAddNote = async (e) => {
    e.preventDefault();
    if (!newNote.trim() || !activeClient) return;

    try {
      setSaving(true);
      const response = await fetch(
        `${API_BASE_URL}/api/advisor/clients/${activeClient.id}/notes`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            note: newNote.trim(),
            category: newCategory,
          }),
        }
      );

      if (!response.ok) throw new Error('Failed to add note');
      const created = await response.json();
      setNotes(prev => [created, ...prev]);
      setNewNote('');
      setNewCategory('general');
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const handleUpdateNote = async (noteId) => {
    try {
      const response = await fetch(
        `${API_BASE_URL}/api/advisor/clients/${activeClient.id}/notes/${noteId}`,
        {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            note: editText,
            category: editCategory,
          }),
        }
      );

      if (!response.ok) throw new Error('Failed to update note');
      const updated = await response.json();
      setNotes(prev => prev.map(n => n.id === noteId ? updated : n));
      setEditingId(null);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleDeleteNote = async (noteId) => {
    if (!confirm('Delete this note?')) return;

    try {
      const response = await fetch(
        `${API_BASE_URL}/api/advisor/clients/${activeClient.id}/notes/${noteId}`,
        {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) throw new Error('Failed to delete note');
      setNotes(prev => prev.filter(n => n.id !== noteId));
    } catch (err) {
      setError(err.message);
    }
  };

  const handleTogglePin = async (note) => {
    try {
      const response = await fetch(
        `${API_BASE_URL}/api/advisor/clients/${activeClient.id}/notes/${note.id}`,
        {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            isPinned: !note.isPinned,
          }),
        }
      );

      if (!response.ok) throw new Error('Failed to update note');
      const updated = await response.json();
      setNotes(prev => prev.map(n => n.id === note.id ? updated : n));
    } catch (err) {
      setError(err.message);
    }
  };

  const startEdit = (note) => {
    setEditingId(note.id);
    setEditText(note.note);
    setEditCategory(note.category);
  };

  const getCategoryInfo = (categoryValue) => {
    return CATEGORIES.find(c => c.value === categoryValue) || CATEGORIES[0];
  };

  const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffDays = Math.floor((now - date) / (1000 * 60 * 60 * 24));

    if (diffDays === 0) {
      return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
    } else if (diffDays === 1) {
      return 'Yesterday';
    } else if (diffDays < 7) {
      return date.toLocaleDateString('en-US', { weekday: 'short' });
    } else {
      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    }
  };

  if (!activeClient) {
    return <div className="notes-panel empty">Select a client to view notes</div>;
  }

  return (
    <div className="notes-panel">
      <div className="notes-header">
        <h3>Client Notes</h3>
        <div className="notes-filter">
          <select
            value={filterCategory}
            onChange={(e) => setFilterCategory(e.target.value)}
          >
            <option value="">All Categories</option>
            {CATEGORIES.map(cat => (
              <option key={cat.value} value={cat.value}>{cat.label}</option>
            ))}
          </select>
        </div>
      </div>

      {error && (
        <div className="notes-error">
          {error}
          <button onClick={() => setError(null)}>Dismiss</button>
        </div>
      )}

      <form className="add-note-form" onSubmit={handleAddNote}>
        <textarea
          value={newNote}
          onChange={(e) => setNewNote(e.target.value)}
          placeholder="Add a note about this client..."
          rows={3}
        />
        <div className="add-note-actions">
          <select
            value={newCategory}
            onChange={(e) => setNewCategory(e.target.value)}
          >
            {CATEGORIES.map(cat => (
              <option key={cat.value} value={cat.value}>{cat.label}</option>
            ))}
          </select>
          <button
            type="submit"
            className="btn btn-primary"
            disabled={!newNote.trim() || saving}
          >
            {saving ? 'Saving...' : 'Add Note'}
          </button>
        </div>
      </form>

      <div className="notes-list">
        {loading ? (
          <div className="notes-loading">Loading notes...</div>
        ) : notes.length === 0 ? (
          <div className="notes-empty">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
              <polyline points="14 2 14 8 20 8" />
              <line x1="16" y1="13" x2="8" y2="13" />
              <line x1="16" y1="17" x2="8" y2="17" />
            </svg>
            <p>No notes yet</p>
            <span>Add your first note above</span>
          </div>
        ) : (
          notes.map(note => {
            const catInfo = getCategoryInfo(note.category);
            const isEditing = editingId === note.id;

            return (
              <div
                key={note.id}
                className={`note-card ${note.isPinned ? 'pinned' : ''}`}
              >
                {note.isPinned && (
                  <div className="pin-indicator">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M16 4h2v6l3 3v2h-6v6l-1 1-1-1v-6H7v-2l3-3V4h2V2h4v2z" />
                    </svg>
                  </div>
                )}

                <div className="note-header">
                  <span
                    className="note-category-badge"
                    style={{ backgroundColor: catInfo.color }}
                  >
                    {catInfo.label}
                  </span>
                  <span className="note-date">{formatDate(note.createdAt)}</span>
                </div>

                {isEditing ? (
                  <div className="note-edit-form">
                    <textarea
                      value={editText}
                      onChange={(e) => setEditText(e.target.value)}
                      rows={3}
                    />
                    <div className="note-edit-actions">
                      <select
                        value={editCategory}
                        onChange={(e) => setEditCategory(e.target.value)}
                      >
                        {CATEGORIES.map(cat => (
                          <option key={cat.value} value={cat.value}>{cat.label}</option>
                        ))}
                      </select>
                      <button
                        className="btn btn-sm btn-primary"
                        onClick={() => handleUpdateNote(note.id)}
                      >
                        Save
                      </button>
                      <button
                        className="btn btn-sm btn-secondary"
                        onClick={() => setEditingId(null)}
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                ) : (
                  <p className="note-content">{note.note}</p>
                )}

                {!isEditing && (
                  <div className="note-actions">
                    <button
                      className="note-action-btn"
                      onClick={() => handleTogglePin(note)}
                      title={note.isPinned ? 'Unpin' : 'Pin'}
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill={note.isPinned ? "currentColor" : "none"} stroke="currentColor" strokeWidth="2">
                        <path d="M16 4h2v6l3 3v2h-6v6l-1 1-1-1v-6H7v-2l3-3V4h2V2h4v2z" />
                      </svg>
                    </button>
                    <button
                      className="note-action-btn"
                      onClick={() => startEdit(note)}
                      title="Edit"
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                        <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
                      </svg>
                    </button>
                    <button
                      className="note-action-btn delete"
                      onClick={() => handleDeleteNote(note.id)}
                      title="Delete"
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <polyline points="3 6 5 6 21 6" />
                        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                      </svg>
                    </button>
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
