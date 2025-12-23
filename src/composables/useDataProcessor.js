import { detectColumns } from '../utils/columnDetector.js';
import { aggregateByMonth, aggregateByDay, aggregateByCategory, calculateTotals } from '../utils/dataAggregator.js';

export function useDataProcessor() {
  const logs = [];

  function log(message) {
    logs.push(message);
  }

  function processData(filesData) {
    log('Starting data processing...');

    const allData = filesData.flatMap(f => f.data);
    log(`Total rows across all files: ${allData.length}`);

    if (allData.length === 0) {
      return { data: null, logs };
    }

    const sample = allData[0];
    const headers = Object.keys(sample);

    log('Detecting column types...');
    const columns = detectColumns(headers);

    if (columns.dateCol) log(`Date column: "${columns.dateCol}"`);
    if (columns.amountCol) log(`Amount column: "${columns.amountCol}"`);
    if (columns.debitCol) log(`Debit column: "${columns.debitCol}"`);
    if (columns.creditCol) log(`Credit column: "${columns.creditCol}"`);
    if (columns.categoryCol) log(`Category column: "${columns.categoryCol}"`);
    if (columns.descriptionCol) log(`Description column: "${columns.descriptionCol}"`);

    const transactions = [];

    // Process each row
    allData.forEach((row, idx) => {
      // Get date
      let date = null;
      if (columns.dateCol && row[columns.dateCol]) {
        date = new Date(row[columns.dateCol]);
        if (isNaN(date)) {
          // Try different date formats
          const dateStr = row[columns.dateCol].toString();
          date = new Date(dateStr.replace(/(\d{2})\/(\d{2})\/(\d{4})/, '$3-$1-$2'));
        }
      }

      if (!date || isNaN(date)) {
        if (idx < 5) log(`⚠ Row ${idx}: Invalid date "${row[columns.dateCol]}"`);
        return;
      }

      // Get amount - check for separate debit/credit columns first
      let amount = 0;
      let isExpense = false;

      if (columns.debitCol && columns.creditCol) {
        // Separate debit/credit columns
        const debitVal = parseFloat(String(row[columns.debitCol] || '0').replace(/[$,]/g, ''));
        const creditVal = parseFloat(String(row[columns.creditCol] || '0').replace(/[$,]/g, ''));

        if (!isNaN(debitVal) && debitVal !== 0) {
          amount = -Math.abs(debitVal); // Debits are expenses (negative)
          isExpense = true;
        } else if (!isNaN(creditVal) && creditVal !== 0) {
          amount = Math.abs(creditVal); // Credits are income (positive)
          isExpense = false;
        }
      } else if (columns.amountCol) {
        // Single amount column
        let rawAmount = String(row[columns.amountCol] || '0').replace(/[$,]/g, '');
        amount = parseFloat(rawAmount);

        if (isNaN(amount)) {
          if (idx < 5) log(`⚠ Row ${idx}: Invalid amount "${row[columns.amountCol]}"`);
          return;
        }

        // Determine if expense or income
        if (amount < 0) {
          isExpense = true;
        } else if (amount > 0) {
          // Check if this should be classified as expense based on category
          const category = (row[columns.categoryCol] || '').toLowerCase();
          const description = (row[columns.descriptionCol] || '').toLowerCase();
          const combined = category + ' ' + description;

          // Common expense keywords
          const expenseKeywords = ['expense', 'payment', 'purchase', 'debit', 'withdrawal',
                                   'grocery', 'restaurant', 'gas', 'utility', 'rent',
                                   'subscription', 'shopping', 'amazon', 'store'];
          const incomeKeywords = ['income', 'salary', 'deposit', 'credit', 'paycheck',
                                 'payment received', 'transfer from', 'refund'];

          const hasExpenseKeyword = expenseKeywords.some(kw => combined.includes(kw));
          const hasIncomeKeyword = incomeKeywords.some(kw => combined.includes(kw));

          if (hasExpenseKeyword && !hasIncomeKeyword) {
            isExpense = true;
            amount = -Math.abs(amount); // Make it negative
          } else if (hasIncomeKeyword) {
            isExpense = false;
            amount = Math.abs(amount); // Keep positive
          }
        }
      }

      if (amount === 0) return;

      // Get category
      let category = row[columns.categoryCol] || row[columns.descriptionCol] || 'Uncategorized';
      category = String(category).trim();

      const month = date.toLocaleDateString('en-US', { year: 'numeric', month: 'short' });
      const day = date.toISOString().split('T')[0];

      transactions.push({
        date: date,
        amount: amount,
        category: category,
        month: month,
        day: day,
        isExpense: isExpense || amount < 0
      });
    });

    // Sort transactions by date
    transactions.sort((a, b) => a.date - b.date);

    log(`Processed transactions: ${transactions.length}`);

    // Aggregate data
    const monthlyData = aggregateByMonth(transactions);
    const dailyData = aggregateByDay(transactions);
    const categories = aggregateByCategory(transactions);
    const totals = calculateTotals(transactions);

    log(`Total Income: $${totals.income.toFixed(2)}`);
    log(`Total Expenses: $${totals.expenses.toFixed(2)}`);
    log(`Net Balance: $${totals.balance.toFixed(2)}`);

    return {
      data: {
        transactions,
        monthlyData,
        dailyData,
        categories,
        totals
      },
      logs
    };
  }

  return {
    processData
  };
}
