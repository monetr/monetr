(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[3967],{7376:(e,t,n)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/development/credentials",function(){return n(9484)}])},9484:(e,t,n)=>{"use strict";n.r(t),n.d(t,{default:()=>h,useTOC:()=>u});var o=n(2540),i=n(7933),r=n(3904),a=n(8439);let s={src:"/_next/static/media/stripe_test_mode.603e08de.png",height:131,width:201,blurDataURL:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAgAAAAFCAMAAABPT11nAAAAJFBMVEX9+fj////l6ez76+D09PTy7u/b3Pfd4OXW2dn42sedmP+Xkf61opmtAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAJklEQVR4nC3BCQ4AIAwCMBCY1///a0zWYo1vAlW2E+KeSFtEI9geDOMAc/HMPgAAAAAASUVORK5CYII=",blurWidth:8,blurHeight:5},d={src:"/_next/static/media/stripe_keys.77334ee2.png",height:208,width:740,blurDataURL:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAgAAAACCAMAAABSSm3fAAAABlBMVEX7+/zv7/DvgxLkAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAD0lEQVR4nGNgZGSAABgNAAA7AAQ6zPVYAAAAAElFTkSuQmCC",blurWidth:8,blurHeight:2},l={src:"/_next/static/media/stripe_new_webhook.cf4f5ed0.png",height:720,width:1280,blurDataURL:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAgAAAAFCAMAAABPT11nAAAACVBMVEX1+Prt8PT+/v43CJF+AAAACXBIWXMAAAsTAAALEwEAmpwYAAAAH0lEQVR4nFXIsQkAIAAEsdzvP7QgWNiFyAQXxfbmQxwCxQAfxHIyqAAAAABJRU5ErkJggg==",blurWidth:8,blurHeight:5},c={src:"/_next/static/media/stripe_created_webhook.0f130653.png",height:184,width:740,blurDataURL:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAgAAAACCAMAAABSSm3fAAAADFBMVEX19ffu7+7p6uv8/f2UyYpKAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAF0lEQVR4nGNgYmJkYmJiZGBgZAABZmYAAM0AFDWNU3wAAAAASUVORK5CYII=",blurWidth:8,blurHeight:2};function u(e){return[{value:"Plaid",id:"plaid",depth:2},{value:"OAuth",id:"oauth",depth:3},{value:"ngrok",id:"ngrok",depth:2},{value:"Stripe",id:"stripe",depth:2},{value:"Secret key",id:"secret-key",depth:3},{value:"Webhook secret",id:"webhook-secret",depth:3},{value:"ReCAPTCHA",id:"recaptcha",depth:2},{value:"Sentry",id:"sentry",depth:2},{value:"SMTP",id:"smtp",depth:2},{value:"Google Cloud KMS",id:"google-cloud-kms",depth:2}]}let h=(0,i.e)(function(e){let{toc:t=u(e)}=e,n={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",img:"img",li:"li",ol:"ol",p:"p",pre:"pre",span:"span",strong:"strong",ul:"ul",...(0,a.R)(),...e.components};return(0,o.jsxs)(o.Fragment,{children:[(0,o.jsx)(n.h1,{children:"3rd Party API credentials for development"}),"\n",(0,o.jsxs)(n.p,{children:["To work on all of the features monetr provides locally, you will need access to several sets of API credentials. These\ncredentials are outlines here in order of significance. monetr or people representing monetr ",(0,o.jsx)(n.strong,{children:"will not"})," provide any of\nthese credentials to you. You are responsible for gaining access to these credentials on your own. None of the\ncredentials require that you pay for them for development purposes."]}),"\n",(0,o.jsx)(n.h2,{id:t[0].id,children:t[0].value}),"\n",(0,o.jsxs)(n.p,{children:["Plaid credentials are ",(0,o.jsx)(n.strong,{children:"required"})," for local development at this time. Until manual accounts are fully supported, only\nlive bank accounts can be used for budgeting within monetr. It is recommended to use Sandbox credentials from Plaid for\nlocal development. The “development” credentials (as Plaid designates them) are for live bank accounts, however they can\nonly be used a limited number of times."]}),"\n",(0,o.jsxs)(n.ol,{children:["\n",(0,o.jsxs)(n.li,{children:["\n",(0,o.jsxs)(n.p,{children:["Start by creating a Plaid account at: ",(0,o.jsx)(n.a,{href:"https://dashboard.plaid.com/signup",children:"Plaid Sign Up"})]}),"\n"]}),"\n",(0,o.jsxs)(n.li,{children:["\n",(0,o.jsxs)(n.p,{children:["Fill out the form to the best of your abilities. Please do not use ",(0,o.jsx)(n.code,{children:"monetr"})," for the company name."]}),"\n"]}),"\n",(0,o.jsxs)(n.li,{children:["\n",(0,o.jsxs)(n.p,{children:["Once you have created your Plaid account, you can find your credentials\nhere: ",(0,o.jsx)(n.a,{href:"https://dashboard.plaid.com/team/keys",children:"Plaid Keys"})]}),"\n"]}),"\n"]}),"\n",(0,o.jsxs)(n.p,{children:["For monetr you will need your ",(0,o.jsx)(n.code,{children:"client_id"})," as well as your ",(0,o.jsx)(n.code,{children:"sandbox"})," secret."]}),"\n",(0,o.jsxs)(n.p,{children:["Add your credentials to the file ",(0,o.jsx)(n.code,{children:"$HOME/.monetr/development.env"}),":"]}),"\n",(0,o.jsx)(n.pre,{tabIndex:"0","data-language":"env","data-word-wrap":"","data-filename":"$HOME/.monetr/development.env",children:(0,o.jsxs)(n.code,{children:[(0,o.jsx)(n.span,{children:(0,o.jsx)(n.span,{children:"PLAID_CLIENT_ID=..."})}),"\n",(0,o.jsx)(n.span,{children:(0,o.jsx)(n.span,{children:"PLAID_CLIENT_SECRET=..."})})]})}),"\n",(0,o.jsx)(n.h3,{id:t[1].id,children:t[1].value}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.strong,{children:"TODO"})}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.a,{href:"https://github.com/monetr/monetr/issues/806",children:(0,o.jsx)(n.img,{src:"https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.github.com%2Frepos%2Fmonetr%2Fmonetr%2Fissues%2F806&query=%24.title&logo=github&label=docs",alt:"GitHub issue/pull request detail"})})}),"\n",(0,o.jsx)(n.h2,{id:t[2].id,children:t[2].value}),"\n",(0,o.jsx)(n.p,{children:"ngrok is used to test code for webhooks. It allows requests to be made to your local development instance from an\nexternal endpoint. You can use ngrok without an API key; however, the tunnels will only last a short amount of time, and\nthe external endpoint will change each time. This might cause difficulty if you plan on working on webhook related\nfeatures. It is recommended to sign up for the free plan of ngrok and use the API key they provide you."}),"\n",(0,o.jsxs)(n.p,{children:["You can sign up for ngrok here: ",(0,o.jsx)(n.a,{href:"https://dashboard.ngrok.com/signup",children:"ngrok Sign Up"})]}),"\n",(0,o.jsx)(n.p,{children:"Add your token and desired ngrok domain to your development environment file:"}),"\n",(0,o.jsx)(n.pre,{tabIndex:"0","data-language":"env","data-word-wrap":"","data-filename":"$HOME/.monetr/development.env",children:(0,o.jsxs)(n.code,{children:[(0,o.jsx)(n.span,{children:(0,o.jsx)(n.span,{children:"NGROK_AUTH=..."})}),"\n",(0,o.jsx)(n.span,{children:(0,o.jsx)(n.span,{children:"NGROK_HOSTNAME=..."})})]})}),"\n",(0,o.jsx)(n.h2,{id:t[3].id,children:t[3].value}),"\n",(0,o.jsxs)(n.p,{children:["If you want to work on billing related features, you can also provide Stripe credentials to the local development\nenvironment. It is required to provide ngrok credentials along-side Stripe for local development. You can sign up for a\nStripe account here: ",(0,o.jsx)(n.a,{href:"https://dashboard.stripe.com/register",children:"Stripe Sign Up"})]}),"\n",(0,o.jsx)(n.p,{children:"You will need two sets of keys to work with Stripe."}),"\n",(0,o.jsxs)(n.ul,{children:["\n",(0,o.jsxs)(n.li,{children:["A ",(0,o.jsx)(n.strong,{children:"test mode"})," Stripe secret key. (Not the public key)"]}),"\n",(0,o.jsx)(n.li,{children:"A webhook secret, configured for your ngrok endpoint and with the proper scopes selected."}),"\n"]}),"\n",(0,o.jsx)(n.p,{children:"Once you have made a Stripe account you can follow this guide to retrieve your keys."}),"\n",(0,o.jsx)(n.h3,{id:t[4].id,children:t[4].value}),"\n",(0,o.jsxs)(n.p,{children:["Navigate to your ",(0,o.jsx)(n.a,{href:"https://dashboard.stripe.com/test/apikeys",children:"Stripe API Keys"})," page within the dashboard. Make sure you\nare in\n“Test mode”."]}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.img,{alt:"Stripe Test Mode",placeholder:"blur",src:s})}),"\n",(0,o.jsxs)(n.p,{children:["You will need to click ",(0,o.jsx)(n.code,{children:"Reveal test key"})," in order to retrieve the API key."]}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.img,{alt:"Stripe Keys",placeholder:"blur",src:d})}),"\n",(0,o.jsx)(n.h3,{id:t[5].id,children:t[5].value}),"\n",(0,o.jsxs)(n.p,{children:["On the ",(0,o.jsx)(n.a,{href:"https://dashboard.stripe.com/test/webhooks",children:"Stripe Webhooks"})," page click ",(0,o.jsx)(n.code,{children:"+ Add endpoint"}),"."]}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.img,{alt:"New Stripe Webhook",placeholder:"blur",src:l})}),"\n",(0,o.jsxs)(n.p,{children:["Enter your ngrok base URL here with the suffix: ",(0,o.jsx)(n.code,{children:"/api/stripe/webhook"})]}),"\n",(0,o.jsx)(n.p,{children:"Then you can add events that you need to work with. At a minimum the following events should be added as monetr requires\nthem."}),"\n",(0,o.jsx)(n.pre,{tabIndex:"0","data-language":"text","data-word-wrap":"","data-filename":"Stripe Webhook Events",children:(0,o.jsxs)(n.code,{children:[(0,o.jsx)(n.span,{children:(0,o.jsx)(n.span,{children:"checkout.session.completed"})}),"\n",(0,o.jsx)(n.span,{children:(0,o.jsx)(n.span,{children:"customer.deleted"})}),"\n",(0,o.jsx)(n.span,{children:(0,o.jsx)(n.span,{children:"customer.subscription.created"})}),"\n",(0,o.jsx)(n.span,{children:(0,o.jsx)(n.span,{children:"customer.subscription.deleted"})}),"\n",(0,o.jsx)(n.span,{children:(0,o.jsx)(n.span,{children:"customer.subscription.updated"})})]})}),"\n",(0,o.jsxs)(n.p,{children:["Once the webhook endpoint has been created click ",(0,o.jsx)(n.code,{children:"Reveal"})," under Signing Secret to retrieve the secret for the webhook\nendpoint."]}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.img,{alt:"Stripe Created Webhook",placeholder:"blur",src:c})}),"\n",(0,o.jsx)(n.h2,{id:t[6].id,children:t[6].value}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.strong,{children:"TODO"})}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.a,{href:"https://github.com/monetr/monetr/issues/805",children:(0,o.jsx)(n.img,{src:"https://img.shields.io/github/issues/detail/state/monetr/monetr/805?label=%23805%20-%20docs%3A%20Document%20ReCAPTCHA%20credentials.&logo=github",alt:"GitHub issue/pull request detail"})})}),"\n",(0,o.jsx)(n.h2,{id:t[7].id,children:t[7].value}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.strong,{children:"TODO"})}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.a,{href:"https://github.com/monetr/monetr/issues/856",children:(0,o.jsx)(n.img,{src:"https://img.shields.io/github/issues/detail/state/monetr/monetr/856?label=%23856%20-%20docs%3A%20Document%20Sentry%20credentials.&logo=github",alt:"GitHub issue/pull request detail"})})}),"\n",(0,o.jsx)(n.h2,{id:t[8].id,children:t[8].value}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.strong,{children:"TODO"})}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.a,{href:"https://github.com/monetr/monetr/issues/857",children:(0,o.jsx)(n.img,{src:"https://img.shields.io/github/issues/detail/state/monetr/monetr/857?label=%23857%20-%20docs%3A%20Document%20SMTP%20credentials.&logo=github",alt:"GitHub issue/pull request detail"})})}),"\n",(0,o.jsx)(n.h2,{id:t[9].id,children:t[9].value}),"\n",(0,o.jsx)(n.p,{children:"Google Cloud KMS support is currently being added to improve the security of storing encrypted secrets in monetr.\nDocumentation to follow."}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.strong,{children:"TODO"})}),"\n",(0,o.jsx)(n.p,{children:(0,o.jsx)(n.a,{href:"https://github.com/monetr/monetr/issues/857",children:(0,o.jsx)(n.img,{src:"https://img.shields.io/github/issues/detail/state/monetr/monetr/936?label=%23936%20-%20docs%3A%20Document%20Google%20Cloud%20KMS%20credentials&logo=github",alt:"GitHub issue/pull request detail"})})})]})},"/documentation/development/credentials",{filePath:"src/pages/documentation/development/credentials.mdx",timestamp:1733607095e3,pageMap:r.O,frontMatter:{},title:"3rd Party API credentials for development"},"undefined"==typeof RemoteContent?u:RemoteContent.useTOC)},8439:(e,t,n)=>{"use strict";n.d(t,{R:()=>d});var o=n(3023),i=n(8209),r=n.n(i),a=n(3696);let s={img:e=>(0,a.createElement)("object"==typeof e.src?r():"img",e)},d=e=>(0,o.R)({...s,...e})},7933:(e,t,n)=>{"use strict";n.d(t,{e:()=>l});var o=n(2540),i=n(2922),r=n(8808);let a=(0,n(3696).createContext)({}),s=a.Provider;a.displayName="SSG";var d=n(8439);function l(e,t,n,o){let r=globalThis[i.VZ];return r.route=t,r.pageMap=n.pageMap,r.context[t]={Content:e,pageOpts:n,useTOC:o},c}function c({__nextra_pageMap:e=[],__nextra_dynamic_opts:t,...n}){let a=globalThis[i.VZ],{Layout:d,themeConfig:l}=a,{route:c,locale:h}=(0,r.r)(),p=a.context[c];if(!p)throw Error(`No content found for the "${c}" route. Please report it as a bug.`);let{pageOpts:m,useTOC:g,Content:f}=p;if(c.startsWith("/["))m.pageMap=e;else for(let{route:t,children:n}of e){let e=t.split("/").slice(h?2:1);(function e(t,[n,...o]){for(let i of t)if("children"in i&&n===i.name)return o.length?e(i.children,o):i})(m.pageMap,e).children=n}if(t){let{title:e,frontMatter:n}=t;m={...m,title:e,frontMatter:n}}return(0,o.jsx)(d,{themeConfig:l,pageOpts:m,pageProps:n,children:(0,o.jsx)(s,{value:n,children:(0,o.jsx)(u,{useTOC:g,children:(0,o.jsx)(f,{...n})})})})}function u({children:e,useTOC:t}){let{wrapper:n}=(0,d.R)();return(0,o.jsx)(h,{useTOC:t,wrapper:n,children:e})}function h({children:e,useTOC:t,wrapper:n,...i}){let r=t(i);return n?(0,o.jsx)(n,{toc:r,children:e}):e}},3904:(e,t,n)=>{"use strict";n.d(t,{O:()=>o});let o=[{data:{index:{type:"page",title:"monetr",display:"hidden",theme:{layout:"raw"}},about:{type:"page",title:"About",theme:{layout:"raw"}},pricing:{type:"page",title:"Pricing",theme:{layout:"raw"}},blog:{type:"page",title:"Blog",theme:{layout:"raw"}},documentation:{type:"page",title:"Documentation"},contact:{type:"page",title:"Contact",display:"hidden"},policy:{type:"page",title:"Policies",display:"hidden"}}},{name:"about",route:"/about",frontMatter:{title:"About"}},{name:"blog",route:"/blog",frontMatter:{title:"Blog"}},{name:"contact",route:"/contact",frontMatter:{sidebarTitle:"Contact"}},{name:"documentation",route:"/documentation",children:[{data:{index:"Introduction","-- Help":{type:"separator",title:"Help"},use:"Using monetr","-- Installation":{type:"separator",title:"Installation"},install:"",configure:"","-- Contributing":{type:"separator",title:"Contributing"},development:""}},{name:"configure",route:"/documentation/configure",children:[{name:"cors",route:"/documentation/configure/cors",frontMatter:{title:"CORS"}},{name:"email",route:"/documentation/configure/email",frontMatter:{sidebarTitle:"Email"}},{name:"kms",route:"/documentation/configure/kms",frontMatter:{title:"Key Management"}},{name:"links",route:"/documentation/configure/links",frontMatter:{sidebarTitle:"Links"}},{name:"logging",route:"/documentation/configure/logging",frontMatter:{sidebarTitle:"Logging"}},{name:"plaid",route:"/documentation/configure/plaid",frontMatter:{sidebarTitle:"Plaid"}},{name:"postgres",route:"/documentation/configure/postgres",frontMatter:{sidebarTitle:"Postgres"}},{name:"recaptcha",route:"/documentation/configure/recaptcha",frontMatter:{title:"ReCAPTCHA"}},{name:"redis",route:"/documentation/configure/redis",frontMatter:{sidebarTitle:"Redis"}},{name:"security",route:"/documentation/configure/security",frontMatter:{sidebarTitle:"Security"}},{name:"sentry",route:"/documentation/configure/sentry",frontMatter:{sidebarTitle:"Sentry"}},{name:"server",route:"/documentation/configure/server",frontMatter:{sidebarTitle:"Server"}},{name:"storage",route:"/documentation/configure/storage",frontMatter:{sidebarTitle:"Storage"}}]},{name:"configure",route:"/documentation/configure",frontMatter:{title:"Configuration",description:"Learn how to configure your self-hosted monetr installation using the comprehensive YAML configuration file. Explore detailed guides for customizing server, database, email, security, and more."}},{name:"development",route:"/documentation/development",children:[{data:{documentation:"",code_of_conduct:"",build:"",local_development:"",credentials:""}},{name:"build",route:"/documentation/development/build",frontMatter:{sidebarTitle:"Build"}},{name:"code_of_conduct",route:"/documentation/development/code_of_conduct",frontMatter:{sidebarTitle:"Code of Conduct"}},{name:"credentials",route:"/documentation/development/credentials",frontMatter:{sidebarTitle:"Credentials"}},{name:"documentation",route:"/documentation/development/documentation",frontMatter:{sidebarTitle:"Documentation"}},{name:"local_development",route:"/documentation/development/local_development",frontMatter:{sidebarTitle:"Local Development"}}]},{name:"development",route:"/documentation/development",frontMatter:{title:"Contributing",description:"Guides on how to contribute to monetr, make changes to the application's code."}},{name:"index",route:"/documentation",frontMatter:{title:"Documentation",description:"Explore the monetr documentation to learn how to get started, host the application, and contribute to development. Find all the resources you need to effectively manage your finances with monetr."}},{name:"install",route:"/documentation/install",children:[{data:{docker:"Docker Compose"}},{name:"docker",route:"/documentation/install/docker",frontMatter:{title:"Self-Host with Docker Compose",description:"Learn how to self-host monetr using Docker Compose. Follow step-by-step instructions to set up monetr, manage updates, and troubleshoot common issues for a seamless self-hosting experience."}}]},{name:"install",route:"/documentation/install",frontMatter:{title:"Self-Hosted Installation",description:"Learn how to self-host monetr for free using Docker or Podman. Explore the benefits of self-hosting and get an overview of installation requirements and options."}},{name:"use",route:"/documentation/use",children:[{data:{getting_started:"Getting Started",funding_schedule:"Funding Schedules",expense:"Expenses",goal:"Goals",transactions:"Transactions",free_to_use:"Free-To-Use",security:"Security"}},{name:"billing",route:"/documentation/use/billing",frontMatter:{title:"Billing",description:"Learn about monetr's billing process, including the 30-day free trial, subscription details, and how to manage or cancel your subscription. Stay informed about payments, access, and managing your account."}},{name:"expense",route:"/documentation/use/expense",frontMatter:{title:"Expenses",description:"Learn how to manage recurring expenses like rent, subscriptions, and credit card payments with monetr. This guide covers creating, tracking, and optimizing expenses to ensure consistent budgeting and predictable Free-To-Use funds."}},{name:"free_to_use",route:"/documentation/use/free_to_use",frontMatter:{sidebarTitle:"Free to Use"}},{name:"funding_schedule",route:"/documentation/use/funding_schedule",frontMatter:{title:"Funding Schedules",description:"Discover how to set up and optimize funding schedules in monetr to manage your budgets effectively. Learn how funding schedules allocate funds for recurring expenses, ensure consistent budgeting, and maintain predictable Free-To-Use funds with every paycheck."}},{name:"getting_started",route:"/documentation/use/getting_started",frontMatter:{title:"Getting Started",description:"Learn how to set up monetr for effective financial management. This guide walks you through connecting your bank account via Plaid or setting up a manual budget, configuring budgets, and creating a funding schedule to take control of your finances."}},{name:"goal",route:"/documentation/use/goal",frontMatter:{title:"Goals",description:"Learn how to use monetr's Goals feature to save for one-time financial targets like vacations, loans, or down payments. Understand how Goals track contributions and spending, helping you plan effectively and meet your financial objectives without over-funding."}},{name:"security",route:"/documentation/use/security",children:[{name:"user_password",route:"/documentation/use/security/user_password",frontMatter:{sidebarTitle:"User Password"}}]},{name:"transactions",route:"/documentation/use/transactions",frontMatter:{sidebarTitle:"Transactions"}}]},{name:"use",route:"/documentation/use",frontMatter:{title:"Using monetr",description:"Discover how to use monetr to effectively manage your finances. Explore guides on setting up your account, managing recurring expenses, creating funding schedules, planning savings goals, and customizing your budget."}}]},{name:"index",route:"/",frontMatter:{title:"monetr: Take Control of Your Finances",description:"Take control of your finances, paycheck by paycheck, with monetr. Put aside what you need, spend what you want, and confidently manage your money with ease. Always know you’ll have enough for your bills and what’s left to save or spend."}},{name:"policy",route:"/policy",children:[{data:{terms:{title:"Terms & Conditions",theme:{sidebar:!1}},privacy:{title:"Privacy Policy",theme:{sidebar:!1}}}},{name:"privacy",route:"/policy/privacy",frontMatter:{sidebarTitle:"Privacy"}},{name:"terms",route:"/policy/terms",frontMatter:{sidebarTitle:"Terms"}}]},{name:"pricing",route:"/pricing",frontMatter:{title:"Pricing"}}]}},e=>{var t=t=>e(e.s=t);e.O(0,[636,6593,8792],()=>t(7376)),_N_E=e.O()}]);