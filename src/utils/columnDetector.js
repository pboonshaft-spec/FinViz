export function detectColumns(headers) {
  let dateCol = null;
  let amountCol = null;
  let debitCol = null;
  let creditCol = null;
  let categoryCol = null;
  let descriptionCol = null;

  headers.forEach(h => {
    const lower = h.toLowerCase().trim();

    if (lower.includes('date') || lower.includes('posted')) {
      dateCol = h;
    }
    if (lower === 'amount' || lower.includes('amount')) {
      amountCol = h;
    }
    if (lower.includes('debit') || lower.includes('withdrawal')) {
      debitCol = h;
    }
    if (lower.includes('credit') || lower.includes('deposit')) {
      creditCol = h;
    }
    if (lower.includes('category') || lower === 'type') {
      categoryCol = h;
    }
    if (lower.includes('description') || lower.includes('merchant') || lower.includes('name')) {
      descriptionCol = h;
    }
  });

  return {
    dateCol,
    amountCol,
    debitCol,
    creditCol,
    categoryCol,
    descriptionCol
  };
}
