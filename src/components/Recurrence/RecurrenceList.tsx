import { Checkbox, List as ListUI, ListItem, ListItemIcon, ListItemText } from "@material-ui/core";
import getRecurrencesForDate from "components/Recurrence/getRecurrencesForDate";
import Recurrence from "components/Recurrence/Recurrence";

import { List } from 'immutable';
import moment from "moment";
import React, { Component } from "react";

export interface PropTypes {
  date: moment.Moment;
  onChange: { (value: Recurrence): void }
}

interface State {
  rules: List<Recurrence>;
  selectedIndex?: number;
}

// RecurrenceList generates a list of possible recurrence rules based on the provided date. When a recurrence is
// selected the onChange function will be called with a string value representing an RRule.
export class RecurrenceList extends Component<PropTypes, State> {

  state = {
    selectedIndex: null,
    rules: List<Recurrence>(),
  };

  componentDidMount() {
    const { date } = this.props;
    this.setState({
      rules: getRecurrencesForDate(date)
    });
  }

  selectItem = (index: number) => () => {
    const { onChange } = this.props;
    const { rules } = this.state;

    return this.setState({
      selectedIndex: index,
    }, () => onChange(rules.get(index)));
  };

  renderItems = () => {
    const { rules, selectedIndex } = this.state;

    return rules.map((rule, index) => (
      <ListItem key={ rule.name } dense button onClick={ this.selectItem(index) }>
        <ListItemIcon>
          <Checkbox
            edge="start"
            checked={ selectedIndex === index }
            tabIndex={ -1 }
            color="primary"
          />
        </ListItemIcon>
        <ListItemText>
          { rule.name }
        </ListItemText>
      </ListItem>
    ));
  };

  render() {

    return (
      <ListUI dense>
        { this.renderItems() }
      </ListUI>
    );
  }
}
