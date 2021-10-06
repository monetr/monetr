import { render } from "@testing-library/react";
import { TransactionDetailView } from "components/Transactions/TransactionDetail";
import Transaction from "data/Transaction";
import { Map } from 'immutable';
import moment from "moment";


describe('transaction detail view', () => {
  it('will render', () => {
    render(<TransactionDetailView
      transaction={ new Transaction({
        name: 'Dumb Stuff',
        date: moment(),
      }) }
      spending={ new Map() }
    />);

    // Make sure it's actually there.
    expect(document.querySelector('.transaction-detail')).not.toBeEmptyDOMElement();
  });

  it('will not render', () => {
    render(<TransactionDetailView
      transaction={ null }
      spending={ new Map() }
    />);

    // Make sure it's actually there.
    expect(document.querySelector('.transaction-detail')).toBeNull();
  })
});
