import { Card, List, ListItem, ListItemText, Typography } from "@material-ui/core";
import React, { Component } from "react";
import { connect } from "react-redux";
import { getLinks } from "shared/links/selectors/getLinks";
import { Map } from 'immutable';
import Link from "data/Link";

interface WithConnectionPropTypes {
  links: Map<number, Link>;
}

export class AccountView extends Component<WithConnectionPropTypes, any> {

  render() {
    return (
      <div className="minus-nav">
        <div className="flex flex-col h-full p-10 max-h-full">
          <div className="grid grid-cols-3 gap-4 flex-grow">
            <div className="col-span-1">
              <Card elevation={ 4 } className="w-full goals-list ">
                <div className="w-full text-center pt-5">
                  <Typography
                    variant="h5"
                  >
                    Linked Bank Accounts
                  </Typography>
                </div>
                <List className="w-full">
                  { this.props.links.map(link => (
                    <ListItem key={ link.linkId } button>
                      <ListItemText>
                        { link.getName() }
                      </ListItemText>
                    </ListItem>
                  )).valueSeq().toArray() }
                </List>
              </Card>
            </div>
            <div className="col-span-2">
              <Card elevation={ 4 } className="w-full goals-list ">
                <div className="h-full flex justify-center items-center">
                  <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-2">
                    <Typography
                      className="opacity-50"
                      variant="h3"
                    >
                      Account things
                    </Typography>
                  </div>
                </div>
              </Card>
            </div>
          </div>
        </div>
      </div>
    );
  }
}

export default connect(
  (state) => ({
    links: getLinks(state),
  }),
  {}
)(AccountView);
