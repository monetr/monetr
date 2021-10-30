import React, { Component, Fragment } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Snackbar,
  Typography
} from "@mui/material";
import Link from "models/Link";
import { getLink } from "shared/links/selectors/getLink";
import { connect } from "react-redux";
import { Close } from "@mui/icons-material";
import removeLink from "shared/links/actions/removeLink";
import classnames from "classnames";

interface PropTypes {
  open: boolean;
  onClose: () => void;
  linkId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  link: Link;
  removeLink: (link: Link) => Promise<void>;
}

interface State {
  loading: boolean;
  error: string | null;
}

class RemoveLinkConfirmationDialog extends Component<WithConnectionPropTypes, State> {

  state = {
    loading: false,
    error: null,
  };

  doRemoveLink = (): Promise<void> => {
    this.setState({
      loading: true,
    });

    return this.props.removeLink(this.props.link)
      // If we successfully remove the link then close this dialog.
      .then(() => this.props.onClose())
      // If it fails then show an error.
      .catch(error => this.setState({
        error: error.response.data.error,
      }))
      // If the request failed then we will hit this, this will remove the loading state but will not close the dialog.
      .finally(() => this.setState({
        loading: false,
      }));
  };

  renderErrorMaybe = () => {
    const { error } = this.state;

    if (!error) {
      return null;
    }

    const onClose = () => this.setState({ error: null });

    return (
      <Snackbar open autoHideDuration={ 6000 } onClose={ onClose }>
        <Alert onClose={ onClose } severity="error">
          { error }
        </Alert>
      </Snackbar>
    )
  };

  render() {
    const { open, onClose, link } = this.props;
    const { loading } = this.state;

    return (
      <Fragment>
        { this.renderErrorMaybe() }

        <Dialog open={ open } onClose={ onClose }>
          <DialogTitle>
            <div className="flex items-center">
              <span className="text-2xl flex-auto">
                Remove { link.getName() }
              </span>
              <IconButton
                disabled={ loading }
                className="flex-none"
                onClick={ onClose }
              >
                <Close/>
              </IconButton>
            </div>
          </DialogTitle>
          <DialogContent>
            <Typography>
              Are you sure you want to remove the <b>{ link.getName() }</b> link? This cannot be undone.
            </Typography>
            { link.getIsPlaid() && <Typography>You can also convert this link to be manual instead.</Typography> }
          </DialogContent>
          <DialogActions>
            <Button
              disabled={ loading }
              onClick={ onClose }
            >
              Cancel
            </Button>
            <Button
              disabled={ loading }
              onClick={ this.doRemoveLink }
              className={ classnames({
                "text-red-500": !loading,
              }) }
            >
              Remove
            </Button>
          </DialogActions>
        </Dialog>
      </Fragment>
    )
  }
}

export default connect(
  (state, props: PropTypes) => ({
    link: getLink(props.linkId)(state),
  }),
  {
    removeLink,
  },
)(RemoveLinkConfirmationDialog);
