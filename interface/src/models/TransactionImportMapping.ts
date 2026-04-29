// FieldRef references either a named column from the input file or a derived value computed at processing time, never
// both. The discriminated union enforces this at compile time: setting both `name` and `derivedKind` is a type error.
export enum DerivedKind {
  RowNumber = 'rowNumber',
  RowNumberPerDay = 'rowNumberPerDay',
  RowNumberPerDayPerAmount = 'rowNumberPerDayPerAmount',
}

export type FieldRef = { name: string; derivedKind?: never } | { name?: never; derivedKind: DerivedKind };

// IDSpec is not a union; both kinds use the same shape. Kind chooses the strategy used to build the unique identifier
// from the supplied fields.
export enum IDSpecKind {
  Native = 'native',
  Hashed = 'hashed',
}

export interface IDSpec {
  kind: IDSpecKind;
  fields: FieldRef[];
}

// AmountSpec is a three-way discriminated union on `kind`:
//   - sign:   one field, no credit/debit
//   - type:   two fields plus a credit and debit string for the type column
//   - column: two fields, one debit and one credit, no credit/debit strings
export enum AmountKind {
  Sign = 'sign',
  Type = 'type',
  Column = 'column',
}

interface AmountSpecBase {
  invert?: boolean;
  fields: FieldRef[];
}

export interface SignAmountSpec extends AmountSpecBase {
  kind: AmountKind.Sign;
  credit?: never;
  debit?: never;
}

export interface TypeAmountSpec extends AmountSpecBase {
  kind: AmountKind.Type;
  credit: string;
  debit: string;
}

export interface ColumnAmountSpec extends AmountSpecBase {
  kind: AmountKind.Column;
  credit?: never;
  debit?: never;
}

export type AmountSpec = SignAmountSpec | TypeAmountSpec | ColumnAmountSpec;

export interface DateSpec {
  fields: FieldRef[];
  format: string;
}

export interface PostedSpec {
  fields: FieldRef[];
  posted?: string;
}

// BalanceSpec is a discriminated union on `kind`:
//   - none / sum: fields must be absent
//   - field:      fields contains exactly the column to read the balance from
export enum BalanceKind {
  None = 'none',
  Field = 'field',
  Sum = 'sum',
}

export interface NoneOrSumBalanceSpec {
  kind: BalanceKind.None | BalanceKind.Sum;
  fields?: never;
}

export interface FieldBalanceSpec {
  kind: BalanceKind.Field;
  fields: FieldRef[];
}

export type BalanceSpec = NoneOrSumBalanceSpec | FieldBalanceSpec;

interface TransactionImportMapping {
  id: IDSpec;
  amount: AmountSpec;
  memo: FieldRef;
  merchant?: FieldRef;
  date: DateSpec;
  posted?: PostedSpec;
  balance: BalanceSpec;
  headers: string[];
}

export default TransactionImportMapping;
