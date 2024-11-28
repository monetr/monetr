(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[857],{4194:(e,t,n)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/configure/cors",function(){return n(3706)}])},3706:(e,t,n)=>{"use strict";n.r(t),n.d(t,{default:()=>l,useTOC:()=>a});var r=n(2540),i=n(7933),o=n(931),s=n(8439);function a(e){return[]}let l=(0,i.e)(function(e){let t={code:"code",h1:"h1",p:"p",pre:"pre",span:"span",strong:"strong",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,s.R)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(t.h1,{children:"CORS (Cross Origin Resource Sharing)"}),"\n",(0,r.jsx)(t.p,{children:"monetr generally is hosted on a single domain name and thus does not require CORS, however if your self hosting setup\nrequires that your monetr instance be accessible from another domain name then you must configure CORS."}),"\n",(0,r.jsx)(t.p,{children:"Below is an example of the CORS configuration block:"}),"\n",(0,r.jsx)(t.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,r.jsxs)(t.code,{children:[(0,r.jsxs)(t.span,{children:[(0,r.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"cors"}),(0,r.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,r.jsx)(t.span,{children:(0,r.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"  # allowedOrigins determines the value of the `Access-Control-Allow-Origin` response header. In monetr this defaults to"})}),"\n",(0,r.jsx)(t.span,{children:(0,r.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"  # an empty list. This default forbids all cross origin access."})}),"\n",(0,r.jsxs)(t.span,{children:[(0,r.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  allowedOrigins"}),(0,r.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "})]}),"\n",(0,r.jsxs)(t.span,{children:[(0,r.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:"    - "}),(0,r.jsx)(t.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"https://your.monetr.local"})]}),"\n",(0,r.jsx)(t.span,{children:(0,r.jsx)(t.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"  # Enable debug logging to help diagnose CORS issues."})}),"\n",(0,r.jsxs)(t.span,{children:[(0,r.jsx)(t.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  debug"}),(0,r.jsx)(t.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,r.jsx)(t.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"true"})]})]})}),"\n",(0,r.jsxs)(t.table,{children:[(0,r.jsx)(t.thead,{children:(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.th,{children:(0,r.jsx)(t.strong,{children:"Name"})}),(0,r.jsx)(t.th,{children:(0,r.jsx)(t.strong,{children:"Type"})}),(0,r.jsx)(t.th,{children:(0,r.jsx)(t.strong,{children:"Default"})}),(0,r.jsx)(t.th,{children:(0,r.jsx)(t.strong,{children:"Description"})})]})}),(0,r.jsxs)(t.tbody,{children:[(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:(0,r.jsx)(t.code,{children:"allowedOrigins"})}),(0,r.jsx)(t.td,{children:"Array"}),(0,r.jsx)(t.td,{children:(0,r.jsx)(t.code,{children:"[]"})}),(0,r.jsx)(t.td,{children:"Other origins that are allowed to access your monetr server."})]}),(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:(0,r.jsx)(t.code,{children:"debug"})}),(0,r.jsx)(t.td,{children:"Boolean"}),(0,r.jsx)(t.td,{children:(0,r.jsx)(t.code,{children:"false"})}),(0,r.jsx)(t.td,{children:"Debug logging for helping diagnose CORS issues."})]})]})]}),"\n",(0,r.jsx)(t.p,{children:"The following environment variables can be used to configure CORS options:"}),"\n",(0,r.jsxs)(t.table,{children:[(0,r.jsx)(t.thead,{children:(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.th,{children:"Variable"}),(0,r.jsx)(t.th,{children:"Config File Field"})]})}),(0,r.jsxs)(t.tbody,{children:[(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:(0,r.jsx)(t.code,{children:"MONETR_CORS_ALLOWED_ORIGINS"})}),(0,r.jsx)(t.td,{children:(0,r.jsx)(t.code,{children:"cors.allowedOrigins"})})]}),(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:(0,r.jsx)(t.code,{children:"MONETR_CORS_DEBUG"})}),(0,r.jsx)(t.td,{children:(0,r.jsx)(t.code,{children:"cors.debug"})})]})]})]})]})},"/documentation/configure/cors",{filePath:"src/pages/documentation/configure/cors.mdx",timestamp:1732468966e3,pageMap:o.O,frontMatter:{title:"CORS"},title:"CORS"},"undefined"==typeof RemoteContent?a:RemoteContent.useTOC)},8439:(e,t,n)=>{"use strict";n.d(t,{R:()=>l});var r=n(3023),i=n(8209),o=n.n(i),s=n(3696);let a={img:e=>(0,s.createElement)("object"==typeof e.src?o():"img",e)},l=e=>(0,r.R)({...a,...e})},7933:(e,t,n)=>{"use strict";n.d(t,{e:()=>d});var r=n(2540),i=n(2922),o=n(8808);let s=(0,n(3696).createContext)({}),a=s.Provider;s.displayName="SSG";var l=n(8439);function d(e,t,n,r){let o=globalThis[i.VZ];return o.route=t,o.pageMap=n.pageMap,o.context[t]={Content:e,pageOpts:n,useTOC:r},c}function c({__nextra_pageMap:e=[],__nextra_dynamic_opts:t,...n}){let s=globalThis[i.VZ],{Layout:l,themeConfig:d}=s,{route:c,locale:h}=(0,o.r)(),m=s.context[c];if(!m)throw Error(`No content found for the "${c}" route. Please report it as a bug.`);let{pageOpts:g,useTOC:p,Content:f}=m;if(c.startsWith("/["))g.pageMap=e;else for(let{route:t,children:n}of e){let e=t.split("/").slice(h?2:1);(function e(t,[n,...r]){for(let i of t)if("children"in i&&n===i.name)return r.length?e(i.children,r):i})(g.pageMap,e).children=n}if(t){let{title:e,frontMatter:n}=t;g={...g,title:e,frontMatter:n}}return(0,r.jsx)(l,{themeConfig:d,pageOpts:g,pageProps:n,children:(0,r.jsx)(a,{value:n,children:(0,r.jsx)(u,{useTOC:p,children:(0,r.jsx)(f,{...n})})})})}function u({children:e,useTOC:t}){let{wrapper:n}=(0,l.R)();return(0,r.jsx)(h,{useTOC:t,wrapper:n,children:e})}function h({children:e,useTOC:t,wrapper:n,...i}){let o=t(i);return n?(0,r.jsx)(n,{toc:o,children:e}):e}},931:(e,t,n)=>{"use strict";n.d(t,{O:()=>r});let r=[{data:{index:{type:"page",title:"monetr",display:"hidden",theme:{layout:"raw"}},about:{type:"page",title:"About",theme:{layout:"raw"}},pricing:{type:"page",title:"Pricing",theme:{layout:"raw"}},blog:{type:"page",title:"Blog",theme:{layout:"raw"}},documentation:{type:"page",title:"Documentation"},contact:{type:"page",title:"Contact",display:"hidden"},policy:{type:"page",title:"Policies",display:"hidden"}}},{name:"about",route:"/about",frontMatter:{title:"About"}},{name:"blog",route:"/blog",frontMatter:{title:"Blog"}},{name:"contact",route:"/contact",frontMatter:{sidebarTitle:"Contact"}},{name:"documentation",route:"/documentation",children:[{data:{index:"Introduction","-- Help":{type:"separator",title:"Help"},use:"Using monetr","-- Installation":{type:"separator",title:"Installation"},install:"",configure:"","-- Contributing":{type:"separator",title:"Contributing"},development:""}},{name:"configure",route:"/documentation/configure",children:[{name:"captcha",route:"/documentation/configure/captcha",frontMatter:{title:"ReCAPTCHA"}},{name:"cors",route:"/documentation/configure/cors",frontMatter:{title:"CORS"}},{name:"email",route:"/documentation/configure/email",frontMatter:{sidebarTitle:"Email"}},{name:"kms",route:"/documentation/configure/kms",frontMatter:{title:"Key Management"}},{name:"links",route:"/documentation/configure/links",frontMatter:{sidebarTitle:"Links"}},{name:"logging",route:"/documentation/configure/logging",frontMatter:{sidebarTitle:"Logging"}},{name:"plaid",route:"/documentation/configure/plaid",frontMatter:{sidebarTitle:"Plaid"}},{name:"postgres",route:"/documentation/configure/postgres",frontMatter:{sidebarTitle:"Postgres"}},{name:"redis",route:"/documentation/configure/redis",frontMatter:{sidebarTitle:"Redis"}},{name:"security",route:"/documentation/configure/security",frontMatter:{sidebarTitle:"Security"}},{name:"sentry",route:"/documentation/configure/sentry",frontMatter:{sidebarTitle:"Sentry"}},{name:"server",route:"/documentation/configure/server",frontMatter:{sidebarTitle:"Server"}},{name:"storage",route:"/documentation/configure/storage",frontMatter:{sidebarTitle:"Storage"}}]},{name:"configure",route:"/documentation/configure",frontMatter:{title:"Configuration",description:"Configure self-hosted monetr servers"}},{name:"development",route:"/documentation/development",children:[{data:{documentation:"",code_of_conduct:"",build:"",local_development:"",credentials:""}},{name:"build",route:"/documentation/development/build",frontMatter:{sidebarTitle:"Build"}},{name:"code_of_conduct",route:"/documentation/development/code_of_conduct",frontMatter:{sidebarTitle:"Code of Conduct"}},{name:"credentials",route:"/documentation/development/credentials",frontMatter:{sidebarTitle:"Credentials"}},{name:"documentation",route:"/documentation/development/documentation",frontMatter:{sidebarTitle:"Documentation"}},{name:"local_development",route:"/documentation/development/local_development",frontMatter:{sidebarTitle:"Local Development"}}]},{name:"development",route:"/documentation/development",frontMatter:{title:"Contributing",description:"Guides on how to contribute to monetr, make changes to the application's code."}},{name:"index",route:"/documentation",frontMatter:{title:"Documentation",description:"Guides on how to use, self-host, or develop against monetr."}},{name:"install",route:"/documentation/install",children:[{name:"docker",route:"/documentation/install/docker",frontMatter:{title:"Self-Host via Docker",description:"Self-host monetr via Docker containers"}}]},{name:"install",route:"/documentation/install",frontMatter:{title:"Self-Host Installation",description:"Options on how to run monetr yourself for free."}},{name:"use",route:"/documentation/use",children:[{data:{starting_fresh:"Starting Fresh",funding_schedule:"Funding Schedules",expense:"Expenses",goal:"Goals",free_to_use:"Free-To-Use",security:"Security"}},{name:"expense",route:"/documentation/use/expense",frontMatter:{title:"Expenses",description:"Keep track of your regular or planned spending easily using expenses."}},{name:"free_to_use",route:"/documentation/use/free_to_use",frontMatter:{sidebarTitle:"Free to Use"}},{name:"funding_schedule",route:"/documentation/use/funding_schedule",frontMatter:{title:"Funding Schedules",description:"Contribute to your budgets on a regular basis, like every time you get paid."}},{name:"goal",route:"/documentation/use/goal",frontMatter:{sidebarTitle:"Goal"}},{name:"security",route:"/documentation/use/security",children:[{name:"user_password",route:"/documentation/use/security/user_password",frontMatter:{sidebarTitle:"User Password"}}]},{name:"starting_fresh",route:"/documentation/use/starting_fresh",frontMatter:{sidebarTitle:"Starting Fresh"}}]},{name:"use",route:"/documentation/use",frontMatter:{title:"Using monetr",description:"How to use and get the most out of monetr"}}]},{name:"index",route:"/",frontMatter:{title:"monetr",description:"Always know what you can spend. Put a bit of money aside every time you get paid. Always be sure you'll have enough to cover your bills, and know what you have left-over to save or spend on whatever you'd like."}},{name:"policy",route:"/policy",children:[{data:{terms:{title:"Terms & Conditions",theme:{sidebar:!1}},privacy:{title:"Privacy Policy",theme:{sidebar:!1}}}},{name:"privacy",route:"/policy/privacy",frontMatter:{sidebarTitle:"Privacy"}},{name:"terms",route:"/policy/terms",frontMatter:{sidebarTitle:"Terms"}}]},{name:"pricing",route:"/pricing",frontMatter:{title:"Pricing"}}]}},e=>{var t=t=>e(e.s=t);e.O(0,[636,593,792],()=>t(4194)),_N_E=e.O()}]);