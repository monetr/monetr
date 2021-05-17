import {
  Box,
  Button,
  Card,
  CardActions,
  CardContent,
  CardHeader,
  CircularProgress,
  Container,
  Grid,
  Grow,
  Snackbar,
  TextField
} from "@material-ui/core";
import { Alert, AlertTitle } from "@material-ui/lab";
import { Formik } from "formik";
import PropTypes from "prop-types";
import React, { Component } from "react";
import ReCAPTCHA from "react-google-recaptcha";
import { connect } from "react-redux";
import { Link as RouterLink, withRouter } from "react-router-dom";
import { bindActionCreators } from "redux";
import bootstrapLogin from "shared/authentication/actions/bootstrapLogin";
import {
  getReCAPTCHAKey,
  getRequireBetaCode,
  getRequireLegalName,
  getRequirePhoneNumber,
  getShouldVerifyRegister,
  getSignUpAllowed
} from "shared/bootstrap/selectors";
import request from "shared/util/request";
import * as Sentry from "@sentry/react";

export class SignUpView extends Component {
  state = {
    verification: null,
    error: null,
    loading: false,
  };

  static propTypes = {
    ReCAPTCHAKey: PropTypes.string,
    bootstrapLogin: PropTypes.func.isRequired,
    history: PropTypes.shape({
      push: PropTypes.func.isRequired,
    }).isRequired,
    requireLegalName: PropTypes.bool.isRequired,
    requirePhoneNumber: PropTypes.bool.isRequired,
    setToken: PropTypes.func.isRequired,
    verifyRegister: PropTypes.bool.isRequired,
    requireBetaCode: PropTypes.bool.isRequired,
  }

  submitRegister = values => {
    this.setState({
      error: null,
      loading: true,
    });
    const { verification } = this.state;
    const { bootstrapLogin } = this.props;
    return request().post('/authentication/register', {
      email: values.email,
      password: values.password,
      firstName: values.firstName,
      lastName: values.lastName,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      captcha: verification,
      betaCode: values.betaCode,
    })
      .then(result => {
        if (result.data.token) {
          return bootstrapLogin(result.data.token, result.data.user, result.data.isActive)
            .then(() => {
              if (result && result.data.nextUrl) {
                console.log(`going to ${ result.data.nextUrl }`);
                this.props.history.push(result.data.nextUrl);
                return
              }

              this.props.history.push('/');
            });
        }
      })
      .catch(error => {
        Sentry.captureException(error);
        if (error.response.data.error) {
          this.setState({
            error: error.response.data.error,
            loading: false,
          });
          return error;
        }

        this.setState({
          error: 'Failed to sign up.',
          loading: false,
        });
      });
  };

  renderReCAPTCHAMaybe = () => {
    const { verifyRegister, ReCAPTCHAKey } = this.props;
    if (!verifyRegister) {
      return null;
    }

    return (
      <Grid item xs={ 12 }>
        <div className="w-full flex justify-center items-center">
          { this.state.loading && <CircularProgress/> }
          { !this.state.loading && <ReCAPTCHA
            sitekey={ ReCAPTCHAKey }
            onChange={ value => this.setState({ verification: value }) }
          /> }
        </div>
      </Grid>
    )
  };

