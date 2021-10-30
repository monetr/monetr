import { connect } from "react-redux";
import { getReCAPTCHAKey } from "shared/bootstrap/selectors";
import React, { Component } from "react";
import { CircularProgress } from "@mui/material";
import ReCAPTCHA from "react-google-recaptcha";

export interface PropTypes {
  show?: boolean;
  loading?: boolean;
  onVerify: (verification: string) => void;
}

interface WithConnectionPropTypes extends PropTypes {
  ReCAPTCHAKey: string | null;
}

class CaptchaMaybe extends Component<WithConnectionPropTypes, any> {

  render() {
    const { show, ReCAPTCHAKey, loading } = this.props;

    if (!show || !ReCAPTCHAKey) {
      return null;
    }

    const { onVerify } = this.props;

    return (
      <div className="flex items-center justify-center w-full">
        { !loading && <ReCAPTCHA
          sitekey={ ReCAPTCHAKey }
          onChange={ onVerify }
        /> }
        { loading && <CircularProgress/> }
      </div>
    );
  }
}

export default connect(
  state => ({
    ReCAPTCHAKey: getReCAPTCHAKey(state),
  }),
  {},
)(CaptchaMaybe);
