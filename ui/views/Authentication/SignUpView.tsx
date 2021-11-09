import React, { Component, Fragment } from 'react';
import { RouteComponentProps, withRouter } from 'react-router-dom';
import { connect } from 'react-redux';
import bootstrapLogin from 'shared/authentication/actions/bootstrapLogin';
import request from 'shared/util/request';
import User from 'models/User';
import {
  getInitialPlan,
  getReCAPTCHAKey,
  getRequireBetaCode,
  getShouldVerifyRegister,
  getStripePublicKey
} from 'shared/bootstrap/selectors';
import ReCAPTCHA from 'react-google-recaptcha';
import classnames from 'classnames';
import {
  Alert,
  AlertTitle,
  Button,
  Checkbox,
  CircularProgress,
  FormControl,
  FormControlLabel,
  FormGroup,
  Snackbar,
  TextField
} from '@mui/material';
import { Formik, FormikHelpers } from 'formik';
import { AppState } from 'store';
import verifyEmailAddress from 'util/verifyEmailAddress';
import AfterEmailVerificationSent from 'views/Authentication/AfterEmailVerificationSent';

import Logo from 'assets';

interface SignUpValues {
  agree: boolean;
  betaCode: string;
  email: string;
  firstName: string;
  lastName: string;
  password: string;
  verifyPassword: string;
}

interface State {
  error: string | null;
  message: string | null;
  loading: boolean;
  verification: string | null;
  successful: boolean;
}

interface WithConnectionPropTypes extends RouteComponentProps {
  ReCAPTCHAKey: string | null;
  bootstrapLogin: (token: string, user: User, subscriptionIsActive: boolean) => Promise<void>;
  initialPlan: { price: number, freeTrialDays: number } | null;
  requireBetaCode: boolean;
  stripePublicKey: string | null;
  verifyRegister: boolean;
}

class SignUpView extends Component<WithConnectionPropTypes, State> {

  state = {
    verification: null,
    loading: false,
    error: null,
    message: null,
    successful: false,
  };

