export function aggregateByMonth(transactions) {
  const monthlyData = {};

  transactions.forEach(transaction => {
    const month = transaction.month;

    if (!monthlyData[month]) {
      monthlyData[month] = { income: 0, expenses: 0, net: 0 };
    }

    if (transaction.amount > 0) {
      monthlyData[month].income += transaction.amount;
    } else {
      monthlyData[month].expenses += Math.abs(transaction.amount);
    }
    monthlyData[month].net += transaction.amount;
  });

  return monthlyData;
}

export function aggregateByDay(transactions) {
  const dailyData = {};

  transactions.forEach(transaction => {
    const day = transaction.day;
    dailyData[day] = (dailyData[day] || 0) + transaction.amount;
  });

  return dailyData;
}

export function aggregateByCategory(transactions) {
  const categories = {};

  transactions.forEach(transaction => {
    // Only count expenses in categories
    if (transaction.amount < 0) {
      const category = transaction.category;
      categories[category] = (categories[category] || 0) + Math.abs(transaction.amount);
    }
  });

  return categories;
}

export function calculateTotals(transactions) {
  const totals = {
    income: 0,
    expenses: 0,
    balance: 0
  };

  transactions.forEach(transaction => {
    if (transaction.amount > 0) {
      totals.income += transaction.amount;
    } else {
      totals.expenses += Math.abs(transaction.amount);
    }
  });

  totals.balance = totals.income - totals.expenses;

  return totals;
}