  cannotSubmit = (isSubmitting, values) => {
    const { verifyRegister } = this.props;
    const { verification } = this.state;
    return isSubmitting || !values.email || !values.password || !values.confirmPassword || !values.firstName || (verifyRegister && !verification)
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
    )
  };

  onValidate = (values) => {
    const { requireLegalName } = this.props;
    const errors = {};
    if (!values.email) {
      errors.email = 'An email is required to sign up.';
    }

    if (!values.password || values.password.length < 8) {
      errors.password = 'Your password must be at least 8 characters.';
    }

    if (values.password !== values.confirmPassword) {
      errors.confirmPassword = 'Passwords must match.';
    }

    if (!values.firstName || values.firstName.length === 0) {
      errors.firstName = 'First name is required.';
    }

    if (requireLegalName && (!values.lastName || values.lastName.length === 0)) {
      errors.lastName = 'Last name is required.';
    }

    if (this.props.requireBetaCode && !values.betaCode) {
      errors.betaCode = 'Beta Code is required.'
    }

    return errors;
  };

  onSubmit = (values, { setSubmitting }) => {
    return this.submitRegister(values)
      .finally(() => setSubmitting(false));
  };

  render() {
    return (
      <div className="register-view">
        { this.renderErrorMaybe() }
        <Formik
          initialValues={ {
            email: '',
            password: '',
            confirmPassword: '',
            firstName: '',
            lastName: '',
            betaCode: '',
          } }
          validate={ this.onValidate }
          onSubmit={ this.onSubmit }
        >
          { ({
               values,
               errors,
               touched,
               handleChange,
               handleBlur,
               handleSubmit,
               isSubmitting,
               /* and other goodies */
             }) => (
            <Box m={ 6 }>
              <Container maxWidth="xs" className={ "login-root" }>
                <Grow in>
                  <Card>
                    <CardHeader title="monetr" subheader="Sign Up"/>
                    <CardContent>
                      <Grid container spacing={ 1 }>
                        <Grid item xs={ 12 }>
                          <TextField
                            autoFocus={ true }
                            fullWidth
                            id="email"
                            label="Email"
                            name="email"
                            value={ values.email }
                            onChange={ handleChange }
                            onBlur={ handleBlur }
                            error={ values.email && touched.email && !!errors.email }
                            helperText={ values.email && touched.email && errors.email }
                            disabled={ isSubmitting }
                          />
                        </Grid>
                        <Grid item xs={ 12 }>
                          <TextField
                            fullWidth
                            id="password"
                            label="Password"
                            name="password"
                            type="password"
                            value={ values.password }
                            onChange={ handleChange }
                            onBlur={ handleBlur }
                            error={ values.password && touched.password && !!errors.password }
                            helperText={ values.password && touched.password && errors.password }
                            disabled={ isSubmitting }
                          />
                        </Grid>
                        <Grid item xs={ 12 }>
                          <TextField
                            fullWidth
                            id="confirmPassword"
                            label="Confirm Password"
                            name="confirmPassword"
                            type="password"
                            value={ values.confirmPassword }
                            onChange={ handleChange }
                            onBlur={ handleBlur }
                            error={ values.confirmPassword && touched.confirmPassword && !!errors.confirmPassword }
                            helperText={ values.confirmPassword && touched.confirmPassword && errors.confirmPassword }
                            disabled={ isSubmitting }
                          />
                        </Grid>
                        <Grid item xs={ 6 }>
                          <TextField
                            fullWidth
                            id="firstName"
                            label="First Name"
                            name="firstName"
                            value={ values.firstName }
                            onChange={ handleChange }
                            onBlur={ handleBlur }
                            error={ values.firstName && touched.firstName && !!errors.firstName }
                            helperText={ values.firstName && touched.firstName && errors.firstName }
                            disabled={ isSubmitting }
                          />
                        </Grid>
                        <Grid item xs={ 6 }>
                          <TextField
                            fullWidth
                            id="lastName"
                            label={ this.props.requireLegalName ? "Last Name" : "Last Name (optional)" }
                            name="lastName"
                            value={ values.lastName }
                            onChange={ handleChange }
                            onBlur={ handleBlur }
                            error={ values.lastName && touched.lastName && !!errors.lastName }
                            helperText={ values.lastName && touched.lastName && errors.lastName }
                            disabled={ isSubmitting }
                          />
                        </Grid>
                        { this.props.requireBetaCode && <Grid item xs={ 12 }>
                          <TextField
                            fullWidth
                            id="betaCode"
                            label="Beta Code"
                            name="betaCode"
                            type="betaCode"
                            value={ values.betaCode }
                            onChange={ handleChange }
                            onBlur={ handleBlur }
                            error={ values.betaCode && touched.betaCode && !!errors.betaCode }
                            helperText={ values.betaCode && touched.betaCode && errors.betaCode }
                            disabled={ isSubmitting }
                          />
                        </Grid> }
                        { this.renderReCAPTCHAMaybe() }
                      </Grid>
                    </CardContent>
                    <CardActions>
                      <Button
                        to="/login"
                        component={ RouterLink }
                      >
                        Cancel
                      </Button>
                      <div style={ { marginLeft: 'auto' } }/>
                      <Button
                        to="/register"
                        component={ RouterLink }
                        variant="outlined"
                        color="primary"
                        onClick={ handleSubmit }
                        disabled={ this.cannotSubmit(isSubmitting, values) }
                      >
                        Register
                      </Button>
                    </CardActions>
                  </Card>
                </Grow>
              </Container>
            </Box>
          ) }
        </Formik>
      </div>
    )
  }
}

export default connect(
  state => ({
    allowSignUp: getSignUpAllowed(state),
    verifyRegister: getShouldVerifyRegister(state),
    ReCAPTCHAKey: getReCAPTCHAKey(state),
    requireLegalName: getRequireLegalName(state),
    requirePhoneNumber: getRequirePhoneNumber(state),
    requireBetaCode: getRequireBetaCode(state),
  }),
  dispatch => bindActionCreators({
    bootstrapLogin,
  }, dispatch),
)(withRouter(SignUpView));
