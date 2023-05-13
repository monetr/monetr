import { Meta, StoryObj } from "@storybook/react";
import MLayout from "components/MLayout";
import moment from "moment";
import React from "react";
import TransactionsNew from "./transactions-new";

export default {
  title: 'Pages/Transactions',
  component: TransactionsNew,
} as Meta<typeof TransactionsNew>;

export const Default: StoryObj = {
  name: 'Default',
  render: () => (
    <MLayout>
      <TransactionsNew />
    </MLayout>
  ),
  args: {
    requests: [
      {
        method: 'GET',
        path: '/api/config',
        status: 200,
        response: {
          billingEnabled: true,
        }
      },
      {
        method: 'GET',
        path: '/api/links',
        status: 200,
        response: [
          {
            linkId: 1,
          }
        ]
      },
      {
        method: 'GET',
        path: '/api/bank_accounts',
        status: 200,
        response: [
          {
            bankAccountId: 1,
            linkId: 1,
          }
        ]
      },
      {
        method: 'GET',
        path: '/api/bank_accounts/1/balances',
        status: 200,
        response: {
          bankAccountId: 1,
          available: 100000,
          current: 99000,
          safe: 70000,
          expenses: 10000,
          goals: 19000,
        }
      },
      {
        method: 'GET',
        path: '/api/bank_accounts/1/spending',
        status: 200,
        response: [
          {
            spendingId: 1,
            bankAccountId: 1,
            fundingScheduleId: null,
            name: 'Car Saving',
            description: 'I want to save for a car',
            spendingType: 1,
            targetAmount: 100000,
            currentAmount: 19000,
            usedAmount: 0,
          }
        ]
      },
      {
        method: 'GET',
        path: '/api/bank_accounts/1/transactions',
        status: 200,
        response: [
          {
            transactionId: 1,
            bankAccountId: 1,
            amount: 1200,
            spendingId: null,
            spendingAmount: null,
            categories: [],
            originalCategories: [],
            date: moment().toISOString(),
            authorizedDate: null,
            name: null,
            originalName: 'ACH 1239 - SOME BANK TRANSFER',
            merchantName: null,
            originalMerchantName: null,
            isPending: false,
            createdAt: moment().toISOString(),
          },
          {
            transactionId: 2,
            bankAccountId: 1,
            amount: 1200,
            spendingId: null,
            spendingAmount: null,
            categories: [],
            originalCategories: [],
            date: moment().toISOString(),
            authorizedDate: null,
            name: null,
            originalName: 'WIRE 1239 - SOME BANK WIRE',
            merchantName: null,
            originalMerchantName: null,
            isPending: false,
            createdAt: moment().toISOString(),
          },
        ]
      }
    ]
  }
};
