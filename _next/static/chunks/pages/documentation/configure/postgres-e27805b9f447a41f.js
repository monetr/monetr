(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[4583],{7446:(e,t,n)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/configure/postgres",function(){return n(2556)}])},2556:(e,t,n)=>{"use strict";n.r(t),n.d(t,{default:()=>u,useTOC:()=>s});var o=n(2540),r=n(7933),i=n(3904),a=n(8439);function s(e){return[]}let u=(0,r.e)(function(e){let t={h1:"h1",...(0,a.R)(),...e.components};return(0,o.jsx)(t.h1,{children:"PostgreSQL Configuration"})},"/documentation/configure/postgres",{filePath:"src/pages/documentation/configure/postgres.mdx",timestamp:1732423685e3,pageMap:i.O,frontMatter:{},title:"PostgreSQL Configuration"},"undefined"==typeof RemoteContent?s:RemoteContent.useTOC)},8439:(e,t,n)=>{"use strict";n.d(t,{R:()=>u});var o=n(3023),r=n(8209),i=n.n(r),a=n(3696);let s={img:e=>(0,a.createElement)("object"==typeof e.src?i():"img",e)},u=e=>(0,o.R)({...s,...e})},7933:(e,t,n)=>{"use strict";n.d(t,{e:()=>c});var o=n(2540),r=n(2922),i=n(8808);let a=(0,n(3696).createContext)({}),s=a.Provider;a.displayName="SSG";var u=n(8439);function c(e,t,n,o){let i=globalThis[r.VZ];return i.route=t,i.pageMap=n.pageMap,i.context[t]={Content:e,pageOpts:n,useTOC:o},l}function l({__nextra_pageMap:e=[],__nextra_dynamic_opts:t,...n}){let a=globalThis[r.VZ],{Layout:u,themeConfig:c}=a,{route:l,locale:m}=(0,i.r)(),g=a.context[l];if(!g)throw Error(`No content found for the "${l}" route. Please report it as a bug.`);let{pageOpts:f,useTOC:p,Content:h}=g;if(l.startsWith("/["))f.pageMap=e;else for(let{route:t,children:n}of e){let e=t.split("/").slice(m?2:1);(function e(t,[n,...o]){for(let r of t)if("children"in r&&n===r.name)return o.length?e(r.children,o):r})(f.pageMap,e).children=n}if(t){let{title:e,frontMatter:n}=t;f={...f,title:e,frontMatter:n}}return(0,o.jsx)(u,{themeConfig:c,pageOpts:f,pageProps:n,children:(0,o.jsx)(s,{value:n,children:(0,o.jsx)(d,{useTOC:p,children:(0,o.jsx)(h,{...n})})})})}function d({children:e,useTOC:t}){let{wrapper:n}=(0,u.R)();return(0,o.jsx)(m,{useTOC:t,wrapper:n,children:e})}function m({children:e,useTOC:t,wrapper:n,...r}){let i=t(r);return n?(0,o.jsx)(n,{toc:i,children:e}):e}},3904:(e,t,n)=>{"use strict";n.d(t,{O:()=>o});let o=[{data:{index:{type:"page",title:"monetr",display:"hidden",theme:{layout:"raw"}},about:{type:"page",title:"About",theme:{layout:"raw"}},pricing:{type:"page",title:"Pricing",theme:{layout:"raw"}},blog:{type:"page",title:"Blog",theme:{layout:"raw"}},documentation:{type:"page",title:"Documentation"},contact:{type:"page",title:"Contact",display:"hidden"},policy:{type:"page",title:"Policies",display:"hidden"}}},{name:"about",route:"/about",frontMatter:{title:"About"}},{name:"blog",route:"/blog",frontMatter:{title:"Blog"}},{name:"contact",route:"/contact",frontMatter:{sidebarTitle:"Contact"}},{name:"documentation",route:"/documentation",children:[{data:{index:"Introduction","-- Help":{type:"separator",title:"Help"},use:"Using monetr","-- Installation":{type:"separator",title:"Installation"},install:"",configure:"","-- Contributing":{type:"separator",title:"Contributing"},development:""}},{name:"configure",route:"/documentation/configure",children:[{name:"captcha",route:"/documentation/configure/captcha",frontMatter:{title:"ReCAPTCHA"}},{name:"cors",route:"/documentation/configure/cors",frontMatter:{title:"CORS"}},{name:"email",route:"/documentation/configure/email",frontMatter:{sidebarTitle:"Email"}},{name:"kms",route:"/documentation/configure/kms",frontMatter:{title:"Key Management"}},{name:"links",route:"/documentation/configure/links",frontMatter:{sidebarTitle:"Links"}},{name:"logging",route:"/documentation/configure/logging",frontMatter:{sidebarTitle:"Logging"}},{name:"plaid",route:"/documentation/configure/plaid",frontMatter:{sidebarTitle:"Plaid"}},{name:"postgres",route:"/documentation/configure/postgres",frontMatter:{sidebarTitle:"Postgres"}},{name:"redis",route:"/documentation/configure/redis",frontMatter:{sidebarTitle:"Redis"}},{name:"security",route:"/documentation/configure/security",frontMatter:{sidebarTitle:"Security"}},{name:"sentry",route:"/documentation/configure/sentry",frontMatter:{sidebarTitle:"Sentry"}},{name:"server",route:"/documentation/configure/server",frontMatter:{sidebarTitle:"Server"}},{name:"storage",route:"/documentation/configure/storage",frontMatter:{sidebarTitle:"Storage"}}]},{name:"configure",route:"/documentation/configure",frontMatter:{title:"Configuration",description:"Learn how to configure your self-hosted monetr installation using the comprehensive YAML configuration file. Explore detailed guides for customizing server, database, email, security, and more."}},{name:"development",route:"/documentation/development",children:[{data:{documentation:"",code_of_conduct:"",build:"",local_development:"",credentials:""}},{name:"build",route:"/documentation/development/build",frontMatter:{sidebarTitle:"Build"}},{name:"code_of_conduct",route:"/documentation/development/code_of_conduct",frontMatter:{sidebarTitle:"Code of Conduct"}},{name:"credentials",route:"/documentation/development/credentials",frontMatter:{sidebarTitle:"Credentials"}},{name:"documentation",route:"/documentation/development/documentation",frontMatter:{sidebarTitle:"Documentation"}},{name:"local_development",route:"/documentation/development/local_development",frontMatter:{sidebarTitle:"Local Development"}}]},{name:"development",route:"/documentation/development",frontMatter:{title:"Contributing",description:"Guides on how to contribute to monetr, make changes to the application's code."}},{name:"index",route:"/documentation",frontMatter:{title:"Documentation",description:"Explore the monetr documentation to learn how to get started, host the application, and contribute to development. Find all the resources you need to effectively manage your finances with monetr."}},{name:"install",route:"/documentation/install",children:[{data:{docker:"Docker Compose"}},{name:"docker",route:"/documentation/install/docker",frontMatter:{title:"Self-Host with Docker Compose",description:"Learn how to self-host monetr using Docker Compose. Follow step-by-step instructions to set up monetr, manage updates, and troubleshoot common issues for a seamless self-hosting experience."}}]},{name:"install",route:"/documentation/install",frontMatter:{title:"Self-Hosted Installation",description:"Learn how to self-host monetr for free using Docker or Podman. Explore the benefits of self-hosting and get an overview of installation requirements and options."}},{name:"use",route:"/documentation/use",children:[{data:{getting_started:"Getting Started",funding_schedule:"Funding Schedules",expense:"Expenses",goal:"Goals",transactions:"Transactions",free_to_use:"Free-To-Use",security:"Security"}},{name:"billing",route:"/documentation/use/billing",frontMatter:{title:"Billing",description:"Learn about monetr's billing process, including the 30-day free trial, subscription details, and how to manage or cancel your subscription. Stay informed about payments, access, and managing your account."}},{name:"expense",route:"/documentation/use/expense",frontMatter:{title:"Expenses",description:"Learn how to manage recurring expenses like rent, subscriptions, and credit card payments with monetr. This guide covers creating, tracking, and optimizing expenses to ensure consistent budgeting and predictable Free-To-Use funds."}},{name:"free_to_use",route:"/documentation/use/free_to_use",frontMatter:{sidebarTitle:"Free to Use"}},{name:"funding_schedule",route:"/documentation/use/funding_schedule",frontMatter:{title:"Funding Schedules",description:"Discover how to set up and optimize funding schedules in monetr to manage your budgets effectively. Learn how funding schedules allocate funds for recurring expenses, ensure consistent budgeting, and maintain predictable Free-To-Use funds with every paycheck."}},{name:"getting_started",route:"/documentation/use/getting_started",frontMatter:{title:"Getting Started",description:"Learn how to set up monetr for effective financial management. This guide walks you through connecting your bank account via Plaid or setting up a manual budget, configuring budgets, and creating a funding schedule to take control of your finances."}},{name:"goal",route:"/documentation/use/goal",frontMatter:{title:"Goals",description:"Learn how to use monetr's Goals feature to save for one-time financial targets like vacations, loans, or down payments. Understand how Goals track contributions and spending, helping you plan effectively and meet your financial objectives without overfunding."}},{name:"security",route:"/documentation/use/security",children:[{name:"user_password",route:"/documentation/use/security/user_password",frontMatter:{sidebarTitle:"User Password"}}]},{name:"transactions",route:"/documentation/use/transactions",frontMatter:{sidebarTitle:"Transactions"}}]},{name:"use",route:"/documentation/use",frontMatter:{title:"Using monetr",description:"Discover how to use monetr to effectively manage your finances. Explore guides on setting up your account, managing recurring expenses, creating funding schedules, planning savings goals, and customizing your budget."}}]},{name:"index",route:"/",frontMatter:{title:"monetr: Take Control of Your Finances",description:"Take control of your finances, paycheck by paycheck, with monetr. Put aside what you need, spend what you want, and confidently manage your money with ease. Always know you’ll have enough for your bills and what’s left to save or spend."}},{name:"policy",route:"/policy",children:[{data:{terms:{title:"Terms & Conditions",theme:{sidebar:!1}},privacy:{title:"Privacy Policy",theme:{sidebar:!1}}}},{name:"privacy",route:"/policy/privacy",frontMatter:{sidebarTitle:"Privacy"}},{name:"terms",route:"/policy/terms",frontMatter:{sidebarTitle:"Terms"}}]},{name:"pricing",route:"/pricing",frontMatter:{title:"Pricing"}}]}},e=>{var t=t=>e(e.s=t);e.O(0,[636,6593,8792],()=>t(7446)),_N_E=e.O()}]);