(self.webpackChunk_monetr_stories=self.webpackChunk_monetr_stories||[]).push([["interface_src_pages_expense_details_stories_tsx~interface_src_pages_new_stories_tsx~interface~2cd1a9"],{"../interface/src/pages/register.tsx":function(e,t,s){"use strict";Object.defineProperty(t,"__esModule",{value:!0});!function(e,t){for(var s in t)Object.defineProperty(e,s,{enumerable:!0,get:t[s]})}(t,{RegisterSuccessful:function(){return j},default:function(){return N}});var a=s("../node_modules/react/jsx-runtime.js"),r=s("../node_modules/react/index.js"),l=s("../node_modules/react-router-dom/dist/index.js"),i=s("../node_modules/@tanstack/react-query/build/lib/index.mjs"),n=s("../node_modules/notistack/notistack.esm.js"),o=s.ir(s("../interface/src/components/MButton.tsx")),c=s.ir(s("../interface/src/components/MCaptcha.tsx")),m=s.ir(s("../interface/src/components/MForm.tsx")),d=s.ir(s("../interface/src/components/MLink.tsx")),u=s.ir(s("../interface/src/components/MLogo.tsx")),f=s.ir(s("../interface/src/components/MSpan.tsx")),x=s.ir(s("../interface/src/components/MTextField.tsx")),p=s("../interface/src/hooks/useAppConfiguration.ts"),h=s.ir(s("../interface/src/hooks/useSignUp.ts")),g=s.ir(s("../interface/src/util/verifyEmailAddress.ts"));let w={firstName:"",lastName:"",email:"",password:"",confirmPassword:""};function b(e){let t={};return e?.firstName.length<2&&(t.firstName="First name must have at least 2 characters."),e?.lastName.length<2&&(t.lastName="Last name must have at least 2 characters."),0===e?.email.length&&(t.email="Email must be provided."),e?.email&&!(0,g.default)(e?.email)&&(t.email="Email must be valid."),e?.password.length<8&&(t.password="Password must be at least 8 characters long."),e?.confirmPassword!==e?.password&&(t.confirmPassword="Password confirmation must match."),t}function j(){return(0,a.jsxs)("div",{className:"w-full h-full flex justify-center items-center flex-col",children:[(0,a.jsx)(u.default,{className:"h-24 w-24"}),(0,a.jsx)(f.default,{size:"xl",weight:"medium",className:"max-w-md text-center",children:"A verification message has been sent to your email address, please verify your email."})]})}function N(){let{enqueueSnackbar:e}=(0,n.useSnackbar)(),t=(0,p.useAppConfiguration)(),s=(0,h.default)(),g=(0,l.useNavigate)(),N=(0,i.useQueryClient)(),[y,v]=(0,r.useState)(!1);async function _(t,a){return a.setSubmitting(!0),s({betaCode:t.betaCode,captcha:t.captcha,email:t.email,firstName:t.firstName,lastName:t.lastName,password:t.password,timezone:Intl.DateTimeFormat().resolvedOptions().timeZone}).then(e=>e.requireVerification?v(!0):N.invalidateQueries(["/users/me"]).then(()=>e.nextUrl?g(e.nextUrl):g("/"))).catch(t=>{let s=t.response.data.error||"Failed to sign up.";e(s,{variant:"error",disableWindowBlurListener:!0})}).finally(()=>a.setSubmitting(!1))}return y?(0,a.jsx)(j,{}):(0,a.jsx)("div",{className:"w-full h-full flex pt-10 md:pt-0 md:pb-10 md:justify-center items-center flex-col gap-1 px-5 overflow-y-auto py-4",children:(0,a.jsxs)(m.default,{initialValues:w,validate:b,onSubmit:_,className:"flex flex-col md:w-1/2 lg:w-1/3 xl:w-1/4 items-center",children:[(0,a.jsx)("div",{className:"max-w-[96px] w-full",children:(0,a.jsx)(u.default,{})}),(0,a.jsxs)("div",{className:"flex flex-col items-center text-center",children:[(0,a.jsx)(f.default,{className:"text-5xl",children:"Get Started"}),(0,a.jsx)(f.default,{color:"subtle",className:"text-lg",children:"Create your monetr account now"})]}),(0,a.jsxs)("div",{className:"flex flex-col sm:flex-row gap-2.5 w-full",children:[(0,a.jsx)(x.default,{"data-testid":"register-first-name",autoFocus:!0,label:"First Name",name:"firstName",type:"text",required:!0,className:"w-full"}),(0,a.jsx)(x.default,{"data-testid":"register-last-name",label:"Last Name",name:"lastName",type:"text",required:!0,className:"w-full"})]}),(0,a.jsx)(x.default,{"data-testid":"register-email",label:"Email Address",name:"email",type:"email",required:!0,className:"w-full"}),(0,a.jsx)(x.default,{autoComplete:"new-password",className:"w-full","data-testid":"register-password",label:"Password",name:"password",required:!0,type:"password"}),(0,a.jsx)(x.default,{autoComplete:"new-password",className:"w-full","data-testid":"register-confirm-password",label:"Confirm Password",name:"confirmPassword",required:!0,type:"password"}),(0,a.jsx)(function(){return t?.requireBetaCode?(0,a.jsx)(x.default,{label:"Beta Code",name:"betaCode",type:"text",required:!0,uppercasetext:!0,className:"w-full md:w-1/2 lg:w-1/3 xl:w-1/4"}):null},{}),(0,a.jsx)(c.default,{className:"mb-1",name:"captcha",show:!!t?.verifyRegister}),(0,a.jsx)(o.default,{"data-testid":"register-submit",className:"w-full mt-1",color:"primary",role:"form",type:"submit",variant:"solid",children:"Sign Up"}),(0,a.jsx)("div",{className:"mt-1 flex justify-center gap-1 flex-col md:flex-row items-center",children:(0,a.jsxs)(f.default,{className:"gap-1 inline-block text-center",size:"sm",color:"subtle",component:"p",children:["By signing up you agree to monetr's\xa0",(0,a.jsx)("a",{target:"_blank",className:"text-dark-monetr-blue hover:underline focus:ring-2 focus:ring-dark-monetr-blue focus:underline",href:"https://github.com/monetr/legal/blob/main/TERMS_OF_USE.md",children:"Terms of Use"})," and\xa0",(0,a.jsx)("a",{target:"_blank",className:"text-dark-monetr-blue hover:underline focus:ring-2 focus:ring-dark-monetr-blue focus:underline",href:"https://github.com/monetr/legal/blob/main/PRIVACY.md",children:"Privacy Policy"})]})}),(0,a.jsxs)("div",{className:"mt-1 flex justify-center gap-1 flex-col md:flex-row items-center",children:[(0,a.jsx)(f.default,{color:"subtle",className:"text-sm",children:"Already have an account?"}),(0,a.jsx)(d.default,{to:"/login",size:"sm",children:"Sign in instead"})]})]})})}},"../interface/src/hooks/useSignUp.ts":function(e,t,s){"use strict";Object.defineProperty(t,"__esModule",{value:!0}),Object.defineProperty(t,"default",{enumerable:!0,get:function(){return r}});var a=s.ir(s("../interface/src/util/request.ts"));function r(){return async e=>(0,a.default)().post("/authentication/register",e).then(e=>e.data)}}}]);