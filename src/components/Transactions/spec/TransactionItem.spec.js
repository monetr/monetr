import { render } from '@testing-library/react'
import { TransactionItem } from "components/Transactions/TransactionItem";
import Expense from "data/Expense";
import Transaction from "data/Transaction";
import moment from "moment";


describe('transaction item', () => {
  it('will render', () => {
    render(<TransactionItem
      transactionId={ 123 }
      transaction={ new Transaction({
        name: 'Dumb Stuff',
        date: moment(),
      }) }
    />);

    // Make sure it's actually there.
    expect(document.querySelector('.transactions-item')).not.toBeEmptyDOMElement();
    expect(document.querySelector('.transaction-item-name').textContent).toBe('Dumb Stuff');
  });

  it('will render with expense', () => {
    render(<TransactionItem
      transactionId={ 123 }
      transaction={ new Transaction({
        name: 'Dumb Stuff',
        date: moment(),
      }) }
      expense={ new Expense({
        name: 'Dumb Stuff Budget'
      }) }
    />);

    expect(document.querySelector('.transaction-expense-name').textContent).toBe('Spent From Dumb Stuff Budget');
  });

  it('will render a deposit', () => {
    render(<TransactionItem
      transactionId={ 123 }
      transaction={ new Transaction({
        name: 'Dumb Stuff',
        date: moment(),
        amount: -100, // $1.00
      }) }
    />);

    expect(document.querySelector('.transaction-expense-name').textContent).toBe('Deposited Into Safe-To-Spend');
    expect(document.querySelector('.amount').textContent).toBe('+ $1.00');
  });
});
