import React, { Component, Fragment } from 'react';
import { Theme } from '@mui/material';
import { WithTheme, withTheme } from '@mui/styles';

interface PropTypes extends WithTheme<Theme> {
  children: React.ReactNode;
}

class TextWithLine extends Component<PropTypes> {

  render() {
    return (
      <Fragment>
        <div className="relative w-full border-t border-gray-400" style={ {
          top: '1.2em',
        } } />
        <div className="relative flex justify-center inline w-full">
          <span className="relative p-1.5" style={ {
            backgroundColor: this.props.theme.palette.background.default,
          } }>
            { this.props.children }
          </span>
        </div>
      </Fragment>
    );
  }
}

export default withTheme<Theme, typeof TextWithLine>(TextWithLine);
