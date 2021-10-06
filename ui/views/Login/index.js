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
  TextField,
} from "@material-ui/core";
import { Alert, AlertTitle } from "@material-ui/lab";
import { Formik } from 'formik';
import PropTypes from "prop-types";
import React, { Component } from 'react';
import ReCAPTCHA from "react-google-recaptcha";
import { connect } from 'react-redux';
import { Link as RouterLink, withRouter } from 'react-router-dom';
import { bindActionCreators } from "redux";
import bootstrapLogin from "shared/authentication/actions/bootstrapLogin";
import { getReCAPTCHAKey, getShouldVerifyLogin, getSignUpAllowed } from "shared/bootstrap/selectors";
import request from "shared/util/request";

import './styles/login.scss';

export class LoginView extends Component {
  state = {
    verification: null,
    error: null,
    loading: false,
  };

  static propTypes = {
    allowSignUp: PropTypes.bool.isRequired,
    verifyLogin: PropTypes.bool.isRequired,
    ReCAPTCHAKey: PropTypes.string,
    bootstrapLogin: PropTypes.func.isRequired,
    history: PropTypes.instanceOf(History).isRequired,
  };

  submitLogin = values => {
    this.setState({
      error: null,
      loading: true,
    });
    return request().post('/authentication/login', {
      email: values.email,
      password: values.password,
      captcha: this.state.verification,
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
        if (error.response.data.error) {
          this.setState({
            error: error.response.data.error,
            loading: false,
          });
        } else {
          throw error;
        }
      });
  };

  renderReCAPTCHAMaybe = () => {
    const { verifyLogin, ReCAPTCHAKey } = this.props;
    if (!verifyLogin) {
      return null;
    }

    return (
      <Grid item xs={ 12 }>
        <div className="w-full flex justify-center items-center">
          { !this.state.loading && <ReCAPTCHA
            sitekey={ ReCAPTCHAKey }
            onChange={ value => this.setState({ verification: value }) }
          /> }
          { this.state.loading && <CircularProgress/> }
        </div>
      </Grid>
    )
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

  render() {
    return (
      <div className="login-view">
        { this.renderErrorMaybe() }
        <Formik
          initialValues={ {
            email: '',
            password: '',
          } }
          onSubmit={ (values, { setSubmitting }) => {
            this.submitLogin(values)
              .finally(() => setSubmitting(false));
          } }
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
                    <CardHeader title="monetr" subheader="Login"/>
                    <CardContent>
                      <Grid container spacing={ 1 }>
                        <Grid item xs={ 12 }>
                          <TextField
                            fullWidth
                            id="email"
                            label="Email"
                            name="email"
                            value={ values.email }
                            onChange={ handleChange }
                            error={ touched.email && !!errors.email }
                            helperText={ touched.email && errors.email }
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
                            error={ touched.password && !!errors.password }
                            helperText={ touched.password && errors.password }
                            disabled={ isSubmitting }
                          />
                        </Grid>
                        { this.renderReCAPTCHAMaybe() }
                      </Grid>
                    </CardContent>
                    <CardActions>
                      <div style={ { marginLeft: 'auto' } }/>
                      { this.props.allowSignUp &&
                      <Button
                        to="/register"
                        component={ RouterLink }
                        variant="outlined"
                        color="secondary"
                        disabled={ isSubmitting }
                      >
                        Sign Up
                      </Button>
                      }
                      <Button
                        onClick={ handleSubmit }
                        variant="outlined"
                        color="primary"
                        disabled={ isSubmitting }
                      >
                        Login
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
    verifyLogin: getShouldVerifyLogin(state),
    ReCAPTCHAKey: getReCAPTCHAKey(state),
  }),
  dispatch => bindActionCreators({
    bootstrapLogin,
  }, dispatch),
)(withRouter(LoginView));
