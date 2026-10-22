import { Fragment } from 'react';
import { useRoute } from 'wouter';

import BankSidebar from '@monetr/interface/components/Layout/BankSidebar';
import BudgetingSidebar from '@monetr/interface/components/Layout/BudgetingSidebar';

// Both sidebars render here instead of the budgeting one being in BudgetLayout. This way on mobile they move together
// as one drawer. A secondary sidebar for another route should go here too, with data-sidebar-secondary on it.
export default function Sidebar(): React.JSX.Element {
  const [isBankRoute] = useRoute('/bank/:bankAccountId/*');

  return (
    <Fragment>
      <BankSidebar />
      {isBankRoute && <BudgetingSidebar />}
    </Fragment>
  );
}
