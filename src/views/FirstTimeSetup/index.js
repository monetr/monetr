import { Box, Button, CircularProgress, Container, Grid, Grow, Paper, Typography } from "@material-ui/core";
import { List } from "immutable";
import PropTypes from "prop-types";
import React, { Component, Fragment } from "react";
import { connect } from "react-redux";
import { withRouter } from "react-router-dom";
import { bindActionCreators } from "redux";
import logout from "shared/authentication/actions/logout";
import fetchLinks from "shared/links/actions/fetchLinks";
import request from "shared/util/request";
import { PlaidConnectButton } from "views/FirstTimeSetup/PlaidConnectButton";


export class FirstTimeSetup extends Component {

  STEP = {
    INTRO: 0,
    MANUAL: 1,
  };


  state = {
    step: this.STEP.INTRO,
    loading: true,
    linkToken: '',
  };

  static propTypes = {
    logout: PropTypes.func.isRequired,
    fetchLinks: PropTypes.func.isRequired,
  }

  componentDidMount() {
    request()
      .get('/api/plaid/link/token/new')
      .then(result => {
        this.setState({
          loading: false,
          linkToken: result.data.linkToken,
        });
      })
      .catch(error => {
        alert(error);
      });
  }

  doCancel = () => {
    this.props.logout();
  }

  plaidLinkSuccess = (token, metadata) => {
    request().post('/api/plaid/link/token/callback', {
      publicToken: token,
      institutionId: metadata.institution.institution_id,
      institutionName: metadata.institution.name,
      accountIds: new List(metadata.accounts).map(account => account.id).toArray()
    })
      .then(result => {
        return this.props.fetchLinks();
      })
      .catch(error => {
        console.error(error);
      })
  };

  renderPlaidLink = () => {
    const { loading, linkToken } = this.state;
    if (loading) {
      return <CircularProgress style={ { float: 'right' } }/>;
    }

    if (linkToken.length > 0) {
      return (
        <PlaidConnectButton token={ linkToken } onSuccess={ this.plaidLinkSuccess }/>
      )
    }

    return <Typography>Something went wrong...</Typography>;
  };

  renderIntro = () => {
    return (
      <Fragment>
        <Grid item xs={ 12 }>
          <Typography variant="h5">Welcome to Harder Than It Needs To Be</Typography>
          <Typography>
            To continue you will need to setup a bank account. You can setup a manual bank account where all
            transactions and balances are entered manually. Or you can link your bank account with our
            application and we can automatically import and maintain this data for you.
          </Typography>
          <Typography>
            If you do not want to setup your account at this time click cancel and you will be logged out.
          </Typography>
        </Grid>
        <Grid item xs={ 6 }>
          <Button variant="outlined" onClick={ this.doCancel }>Cancel</Button>
        </Grid>
        <Grid item xs={ 3 }>
          <Button variant="outlined" onClick={ () => {
            this.setState({
              step: this.STEP.MANUAL,
            });
          } }>Manual</Button>
        </Grid>
        <Grid item xs={ 3 }>
          { this.renderPlaidLink() }
        </Grid>
      </Fragment>
    )
  };

  renderManualStep = () => {
    return (
      <Fragment>
        <Grid item xs={ 12 }>
          <Typography variant="h5">Welcome to Harder Than It Needs To Be</Typography>
          <Typography>
            What do you want to name the bank for your manual account?
          </Typography>
        </Grid>
        <Grid item xs={ 6 }>
          <Button variant="outlined" onClick={ () => {
            this.setState(prevState => ({
              step: prevState.step - 1,
            }));
          } }>Back</Button>
        </Grid>
        <Grid item xs={ 6 }>
          <Button style={ { float: 'right' } } onClick={ () => {
            this.setState(prevState => ({
              step: prevState.step - 1,
            }));
          } }>Continue</Button>
        </Grid>
      </Fragment>
    );
  };

  render() {
    const { step } = this.state;
    return (
      <Box m={ 12 }>
        <Container maxWidth="sm">
          <Grow in>
            <Paper elevation={ 3 }>
              <Box m={ 4 }>
                <Grid container spacing={ 4 }>
                  { step === this.STEP.INTRO && this.renderIntro() }
                  { step === this.STEP.MANUAL && this.renderManualStep() }
                </Grid>
              </Box>
            </Paper>
          </Grow>
        </Container>
      </Box>
    );
  }
}

export default connect(
  state => ({}),
  dispatch => bindActionCreators({
    logout,
    fetchLinks,
  }, dispatch),
)(withRouter(FirstTimeSetup));
