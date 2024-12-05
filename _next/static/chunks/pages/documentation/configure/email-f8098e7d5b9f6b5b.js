(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[474],{9856:(i,s,e)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/configure/email",function(){return e(8773)}])},8773:(i,s,e)=>{"use strict";e.r(s),e.d(s,{default:()=>d,useTOC:()=>a});var n=e(2540),r=e(7933),h=e(3904),l=e(8439),t=e(1785);function a(i){return[{value:"Email Verification Configuration",id:"email-verification-configuration",depth:2},{value:"Forgot Password Configuration",id:"forgot-password-configuration",depth:2},{value:"SMTP Configuration",id:"smtp-configuration",depth:2}]}let d=(0,r.e)(function(i){let{toc:s=a(i)}=i,e={a:"a",code:"code",em:"em",h1:"h1",h2:"h2",p:"p",pre:"pre",span:"span",strong:"strong",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,l.R)(),...i.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(e.h1,{children:"Email/SMTP Configuration"}),"\n",(0,n.jsx)(e.p,{children:"monetr supports sending email notifications (and email verification) if SMTP is configured. Currently emails can be sent\nwhen a user creates a new account, forgets their password, or changes their password."}),"\n",(0,n.jsxs)(e.p,{children:["All email features require that ",(0,n.jsx)(e.code,{children:"enabled"})," is set to ",(0,n.jsx)(e.code,{children:"true"})," and a valid ",(0,n.jsx)(e.code,{children:"smtp"})," config is provided. monetr does not\nsupport specific email APIs and has no plans to. Several email providers offer an SMTP relay, this is monetr’s preferred\nmethod of sending emails as it is the most flexible."]}),"\n",(0,n.jsx)(e.p,{children:"Below is an example of the email/SMTP configuration block:"}),"\n",(0,n.jsx)(e.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,n.jsxs)(e.code,{children:[(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"email"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  enabled"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"true"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  domain"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"example.com"'})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  verification"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": { "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"..."}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:" }   "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"# Email verification configuration"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  forgotPassword"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": { "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"..."}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:" } "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"# Password reset via email link"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  smtp"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": { "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"..."}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:" }           "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"# SMTP configuration"})]})]})}),"\n",(0,n.jsxs)(e.table,{children:[(0,n.jsx)(e.thead,{children:(0,n.jsxs)(e.tr,{children:[(0,n.jsx)(e.th,{children:(0,n.jsx)(e.strong,{children:"Name"})}),(0,n.jsx)(e.th,{children:(0,n.jsx)(e.strong,{children:"Type"})}),(0,n.jsx)(e.th,{children:(0,n.jsx)(e.strong,{children:"Default"})}),(0,n.jsx)(e.th,{children:(0,n.jsx)(e.strong,{children:"Description"})})]})}),(0,n.jsxs)(e.tbody,{children:[(0,n.jsxs)(e.tr,{children:[(0,n.jsx)(e.td,{children:(0,n.jsx)(e.code,{children:"enabled"})}),(0,n.jsx)(e.td,{children:"Boolean"}),(0,n.jsx)(e.td,{children:(0,n.jsx)(e.code,{children:"false"})}),(0,n.jsx)(e.td,{children:"Are email notifications enabled on this server?"})]}),(0,n.jsxs)(e.tr,{children:[(0,n.jsx)(e.td,{children:(0,n.jsx)(e.code,{children:"domain"})}),(0,n.jsx)(e.td,{children:"String"}),(0,n.jsx)(e.td,{}),(0,n.jsxs)(e.td,{children:["Email domain used to send emails, emails will always be sent from ",(0,n.jsx)(e.code,{children:"no-reply@{DOMAIN}"}),"."]})]})]})]}),"\n",(0,n.jsx)(e.h2,{id:s[0].id,children:s[0].value}),"\n",(0,n.jsx)(e.p,{children:"If you want to require users to verify their email address when they create a new login on monetr, you can enable email\nverification. This will email users a link that they must click after creating their login, the link’s lifetime can be\ncustomized if needed."}),"\n",(0,n.jsx)(e.p,{children:"An example of the email verification config:"}),"\n",(0,n.jsx)(e.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,n.jsxs)(e.code,{children:[(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"email"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  verification"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    enabled"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"true"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"      # Can be true or false"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    tokenLifetime"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"10m"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:" # Duration that the verification link should be valid"})]})]})}),"\n",(0,n.jsxs)(e.p,{children:["The token lifetime is parsed using ",(0,n.jsx)(e.a,{href:"https://pkg.go.dev/time#ParseDuration",children:(0,n.jsx)(e.code,{children:"time.ParseDuration(...)"})}),", any value that\ncan be parsed using that function is a valid configuration value."]}),"\n",(0,n.jsxs)(e.table,{children:[(0,n.jsx)(e.thead,{children:(0,n.jsxs)(e.tr,{children:[(0,n.jsx)(e.th,{children:(0,n.jsx)(e.strong,{children:"Name"})}),(0,n.jsx)(e.th,{children:(0,n.jsx)(e.strong,{children:"Type"})}),(0,n.jsx)(e.th,{children:(0,n.jsx)(e.strong,{children:"Default"})}),(0,n.jsx)(e.th,{children:(0,n.jsx)(e.strong,{children:"Description"})})]})}),(0,n.jsxs)(e.tbody,{children:[(0,n.jsxs)(e.tr,{children:[(0,n.jsx)(e.td,{children:(0,n.jsx)(e.code,{children:"enabled"})}),(0,n.jsx)(e.td,{children:"Boolean"}),(0,n.jsx)(e.td,{children:(0,n.jsx)(e.code,{children:"false"})}),(0,n.jsx)(e.td,{children:"Is email verification enabled/required on this server?"})]}),(0,n.jsxs)(e.tr,{children:[(0,n.jsx)(e.td,{children:(0,n.jsx)(e.code,{children:"tokenLifetime"})}),(0,n.jsx)(e.td,{children:"Duration"}),(0,n.jsx)(e.td,{children:(0,n.jsx)(e.code,{children:"10m"})}),(0,n.jsx)(e.td,{children:"How long should the link in the verification email be valid?"})]})]})]}),"\n",(0,n.jsx)(e.h2,{id:s[1].id,children:s[1].value}),"\n",(0,n.jsxs)(e.p,{children:["If you ever lose your password and need to reset it, the easiest way is by using the forgot password form. This will\nsend an email to the user (if a user with that email exists) that includes a link to reset their password. Similar to\nthe ",(0,n.jsx)(e.a,{href:"#email-verification-configuration",children:"Email Verification Configuration"}),", this also only requires an ",(0,n.jsx)(e.code,{children:"enabled"})," and\n",(0,n.jsx)(e.code,{children:"tokenLifetime"})," value."]}),"\n",(0,n.jsx)(e.p,{children:"Example of the forgot password configuration:"}),"\n",(0,n.jsx)(e.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,n.jsxs)(e.code,{children:[(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"email"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  forgotPassword"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    enabled"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"true"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"      # Can be true or false"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    tokenLifetime"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"10m"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:" # Duration that the password reset link should be valid"})]})]})}),"\n",(0,n.jsxs)(e.table,{children:[(0,n.jsx)(e.thead,{children:(0,n.jsxs)(e.tr,{children:[(0,n.jsx)(e.th,{children:(0,n.jsx)(e.strong,{children:"Name"})}),(0,n.jsx)(e.th,{children:(0,n.jsx)(e.strong,{children:"Type"})}),(0,n.jsx)(e.th,{children:(0,n.jsx)(e.strong,{children:"Default"})}),(0,n.jsx)(e.th,{children:(0,n.jsx)(e.strong,{children:"Description"})})]})}),(0,n.jsxs)(e.tbody,{children:[(0,n.jsxs)(e.tr,{children:[(0,n.jsx)(e.td,{children:(0,n.jsx)(e.code,{children:"enabled"})}),(0,n.jsx)(e.td,{children:"Boolean"}),(0,n.jsx)(e.td,{children:(0,n.jsx)(e.code,{children:"false"})}),(0,n.jsx)(e.td,{children:"Are users allowed to reset their password via forgot password?"})]}),(0,n.jsxs)(e.tr,{children:[(0,n.jsx)(e.td,{children:(0,n.jsx)(e.code,{children:"tokenLifetime"})}),(0,n.jsx)(e.td,{children:"Duration"}),(0,n.jsx)(e.td,{children:(0,n.jsx)(e.code,{children:"10m"})}),(0,n.jsx)(e.td,{children:"How long should the password reset link be valid?"})]})]})]}),"\n",(0,n.jsx)(e.h2,{id:s[2].id,children:s[2].value}),"\n",(0,n.jsxs)(e.p,{children:["monetr only supports ",(0,n.jsx)(e.a,{href:"https://datatracker.ietf.org/doc/html/rfc4616",children:"PLAIN SMTP authentication"})," at this time. You can\nobtain all of the necessary details from your preferred email provider."]}),"\n",(0,n.jsx)(t.P,{type:"info",children:(0,n.jsxs)(e.p,{children:["monetr’s SMTP implementation ",(0,n.jsx)(e.em,{children:"requires"})," TLS. Your email provider must support TLS on whatever port specified below."]})}),"\n",(0,n.jsx)(e.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,n.jsxs)(e.code,{children:[(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"email"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  smtp"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    identity"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:" # SMTP Identity"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    username"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:" # SMTP Username"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    password"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:" # SMTP Password or app password depending on provider"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    host"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"     # Domain name of the SMTP server, no protocol or port specified"})]}),"\n",(0,n.jsxs)(e.span,{children:[(0,n.jsx)(e.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    port"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"587"}),(0,n.jsx)(e.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"       # Use the port specified by your provider, could be 587, 465 or 25"})]})]})})]})},"/documentation/configure/email",{filePath:"src/pages/documentation/configure/email.mdx",timestamp:1732468966e3,pageMap:h.O,frontMatter:{},title:"Email/SMTP Configuration"},"undefined"==typeof RemoteContent?a:RemoteContent.useTOC)}},i=>{var s=s=>i(i.s=s);i.O(0,[5684,636,6593,8792],()=>s(9856)),_N_E=i.O()}]);