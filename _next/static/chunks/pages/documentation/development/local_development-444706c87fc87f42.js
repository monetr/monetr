(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[2621],{5462:function(e,n,s){(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/development/local_development",function(){return s(1343)}])},1343:function(e,n,s){"use strict";s.r(n),s.d(n,{__toc:function(){return a},default:function(){return h}});var i=s(4246),l=s(9304),o=s(1441),t={src:"/_next/static/media/IntellJ_IDEA_Go_Debugging.5ce4efa4.png",height:887,width:1081,blurDataURL:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAgAAAAHCAYAAAA1WQxeAAAAxElEQVR42hWNS1LDMBBEW2NJjhOcpKhwETgDsOBUnIc1d+AeLMAUkIWV6DPSTJxedtV7zzw9P75O088LFz45Z4lrQ5hn2YzjBqrv9nv6u08xPdzttphzQUkRwzCgckWr/GX/f6dIhlBuD7w77LubcYuccwshuJJTtN73JK1BRYhzoiaCc4zauMAYQ5aXR2EQwyzSKkQVRCTdMgBi957WBgpu1a2cxYkrSikdEV0Na9sZvEHlc2mmo4i5YtY57ft+5b3/uABa3G/8vsWvkwAAAABJRU5ErkJggg==",blurWidth:8,blurHeight:7},r=s(2961);let a=[{depth:2,value:"Prerequisites",id:"prerequisites"},{depth:2,value:"Clone the repository",id:"clone-the-repository"},{depth:2,value:"Dependencies",id:"dependencies"},{depth:3,value:"Required",id:"required"},{depth:3,value:"Optional",id:"optional"},{depth:3,value:"Mac Specific",id:"mac-specific"},{depth:2,value:"Configuration & Credentials",id:"configuration--credentials"},{depth:3,value:"CMake",id:"cmake"},{depth:3,value:"Teller",id:"teller"},{depth:2,value:"Starting It Up",id:"starting-it-up"},{depth:2,value:"Working",id:"working"},{depth:3,value:"Local Services",id:"local-services"},{depth:3,value:"Debugging",id:"debugging"},{depth:3,value:"Running Tests",id:"running-tests"},{depth:3,value:"Running Storybook",id:"running-storybook"},{depth:2,value:"Cleaning Up",id:"cleaning-up"},{depth:3,value:"Shutting Down Development Environment",id:"shutting-down-development-environment"},{depth:3,value:"Completely Clean up",id:"completely-clean-up"}];function c(e){let n=Object.assign({h1:"h1",p:"p",h2:"h2",a:"a",pre:"pre",code:"code",span:"span",h3:"h3",ul:"ul",li:"li",em:"em",img:"img",hr:"hr",strong:"strong"},(0,o.a)(),e.components);return(0,i.jsxs)(i.Fragment,{children:[(0,i.jsx)(n.h1,{children:"Local Development"}),"\n",(0,i.jsx)(n.p,{children:"This guide walks you through setting up a local development environment for monetr on macOS or Linux. If you are using\nWindows, it is still possible to run the development environment locally. However, it is not documented at this time."}),"\n",(0,i.jsx)(n.h2,{id:"prerequisites",children:"Prerequisites"}),"\n",(0,i.jsxs)(n.p,{children:["At the time of writing this, monetr requires Plaid credentials for development. Among other credentials, documentation\non how to retrieve them can be found here: ",(0,i.jsx)(n.a,{href:"credentials",children:"Developing > Credentials"})]}),"\n",(0,i.jsx)(n.h2,{id:"clone-the-repository",children:"Clone the repository"}),"\n",(0,i.jsxs)(n.p,{children:["To get started, clone the monetr repository from ",(0,i.jsx)(n.a,{href:"https://github.com/monetr/monetr",children:"GitHub"})," or from your fork."]}),"\n",(0,i.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Shell",children:(0,i.jsxs)(n.code,{"data-language":"shell","data-theme":"default",children:[(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"git"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"clone"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"https://github.com/monetr/monetr.git"})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"cd"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"monetr"})]})]})}),"\n",(0,i.jsx)(n.p,{children:"The rest of the commands in this guide are issued from this directory."}),"\n",(0,i.jsx)(n.h2,{id:"dependencies",children:"Dependencies"}),"\n",(0,i.jsx)(n.p,{children:"monetr does require a few tools to be installed locally in order to develop or build it. These tools are outlines below:"}),"\n",(0,i.jsx)(n.h3,{id:"required",children:"Required"}),"\n",(0,i.jsxs)(n.ul,{children:["\n",(0,i.jsxs)(n.li,{children:["Node (",(0,i.jsx)(n.code,{children:">= 16.0.0"}),")"]}),"\n",(0,i.jsxs)(n.li,{children:["npm (",(0,i.jsx)(n.code,{children:">= 8.0.0"}),")"]}),"\n",(0,i.jsxs)(n.li,{children:["git (",(0,i.jsx)(n.code,{children:">= 2.0.0"}),")"]}),"\n",(0,i.jsxs)(n.li,{children:["Go (",(0,i.jsx)(n.code,{children:">= 1.20.0"}),")"]}),"\n",(0,i.jsxs)(n.li,{children:["CMake (",(0,i.jsx)(n.code,{children:">= 3.23.0"}),")"]}),"\n",(0,i.jsxs)(n.li,{children:["GNUMake (",(0,i.jsx)(n.code,{children:">= 4.0"}),")"]}),"\n"]}),"\n",(0,i.jsx)(n.p,{children:"The tools above are the minimum tools required in order to build and work on monetr locally. But if you intend to run\nthe complete development environment locally or if you plan on creating release builds of monetr you will also need:"}),"\n",(0,i.jsx)(n.h3,{id:"optional",children:"Optional"}),"\n",(0,i.jsxs)(n.ul,{children:["\n",(0,i.jsxs)(n.li,{children:["Docker (",(0,i.jsx)(n.code,{children:">= 20.0.0"}),"): Docker (and Docker Compose) are used to run the local development environment for monetr,\nallowing you to have the entire application and all of it's features with hot-reloading."]}),"\n",(0,i.jsxs)(n.li,{children:["Ruby (",(0,i.jsx)(n.code,{children:">= 2.7"}),"): Ruby is required to run ",(0,i.jsx)(n.code,{children:"licensed"})," which is used to generate third-party-notice files, these show all\nof the dependencies of monetr and their licenses. This is embeded at build time for releases."]}),"\n",(0,i.jsxs)(n.li,{children:["Kubectl (",(0,i.jsx)(n.code,{children:">= 1.23.0"}),"): Kubectl is used to deploy monetr to a Kubernetes cluster. At the moment this is ",(0,i.jsx)(n.em,{children:"only"})," used in\nCI/CD pipelines for deployying monetr to the Staging and Production clusters."]}),"\n"]}),"\n",(0,i.jsx)(n.h3,{id:"mac-specific",children:"Mac Specific"}),"\n",(0,i.jsxs)(n.p,{children:["macOS can ship with a version of ",(0,i.jsx)(n.code,{children:"make"})," that is outdated. It is recommended that you use ",(0,i.jsx)(n.code,{children:"brew"})," or any other preferred\nmethod to install the most recent version of GNUMake on your Mac. This will not break anything that is already using\nmake, but will make sure that your version is compatible with the monetr Makefiles."]}),"\n",(0,i.jsx)(n.p,{children:"For example; you should see something like this for your make version."}),"\n",(0,i.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Shell",children:(0,i.jsxs)(n.code,{"data-language":"shell","data-theme":"default",children:[(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"make"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"-v"})]}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-comment)"},children:"# GNU Make 4.3"})}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-comment)"},children:"# Built for x86_64-apple-darwin20.1.0"})}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-comment)"},children:"# Copyright (C) 1988-2020 Free Software Foundation, Inc."})}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-comment)"},children:"# License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>"})}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-comment)"},children:"# This is free software: you are free to change and redistribute it."})}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-comment)"},children:"# There is NO WARRANTY, to the extent permitted by law."})})]})}),"\n",(0,i.jsx)(n.h2,{id:"configuration--credentials",children:"Configuration & Credentials"}),"\n",(0,i.jsxs)(n.p,{children:["At the moment monetr requires at least Plaid credentials in order to run properly, even for development. You can read\nmore about obtaining these credentials here: ",(0,i.jsx)(n.a,{href:"credentials",children:"Credentials"})]}),"\n",(0,i.jsx)(n.p,{children:"The makefile will look for these development credentials and some configuration options in the following path:"}),"\n",(0,i.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"monetr development env file",children:(0,i.jsx)(n.code,{"data-language":"shell","data-theme":"default",children:(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"$HOME/.monetr/development.env"})})})}),"\n",(0,i.jsx)(n.p,{children:"You can create the file manually like this:"}),"\n",(0,i.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Manually creating the development env file",children:(0,i.jsxs)(n.code,{"data-language":"shell","data-theme":"default",children:[(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"mkdir"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" $HOME"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"/.monetr"})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"touch"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" $HOME"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"/.monetr/development.env"})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"vim"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" $HOME"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"/.monetr/development.env"})]})]})}),"\n",(0,i.jsxs)(n.p,{children:["Once you've opened this file you'll need to provide the Plaid Client ID as ",(0,i.jsx)(n.code,{children:"PLAID_CLIENT_ID"})," and Plaid Client Secret as\n",(0,i.jsx)(n.code,{children:"PLAID_CLIENT_SECRET"})," here."]}),"\n",(0,i.jsx)(n.h3,{id:"cmake",children:"CMake"}),"\n",(0,i.jsxs)(n.p,{children:["monetr uses CMake as the primary build system, it is also used for running the tasks for local development. If you want\nto use a custom CMake configuration you can create a ",(0,i.jsx)(n.code,{children:"CMakeUserPresets.json"})," file in the project directory (it will be\nignored by Git) and populate it with the configuration variables you want to override. If you want to always use a\nspecific preset when running general make targets you can create a ",(0,i.jsx)(n.code,{children:".cmakepreset"})," file in the project directory. Inside\nthat file you can put the name of the preset you want to use by default. Or you can pass ",(0,i.jsx)(n.code,{children:"CMAKE_PRESET=..."})," when running\na make target."]}),"\n",(0,i.jsx)(r.UW,{type:"info",children:(0,i.jsxs)(n.p,{children:[(0,i.jsx)(n.code,{children:"CMAKE_PRESET"})," is ignored for some make targets, such as ",(0,i.jsx)(n.code,{children:"release"}),", ",(0,i.jsx)(n.code,{children:"test"})," and ",(0,i.jsx)(n.code,{children:"deploy"}),". As each of those targets have\na specific preset that they are expected to use."]})}),"\n",(0,i.jsx)(n.h3,{id:"teller",children:"Teller"}),"\n",(0,i.jsxs)(n.p,{children:["If you are working with Teller.io locally then you may need to configure certificates if you are not using their sandbox\nenvironment for testing. Or you may wish to use certificates even for sandbox development. If you want to do this you\nwill need to create a ",(0,i.jsx)(n.code,{children:"CMakeUserPresets.json"})," file with a preset that has the following cache variables:"]}),"\n",(0,i.jsxs)(n.ul,{children:["\n",(0,i.jsxs)(n.li,{children:[(0,i.jsx)(n.code,{children:"TELLER_ENVIRONMENT"}),": One of Teller's supported environments, Sandbox is recommended."]}),"\n",(0,i.jsxs)(n.li,{children:[(0,i.jsx)(n.code,{children:"TELLER_CERTIFICATE"}),": A path to the ",(0,i.jsx)(n.code,{children:".pem"})," certificate file from teller.io on your computer."]}),"\n",(0,i.jsxs)(n.li,{children:[(0,i.jsx)(n.code,{children:"TELLER_PRIVATE_KEY"}),": A path to the ",(0,i.jsx)(n.code,{children:".pem"})," private key file from teller.io on your computer."]}),"\n"]}),"\n",(0,i.jsx)(n.p,{children:"For example:"}),"\n",(0,i.jsx)(n.pre,{"data-language":"json","data-theme":"default",filename:"CMakeUserPresets.json",children:(0,i.jsxs)(n.code,{"data-language":"json","data-theme":"default",children:[(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"{"})}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"  "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"version"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-constant)"},children:"5"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"  "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"cmakeMinimumRequired"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" {"})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"    "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"major"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-constant)"},children:"3"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"    "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"minor"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-constant)"},children:"23"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"    "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"patch"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-constant)"},children:"0"})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"  }"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"  "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"configurePresets"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" ["})]}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"    {"})}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"      "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"name"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"example"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"      "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"displayName"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"Example monetr config"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"      "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"description"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"Example monetr config for local development with Teller.io"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"      "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"generator"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"Unix Makefiles"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"      "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"binaryDir"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"${sourceDir}/build"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"      "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"inherits"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"default"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"      "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"cacheVariables"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" {"})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"        "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"TELLER_ENVIRONMENT"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"sandbox"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"        "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"TELLER_CERTIFICATE"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"~/.monetr/teller/certificate.pem"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"        "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"TELLER_PRIVATE_KEY"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"~/.monetr/teller/private_key.pem"'})]}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"      }"})}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"    }"})}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"  ]"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"  "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"buildPresets"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" ["})]}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"    {"})}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"      "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"name"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"example"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:","})]}),"\n",(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"      "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-keyword)"},children:'"configurePreset"'}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-punctuation)"},children:":"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string-expression)"},children:'"example"'})]}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"    }"})}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"  ]"})}),"\n",(0,i.jsx)(n.span,{className:"line",children:(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:"}"})})]})}),"\n",(0,i.jsxs)(n.p,{children:["When running ",(0,i.jsx)(n.code,{children:"make develop CMAKE_PROFILE=example"})," CMake will copy and expose those two files such that the monetr API\ncan access them inside the container environment."]}),"\n",(0,i.jsx)(n.h2,{id:"starting-it-up",children:"Starting It Up"}),"\n",(0,i.jsx)(n.p,{children:"With the above requirements installed. You should be able to spin up the local development environment that runs inside\nof Docker compose."}),"\n",(0,i.jsxs)(n.p,{children:["This command will also load any of the environment variables specified in the development env file (mentioned above)\ninto the ",(0,i.jsx)(n.code,{children:"monetr"})," container where the API is running."]}),"\n",(0,i.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Shell",children:(0,i.jsx)(n.code,{"data-language":"shell","data-theme":"default",children:(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"make"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"develop"})]})})}),"\n",(0,i.jsxs)(n.p,{children:["This will install node modules in the project's directory, as well as start up the containers needed for monetr to run\nlocally. This command will exit automatically once all the containers are healthy. If you want to follow along while it\nis starting up you can use the ",(0,i.jsx)(n.code,{children:"make logs"})," command in another terminal inside the project directory."]}),"\n",(0,i.jsx)(n.h2,{id:"working",children:"Working"}),"\n",(0,i.jsxs)(n.p,{children:["Congratulations, you should now have monetr running locally on your computer. The ",(0,i.jsx)(n.code,{children:"develop"})," task will print out some\nbasic information for you. But if you missed it, you can access the development version of monetr at ",(0,i.jsx)(n.code,{children:"http://monetr.local"}),"."]}),"\n",(0,i.jsxs)(n.p,{children:["If you are working on documentation then that can be accessed at ",(0,i.jsx)(n.code,{children:"http://monetr.local/documentation"}),"."]}),"\n",(0,i.jsx)(n.p,{children:"Almost all of monetr's code is setup to hot-reload as you make changes. The documentation, Go code and React UI will all\nautomatically reload as changes are made. Changes to the Go code will not invoke a browser refresh of any sort though,\nso to observe a new behavior in the API you will need to refresh or make the API call again."}),"\n",(0,i.jsx)(r.UW,{type:"info",children:(0,i.jsxs)(n.p,{children:["If you want to disable hot reloading of the Go code, you can include ",(0,i.jsx)(n.code,{children:"DISABLE_GO_RELOAD=true"})," in your env variables\nwhen you run ",(0,i.jsx)(n.code,{children:"make develop"}),"."]})}),"\n",(0,i.jsx)(n.h3,{id:"local-services",children:"Local Services"}),"\n",(0,i.jsx)(n.p,{children:"As part of the local development stack, several services are run to support monetr. These services include:"}),"\n",(0,i.jsxs)(n.ul,{children:["\n",(0,i.jsxs)(n.li,{children:[(0,i.jsx)(n.a,{href:"https://github.com/minio/minio",children:"minio"})," As an S3 storage backend. The console is accessible via\n",(0,i.jsx)(n.code,{children:"http://localhost:9001"})," when the local environment is running."]}),"\n",(0,i.jsxs)(n.li,{children:[(0,i.jsx)(n.a,{href:"https://github.com/nsmithuk/local-kms",children:"local-kms"})," An AWS KMS compatible local development API. This is used for\nencrypting secrets in the local development environment."]}),"\n",(0,i.jsxs)(n.li,{children:[(0,i.jsx)(n.a,{href:"https://github.com/mailhog/MailHog",children:"mailhog"})," An SMTP server that allows emails to be sent without really sending\nthem. This is used to validate and test communication functionality locally. This service is accessible via\n",(0,i.jsx)(n.code,{children:"https://monetr.local/mail"})," when the local environment is running."]}),"\n"]}),"\n",(0,i.jsx)(n.h3,{id:"debugging",children:"Debugging"}),"\n",(0,i.jsxs)(n.p,{children:["The monetr container running the API has ",(0,i.jsx)(n.a,{href:"https://github.com/go-delve/delve",children:"delve"})," included. If you prefer to work\nusing a step-debugger you can connect your editor to it. You will need to reconnect your editor each time it reloads,\nbut it is very easy to make your changes and then hit ",(0,i.jsx)(n.em,{children:"debug"})," and let your breakpoints be hit."]}),"\n",(0,i.jsxs)(n.p,{children:["Delve is available via port ",(0,i.jsx)(n.code,{children:"2345"})," on ",(0,i.jsx)(n.code,{children:"localhost"}),". I'm not sure what the configuration will be for every editor to\nconnect to it; but this is a screenshot of IntellJ IDEA's configuration for remote debugging."]}),"\n",(0,i.jsx)(n.p,{children:(0,i.jsx)(n.img,{alt:"IntellJ IDEA Configuration",placeholder:"blur",src:t})}),"\n",(0,i.jsx)(n.h3,{id:"running-tests",children:"Running Tests"}),"\n",(0,i.jsx)(n.p,{children:"monetr requires a PostgreSQL instance to be available for Go tests to be run. At the moment there isn't a shorthand\nscript to provision this instance. But an easy way to do so is this:"}),"\n",(0,i.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Shell",children:(0,i.jsx)(n.code,{"data-language":"shell","data-theme":"default",children:(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"docker"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"run"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"-e"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"POSTGRES_HOST_AUTH_METHOD=trust"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"--name"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"postgres"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"--rm"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"-d"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"-p"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-constant)"},children:"5432"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:":5432"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"postgres:14"})]})})}),"\n",(0,i.jsxs)(n.p,{children:["This will start a PostgreSQL instance in Docker (or remove an existing one) and make it available on ",(0,i.jsx)(n.code,{children:"locahost:5432"})," as\nwell as not require authentication. This makes it easy for tests to target it."]}),"\n",(0,i.jsxs)(n.p,{children:["If tests are run via ",(0,i.jsx)(n.code,{children:"make"})," then nothing more needs to be done. However, if you want to run tests directly from your\neditor or other tools you will need to run the database migrations."]}),"\n",(0,i.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Shell",children:(0,i.jsx)(n.code,{"data-language":"shell","data-theme":"default",children:(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"make"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"migrate"})]})})}),"\n",(0,i.jsx)(n.p,{children:"Will run all schema migrations on the PostgreSQL database on your localhost container."}),"\n",(0,i.jsx)(n.hr,{}),"\n",(0,i.jsxs)(n.p,{children:["Tests can be run using the ",(0,i.jsx)(n.code,{children:"go test"})," CLI, or all tests can be run using:"]}),"\n",(0,i.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Shell",children:(0,i.jsx)(n.code,{"data-language":"shell","data-theme":"default",children:(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"make"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"test"})]})})}),"\n",(0,i.jsx)(n.h3,{id:"running-storybook",children:"Running Storybook"}),"\n",(0,i.jsx)(n.p,{children:"monetr now provides a storybook setup for working on UI components outside of running the entire application locally. To\nstart storybook you can run the following command."}),"\n",(0,i.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Shell",children:(0,i.jsx)(n.code,{"data-language":"shell","data-theme":"default",children:(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"make"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"storybook"})]})})}),"\n",(0,i.jsx)(n.p,{children:"This will kick off the storybook server and build process. Once it is ready it will open the storybook in your default\nbrowser. You can then make changes to the components in the stories and see the changes in real time without needing to\nrun the entire application stack locally."}),"\n",(0,i.jsxs)(n.p,{children:[(0,i.jsx)(n.strong,{children:"NOTE"})," At the moment storybook does not work with the CMake development tooling."]}),"\n",(0,i.jsx)(n.h2,{id:"cleaning-up",children:"Cleaning Up"}),"\n",(0,i.jsx)(n.p,{children:"Once you have finished your work and you want to take the local development environment down you have a few options."}),"\n",(0,i.jsx)(n.h3,{id:"shutting-down-development-environment",children:"Shutting Down Development Environment"}),"\n",(0,i.jsx)(n.p,{children:"If you want to completely shut everything down then you can run the following command:"}),"\n",(0,i.jsx)(r.UW,{type:"warning",children:(0,i.jsx)(n.p,{children:"This will delete all of your local development data, including any Plaid links, expenses, goals, etc..."})}),"\n",(0,i.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Shell",children:(0,i.jsx)(n.code,{"data-language":"shell","data-theme":"default",children:(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"make"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"shutdown"})]})})}),"\n",(0,i.jsx)(n.p,{children:"This removes the Plaid links that are active, takes down the Docker compose containers, removes their volumes."}),"\n",(0,i.jsx)(n.h3,{id:"completely-clean-up",children:"Completely Clean up"}),"\n",(0,i.jsxs)(n.p,{children:["If you want to completely start fresh you can run the following make task. This will shut down the local development\nenvironment if it is running, but it will also delete any files created or generated during development. This deletes\nyour ",(0,i.jsx)(n.code,{children:"node_modules"})," folder, any submodules, and generated UI code."]}),"\n",(0,i.jsx)(n.pre,{"data-language":"shell","data-theme":"default",filename:"Shell",children:(0,i.jsx)(n.code,{"data-language":"shell","data-theme":"default",children:(0,i.jsxs)(n.span,{className:"line",children:[(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-function)"},children:"make"}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-color-text)"},children:" "}),(0,i.jsx)(n.span,{style:{color:"var(--shiki-token-string)"},children:"clean"})]})})}),"\n",(0,i.jsx)(n.p,{children:"This should leave the project directory in a state similar to when it was initially cloned."})]})}var h=(0,l.j)({MDXContent:function(){let e=arguments.length>0&&void 0!==arguments[0]?arguments[0]:{},{wrapper:n}=Object.assign({},(0,o.a)(),e.components);return n?(0,i.jsx)(n,{...e,children:(0,i.jsx)(c,{...e})}):c(e)},pageOpts:{filePath:"src/pages/documentation/development/local_development.mdx",route:"/documentation/development/local_development",timestamp:1707690584e3,title:"Local Development",headings:a},pageNextRoute:"/documentation/development/local_development"})}},function(e){e.O(0,[9304,9774,2888,179],function(){return e(e.s=5462)}),_N_E=e.O()}]);