import { Box, Button, CircularProgress, Container, Grid, Grow, Paper, TextField, Typography } from "@material-ui/core";
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
import fetchBankAccounts from "shared/bankAccounts/actions/fetchBankAccounts";
import fetchSpending from "shared/spending/actions/fetchSpending";
import { fetchFundingSchedulesIfNeeded } from "shared/fundingSchedules/actions/fetchFundingSchedulesIfNeeded";
import fetchInitialTransactionsIfNeeded from "shared/transactions/actions/fetchInitialTransactionsIfNeeded";
import fetchBalances from "shared/balances/actions/fetchBalances";
import { Formik } from "formik";
import createLink from "shared/links/actions/createLink";
import Link, { LinkType } from "data/Link";


export class FirstTimeSetup extends Component {

  STEP = {
    INTRO: 0,
    MANUAL: 1,
  };


  state = {
    step: this.STEP.INTRO,
    loading: true,
    error: false,
    linkToken: '',
    longPollAttempts: 0,
    linkId: 0,
  };

  static propTypes = {
    logout: PropTypes.func.isRequired,
    fetchLinks: PropTypes.func.isRequired,
    fetchBankAccounts: PropTypes.func.isRequired,
    fetchSpending: PropTypes.func.isRequired,
    fetchFundingSchedulesIfNeeded: PropTypes.func.isRequired,
    fetchInitialTransactionsIfNeeded: PropTypes.func.isRequired,
    fetchBalances: PropTypes.func.isRequired,
    createLink: PropTypes.func.isRequired,
  }

  componentDidMount() {
    request()
      .get('/plaid/link/token/new')
      .then(result => {
        this.setState({
          loading: false,
          linkToken: result.data.linkToken,
        });
      })
      .catch(error => {
        this.setState({
          loading: false,
          error: true,
        })
      });
  }

  doCancel = () => {
    this.props.logout();
  }

  setupManualLink = () => {
    this.setState({
      loading: true,
    });

    return this.props.createLink(new Link({
      name: 'Manual',
      institutionName: 'Manual',
      linkType: LinkType.Manual,
    }))
      .then(result => {
        return Promise.all([
          this.props.fetchLinks(),
          this.props.fetchBankAccounts().then(() => {
            return Promise.all([
              this.props.fetchInitialTransactionsIfNeeded(),
              this.props.fetchFundingSchedulesIfNeeded(),
              this.props.fetchSpending(),
              this.props.fetchBalances(),
            ]);
          }),
        ]);
      });
  };

  plaidLinkSuccess = (token, metadata) => {
    this.setState({
      loading: true,
    });

    request().post('/plaid/link/token/callback', {
      publicToken: token,
      institutionId: metadata.institution.institution_id,
      institutionName: metadata.institution.name,
      accountIds: new List(metadata.accounts).map(account => account.id).toArray()
    })
      .then(result => {
        this.setState({
          linkId: result.data.linkId,
        });

        return this.longPollSetup()
          .then(() => {
            return Promise.all([
              this.props.fetchLinks(),
              this.props.fetchBankAccounts().then(() => {
                return Promise.all([
                  this.props.fetchInitialTransactionsIfNeeded(),
                  this.props.fetchFundingSchedulesIfNeeded(),
                  this.props.fetchSpending(),
                  this.props.fetchBalances(),
                ]);
              }),
            ]);
          });
      })
      .catch(error => {
        console.error(error);
      })
  };

  longPollSetup = () => {
    this.setState(prevState => ({
      longPollAttempts: prevState.longPollAttempts + 1,
    }));

    const { longPollAttempts, linkId } = this.state;
    if (longPollAttempts > 6) {
      return Promise.resolve();
    }

    return request().get(`/plaid/link/setup/wait/${ linkId }`)
      .then(result => {
        return Promise.resolve();
      })
      .catch(error => {
        if (error.response.status === 408) {
          return this.longPollSetup();
        }
      });
  };

  nextStep = () => {
    this.setState(prevState => ({
      step: prevState.step - 1,
    }));
  };

  previousStep = () => {
    this.setState(prevState => ({
      step: prevState.step - 1,
    }));
  };

  onEvent = (thing, stuff) => {
    console.warn({
      thing,
      stuff
    });
  }

  renderPlaidLink = () => {
    const { loading, linkToken } = this.state;
    if (loading) {
      return <CircularProgress style={ { float: 'right' } }/>;
    }

    if (linkToken.length > 0) {
      return (
        <PlaidConnectButton
          token={ linkToken }
          onSuccess={ this.plaidLinkSuccess }
          disabled={ this.state.loading }
          onEvent={ this.onEvent }
          onExit={ this.onEvent }
          onLoad={ this.onEvent }
        />
      )
    }

    return <Typography>Something went wrong...</Typography>;
  };

  renderIntro = () => {
    return (
      <Fragment>
        <Grid item xs={ 12 }>
          <Typography variant="h5">Welcome to monetr!</Typography>
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
          <Button variant="outlined" onClick={ this.setupManualLink }>Manual</Button>
        </Grid>
        <Grid item xs={ 3 }>
          { this.renderPlaidLink() }
        </Grid>
      </Fragment>
    )
  };

  renderManualStep = () => {
    return (
      <Formik
        initialValues={ {
          name: '',
        } }
        onSubmit={ this.setupManualLink }
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
          <Fragment>
            <Grid item xs={ 12 }>
              <Typography variant="h5">Welcome to monetr!</Typography>
              <Typography>
                What do you want to call your first bank account.
                <br/>
                <i>Note: This should be something like "Checking account" as you want to differentiate between separate
                  accounts even within the same bank for easier management of spending.</i>
              </Typography>
            </Grid>
            <Grid item xs={ 12 }>
              <TextField
                fullWidth
                id="name"
                label="Name"
                name="name"
                value={ values.name }
                onChange={ handleChange }
                error={ touched.name && !!errors.name }
                helperText={ touched.name && errors.name }
                disabled={ isSubmitting }
              />
            </Grid>
            <Grid item xs={ 6 }>
              <Button
                disabled={ this.state.loading }
                variant="outlined"
                onClick={ this.previousStep }
              >
                Back
              </Button>
            </Grid>
            <Grid item xs={ 6 }>
              <Button
                color="primary"
                variant="outlined"
                disabled={ this.state.loading }
                style={ { float: 'right' } }
                onClick={ submitForm }
              >
                Continue
              </Button>
            </Grid>
          </Fragment>
        ) }
      </Formik>
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
    fetchBankAccounts,
    fetchSpending,
    fetchFundingSchedulesIfNeeded,
    fetchInitialTransactionsIfNeeded,
    fetchBalances,
    createLink,
  }, dispatch),
)(withRouter(FirstTimeSetup));