  renderErrorMaybe = (): React.ReactNode | null => {
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

  renderMessageMaybe = (): React.ReactNode | null => {
    const { message } = this.state;
    if (!message) {
      return null;
    }

    return (
      <Snackbar open autoHideDuration={ 10000 } onClose={ () => this.setState({ message: null }) }>
        <Alert variant="filled" severity="info">
          { this.state.message }
        </Alert>
      </Snackbar>
    );
  }

  validateInput = (values: SignUpValues): Partial<SignUpValues> => {
    let errors: Partial<SignUpValues> = {};

    if (values.email) {
      if (!verifyEmailAddress(values.email)) {
        errors['email'] = 'Please provide a valid email address.';
      }
    }

    if (values.password) {
      if (values.password.length < 8) {
        errors['password'] = 'Password must be at least 8 characters long.';
      }
    }

    if (values.verifyPassword && values.password !== values.verifyPassword) {
      errors['verifyPassword'] = 'Passwords must match.';
    }

    if (values.firstName && values.firstName.length === 0) {
      errors['firstName'] = 'First name is required.';
    }

    if (values.lastName && values.lastName.length === 0) {
      errors['lastName'] = 'Last name is required.';
    }

    if (this.props.requireBetaCode && !values.betaCode) {
      errors['betaCode'] = 'Beta code is required.';
    }

    return errors;
  };

  submit = (values: SignUpValues, { setSubmitting }: FormikHelpers<SignUpValues>): Promise<any> => {
    this.setState({
      error: null,
      loading: true,
    });

    const { verification } = this.state;
    const { bootstrapLogin } = this.props;

    return request()
      .post('/authentication/register', {
        email: values.email,
        password: values.password,
        firstName: values.firstName,
        lastName: values.lastName,
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
        captcha: verification,
        betaCode: values.betaCode,
        agree: values.agree,
      })
      .then(result => {
        if (result.data.token) {
          return bootstrapLogin(result.data.token, result.data.user, result.data.isActive)
            .then((): Promise<any> => {
              if (!result) {
                this.props.history.push('/');
                return Promise.resolve();
              }

              if (result.data.nextUrl) {
                console.log(`going to ${ result.data.nextUrl }`);
                this.props.history.push(result.data.nextUrl);
                return Promise.resolve();
              }
              this.props.history.push('/');
              return Promise.resolve();
            });
        }

        if (result.data.requireVerification) {
          this.setState({
            successful: true,
          });
        }

        if (result.data.message) {
          this.setState({
            message: result.data.message,
          });
        }

        return Promise.resolve();
      })
      .catch(error => {
        if (error?.response?.data?.error) {
          return this.setState({
            error: error.response.data.error,
          });
        }

        throw error;
      })
      .finally(() => {
        // Clear the submitting state of the form when we are done.
        setSubmitting(false);

        return this.setState({
          verification: null,
          loading: false,
        });
      });
  };

  renderCaptchaMaybe = (): React.ReactNode => {
    const { verifyRegister, ReCAPTCHAKey } = this.props;

    if (!verifyRegister) {
      return null;
    }

    return (
      <div className="w-full flex justify-center items-center pt-1.5 pb-1.5">
        { !this.state.loading && <ReCAPTCHA
          sitekey={ ReCAPTCHAKey }
          onChange={ value => this.setState({ verification: value }) }
        /> }
        { this.state.loading && <CircularProgress/> }
      </div>
    )
  };

  cannotSubmit = (values: SignUpValues): boolean => {
    const { verifyRegister } = this.props;
    const { verification } = this.state;

    const verified = !verifyRegister || verification;
    return !(verified &&
      values.email &&
      values.password &&
      values.firstName &&
      values.lastName &&
      values.agree
    );
  }

  renderSignUpText = (isSubmitting: boolean): React.ReactNode | string => {
    if (isSubmitting) {
      return 'Signing up...'
    }

    const { initialPlan } = this.props;

    if (initialPlan) {
      const suffix = initialPlan.freeTrialDays > 0 ? <span>(Free for { initialPlan.freeTrialDays } days)</span> : '';

      return (
        <Fragment>
          <span className="mr-1">Sign Up For</span>
          <span className="mr-1"> ${ (initialPlan.price / 100).toFixed(2) }</span>
          { suffix }
        </Fragment>
      );
    }

    return 'Sign Up'
  };

  render() {
    const { successful } = this.state;

    if (successful) {
      return <AfterEmailVerificationSent/>;
    }

    const initialValues: SignUpValues = {
      agree: false,
      betaCode: '',
      email: '',
      firstName: '',
      lastName: '',
      password: '',
      verifyPassword: '',
    };

    return (
      <Fragment>
        { this.renderMessageMaybe() }
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
              <div className="flex justify-center w-full h-full max-h-full">
                <div className="w-full p-10 max-w-screen-sm sm:p-0">
                  <div className="flex justify-center w-full mt-5 mb-5">
                    <img src={ Logo } className="w-1/4"/>
                  </div>
                  <div className="w-full">
                    <div className="w-full pb-1.5 pt-1.5">
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
                    <div className="w-full pb-1.5 pt-1.5 grid grid-flow-row gap-2 sm:grid-flow-col">
                      <TextField
                        className="w-full"
                        disabled={ isSubmitting }
                        error={ touched.firstName && !!errors.firstName }
                        helperText={ (touched.firstName && errors.firstName) ? errors.firstName : null }
                        id="login-firstName"
                        label="First Name"
                        name="firstName"
                        onBlur={ handleBlur }
                        onChange={ handleChange }
                        value={ values.firstName }
                        variant="outlined"
                      />
                      <TextField
                        className="w-full"
                        disabled={ isSubmitting }
                        error={ touched.lastName && !!errors.lastName }
                        helperText={ (touched.lastName && errors.lastName) ? errors.lastName : null }
                        id="login-lastName"
                        label="Last Name"
                        name="lastName"
                        onBlur={ handleBlur }
                        onChange={ handleChange }
                        value={ values.lastName }
                        variant="outlined"
                      />
                    </div>
                    <div className="w-full pb-1.5 pt-1.5 grid grid-flow-row gap-2 sm:grid-flow-col">
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
                        autoComplete="new-password"
                      />
                      <TextField
                        className="w-full"
                        disabled={ isSubmitting }
                        error={ touched.verifyPassword && !!errors.verifyPassword }
                        helperText={ (touched.verifyPassword && errors.verifyPassword) ? errors.verifyPassword : null }
                        id="login-verifyPassword"
                        label="Verify Password"
                        name="verifyPassword"
                        onBlur={ handleBlur }
                        onChange={ handleChange }
                        type="password"
                        value={ values.verifyPassword }
                        variant="outlined"
                        autoComplete="new-password"
                      />
                    </div>
                    { this.props.requireBetaCode &&
                    <div className="w-full pt-1.5 pb-1.5">
                      <TextField
                        className="w-full"
                        disabled={ isSubmitting }
                        error={ touched.betaCode && !!errors.betaCode }
                        helperText={ (touched.betaCode && errors.betaCode) ? errors.betaCode : null }
                        id="login-betaCode"
                        label="Beta Code"
                        name="betaCode"
                        onBlur={ handleBlur }
                        onChange={ handleChange }
                        type="betaCode"
                        value={ values.betaCode }
                        variant="outlined"
                      />
                    </div>
                    }
                  </div>
                  { this.renderCaptchaMaybe() }
                  <div className="w-full flex justify-center items-center pt-1.5 pb-1">
                    <FormControl component="fieldset">
                      <FormGroup aria-label="position" row>
                        <FormControlLabel
                          control={
                            <Checkbox
                              color="primary"
                              name="agree"
                              onChange={ handleChange }
                            />
                          }
                          label="I agree to stuff and things"
                          labelPlacement="end"
                          value="end"
                        />
                      </FormGroup>
                    </FormControl>
                  </div>
                  <div className="w-full pt-1.5 flex justify-center pb-10">
                    <Button
                      className="w-1/2 mr-1 min-w-max"
                      color="secondary"
                      disabled={ isSubmitting }
                      onClick={ () => this.props.history.push('/login') }
                      variant="outlined"
                    >
                      Cancel
                    </Button>
                    <Button
                      className="w-1/2 ml-1 min-w-max"
                      color="primary"
                      disabled={ isSubmitting || this.cannotSubmit(values) }
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
                      { this.renderSignUpText(isSubmitting) }
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
  (state: AppState) => ({
    ReCAPTCHAKey: getReCAPTCHAKey(state),
    initialPlan: getInitialPlan(state),
    requireBetaCode: getRequireBetaCode(state),
    stripePublicKey: getStripePublicKey(state),
    verifyRegister: getShouldVerifyRegister(state),
  }),
  {
    bootstrapLogin,
  }
)(withRouter(SignUpView));
