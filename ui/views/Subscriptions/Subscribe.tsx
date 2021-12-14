import request from 'shared/util/request';
import React from 'react';
import { useSelector } from 'react-redux';
import { getInitialPlan } from 'shared/bootstrap/selectors';
import useMountEffect from 'shared/util/useMountEffect';

export default function Subscribe(): JSX.Element {
  const initialPlan = useSelector(getInitialPlan);

  useMountEffect(() => {
    if (initialPlan) {
      request().post(`/billing/create_checkout`, {
        priceId: '',
        cancelPath: '/logout',
      })
        .then(result => window.location.assign(result.data.url))
        .catch(error => alert(error));
    }
  });

  return null;
}


