(self.webpackChunk_monetr_stories=self.webpackChunk_monetr_stories||[]).push([["interface_src_components_MDatePicker_stories_tsx~interface_src_components_MTextField_stories_~845401"],{"../interface/src/components/MTextField.tsx":function(e,r,d){"use strict";Object.defineProperty(r,"__esModule",{value:!0}),Object.defineProperty(r,"default",{enumerable:!0,get:function(){return l}});var a=d("../node_modules/react/jsx-runtime.js");d("../node_modules/react/index.js");var s=d("../node_modules/formik/dist/formik.esm.js"),i=d.ir(d("../interface/src/components/MLabel.tsx")),n=d.ir(d("../interface/src/util/mergeTailwind.ts"));let t={label:null,labelDecorator:e=>null,disabled:!1,uppercasetext:void 0};function l(e=t){let r=(0,s.useFormikContext)();e={...t,...e,disabled:e?.disabled||r?.isSubmitting,error:e?.error||(r?.touched[e?.name]?r?.errors[e?.name]:null)};let{labelDecorator:d,...l}=e,o=d||t.labelDecorator,c=(0,n.default)({"dark:focus:ring-dark-monetr-brand":!e.disabled&&!e.error,"dark:hover:ring-zinc-400":!e.disabled&&!e.error,"dark:ring-dark-monetr-border-string":!e.disabled&&!e.error,"dark:ring-red-500":!e.disabled&&!!e.error,"ring-gray-300":!e.disabled&&!e.error,"ring-red-300":!e.disabled&&!!e.error,uppercase:e.uppercasetext},{"focus:ring-purple-400":!e.error,"focus:ring-red-400":e.error},{"dark:bg-dark-monetr-background":!e.disabled,"dark:text-zinc-200":!e.disabled,"text-gray-900":!e.disabled},{"dark:bg-dark-monetr-background-subtle":e.disabled,"dark:ring-dark-monetr-background-emphasis":e.disabled,"ring-gray-200":e.disabled,"text-gray-500":e.disabled},"block","border-0","focus:ring-2","focus:ring-inset","placeholder:text-gray-400","px-3","py-1.5","ring-1","ring-inset","rounded-lg","shadow-sm","sm:leading-6","text-sm","w-full","dark:caret-zinc-50","min-h-[38px]"),u=(0,n.default)({"pb-[18px]":!e.error},e.className),b=r?.values[e.name];return(0,a.jsxs)("div",{className:u,children:[(0,a.jsx)(i.default,{label:e.label,disabled:e.disabled,htmlFor:e.id,required:e.required,children:(0,a.jsx)(o,{name:e.name,disabled:e.disabled})}),(0,a.jsx)("div",{children:(0,a.jsx)("input",{value:b,onChange:r?.handleChange,onBlur:r?.handleBlur,disabled:r?.isSubmitting||e.disabled,...l,className:c})}),(0,a.jsx)(function(){return e.error?(0,a.jsx)("p",{className:"text-xs font-medium text-red-500 mt-0.5",children:e.error}):null},{})]})}}}]);