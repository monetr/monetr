import React, {PureComponent} from 'react';
import PropTypes from "prop-types";
import {bindActionCreators} from "redux";
import {connect} from "react-redux";
import Application from "./Application";
import {getIsAuthenticated} from "./shared/authentication/selectors";
import bootstrapApplication from "./shared/bootstrap";
import {getIsBootstrapped, getSignUpAllowed} from "./shared/bootstrap/selectors";
import {
  BrowserRouter as Router,
  Switch,
  Route,
  Link,
  Redirect
} from "react-router-dom";
import SignUpView from "./views/SignUp";
import LoginView from "./views/Login";
import bootstrapLogin from "./shared/authentication/actions/bootstrapLogin";

export class Root extends PureComponent {
  state = {
    loading: true,
  };

  static propTypes = {
    isReady: PropTypes.bool.isRequired,
    isAuthenticated: PropTypes.bool.isRequired,
    allowSignUp: PropTypes.bool.isRequired,
    bootstrapApplication: PropTypes.func.isRequired,
    bootstrapLogin: PropTypes.func.isRequired,
  };

  componentDidMount() {
    this.props.bootstrapApplication()
      .then(() => this.props.bootstrapLogin())
      .finally(() => {
        this.setState({
          loading: false
        });
      });
  }

  renderUnauthenticated = () => {
    return (
      <Router>
        <Switch>
          <Route path="/login">
            <LoginView/>
          </Route>
          {this.props.allowSignUp &&
          <Route path="/register">
            <SignUpView/>
          </Route>
          }
          <Route>
            <Redirect to={{pathname: '/login'}}/>
          </Route>
        </Switch>
      </Router>
    )
  };

  renderAuthenticated = () => {
    return (
      <Router>
        <Switch>
          <Route path="/transactions">
            <h1>Transactions</h1>
          </Route>
          <Route path="/expenses">
            <h1>Expenses</h1>
          </Route>
          <Route path="/goals">
            <h1>Goals</h1>
          </Route>
          <Route path="/">
            <h1>Home/Setup</h1>
          </Route>
          <Route>
            <h1>Not found</h1>
          </Route>
        </Switch>
      </Router>
    )
  };

  render() {
    const {isReady, isAuthenticated} = this.props;
    if (!isReady || this.state.loading) {
      return <h1>Loading...</h1>;
    }

    if (!isAuthenticated) {
      return this.renderUnauthenticated();
    }

    return this.renderAuthenticated();
  }
}

export default connect(
  state => ({
    isAuthenticated: getIsAuthenticated(state),
    isReady: getIsBootstrapped(state),
    allowSignUp: getSignUpAllowed(state),
  }),
  dispatch => bindActionCreators({
    bootstrapApplication,
    bootstrapLogin,
  }, dispatch),
)(Root);
