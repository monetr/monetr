(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[425],{2848:(e,i,n)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/configure/logging",function(){return n(314)}])},314:(e,i,n)=>{"use strict";n.r(i),n.d(i,{default:()=>o,useTOC:()=>l});var s=n(2540),t=n(7933),r=n(3786),a=n(8439);function l(e){return[]}let o=(0,t.e)(function(e){let i={a:"a",code:"code",h1:"h1",p:"p",pre:"pre",span:"span",...(0,a.R)(),...e.components};return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)(i.h1,{children:"Logging Configuration"}),"\n",(0,s.jsxs)(i.p,{children:["monetr uses ",(0,s.jsx)(i.a,{href:"https://github.com/sirupsen/logrus",children:"logrus"})," for all of its logging at this time. At the moment it supports\na text and JSON formatter. The text formatter will have colors enabled by default, unless an environment variable ",(0,s.jsx)(i.code,{children:"CI"}),"\nis provided with a non-empty value."]}),"\n",(0,s.jsx)(i.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,s.jsxs)(i.code,{children:[(0,s.jsxs)(i.span,{children:[(0,s.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"logging"}),(0,s.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,s.jsxs)(i.span,{children:[(0,s.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  level"}),(0,s.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,s.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"<panic|fatal|error|warn|info|debug|trace>"'})]}),"\n",(0,s.jsxs)(i.span,{children:[(0,s.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  format"}),(0,s.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,s.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"<text|json>"'})]}),"\n",(0,s.jsxs)(i.span,{children:[(0,s.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  stackDriver"}),(0,s.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,s.jsxs)(i.span,{children:[(0,s.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    enabled"}),(0,s.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,s.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"<true|false>"})]})]})}),"\n",(0,s.jsxs)(i.p,{children:["The default log level for monetr is ",(0,s.jsx)(i.code,{children:"info"}),". Lower log levels can create a lot of noise, ",(0,s.jsx)(i.code,{children:"debug"})," will log each HTTP\nrequest that the server handles, ",(0,s.jsx)(i.code,{children:"trace"})," will log every SQL query that the application performs (except as part of\nbackground job processing)."]}),"\n",(0,s.jsxs)(i.p,{children:["If you are running your application on Google Cloud, it is recommended to enable StackDriver logging, as it will adjust\nthe way some important fields are formatted (when the format is set to ",(0,s.jsx)(i.code,{children:"json"}),") to match StackDriver’s expected patterns.\nMore information on the StackDriver format is available here: ",(0,s.jsx)(i.a,{href:"https://cloud.google.com/logging/docs/structured-logging",children:"Structured\nLogging"}),"."]})]})},"/documentation/configure/logging",{filePath:"src/pages/documentation/configure/logging.mdx",timestamp:1732468966e3,pageMap:r.O,frontMatter:{},title:"Logging Configuration"},"undefined"==typeof RemoteContent?l:RemoteContent.useTOC)}},e=>{var i=i=>e(e.s=i);e.O(0,[5694,636,6593,8792],()=>i(2848)),_N_E=e.O()}]);