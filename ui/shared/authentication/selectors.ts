import { AppState } from 'store';

export const getIsAuthenticated = (state: AppState): boolean => state.authentication.isAuthenticated || false;

export const getSubscriptionIsActive = (state: AppState): boolean => state.authentication.isActive || false;
