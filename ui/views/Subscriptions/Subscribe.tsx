import { getHasSubscription } from 'shared/authentication/selectors';
import request from 'shared/util/request';
import { useSelector } from 'react-redux';
import { getInitialPlan } from 'shared/bootstrap/selectors';
import useMountEffect from 'shared/util/useMountEffect';

export default function Subscribe(): JSX.Element {
  const initialPlan = useSelector(getInitialPlan);
  const hasSubscription = useSelector(getHasSubscription);

  useMountEffect(() => {
    if (initialPlan && !hasSubscription) {
      request().post(`/billing/create_checkout`, {
        priceId: '',
        cancelPath: '/logout',
      })
        .then(result => window.location.assign(result.data.url))
        .catch(error => alert(error));
    } else if (hasSubscription) {
      // If the customer has a subscription then we want to just manage it. This will allow a customer to fix a
      // subscription for a card that has failed payment or something similar.
      request().get('/billing/portal')
        .then(result => window.location.assign(result.data.url))
        .catch(error => alert(error));
    }
  });

  return null;
}


