(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[8952],{4948:(e,i,s)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/configure/plaid",function(){return s(3967)}])},3967:(e,i,s)=>{"use strict";s.r(i),s.d(i,{default:()=>h,useTOC:()=>t});var n=s(2540),d=s(7933),l=s(907),r=s(8439);function t(e){return[]}let h=(0,d.e)(function(e){let i={a:"a",code:"code",h1:"h1",p:"p",pre:"pre",span:"span",strong:"strong",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,r.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(i.h1,{children:"Plaid Configuration"}),"\n",(0,n.jsxs)(i.p,{children:["This guide shows you how to configure Plaid for your self hosted monetr instance. This will require Plaid credentials\nwhich can be obtained by following the ",(0,n.jsx)(i.a,{href:"../development/credentials#plaid",children:"Plaid Credentials Guide"}),"."]}),"\n",(0,n.jsx)(i.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,n.jsxs)(i.code,{children:[(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"plaid"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  enabled"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"<true|false>"})]}),"\n",(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  clientId"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'})]}),"\n",(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  clientSecret"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'})]}),"\n",(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  environment"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"<https://sandbox.plaid.com|https://production.plaid.com>"'})]}),"\n",(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  webhooksEnabled"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"<true|false>"})]}),"\n",(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  webhooksDomain"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'})]}),"\n",(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  oauthDomain"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'})]})]})}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:(0,n.jsx)(i.strong,{children:"Name"})}),(0,n.jsx)(i.th,{children:(0,n.jsx)(i.strong,{children:"Type"})}),(0,n.jsx)(i.th,{children:(0,n.jsx)(i.strong,{children:"Default"})}),(0,n.jsx)(i.th,{children:(0,n.jsx)(i.strong,{children:"Description"})})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"enabled"})}),(0,n.jsx)(i.td,{children:"Boolean"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"true"})}),(0,n.jsx)(i.td,{children:"Are users allowed to create Plaid links on this server? Even if this value is false, it is only considered enabled if the Client ID and Client Secret are also provided."})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"clientId"})}),(0,n.jsx)(i.td,{children:"String"}),(0,n.jsx)(i.td,{}),(0,n.jsx)(i.td,{children:"Your Plaid Client ID obtained from your Plaid account."})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"clientSecret"})}),(0,n.jsx)(i.td,{children:"String"}),(0,n.jsx)(i.td,{}),(0,n.jsx)(i.td,{children:"Your Plaid Client Secret obtained from your Plaid account."})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"environment"})}),(0,n.jsx)(i.td,{children:"String"}),(0,n.jsx)(i.td,{}),(0,n.jsx)(i.td,{children:"Plaid environment URL, must match the environment of your Plaid credentials."})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"webhooksEnabled"})}),(0,n.jsx)(i.td,{children:"Boolean"}),(0,n.jsx)(i.td,{}),(0,n.jsx)(i.td,{children:"If you want to allow Plaid to send updates to monetr via webhooks."})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"webhooksDomain"})}),(0,n.jsx)(i.td,{children:"String"}),(0,n.jsx)(i.td,{}),(0,n.jsx)(i.td,{children:"Required if you want to receive webhooks from Plaid, must be an externally accessible domain name. Plaid also requires HTTPS for all webhooks. Only specify the domain name, not a full URL. Sub-routes are not supported here."})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"oauthDomain"})}),(0,n.jsx)(i.td,{children:"String"}),(0,n.jsx)(i.td,{}),(0,n.jsx)(i.td,{children:"Domain used for OAuth redirect URLs, does not necessarily need to be externally accessible, however must support HTTPS as Plaid will not redirect to a non-HTTPS URL. Only specify the domain name. Sub-routes are not supported."})]})]})]}),"\n",(0,n.jsx)(i.p,{children:"The following environment variables map to the following configuration file fields. Each field is documented below."}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Variable"}),(0,n.jsx)(i.th,{children:"Config File Field"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"MONETR_PLAID_CLIENT_ID"})}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"plaid.clientId"})})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"MONETR_PLAID_CLIENT_SECRET"})}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"plaid.clientSecret"})})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"MONETR_PLAID_ENVIRONMENT"})}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"plaid.environment"})})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"MONETR_PLAID_WEBHOOKS_ENABLED"})}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"plaid.webhooksEnabled"})})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"MONETR_PLAID_WEBHOOKS_DOMAIN"})}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"plaid.webhooksDomain"})})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"MONETR_PLAID_OAUTH_DOMAIN"})}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"plaid.oauthDomain"})})]})]})]})]})},"/documentation/configure/plaid",{filePath:"src/pages/documentation/configure/plaid.mdx",timestamp:1732468966e3,pageMap:l.O,frontMatter:{},title:"Plaid Configuration"},"undefined"==typeof RemoteContent?t:RemoteContent.useTOC)}},e=>{var i=i=>e(e.s=i);e.O(0,[2615,636,6593,8792],()=>i(4948)),_N_E=e.O()}]);