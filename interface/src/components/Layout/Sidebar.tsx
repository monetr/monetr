import { Fragment } from 'react';
import { useRoute } from 'wouter';

import BankSidebar from '@monetr/interface/components/Layout/BankSidebar';
import BudgetingSidebar from '@monetr/interface/components/Layout/BudgetingSidebar';

// The whole navigation panel: the bank rail plus a secondary column when the route has one. Both columns live here,
// rather than the secondary being rendered by BudgetLayout, so that on mobile they translate and trap focus as one
// drawer. Desktop is unaffected; the fragment leaves #root's flex children unchanged.
//
// A future secondary column for some other route gets rendered here with `data-sidebar-secondary` on it; the drawer
// derives its width and close behavior from that attribute.
export default function Sidebar(): React.JSX.Element {
  const [isBankRoute] = useRoute('/bank/:bankAccountId/*');

  return (
    <Fragment>
      <BankSidebar />
      {isBankRoute && <BudgetingSidebar />}
    </Fragment>
  );
}
