import moment from "moment";

export default class Transaction {
    transactionId: number;
    bankAccountId: number;
    amount: number;
    expenseId?: number;
    categories: string[];
    originalCategories: string[];
    date: moment.Moment;
    authorizedDate?: moment.Moment;
    name?: string;
    originalName: string;
    merchantName?: string;
    originalMerchantName?: string;
    isPending: boolean;
    createdAt: moment.Moment;
}
