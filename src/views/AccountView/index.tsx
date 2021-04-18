import { Card, Typography } from "@material-ui/core";
import React, { Component } from "react";
import { connect } from "react-redux";

export class AccountView extends Component<any, any> {

  render() {
    return (
      <div className="minus-nav">
        <div className="flex flex-col h-full p-10 max-h-full">
          <div className="grid grid-cols-3 gap-4 flex-grow">
            <div className="col-span-3">
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
  (state) => ({}),
  {}
)(AccountView);
