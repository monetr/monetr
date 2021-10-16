import { render } from '@testing-library/react';
import { createLocation, createMemoryHistory, MemoryHistory } from 'history';
import mockAxios from 'jest-mock-axios';
import React from 'react';
import { match } from 'react-router';
import { Logout } from 'views/Authentication/Logout';

describe('Logout', () => {
  const path = `/verify/email`;

  let history: MemoryHistory;

  beforeEach(() => {
    history = createMemoryHistory();
    window.alert = jest.fn();
    history.push = jest.fn();
  });

  it('will logout', () => {
    const newMatch: match<object> = {
      isExact: false,
      path,
      url: path,
      params: {},
    };
    const location = createLocation(newMatch.url);

    const logout = jest.fn();

    const verifyEmail = render(<Logout
      history={ history }
      location={ location }
      match={ newMatch }
      logout={ logout }
    />);

    expect(verifyEmail.container).toBeEmptyDOMElement();
    expect(logout).toHaveBeenCalled();
    expect(history.push).toHaveBeenCalledWith('/login');
  });
});