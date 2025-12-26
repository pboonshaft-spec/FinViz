import React, { useState } from 'react';

function DebtForm({ debt, onSubmit, onCancel }) {
  const [formData, setFormData] = useState({
    name: debt?.name || '',
    currentBalance: debt?.currentBalance || '',
    interestRate: debt?.interestRate || '',
    minimumPayment: debt?.minimumPayment || '',
  });

  const handleSubmit = (e) => {
    e.preventDefault();
    onSubmit({
      name: formData.name,
      currentBalance: parseFloat(formData.currentBalance) || 0,
      interestRate: formData.interestRate ? parseFloat(formData.interestRate) : null,
      minimumPayment: formData.minimumPayment ? parseFloat(formData.minimumPayment) : null,
    });
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  return (
    <div className="form-card">
      <form onSubmit={handleSubmit}>
        <div className="form-row">
          <div className="form-group">
            <label htmlFor="name">Debt Name</label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              placeholder="e.g., Chase Credit Card"
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="currentBalance">Current Balance ($)</label>
            <input
              type="number"
              id="currentBalance"
              name="currentBalance"
              value={formData.currentBalance}
              onChange={handleChange}
              placeholder="0.00"
              step="0.01"
              min="0"
              required
            />
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="interestRate">Interest Rate (% APR)</label>
            <input
              type="number"
              id="interestRate"
              name="interestRate"
              value={formData.interestRate}
              onChange={handleChange}
              placeholder="e.g., 18.99"
              step="0.01"
              min="0"
              max="100"
            />
          </div>
          <div className="form-group">
            <label htmlFor="minimumPayment">Minimum Payment ($/mo)</label>
            <input
              type="number"
              id="minimumPayment"
              name="minimumPayment"
              value={formData.minimumPayment}
              onChange={handleChange}
              placeholder="e.g., 25.00"
              step="0.01"
              min="0"
            />
          </div>
        </div>

        <div className="form-actions">
          <button type="button" className="btn btn-secondary" onClick={onCancel}>
            Cancel
          </button>
          <button type="submit" className="btn btn-primary">
            {debt ? 'Update Debt' : 'Add Debt'}
          </button>
        </div>
      </form>
    </div>
  );
}

export default DebtForm;
