import React from 'react';

function ConfirmModal({ isOpen, title, message, options, onSelect }) {
  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={() => onSelect(null)}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <h3>{title}</h3>
        <p>{message}</p>
        <div className="modal-actions">
          {options.map((option, idx) => (
            <button
              key={idx}
              className={`btn ${option.variant === 'danger' ? 'btn-danger' : option.variant === 'primary' ? 'btn-primary' : 'btn-secondary'}`}
              onClick={() => onSelect(option.value)}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}

export default ConfirmModal;
