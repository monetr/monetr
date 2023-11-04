(self.webpackChunk_monetr_stories=self.webpackChunk_monetr_stories||[]).push([["interface_src_modals_TransferModal_stories_tsx~interface_src_pages_expense_details_stories_ts~b352cf"],{"../interface/src/modals/TransferModal.tsx":function(e,n,t){"use strict";Object.defineProperty(n,"__esModule",{value:!0});!function(e,n){for(var t in n)Object.defineProperty(e,t,{enumerable:!0,get:n[t]})}(n,{default:function(){return h},showTransferModal:function(){return S}});var r=t("../node_modules/react/jsx-runtime.js"),s=t("../node_modules/react/index.js"),o=t.ir(t("../node_modules/@ebay/nice-modal-react/lib/esm/index.js")),a=t("../node_modules/@mui/icons-material/esm/index.js"),l=t("../node_modules/formik/dist/formik.esm.js"),i=t("../node_modules/notistack/notistack.esm.js"),u=t.ir(t("../interface/src/components/MAmountField.tsx")),d=t.ir(t("../interface/src/components/MButton.tsx")),m=t.ir(t("../interface/src/components/MForm.tsx")),c=t.ir(t("../interface/src/components/MModal.tsx")),f=t.ir(t("../interface/src/components/MSelectSpending.tsx")),x=t.ir(t("../interface/src/components/MSpan.tsx")),p=t("../interface/src/hooks/balances.ts"),g=t("../interface/src/hooks/spending.ts"),b=t("../interface/src/util/amounts.ts");let j=o.default.create(function(e){let n=(0,o.useModal)(),t=(0,s.useRef)(null),a=(0,g.useTransfer)(),{enqueueSnackbar:l}=(0,i.useSnackbar)(),j=(0,p.useCurrentBalance)(),{result:h}=(0,g.useSpendingSink)(),S={fromSpendingId:e.initialFromSpendingId,toSpendingId:e.initialToSpendingId,amount:0};function _(e){let n={},t=(0,b.friendlyToAmount)(e.amount);if(t<=0&&(n.amount="Amount must be greater than zero"),null===e.fromSpendingId)t>j?.free&&(n.amount="Cannot move more than is available from Free-To-Use");else{let r=h?.find(n=>n.spendingId===e.fromSpendingId);t>r?.currentAmount&&(n.amount=`Cannot move more than is available from ${r?.name}`)}return n}async function y(e,t){if(null===e.toSpendingId&&null===e.fromSpendingId)return t.setFieldError("toSpendingId","Must select a destination and a source"),Promise.resolve();let r=Math.ceil(100*e.amount),s=_(e);return Object.keys(s).length>0?(t.setErrors(s),Promise.resolve()):(t.setSubmitting(!0),a(e.fromSpendingId,e.toSpendingId,r).then(()=>n.remove()).catch(e=>void l(e.response.data.error,{variant:"error",disableWindowBlurListener:!0})).finally(()=>t.setSubmitting(!1)))}return(0,r.jsx)(c.default,{open:n.visible,ref:t,className:"md:max-w-sm",children:(0,r.jsxs)(m.default,{onSubmit:y,initialValues:S,validate:_,className:"h-full flex flex-col gap-2 p-2 justify-between","data-testid":"transfer-modal",children:[(0,r.jsxs)("div",{className:"flex flex-col gap-2",children:[(0,r.jsxs)("div",{className:"flex flex-col items-center",children:[(0,r.jsx)(x.default,{size:"2xl",weight:"semibold",children:"Transfer"}),(0,r.jsx)(x.default,{size:"lg",weight:"medium",color:"subtle",children:"Move funds between your budgets"})]}),(0,r.jsx)(f.default,{excludeFrom:"toSpendingId",label:"From",labelDecorator:v,menuPortalTarget:document.body,name:"fromSpendingId"}),(0,r.jsx)(k,{}),(0,r.jsx)(f.default,{excludeFrom:"fromSpendingId",label:"To",labelDecorator:v,menuPortalTarget:document.body,name:"toSpendingId"}),(0,r.jsx)(u.default,{name:"amount",label:"Amount",placeholder:"Amount to move...",step:"0.01",allowNegative:!1})]}),(0,r.jsxs)("div",{className:"flex justify-end gap-2",children:[(0,r.jsx)(d.default,{color:"secondary",onClick:n.remove,"data-testid":"close-new-expense-modal",children:"Cancel"}),(0,r.jsx)(d.default,{color:"primary",type:"submit",children:"Transfer"})]})]})})});var h=j;function S(e){return o.default.show(j,e)}function k(){let e=(0,l.useFormikContext)();return(0,r.jsx)("a",{className:"w-full flex justify-center mb-1",children:(0,r.jsx)(a.SwapVertOutlined,{onClick:function(){if(e.isSubmitting)return;let{fromSpendingId:n,toSpendingId:t,amount:r}=e.values;e.setValues({fromSpendingId:t,toSpendingId:n,amount:r})},className:"cursor-pointer text-4xl dark:text-dark-monetr-content-subtle hover:dark:text-dark-monetr-content"})})}function v(e){let n=(0,l.useFormikContext)(),t=n.values[e.name],{result:s}=(0,g.useSpendingSink)(),o=(0,p.useCurrentBalance)();if(!t||-1===t){let e=o?.free;return(0,r.jsx)(_,{amount:e})}let a=s?.find(e=>e.spendingId===t);if(!a)return null;let i=a.currentAmount,u=a.targetAmount,d=Math.max(a.targetAmount-a.currentAmount,0);return d>0&&d!=u?(0,r.jsxs)(x.default,{className:"gap-1",children:[(0,r.jsx)(_,{amount:i}),"of",(0,r.jsx)(_,{amount:u}),"\xa0 (",(0,r.jsx)(_,{amount:d}),")"]}):(0,r.jsxs)(x.default,{className:"gap-1",color:"subtle",children:[(0,r.jsx)(_,{amount:i}),"of",(0,r.jsx)(_,{amount:u})]})}function _({amount:e}){let n=(0,l.useFormikContext)(),t=(0,s.useCallback)(()=>{"number"==typeof e&&n?.setFieldValue("amount",(0,b.amountToFriendly)(e))},[n,e]);return(0,r.jsx)(x.default,{size:"sm",weight:"medium",className:"cursor-pointer hover:dark:text-dark-monetr-content-emphasis",onClick:t,children:"number"==typeof e&&(0,b.formatAmount)(e)})}}}]);