import React, { Component, Fragment } from "react";
import { connect } from "react-redux";
import { Link as RouterLink, RouteComponentProps, withRouter } from "react-router-dom";

import User from "data/User";
import bootstrapLogin from "shared/authentication/actions/bootstrapLogin";
import request from "shared/util/request";
import { getReCAPTCHAKey, getShouldVerifyLogin, getSignUpAllowed } from "shared/bootstrap/selectors";

import ReCAPTCHA from "react-google-recaptcha";
import classnames from "classnames";
import { Alert, AlertTitle } from "@material-ui/lab";
import { Button, CircularProgress, Snackbar, TextField } from "@material-ui/core";
import { Formik, FormikHelpers } from "formik";

import Logo from 'assets';


interface LoginValues {
  email: string | null;
  password: string | null;
}

interface State {
  error: string | null;
  loading: boolean;
  verification: string | null;
}

interface WithConnectionPropTypes extends RouteComponentProps {
  ReCAPTCHAKey: string | null;
  bootstrapLogin: (token: string, user: User) => Promise<void>;
  verifyLogin: boolean;
}

class LoginView extends Component<WithConnectionPropTypes, State> {

  state = {
    error: null,
    loading: false,
    verification: null,
  };

  renderErrorMaybe = () => {
    const { error } = this.state;
    if (!error) {
      return null;
    }

    return (
      <Snackbar open autoHideDuration={ 10000 }>
        <Alert variant="filled" severity="error">
          <AlertTitle>Error</AlertTitle>
          { this.state.error }
        </Alert>
      </Snackbar>
    );
  };

  validateInput = (values: LoginValues): Partial<LoginValues> | null => {
    let errors: Partial<LoginValues> = {};

    if (values.email) {
      const re = /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
      if (!re.test(values.email.toLowerCase())) {
        errors['email'] = 'Please provide a valid email address.';
      }
    }

    if (values.password) {
      if (values.password.length < 8) {
        errors['password'] = 'Password must be at least 8 characters long.'
      }
    }

    return errors;
  };

  submit = (values: LoginValues, helpers: FormikHelpers<LoginValues>): Promise<void> => {
    helpers.setSubmitting(true);
    this.setState({
      error: null,
      loading: true,
    });

    return request()
      .post('/authentication/login', {
        captcha: this.state.verification,
        email: values.email,
        password: values.password,
      })
      .then(result => {
        return this.props.bootstrapLogin(result.data.token, result.data.user)
          .then(() => {
            if (result.data.nextUrl) {
              this.props.history.push(result.data.nextUrl);
              return
            }

            this.props.history.push('/');
          });
      })
      .catch(error => {
        if (error?.response?.data?.error) {
          return this.setState({
            error: error.response.data.error,
            loading: false,
          });
        }

        throw error;
      })
      .finally(() =>{
        helpers.setSubmitting(false);
      });
  };

  renderCaptchaMaybe = () => {
    const { verifyLogin, ReCAPTCHAKey } = this.props;
    if (!verifyLogin) {
      return null;
    }

    return (
      <div className="flex items-center justify-center w-full">
        { !this.state.loading && <ReCAPTCHA
          sitekey={ ReCAPTCHAKey }
          onChange={ value => this.setState({ verification: value }) }
        /> }
        { this.state.loading && <CircularProgress/> }
      </div>
    )
  };

  render() {

    const initialValues: LoginValues = {
      email: null,
      password: null,
    }

    const disableForVerification = !this.props.verifyLogin || (this.props.ReCAPTCHAKey && this.state.verification);

    return (
      <Fragment>
        { this.renderErrorMaybe() }
        <Formik
          initialValues={ initialValues }
          validate={ this.validateInput }
          onSubmit={ this.submit }
        >
          { ({
               values,
               errors,
               touched,
               handleChange,
               handleBlur,
               handleSubmit,
               isSubmitting,
               submitForm,
             }) => (
            <form onSubmit={ handleSubmit } className="h-full overflow-y-auto">
              <div className="flex items-center justify-center w-full h-full max-h-full">
                <div className="w-full p-10 xl:w-3/12 lg:w-5/12 md:w-2/3 sm:w-10/12 max-w-screen-sm sm:p-0">
                  <div className="flex justify-center w-full mb-5">
                    <img src={ Logo } className="w-1/3"/>
                  </div>
                  <div className="w-full pb-2.5">
                    <Button
                      className="w-full"
                      color="secondary"
                      component={ RouterLink }
                      disabled={ isSubmitting }
                      to="/register"
                      variant="contained"
                    >
                      Sign Up For monetr
                    </Button>
                  </div>
                  <div className="w-full opacity-50 pb-2.5">
                    <div className="relative w-full border-t border-gray-400 top-5"/>
                    <div className="relative flex justify-center inline w-full">
                      <span className="relative bg-white p-1.5">
                        or sign in with your email
                      </span>
                    </div>
                  </div>
                  <div className="w-full">
                    <div className="w-full pb-2.5">
                      <TextField
                        autoFocus
                        className="w-full"
                        disabled={ isSubmitting }
                        error={ touched.email && !!errors.email }
                        helperText={ (touched.email && errors.email) ? errors.email : null }
                        id="login-email"
                        label="Email"
                        name="email"
                        onBlur={ handleBlur }
                        onChange={ handleChange }
                        value={ values.email }
                        variant="outlined"
                        autoComplete="username"
                      />
                    </div>
                    <div className="w-full pt-2.5 pb-2.5">
                      <TextField
                        className="w-full"
                        disabled={ isSubmitting }
                        error={ touched.password && !!errors.password }
                        helperText={ (touched.password && errors.password) ? errors.password : null }
                        id="login-password"
                        label="Password"
                        name="password"
                        onBlur={ handleBlur }
                        onChange={ handleChange }
                        type="password"
                        value={ values.password }
                        variant="outlined"
                        autoComplete="current-password"
                      />
                    </div>
                  </div>
                  { this.renderCaptchaMaybe() }
                  <div className="w-full pt-2.5 mb-10">
                    <Button
                      className="w-full"
                      color="primary"
                      disabled={ isSubmitting || (!values.password || !values.email || !disableForVerification) }
                      onClick={ submitForm }
                      type="submit"
                      variant="contained"
                    >
                      { isSubmitting && <CircularProgress
                        className={ classnames('mr-2', {
                          'opacity-50': isSubmitting,
                        }) }
                        size="1em"
                        thickness={ 5 }
                      /> }
                      { isSubmitting ? 'Signing In...' : 'Sign In' }
                    </Button>
                  </div>
                </div>
              </div>
            </form>
          ) }
        </Formik>
      </Fragment>
    )
  }
}

export default connect(
  state => ({
    ReCAPTCHAKey: getReCAPTCHAKey(state),
    allowSignUp: getSignUpAllowed(state),
    verifyLogin: getShouldVerifyLogin(state),
  }),
  {
    bootstrapLogin,
  }
)(withRouter(LoginView));
