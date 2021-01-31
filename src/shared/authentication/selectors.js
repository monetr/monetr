import { createSelector } from 'reselect'


export const getIsAuthenticated = state => state.authentication.isAuthenticated;
