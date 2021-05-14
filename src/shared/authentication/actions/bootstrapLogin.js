import User from "data/User";
import request from "shared/util/request";
import { BOOTSTRAP_LOGIN } from "shared/authentication/actions";
import { getAPIUrl } from "shared/bootstrap/selectors";
import Cookies from 'js-cookie'
import { NewClient } from "api/api";

export default function bootstrapLogin(token = null, user = null) {
  return (dispatch, getState) => {
    // eslint-disable-next-line no-undef
    const conf = CONFIG;
    if (token) {
      // Trying to switch over to using cookies, but I don't want to break anything at the moment.
      if (conf.USE_LOCAL_STORAGE) {
        window.localStorage.setItem('M-Token', token);
      } else {
        Cookies.set('M-Token', token, {
          domain: conf.COOKIE_DOMAIN,
          secure: true,
        });
      }
    } else {
      if (conf.USE_LOCAL_STORAGE) {
        token = window.localStorage.getItem('M-Token');
      } else {
        token = Cookies.get('M-Token');
      }
    }

    // If the token is not present at this point then the user is not authenticated. We want to dispatch accordingly and
    // store in redux that the user is not authenticated.
    if (!token) {
      dispatch({
        type: BOOTSTRAP_LOGIN,
        payload: {
          isAuthenticated: false,
          isActive: false,
          token: null,
          user: null,
        }
      });
      return Promise.resolve();
    }

    const apiUrl = getAPIUrl(getState());

    // eslint-disable-next-line no-undef
    if (CONFIG.USE_LOCAL_STORAGE) {
      window.API = NewClient({
        baseURL: apiUrl,
        withCredentials: true,
        headers: {
          'M-Token': token,
        },
      });
    }


    if (!user) {
      // If we do have the token but we don't have the user info then we need to retrieve it using an API call to get
      // our user data from the API.
      return request().get('/users/me')
        .then(result => {
          dispatch({
            type: BOOTSTRAP_LOGIN,
            payload: {
              isAuthenticated: true,
              token: token,
              isActive: result.data.isActive,
              user: new User(result.data.user),
            }
          });

          return result;
        })
        .catch(error => {
          Cookies.remove('M-Token');
          window.localStorage.removeItem('M-Token');
          console.error(error);
        });
    }

    dispatch({
      type: BOOTSTRAP_LOGIN,
      payload: {
        isAuthenticated: true,
        token: token,
        isActive: result.data.isActive,
        user: new User(user),
      }
    });
    return Promise.resolve();
  };
}
