(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[2428],{1094:function(e,t,n){(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/development/credentials",function(){return n(4652)}])},4652:function(e,t,n){"use strict";n.r(t),n.d(t,{__toc:function(){return c},default:function(){return u}});var s=n(4246),i=n(9304),r=n(1441),o={src:"/_next/static/media/stripe_test_mode.603e08de.png",height:131,width:201,blurDataURL:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAgAAAAFCAIAAAD38zoCAAAAX0lEQVR42l3JNw6AMBAAQf//I1QUvIgsQGdwwAmnIwgaVtMtafsOVgorXyibYaNMjjMY50ndqnFy2lys0kbtSnBxhEjKIlclOmkDBy+YG5qwASKSHMMnvlLK98j493QCZqZyNtzv/+UAAAAASUVORK5CYII=",blurWidth:8,blurHeight:5},l={src:"/_next/static/media/stripe_keys.77334ee2.png",height:208,width:740,blurDataURL:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAgAAAACCAIAAADq9gq6AAAAK0lEQVR42iXHwQ0AIAzDwO6/bqMSQYiAci87MhOApPV4N9tRNWb7f4OkpAOS5y7xiD+PUgAAAABJRU5ErkJggg==",blurWidth:8,blurHeight:2},d={src:"/_next/static/media/stripe_new_webhook.cf4f5ed0.png",height:720,width:1280,blurDataURL:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAgAAAAFCAIAAAD38zoCAAAAWUlEQVR42i2KURLFIAgDvf9h2yoQeSjxMdNm8rW7zT3MkHuT5xHvCoHXG7BEfK3IzGG4upTT+Ws7V35j1/lS82iqGwgyST5qAi9art03gTznkKz8GmPYNI8/W050GdOaC8wAAAAASUVORK5CYII=",blurWidth:8,blurHeight:5},a={src:"/_next/static/media/stripe_created_webhook.0f130653.png",height:184,width:740,blurDataURL:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAgAAAACCAIAAADq9gq6AAAAMElEQVR42hXEyQkAMAgEQPsv1uAB6xEwcR5DzMeWq5qaify8qsmhQEQUUJl7952ZB3eyLiPtSgFcAAAAAElFTkSuQmCC",blurWidth:8,blurHeight:2};let c=[{depth:2,value:"Plaid",id:"plaid"},{depth:3,value:"OAuth",id:"oauth"},{depth:2,value:"Teller",id:"teller"},{depth:2,value:"ngrok",id:"ngrok"},{depth:2,value:"Stripe",id:"stripe"},{depth:3,value:"Secret key",id:"secret-key"},{depth:3,value:"Webhook secret",id:"webhook-secret"},{depth:2,value:"ReCAPTCHA",id:"recaptcha"},{depth:2,value:"Sentry",id:"sentry"},{depth:2,value:"SMTP",id:"smtp"},{depth:2,value:"Google Cloud KMS",id:"google-cloud-kms"}];function h(e){let t=Object.assign({h1:"h1",p:"p",strong:"strong",h2:"h2",ol:"ol",li:"li",a:"a",code:"code",h3:"h3",img:"img",ul:"ul",pre:"pre",span:"span"},(0,r.a)(),e.components);return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)(t.h1,{children:"3rd Party API credentials for development"}),"\n",(0,s.jsxs)(t.p,{children:["To work on all of the features monetr provides locally, you will need access to several sets of API credentials. These\ncredentials are outlines here in order of significance. monetr or people representing monetr ",(0,s.jsx)(t.strong,{children:"will not"})," provide any of\nthese credentials to you. You are responsible for gaining access to these credentials on your own. None of the\ncredentials require that you pay for them for development purposes."]}),"\n",(0,s.jsx)(t.h2,{id:"plaid",children:"Plaid"}),"\n",(0,s.jsxs)(t.p,{children:["Plaid credentials are ",(0,s.jsx)(t.strong,{children:"required"}),' for local development at this time. Until manual accounts are fully supported, only\nlive bank accounts can be used for budgeting within monetr. It is recommended to use Sandbox credentials from Plaid for\nlocal development. The "development" credentials (as Plaid designates them) are for live bank accounts, however they can\nonly be used a limited number of times.']}),"\n",(0,s.jsxs)(t.ol,{children:["\n",(0,s.jsxs)(t.li,{children:["\n",(0,s.jsxs)(t.p,{children:["Start by creating a Plaid account at: ",(0,s.jsx)(t.a,{href:"https://dashboard.plaid.com/signup",children:"Plaid Signup"})]}),"\n"]}),"\n",(0,s.jsxs)(t.li,{children:["\n",(0,s.jsxs)(t.p,{children:["Fill out the form to the best of your abilities. Please do not use ",(0,s.jsx)(t.code,{children:"monetr"})," for the company name."]}),"\n"]}),"\n",(0,s.jsxs)(t.li,{children:["\n",(0,s.jsxs)(t.p,{children:["Once you have created your Plaid account, you can find your credentials\nhere: ",(0,s.jsx)(t.a,{href:"https://dashboard.plaid.com/team/keys",children:"Plaid Keys"})]}),"\n"]}),"\n"]}),"\n",(0,s.jsxs)(t.p,{children:["For monetr you will need your ",(0,s.jsx)(t.code,{children:"client_id"})," as well as your ",(0,s.jsx)(t.code,{children:"sandbox"})," secret."]}),"\n",(0,s.jsx)(t.h3,{id:"oauth",children:"OAuth"}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.strong,{children:"TODO"})}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.a,{href:"https://github.com/monetr/monetr/issues/806",children:(0,s.jsx)(t.img,{src:"https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.github.com%2Frepos%2Fmonetr%2Fmonetr%2Fissues%2F806&query=%24.title&logo=github&label=docs",alt:"GitHub issue/pull request detail"})})}),"\n",(0,s.jsx)(t.h2,{id:"teller",children:"Teller"}),"\n",(0,s.jsx)(t.p,{children:"Teller credentials are not required for local development at this time."}),"\n",(0,s.jsxs)(t.p,{children:["If you are working with Teller it is recommended to set your environment to Sandbox via the ",(0,s.jsx)(t.code,{children:"CMakeUserPresets.json"})," file\nin your project directory."]}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.strong,{children:"TODO"})}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.a,{href:"https://github.com/monetr/monetr/issues/1666",children:(0,s.jsx)(t.img,{src:"https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.github.com%2Frepos%2Fmonetr%2Fmonetr%2Fissues%2F1666&query=%24.title&logo=github&label=docs",alt:"GitHub issue/pull request detail"})})}),"\n",(0,s.jsx)(t.h2,{id:"ngrok",children:"ngrok"}),"\n",(0,s.jsx)(t.p,{children:"ngrok is used to test code for webhooks. It allows requests to be made to your local development instance from an\nexternal endpoint. You can use ngrok without an API key; however, the tunnels will only last a short amount of time, and\nthe external endpoint will change each time. This might cause difficulty if you plan on working on webhook related\nfeatures. It is recommended to sign up for the free plan of ngrok and use the API key they provide you."}),"\n",(0,s.jsxs)(t.p,{children:["You can sign up for ngrok here: ",(0,s.jsx)(t.a,{href:"https://dashboard.ngrok.com/signup",children:"ngrok Sign Up"})]}),"\n",(0,s.jsx)(t.h2,{id:"stripe",children:"Stripe"}),"\n",(0,s.jsxs)(t.p,{children:["If you want to work on billing related features, you can also provide Stripe credentials to the local development\nenvironment. It is required to provide ngrok credentials along-side Stripe for local development. You can sign up for a\nStripe account here: ",(0,s.jsx)(t.a,{href:"https://dashboard.stripe.com/register",children:"Stripe Sign Up"})]}),"\n",(0,s.jsx)(t.p,{children:"You will need two sets of keys to work with Stripe."}),"\n",(0,s.jsxs)(t.ul,{children:["\n",(0,s.jsxs)(t.li,{children:["A ",(0,s.jsx)(t.strong,{children:"test mode"})," Stripe secret key. (Not the public key)"]}),"\n",(0,s.jsx)(t.li,{children:"A webhook secret, configured for your ngrok endpoint and with the proper scopes selected."}),"\n"]}),"\n",(0,s.jsx)(t.p,{children:"Once you have made a Stripe account you can follow this guide to retrieve your keys."}),"\n",(0,s.jsx)(t.h3,{id:"secret-key",children:"Secret key"}),"\n",(0,s.jsxs)(t.p,{children:["Navigate to your ",(0,s.jsx)(t.a,{href:"https://dashboard.stripe.com/test/apikeys",children:"Stripe API Keys"}),' page within the dashboard. Make sure you\nare in\n"Test mode".']}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.img,{alt:"Stripe Test Mode",placeholder:"blur",src:o})}),"\n",(0,s.jsxs)(t.p,{children:["You will need to click ",(0,s.jsx)(t.code,{children:"Reveal test key"})," in order to retrieve the API key."]}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.img,{alt:"Stripe Keys",placeholder:"blur",src:l})}),"\n",(0,s.jsx)(t.h3,{id:"webhook-secret",children:"Webhook secret"}),"\n",(0,s.jsxs)(t.p,{children:["On the ",(0,s.jsx)(t.a,{href:"https://dashboard.stripe.com/test/webhooks",children:"Stripe Webhooks"})," page click ",(0,s.jsx)(t.code,{children:"+ Add endpoint"}),"."]}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.img,{alt:"New Stripe Webhook",placeholder:"blur",src:d})}),"\n",(0,s.jsxs)(t.p,{children:["Enter your ngrok base URL here with the suffix: ",(0,s.jsx)(t.code,{children:"/api/stripe/webhook"})]}),"\n",(0,s.jsx)(t.p,{children:"Then you can add events that you need to work with. At a minimum the following events should be added as monetr requires\nthem."}),"\n",(0,s.jsx)(t.pre,{"data-language":"text","data-theme":"default",filename:"Stripe Webhook Events",children:(0,s.jsxs)(t.code,{"data-language":"text","data-theme":"default",children:[(0,s.jsx)(t.span,{className:"line",children:(0,s.jsx)(t.span,{style:{color:"var(--shiki-color-text)"},children:"checkout.session.completed"})}),"\n",(0,s.jsx)(t.span,{className:"line",children:(0,s.jsx)(t.span,{style:{color:"var(--shiki-color-text)"},children:"customer.deleted"})}),"\n",(0,s.jsx)(t.span,{className:"line",children:(0,s.jsx)(t.span,{style:{color:"var(--shiki-color-text)"},children:"customer.subscription.created"})}),"\n",(0,s.jsx)(t.span,{className:"line",children:(0,s.jsx)(t.span,{style:{color:"var(--shiki-color-text)"},children:"customer.subscription.deleted"})}),"\n",(0,s.jsx)(t.span,{className:"line",children:(0,s.jsx)(t.span,{style:{color:"var(--shiki-color-text)"},children:"customer.subscription.updated"})})]})}),"\n",(0,s.jsxs)(t.p,{children:["Once the webhook endpoint has been created click ",(0,s.jsx)(t.code,{children:"Reveal"})," under Signing Secret to retrieve the secret for the webhook\nendpoint."]}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.img,{alt:"Stripe Created Webhook",placeholder:"blur",src:a})}),"\n",(0,s.jsx)(t.h2,{id:"recaptcha",children:"ReCAPTCHA"}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.strong,{children:"TODO"})}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.a,{href:"https://github.com/monetr/monetr/issues/805",children:(0,s.jsx)(t.img,{src:"https://img.shields.io/github/issues/detail/state/monetr/monetr/805?label=%23805%20-%20docs%3A%20Document%20ReCAPTCHA%20credentials.&logo=github",alt:"GitHub issue/pull request detail"})})}),"\n",(0,s.jsx)(t.h2,{id:"sentry",children:"Sentry"}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.strong,{children:"TODO"})}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.a,{href:"https://github.com/monetr/monetr/issues/856",children:(0,s.jsx)(t.img,{src:"https://img.shields.io/github/issues/detail/state/monetr/monetr/856?label=%23856%20-%20docs%3A%20Document%20Sentry%20credentials.&logo=github",alt:"GitHub issue/pull request detail"})})}),"\n",(0,s.jsx)(t.h2,{id:"smtp",children:"SMTP"}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.strong,{children:"TODO"})}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.a,{href:"https://github.com/monetr/monetr/issues/857",children:(0,s.jsx)(t.img,{src:"https://img.shields.io/github/issues/detail/state/monetr/monetr/857?label=%23857%20-%20docs%3A%20Document%20SMTP%20credentials.&logo=github",alt:"GitHub issue/pull request detail"})})}),"\n",(0,s.jsx)(t.h2,{id:"google-cloud-kms",children:"Google Cloud KMS"}),"\n",(0,s.jsx)(t.p,{children:"Google Cloud KMS support is currently being added to improve the security of storing encrypted secrets in monetr.\nDocumentation to follow."}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.strong,{children:"TODO"})}),"\n",(0,s.jsx)(t.p,{children:(0,s.jsx)(t.a,{href:"https://github.com/monetr/monetr/issues/857",children:(0,s.jsx)(t.img,{src:"https://img.shields.io/github/issues/detail/state/monetr/monetr/936?label=%23936%20-%20docs%3A%20Document%20Google%20Cloud%20KMS%20credentials&logo=github",alt:"GitHub issue/pull request detail"})})})]})}var u=(0,i.j)({MDXContent:function(){let e=arguments.length>0&&void 0!==arguments[0]?arguments[0]:{},{wrapper:t}=Object.assign({},(0,r.a)(),e.components);return t?(0,s.jsx)(t,{...e,children:(0,s.jsx)(h,{...e})}):h(e)},pageOpts:{filePath:"src/pages/documentation/development/credentials.mdx",route:"/documentation/development/credentials",timestamp:1707690584e3,title:"3rd Party API credentials for development",headings:c},pageNextRoute:"/documentation/development/credentials"})}},function(e){e.O(0,[9304,9774,2888,179],function(){return e(e.s=1094)}),_N_E=e.O()}]);