(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[474],{9856:(e,s,r)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/configure/email",function(){return r(3244)}])},3244:(e,s,r)=>{"use strict";r.r(s),r.d(s,{__toc:()=>t,default:()=>c});var n=r(2540),i=r(1354),o=r(1369),l=r(4412);let t=[{depth:2,value:"Email Verification Configuration",id:"email-verification-configuration"},{depth:2,value:"Forgot Password Configuration",id:"forgot-password-configuration"},{depth:2,value:"SMTP Configuration",id:"smtp-configuration"}];function a(e){let s=Object.assign({h1:"h1",p:"p",code:"code",pre:"pre",span:"span",table:"table",thead:"thead",tr:"tr",th:"th",strong:"strong",tbody:"tbody",td:"td",h2:"h2",a:"a",em:"em"},(0,o.R)(),e.components);return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(s.h1,{children:"Email/SMTP Configuration"}),"\n",(0,n.jsx)(s.p,{children:"monetr supports sending email notifications (and email verification) if SMTP is configured. Currently emails can be sent\nwhen a user creates a new account, forgets their password, or changes their password."}),"\n",(0,n.jsxs)(s.p,{children:["All email features require that ",(0,n.jsx)(s.code,{children:"enabled"})," is set to ",(0,n.jsx)(s.code,{children:"true"})," and a valid ",(0,n.jsx)(s.code,{children:"smtp"})," config is provided. monetr does not\nsupport specific email APIs and has no plans to. Several email providers offer an SMTP relay, this is monetr's preferred\nmethod of sending emails as it is the most flexible."]}),"\n",(0,n.jsx)(s.p,{children:"Below is an example of the email/SMTP configuration block:"}),"\n",(0,n.jsx)(s.pre,{"data-language":"yaml","data-theme":"default",filename:"config.yaml",children:(0,n.jsxs)(s.code,{"data-language":"yaml","data-theme":"default",children:[(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"email"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"  "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"enabled"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-constant)"},children:"true"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"  "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"domain"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"example.com"'})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"  "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"verification"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" { "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-constant)"},children:"..."}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" }   "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-comment)"},children:"# Email verification configuration"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"  "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"forgotPassword"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" { "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-constant)"},children:"..."}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" } "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-comment)"},children:"# Password reset via email link"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"  "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"smtp"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" { "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-constant)"},children:"..."}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" }           "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-comment)"},children:"# SMTP configuration"})]})]})}),"\n",(0,n.jsxs)(s.table,{children:[(0,n.jsx)(s.thead,{children:(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Name"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Type"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Default"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Description"})})]})}),(0,n.jsxs)(s.tbody,{children:[(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"enabled"})}),(0,n.jsx)(s.td,{children:"Boolean"}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"false"})}),(0,n.jsx)(s.td,{children:"Are email notifications enabled on this server?"})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"domain"})}),(0,n.jsx)(s.td,{children:"String"}),(0,n.jsx)(s.td,{}),(0,n.jsxs)(s.td,{children:["Email domain used to send emails, emails will always be sent from ",(0,n.jsx)(s.code,{children:"no-reply@{DOMAIN}"}),"."]})]})]})]}),"\n",(0,n.jsx)(s.h2,{id:"email-verification-configuration",children:"Email Verification Configuration"}),"\n",(0,n.jsx)(s.p,{children:"If you want to require users to verify their email address when they create a new login on monetr, you can enable email\nverification. This will email users a link that they must click after creating their login, the link's lifetime can be\ncustomized if needed."}),"\n",(0,n.jsx)(s.p,{children:"An example of the email verification config:"}),"\n",(0,n.jsx)(s.pre,{"data-language":"yaml","data-theme":"default",filename:"config.yaml",children:(0,n.jsxs)(s.code,{"data-language":"yaml","data-theme":"default",children:[(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"email"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"  "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"verification"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"    "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"enabled"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-constant)"},children:"true"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"      "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-comment)"},children:"# Can be true or false"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"    "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"tokenLifetime"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-string-expression)"},children:"10m"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-comment)"},children:"# Duration that the verification link should be valid"})]})]})}),"\n",(0,n.jsxs)(s.p,{children:["The token lifetime is parsed using ",(0,n.jsx)(s.a,{href:"https://pkg.go.dev/time#ParseDuration",children:(0,n.jsx)(s.code,{children:"time.ParseDuration(...)"})}),", any value that\ncan be parsed using that function is a valid configuration value."]}),"\n",(0,n.jsxs)(s.table,{children:[(0,n.jsx)(s.thead,{children:(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Name"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Type"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Default"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Description"})})]})}),(0,n.jsxs)(s.tbody,{children:[(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"enabled"})}),(0,n.jsx)(s.td,{children:"Boolean"}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"false"})}),(0,n.jsx)(s.td,{children:"Is email verification enabled/required on this server?"})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"tokenLifetime"})}),(0,n.jsx)(s.td,{children:"Duration"}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"10m"})}),(0,n.jsx)(s.td,{children:"How long should the link in the verification email be valid?"})]})]})]}),"\n",(0,n.jsx)(s.h2,{id:"forgot-password-configuration",children:"Forgot Password Configuration"}),"\n",(0,n.jsxs)(s.p,{children:["If you ever lose your password and need to reset it, the easiest way is by using the forgot password form. This will\nsend an email to the user (if a user with that email exists) that includes a link to reset their password. Similar to\nthe ",(0,n.jsx)(s.a,{href:"#email-verification-configuration",children:"Email Verification Configuration"}),", this also only requires an ",(0,n.jsx)(s.code,{children:"enabled"})," and\n",(0,n.jsx)(s.code,{children:"tokenLifetime"})," value."]}),"\n",(0,n.jsx)(s.p,{children:"Example of the forgot password configuration:"}),"\n",(0,n.jsx)(s.pre,{"data-language":"yaml","data-theme":"default",filename:"config.yaml",children:(0,n.jsxs)(s.code,{"data-language":"yaml","data-theme":"default",children:[(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"email"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"  "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"forgotPassword"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"    "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"enabled"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-constant)"},children:"true"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"      "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-comment)"},children:"# Can be true or false"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"    "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"tokenLifetime"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-string-expression)"},children:"10m"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-comment)"},children:"# Duration that the password reset link should be valid"})]})]})}),"\n",(0,n.jsxs)(s.table,{children:[(0,n.jsx)(s.thead,{children:(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Name"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Type"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Default"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Description"})})]})}),(0,n.jsxs)(s.tbody,{children:[(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"enabled"})}),(0,n.jsx)(s.td,{children:"Boolean"}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"false"})}),(0,n.jsx)(s.td,{children:"Are users allowed to reset their password via forgot password?"})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"tokenLifetime"})}),(0,n.jsx)(s.td,{children:"Duration"}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"10m"})}),(0,n.jsx)(s.td,{children:"How long should the password reset link be valid?"})]})]})]}),"\n",(0,n.jsx)(s.h2,{id:"smtp-configuration",children:"SMTP Configuration"}),"\n",(0,n.jsxs)(s.p,{children:["monetr only supports ",(0,n.jsx)(s.a,{href:"https://datatracker.ietf.org/doc/html/rfc4616",children:"PLAIN SMTP authentication"})," at this time. You can\nobtain all of the necessary details from your preferred email provider."]}),"\n",(0,n.jsx)(l.Pq,{type:"info",children:(0,n.jsxs)(s.p,{children:["monetr's SMTP implementation ",(0,n.jsx)(s.em,{children:"requires"})," TLS. Your email provider must support TLS on whatever port specified below."]})}),"\n",(0,n.jsx)(s.pre,{"data-language":"yaml","data-theme":"default",filename:"config.yaml",children:(0,n.jsxs)(s.code,{"data-language":"yaml","data-theme":"default",children:[(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"email"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"  "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"smtp"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"    "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"identity"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"..."'}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-comment)"},children:"# SMTP Identity"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"    "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"username"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"..."'}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-comment)"},children:"# SMTP Username"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"    "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"password"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"..."'}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-comment)"},children:"# SMTP Password or app password depending on provider"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"    "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"host"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"..."'}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"     "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-comment)"},children:"# Domain name of the SMTP server, no protocol or port specified"})]}),"\n",(0,n.jsxs)(s.span,{className:"line",children:[(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"    "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:"port"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-keyword)"},children:":"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-constant)"},children:"587"}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-color-text)"},children:"       "}),(0,n.jsx)(s.span,{style:{color:"var(--shiki-token-comment)"},children:"# Use the port specified by your provider, could be 587, 465 or 25"})]})]})})]})}let c=(0,i.n)({MDXContent:function(){let e=arguments.length>0&&void 0!==arguments[0]?arguments[0]:{},{wrapper:s}=Object.assign({},(0,o.R)(),e.components);return s?(0,n.jsx)(s,{...e,children:(0,n.jsx)(a,{...e})}):a(e)},pageOpts:{filePath:"src/pages/documentation/configure/email.mdx",route:"/documentation/configure/email",timestamp:1732468966e3,title:"Email/SMTP Configuration",headings:t},pageNextRoute:"/documentation/configure/email"})}},e=>{var s=s=>e(e.s=s);e.O(0,[354,636,593,792],()=>s(9856)),_N_E=e.O()}]);