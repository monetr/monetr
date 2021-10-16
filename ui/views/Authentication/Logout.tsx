import { Component } from 'react';
import { RouteComponentProps, withRouter } from 'react-router-dom';
import { connect } from 'react-redux';

import logout from 'shared/authentication/actions/logout';

interface PropTypes extends RouteComponentProps {
  logout: () => void;
}

export class Logout extends Component<PropTypes, any> {

  componentDidMount() {
    this.props.logout();
    this.props.history.push('/login');
  }

  render() {
    // This is just used to log the user out and redirect. It is not a real component.
    return null;
  }
}

export default connect(
  _ => ({}),
  {
    logout,
  }
)(withRouter(Logout));
