import React, { Component } from 'react';
import { connect } from 'react-redux';
import { RouteComponentProps, withRouter } from 'react-router-dom';
import { CircularProgress, Typography } from '@material-ui/core';
import Logo from 'assets';
import request from 'shared/util/request';

interface WithConnectionPropTypes extends RouteComponentProps {

}

export class VerifyEmail extends Component<WithConnectionPropTypes, any> {

  componentDidMount() {
    const search = this.props.location.search;
    const query = new URLSearchParams(search);
    const token = query.get('token');

    if (!token) {
      this.errorRedirect('Email verification link is not valid.')
      return
    }

    request().post('/authentication/verify', {
      'token': token,
    })
      .then(result => {
        window.alert(result?.data?.message || 'Your email has been verified, please login.')

        this.props.history.push(result?.data?.nextUrl || '/login');
      })
      .catch(error => {
        this.errorRedirect(
          error?.response?.data?.error || 'Failed to verify email address.',
          error?.response?.data?.nextUrl,
        );
      });
  }

  errorRedirect = (message: string, nextUrl: string = '/login') => {
    window.alert(message);
    this.props.history.push(nextUrl);
  }

  render() {
    return (
      <div className="flex items-center justify-center w-full h-full max-h-full">
        <div className="w-full p-10 xl:w-3/12 lg:w-5/12 md:w-2/3 sm:w-10/12 max-w-screen-sm sm:p-0">
          <div className="flex justify-center w-full mb-5">
            <img src={ Logo } className="w-1/3"/>
          </div>
          <div className="w-full pt-2.5 pb-2.5">
            <Typography
              variant="h5"
              className="w-full text-center"
            >
              Verifying email address...
            </Typography>
          </div>
          <div className="w-full pt-2.5 pb-2.5 flex justify-center">
            <CircularProgress/>
          </div>
        </div>
      </div>
    )
  }
}

export default connect(
  state => ({}),
  {},
)(withRouter(VerifyEmail));
