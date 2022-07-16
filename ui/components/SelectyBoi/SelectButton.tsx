import React from 'react';
import { ArrowDropDown } from '@mui/icons-material';
import { Button } from '@mui/material';
import classnames from 'classnames';

import './styles/SelectButton.scss';

interface PropTypes {
  open?: boolean;
  children: JSX.Element | JSX.Element[] | string;
  onClick?: (event: object) => any;
  disabled?: boolean;
}

const SelectButton = (props: PropTypes): JSX.Element => {
  const { children, onClick, disabled, open } = props;
  return (
    <Button className={ classnames('w-full monetr-select-button overflow-hidden', {
      'selected': open,
    }) } onClick={ onClick }
    disabled={ disabled }>
      <div className="w-full flex justify-start overflow-hidden">
        <div className="flex-auto flex justify-start overflow-hidden normal-case text-lg">
          { children }
        </div>
        <div className="flex-none select-dropdown-icon">
          <ArrowDropDown
            className={ classnames('transform transition transition-transform duration-200', {
              'rotate-180': open,
            }) }
          />
        </div>
      </div>
    </Button>
  );
};

export default SelectButton;
