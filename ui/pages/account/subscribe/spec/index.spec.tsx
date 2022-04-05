import SubscribePage from 'pages/account/subscribe/index';
import React from 'react';
import { Bootstrap } from 'shared/bootstrap/actions';
import { configureStore } from 'store';
import testRenderer from 'testutils/renderer';
import { screen } from '@testing-library/react';
import mockAxios from 'jest-mock-axios';
import { cleanupWindowLocationMock, setupWindowLocationMock } from 'testutils/window';
import { Login } from 'shared/authentication/actions';

describe('/accounts/subscribe', () => {
  beforeEach(() => setupWindowLocationMock());
  afterEach(() => cleanupWindowLocationMock());

  it('will render', () => {
    testRenderer(<SubscribePage/>);
    expect(screen.getByText('Getting Stripe ready...')).not.toBeEmptyDOMElement();

    // When we have nothing in state, no requests should be made because we don't know what the plan is or whether or
    // not the current user already has a subscription.
    expect(mockAxios.post).not.toHaveBeenCalled();
    expect(mockAxios.get).not.toHaveBeenCalled();
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it('will request checkout session', () => {
    const store = configureStore();
    store.dispatch({
      type: Bootstrap.Success,
      payload: {
        initialPlan: {
          price: 499,
          freeTrialDays: 30,
        },
        billingEnabled: true,
      }
    });

    testRenderer(<SubscribePage/>, {
      store,
    });
    expect(mockAxios.post).toHaveBeenCalledWith('/billing/create_checkout', {
      cancelPath: '/logout',
      priceId: '',
    });
    let responseObj = {
      data: {
        url: 'https://stripe.com/your/checkout/session',
      }
    };
    mockAxios.mockResponse(responseObj);
    expect(window.location.assign).toHaveBeenCalledWith('https://stripe.com/your/checkout/session');
  });

  it('will fail to create checkout session', () => {
    const store = configureStore();
    store.dispatch({
      type: Bootstrap.Success,
      payload: {
        initialPlan: {
          price: 499,
          freeTrialDays: 30,
        },
        billingEnabled: true,
      }
    });

    testRenderer(<SubscribePage/>, {
      store,
    });
    expect(mockAxios.post).toHaveBeenCalledWith('/billing/create_checkout', {
      cancelPath: '/logout',
      priceId: '',
    });
    mockAxios.mockError({
      response: {
        status: 500,
        data: {
          error: 'Something went very wrong...',
        }
      }
    });
    expect(screen.getByText('Something went very wrong...')).toBeInTheDocument();
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it('will request billing portal if there is a subscription', () => {
    const store = configureStore();
    store.dispatch({
      type: Login.Success,
      payload: {
        user: {
          accountId: 1234,
        },
        hasSubscription: true,
      }
    });

    testRenderer(<SubscribePage/>, {
      store,
    });
    expect(mockAxios.get).toHaveBeenCalledWith('/billing/portal');
    let responseObj = {
      data: {
        url: 'https://stripe.com/your/checkout/session',
      }
    };
    mockAxios.mockResponse(responseObj);
    expect(window.location.assign).toHaveBeenCalledWith('https://stripe.com/your/checkout/session');
  });

  it('will fail to create billing portal', () => {
    const store = configureStore();
    store.dispatch({
      type: Login.Success,
      payload: {
        user: {
          accountId: 1234,
        },
        hasSubscription: true,
      }
    });

    testRenderer(<SubscribePage/>, {
      store,
    });
    expect(mockAxios.get).toHaveBeenCalledWith('/billing/portal');
    mockAxios.mockError({
      response: {
        status: 500,
        data: {
          error: 'Something has gone the not right way...',
        }
      }
    });
    expect(screen.getByText('Something has gone the not right way...')).toBeInTheDocument();
    expect(window.location.assign).not.toHaveBeenCalled();
  });
});
