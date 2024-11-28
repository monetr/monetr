(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[474],{9856:(e,i,t)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/configure/email",function(){return t(8773)}])},8773:(e,i,t)=>{"use strict";t.r(i),t.d(i,{default:()=>d,useTOC:()=>o});var n=t(2540),s=t(7933),r=t(931),a=t(8439),l=t(1785);function o(e){return[{value:"Email Verification Configuration",id:"email-verification-configuration",depth:2},{value:"Forgot Password Configuration",id:"forgot-password-configuration",depth:2},{value:"SMTP Configuration",id:"smtp-configuration",depth:2}]}let d=(0,s.e)(function(e){let{toc:i=o(e)}=e,t={a:"a",code:"code",em:"em",h1:"h1",h2:"h2",p:"p",pre:"pre",span:"span",strong:"strong",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,a.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(t.h1,{children:"Email/SMTP Configuration"}),"\n",(0,n.jsx)(t.p,{children:"monetr supports sending email notifications (and email verification) if SMTP is configured. Currently emails can be sent\nwhen a user creates a new account, forgets their password, or changes their password."}),"\n",(0,n.jsxs)(t.p,{children:["All email features require that ",(0,n.jsx)(t.code,{children:"enabled"})," is set to ",(0,n.jsx)(t.code,{children:"true"})," and a valid ",(0,n.jsx)(t.code,{children:"smtp"})," config is provided. monetr does not\nsupport specific email APIs and has no plans to. Several email providers offer an SMTP relay, this is monetr’s preferred\nmethod of sending emails as it is the most flexible."]}),"\n",(0,n.jsx)(t.p,{children:"Below is an example of the email/SMTP configuration block:"}),"\n",(0,n.jsx)(t.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,n.jsxs)(t.code,{children:[(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"email"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  enabled"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"true"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  domain"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"example.com"'})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  verification"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": { "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"..."}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:" }   "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"# Email verification configuration"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  forgotPassword"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": { "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"..."}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:" } "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"# Password reset via email link"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  smtp"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": { "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"..."}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:" }           "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"# SMTP configuration"})]})]})}),"\n",(0,n.jsxs)(t.table,{children:[(0,n.jsx)(t.thead,{children:(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.th,{children:(0,n.jsx)(t.strong,{children:"Name"})}),(0,n.jsx)(t.th,{children:(0,n.jsx)(t.strong,{children:"Type"})}),(0,n.jsx)(t.th,{children:(0,n.jsx)(t.strong,{children:"Default"})}),(0,n.jsx)(t.th,{children:(0,n.jsx)(t.strong,{children:"Description"})})]})}),(0,n.jsxs)(t.tbody,{children:[(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.td,{children:(0,n.jsx)(t.code,{children:"enabled"})}),(0,n.jsx)(t.td,{children:"Boolean"}),(0,n.jsx)(t.td,{children:(0,n.jsx)(t.code,{children:"false"})}),(0,n.jsx)(t.td,{children:"Are email notifications enabled on this server?"})]}),(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.td,{children:(0,n.jsx)(t.code,{children:"domain"})}),(0,n.jsx)(t.td,{children:"String"}),(0,n.jsx)(t.td,{}),(0,n.jsxs)(t.td,{children:["Email domain used to send emails, emails will always be sent from ",(0,n.jsx)(t.code,{children:"no-reply@{DOMAIN}"}),"."]})]})]})]}),"\n",(0,n.jsx)(t.h2,{id:i[0].id,children:i[0].value}),"\n",(0,n.jsx)(t.p,{children:"If you want to require users to verify their email address when they create a new login on monetr, you can enable email\nverification. This will email users a link that they must click after creating their login, the link’s lifetime can be\ncustomized if needed."}),"\n",(0,n.jsx)(t.p,{children:"An example of the email verification config:"}),"\n",(0,n.jsx)(t.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,n.jsxs)(t.code,{children:[(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"email"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  verification"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    enabled"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"true"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"      # Can be true or false"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    tokenLifetime"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"10m"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:" # Duration that the verification link should be valid"})]})]})}),"\n",(0,n.jsxs)(t.p,{children:["The token lifetime is parsed using ",(0,n.jsx)(t.a,{href:"https://pkg.go.dev/time#ParseDuration",children:(0,n.jsx)(t.code,{children:"time.ParseDuration(...)"})}),", any value that\ncan be parsed using that function is a valid configuration value."]}),"\n",(0,n.jsxs)(t.table,{children:[(0,n.jsx)(t.thead,{children:(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.th,{children:(0,n.jsx)(t.strong,{children:"Name"})}),(0,n.jsx)(t.th,{children:(0,n.jsx)(t.strong,{children:"Type"})}),(0,n.jsx)(t.th,{children:(0,n.jsx)(t.strong,{children:"Default"})}),(0,n.jsx)(t.th,{children:(0,n.jsx)(t.strong,{children:"Description"})})]})}),(0,n.jsxs)(t.tbody,{children:[(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.td,{children:(0,n.jsx)(t.code,{children:"enabled"})}),(0,n.jsx)(t.td,{children:"Boolean"}),(0,n.jsx)(t.td,{children:(0,n.jsx)(t.code,{children:"false"})}),(0,n.jsx)(t.td,{children:"Is email verification enabled/required on this server?"})]}),(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.td,{children:(0,n.jsx)(t.code,{children:"tokenLifetime"})}),(0,n.jsx)(t.td,{children:"Duration"}),(0,n.jsx)(t.td,{children:(0,n.jsx)(t.code,{children:"10m"})}),(0,n.jsx)(t.td,{children:"How long should the link in the verification email be valid?"})]})]})]}),"\n",(0,n.jsx)(t.h2,{id:i[1].id,children:i[1].value}),"\n",(0,n.jsxs)(t.p,{children:["If you ever lose your password and need to reset it, the easiest way is by using the forgot password form. This will\nsend an email to the user (if a user with that email exists) that includes a link to reset their password. Similar to\nthe ",(0,n.jsx)(t.a,{href:"#email-verification-configuration",children:"Email Verification Configuration"}),", this also only requires an ",(0,n.jsx)(t.code,{children:"enabled"})," and\n",(0,n.jsx)(t.code,{children:"tokenLifetime"})," value."]}),"\n",(0,n.jsx)(t.p,{children:"Example of the forgot password configuration:"}),"\n",(0,n.jsx)(t.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,n.jsxs)(t.code,{children:[(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"email"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  forgotPassword"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    enabled"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"true"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"      # Can be true or false"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    tokenLifetime"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"10m"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:" # Duration that the password reset link should be valid"})]})]})}),"\n",(0,n.jsxs)(t.table,{children:[(0,n.jsx)(t.thead,{children:(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.th,{children:(0,n.jsx)(t.strong,{children:"Name"})}),(0,n.jsx)(t.th,{children:(0,n.jsx)(t.strong,{children:"Type"})}),(0,n.jsx)(t.th,{children:(0,n.jsx)(t.strong,{children:"Default"})}),(0,n.jsx)(t.th,{children:(0,n.jsx)(t.strong,{children:"Description"})})]})}),(0,n.jsxs)(t.tbody,{children:[(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.td,{children:(0,n.jsx)(t.code,{children:"enabled"})}),(0,n.jsx)(t.td,{children:"Boolean"}),(0,n.jsx)(t.td,{children:(0,n.jsx)(t.code,{children:"false"})}),(0,n.jsx)(t.td,{children:"Are users allowed to reset their password via forgot password?"})]}),(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.td,{children:(0,n.jsx)(t.code,{children:"tokenLifetime"})}),(0,n.jsx)(t.td,{children:"Duration"}),(0,n.jsx)(t.td,{children:(0,n.jsx)(t.code,{children:"10m"})}),(0,n.jsx)(t.td,{children:"How long should the password reset link be valid?"})]})]})]}),"\n",(0,n.jsx)(t.h2,{id:i[2].id,children:i[2].value}),"\n",(0,n.jsxs)(t.p,{children:["monetr only supports ",(0,n.jsx)(t.a,{href:"https://datatracker.ietf.org/doc/html/rfc4616",children:"PLAIN SMTP authentication"})," at this time. You can\nobtain all of the necessary details from your preferred email provider."]}),"\n",(0,n.jsx)(l.P,{type:"info",children:(0,n.jsxs)(t.p,{children:["monetr’s SMTP implementation ",(0,n.jsx)(t.em,{children:"requires"})," TLS. Your email provider must support TLS on whatever port specified below."]})}),"\n",(0,n.jsx)(t.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,n.jsxs)(t.code,{children:[(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"email"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  smtp"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    identity"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:" # SMTP Identity"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    username"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:" # SMTP Username"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    password"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:" # SMTP Password or app password depending on provider"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    host"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"     # Domain name of the SMTP server, no protocol or port specified"})]}),"\n",(0,n.jsxs)(t.span,{children:[(0,n.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    port"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"587"}),(0,n.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"       # Use the port specified by your provider, could be 587, 465 or 25"})]})]})})]})},"/documentation/configure/email",{filePath:"src/pages/documentation/configure/email.mdx",timestamp:1732468966e3,pageMap:r.O,frontMatter:{},title:"Email/SMTP Configuration"},"undefined"==typeof RemoteContent?o:RemoteContent.useTOC)},1785:(e,i,t)=>{"use strict";t.d(i,{P:()=>o});var n=t(2540),s=t(1750),r=t(6877);let a={default:"\uD83D\uDCA1",error:"\uD83D\uDEAB",info:(0,n.jsx)(r.KS,{className:"_mt-1"}),warning:"⚠️"},l={default:(0,s.A)("_border-orange-100 _bg-orange-50 _text-orange-800 dark:_border-orange-400/30 dark:_bg-orange-400/20 dark:_text-orange-300"),error:(0,s.A)("_border-red-200 _bg-red-100 _text-red-900 dark:_border-red-200/30 dark:_bg-red-900/30 dark:_text-red-200"),info:(0,s.A)("_border-blue-200 _bg-blue-100 _text-blue-900 dark:_border-blue-200/30 dark:_bg-blue-900/30 dark:_text-blue-200"),warning:(0,s.A)("_border-yellow-100 _bg-yellow-50 _text-yellow-900 dark:_border-yellow-200/30 dark:_bg-yellow-700/30 dark:_text-yellow-200")};function o({children:e,type:i="default",emoji:t=a[i]}){return(0,n.jsxs)("div",{className:(0,s.A)("nextra-callout _overflow-x-auto _mt-6 _flex _rounded-lg _border _py-2 ltr:_pr-4 rtl:_pl-4","contrast-more:_border-current contrast-more:dark:_border-current",l[i]),children:[(0,n.jsx)("div",{className:"_select-none _text-xl ltr:_pl-3 ltr:_pr-2 rtl:_pr-3 rtl:_pl-2",style:{fontFamily:'"Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"'},children:t}),(0,n.jsx)("div",{className:"_w-full _min-w-0 _leading-7",children:e})]})}},8439:(e,i,t)=>{"use strict";t.d(i,{R:()=>o});var n=t(3023),s=t(8209),r=t.n(s),a=t(3696);let l={img:e=>(0,a.createElement)("object"==typeof e.src?r():"img",e)},o=e=>(0,n.R)({...l,...e})},7933:(e,i,t)=>{"use strict";t.d(i,{e:()=>d});var n=t(2540),s=t(2922),r=t(8808);let a=(0,t(3696).createContext)({}),l=a.Provider;a.displayName="SSG";var o=t(8439);function d(e,i,t,n){let r=globalThis[s.VZ];return r.route=i,r.pageMap=t.pageMap,r.context[i]={Content:e,pageOpts:t,useTOC:n},h}function h({__nextra_pageMap:e=[],__nextra_dynamic_opts:i,...t}){let a=globalThis[s.VZ],{Layout:o,themeConfig:d}=a,{route:h,locale:u}=(0,r.r)(),p=a.context[h];if(!p)throw Error(`No content found for the "${h}" route. Please report it as a bug.`);let{pageOpts:m,useTOC:k,Content:g}=p;if(h.startsWith("/["))m.pageMap=e;else for(let{route:i,children:t}of e){let e=i.split("/").slice(u?2:1);(function e(i,[t,...n]){for(let s of i)if("children"in s&&t===s.name)return n.length?e(s.children,n):s})(m.pageMap,e).children=t}if(i){let{title:e,frontMatter:t}=i;m={...m,title:e,frontMatter:t}}return(0,n.jsx)(o,{themeConfig:d,pageOpts:m,pageProps:t,children:(0,n.jsx)(l,{value:t,children:(0,n.jsx)(c,{useTOC:k,children:(0,n.jsx)(g,{...t})})})})}function c({children:e,useTOC:i}){let{wrapper:t}=(0,o.R)();return(0,n.jsx)(u,{useTOC:i,wrapper:t,children:e})}function u({children:e,useTOC:i,wrapper:t,...s}){let r=i(s);return t?(0,n.jsx)(t,{toc:r,children:e}):e}},931:(e,i,t)=>{"use strict";t.d(i,{O:()=>n});let n=[{data:{index:{type:"page",title:"monetr",display:"hidden",theme:{layout:"raw"}},about:{type:"page",title:"About",theme:{layout:"raw"}},pricing:{type:"page",title:"Pricing",theme:{layout:"raw"}},blog:{type:"page",title:"Blog",theme:{layout:"raw"}},documentation:{type:"page",title:"Documentation"},contact:{type:"page",title:"Contact",display:"hidden"},policy:{type:"page",title:"Policies",display:"hidden"}}},{name:"about",route:"/about",frontMatter:{title:"About"}},{name:"blog",route:"/blog",frontMatter:{title:"Blog"}},{name:"contact",route:"/contact",frontMatter:{sidebarTitle:"Contact"}},{name:"documentation",route:"/documentation",children:[{data:{index:"Introduction","-- Help":{type:"separator",title:"Help"},use:"Using monetr","-- Installation":{type:"separator",title:"Installation"},install:"",configure:"","-- Contributing":{type:"separator",title:"Contributing"},development:""}},{name:"configure",route:"/documentation/configure",children:[{name:"captcha",route:"/documentation/configure/captcha",frontMatter:{title:"ReCAPTCHA"}},{name:"cors",route:"/documentation/configure/cors",frontMatter:{title:"CORS"}},{name:"email",route:"/documentation/configure/email",frontMatter:{sidebarTitle:"Email"}},{name:"kms",route:"/documentation/configure/kms",frontMatter:{title:"Key Management"}},{name:"links",route:"/documentation/configure/links",frontMatter:{sidebarTitle:"Links"}},{name:"logging",route:"/documentation/configure/logging",frontMatter:{sidebarTitle:"Logging"}},{name:"plaid",route:"/documentation/configure/plaid",frontMatter:{sidebarTitle:"Plaid"}},{name:"postgres",route:"/documentation/configure/postgres",frontMatter:{sidebarTitle:"Postgres"}},{name:"redis",route:"/documentation/configure/redis",frontMatter:{sidebarTitle:"Redis"}},{name:"security",route:"/documentation/configure/security",frontMatter:{sidebarTitle:"Security"}},{name:"sentry",route:"/documentation/configure/sentry",frontMatter:{sidebarTitle:"Sentry"}},{name:"server",route:"/documentation/configure/server",frontMatter:{sidebarTitle:"Server"}},{name:"storage",route:"/documentation/configure/storage",frontMatter:{sidebarTitle:"Storage"}}]},{name:"configure",route:"/documentation/configure",frontMatter:{title:"Configuration",description:"Configure self-hosted monetr servers"}},{name:"development",route:"/documentation/development",children:[{data:{documentation:"",code_of_conduct:"",build:"",local_development:"",credentials:""}},{name:"build",route:"/documentation/development/build",frontMatter:{sidebarTitle:"Build"}},{name:"code_of_conduct",route:"/documentation/development/code_of_conduct",frontMatter:{sidebarTitle:"Code of Conduct"}},{name:"credentials",route:"/documentation/development/credentials",frontMatter:{sidebarTitle:"Credentials"}},{name:"documentation",route:"/documentation/development/documentation",frontMatter:{sidebarTitle:"Documentation"}},{name:"local_development",route:"/documentation/development/local_development",frontMatter:{sidebarTitle:"Local Development"}}]},{name:"development",route:"/documentation/development",frontMatter:{title:"Contributing",description:"Guides on how to contribute to monetr, make changes to the application's code."}},{name:"index",route:"/documentation",frontMatter:{title:"Documentation",description:"Guides on how to use, self-host, or develop against monetr."}},{name:"install",route:"/documentation/install",children:[{name:"docker",route:"/documentation/install/docker",frontMatter:{title:"Self-Host via Docker",description:"Self-host monetr via Docker containers"}}]},{name:"install",route:"/documentation/install",frontMatter:{title:"Self-Host Installation",description:"Options on how to run monetr yourself for free."}},{name:"use",route:"/documentation/use",children:[{data:{starting_fresh:"Starting Fresh",funding_schedule:"Funding Schedules",expense:"Expenses",goal:"Goals",free_to_use:"Free-To-Use",security:"Security"}},{name:"expense",route:"/documentation/use/expense",frontMatter:{title:"Expenses",description:"Keep track of your regular or planned spending easily using expenses."}},{name:"free_to_use",route:"/documentation/use/free_to_use",frontMatter:{sidebarTitle:"Free to Use"}},{name:"funding_schedule",route:"/documentation/use/funding_schedule",frontMatter:{title:"Funding Schedules",description:"Contribute to your budgets on a regular basis, like every time you get paid."}},{name:"goal",route:"/documentation/use/goal",frontMatter:{sidebarTitle:"Goal"}},{name:"security",route:"/documentation/use/security",children:[{name:"user_password",route:"/documentation/use/security/user_password",frontMatter:{sidebarTitle:"User Password"}}]},{name:"starting_fresh",route:"/documentation/use/starting_fresh",frontMatter:{sidebarTitle:"Starting Fresh"}}]},{name:"use",route:"/documentation/use",frontMatter:{title:"Using monetr",description:"How to use and get the most out of monetr"}}]},{name:"index",route:"/",frontMatter:{title:"monetr",description:"Always know what you can spend. Put a bit of money aside every time you get paid. Always be sure you'll have enough to cover your bills, and know what you have left-over to save or spend on whatever you'd like."}},{name:"policy",route:"/policy",children:[{data:{terms:{title:"Terms & Conditions",theme:{sidebar:!1}},privacy:{title:"Privacy Policy",theme:{sidebar:!1}}}},{name:"privacy",route:"/policy/privacy",frontMatter:{sidebarTitle:"Privacy"}},{name:"terms",route:"/policy/terms",frontMatter:{sidebarTitle:"Terms"}}]},{name:"pricing",route:"/pricing",frontMatter:{title:"Pricing"}}]}},e=>{var i=i=>e(e.s=i);e.O(0,[636,593,792],()=>i(9856)),_N_E=e.O()}]);