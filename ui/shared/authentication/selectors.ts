import { AppState } from 'store';

export const getIsAuthenticated = (state: AppState): boolean => state.authentication.isAuthenticated || false;

export const getSubscriptionIsActive = (state: AppState): boolean => state.authentication.isActive || false;

// getHasSubscription should not be used to determine whether or not the user's subscription is _active_. It is intended
// to be used to determine whether or not a subscription already exists for the user.
export const getHasSubscription = (state: AppState): boolean => state.authentication.hasSubscription || false;
