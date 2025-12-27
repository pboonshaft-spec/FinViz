import React, { useState } from 'react';

function AssetForm({ assetTypes, asset, onSubmit, onCancel }) {
  const [formData, setFormData] = useState({
    name: asset?.name || '',
    typeId: asset?.typeId || (assetTypes[0]?.id || 1),
    currentValue: asset?.currentValue || '',
    customReturn: asset?.customReturn || '',
    customVolatility: asset?.customVolatility || '',
  });

  const handleSubmit = (e) => {
    e.preventDefault();
    onSubmit({
      name: formData.name,
      typeId: parseInt(formData.typeId),
      currentValue: parseFloat(formData.currentValue) || 0,
      customReturn: formData.customReturn ? parseFloat(formData.customReturn) : null,
      customVolatility: formData.customVolatility ? parseFloat(formData.customVolatility) : null,
    });
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const selectedType = assetTypes.find(t => t.id === parseInt(formData.typeId));

  return (
    <div className="form-card">
      <form onSubmit={handleSubmit}>
        <div className="form-row">
          <div className="form-group">
            <label htmlFor="name">Asset Name</label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              placeholder="e.g., Vanguard 401k"
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="typeId">Asset Type</label>
            <select
              id="typeId"
              name="typeId"
              value={formData.typeId}
              onChange={handleChange}
            >
              {assetTypes.map(type => (
                <option key={type.id} value={type.id}>
                  {type.name} ({type.defaultReturn}% return)
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="currentValue">Current Value ($)</label>
            <input
              type="number"
              id="currentValue"
              name="currentValue"
              value={formData.currentValue}
              onChange={handleChange}
              placeholder="0.00"
              step="0.01"
              min="0"
              required
            />
          </div>
        </div>

        <div className="form-section">
          <h4>Custom Projections (Optional)</h4>
          <p className="form-hint">
            Leave blank to use defaults: {selectedType?.defaultReturn || 10}% return, {selectedType?.defaultVolatility || 15}% volatility
          </p>
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="customReturn">Expected Return (%)</label>
              <input
                type="number"
                id="customReturn"
                name="customReturn"
                value={formData.customReturn}
                onChange={handleChange}
                placeholder={selectedType?.defaultReturn || '10'}
                step="0.1"
              />
            </div>
            <div className="form-group">
              <label htmlFor="customVolatility">Volatility (%)</label>
              <input
                type="number"
                id="customVolatility"
                name="customVolatility"
                value={formData.customVolatility}
                onChange={handleChange}
                placeholder={selectedType?.defaultVolatility || '15'}
                step="0.1"
              />
            </div>
          </div>
        </div>

        <div className="form-actions">
          <button type="button" className="btn btn-secondary" onClick={onCancel}>
            Cancel
          </button>
          <button type="submit" className="btn btn-primary">
            {asset ? 'Update Asset' : 'Add Asset'}
          </button>
        </div>
      </form>
    </div>
  );
}

export default AssetForm;
