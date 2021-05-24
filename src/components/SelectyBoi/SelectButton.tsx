import React, { Component } from "react";
import { Button } from "@material-ui/core";
import { ArrowDropDown } from "@material-ui/icons";

import './styles/SelectButton.scss';

export interface Props {
  children: JSX.Element | JSX.Element[];
  onClick?: (event: object) => any;
  disabled?: boolean;
}

export default class SelectButton extends Component<Props, any> {
  render() {
    const { children, onClick, disabled } = this.props;
    return (
      <Button className="w-full monetr-select-button overflow-hidden" onClick={ onClick } disabled={ disabled }>
        <div className="w-full flex justify-start overflow-hidden">
          <div className="flex-auto flex justify-start overflow-hidden">
            { children }
          </div>
          <div className="flex-none select-dropdown-icon">
            <ArrowDropDown/>
          </div>
        </div>
      </Button>
    )
  }
}
