import { Checkbox, ListItem, ListItemIcon, Typography } from '@material-ui/core';
import Spending from 'data/Spending';
import React, { Component } from 'react';
import { connect } from 'react-redux';

export interface PropTypes {
  goalId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  isSelected: boolean;
  goal: Spending;
}

export class GoalRow extends Component<WithConnectionPropTypes, any> {

  render() {
    const { isSelected, goal } = this.props;

    return (
      <ListItem>
        <ListItemIcon>
          <Checkbox
            edge="start"
            checked={ isSelected }
            tabIndex={ -1 }
            color="primary"
          />
        </ListItemIcon>
        <div>
          <div>
            <Typography>{ goal.name }</Typography>
          </div>
        </div>
      </ListItem>
    );
  }
}

export default connect(
  (state, props: PropTypes) => {
    return {}
  },
  {}
)(GoalRow);
