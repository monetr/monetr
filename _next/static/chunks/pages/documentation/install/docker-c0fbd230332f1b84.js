(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[2857],{2162:(e,s,i)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/install/docker",function(){return i(5345)}])},5345:(e,s,i)=>{"use strict";i.r(s),i.d(s,{default:()=>p,useTOC:()=>c});var n=i(2540),r=i(7933),t=i(3904),a=i(8439),o=i(1785),l=i(1750),d=i(3696);function h({children:e,className:s,style:i,...r}){let t=(0,d.useId)().replaceAll(":","");return(0,n.jsx)("div",{className:(0,l.A)("nextra-steps _ms-4 _mb-12 _border-s _border-gray-200 _ps-6","dark:_border-neutral-800",s),style:{...i,"--counter-id":t},...r,children:e})}function c(e){let s={code:"code",...(0,a.R)()};return[{value:"Image Tags",id:"image-tags",depth:2},{value:(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(s.code,{children:"latest"})," Tag"]}),id:"latest-tag",depth:3},{value:"Versioned Tags",id:"versioned-tags",depth:3},{value:"Docker Compose",id:"docker-compose",depth:2},{value:"Clone The Repository",id:"clone-the-repository",depth:3},{value:"Start The Server",id:"start-the-server",depth:3},{value:"Open monetr",id:"open-monetr",depth:3},{value:"Updating Via Docker Compose",id:"updating-via-docker-compose",depth:2},{value:"Update The Repository",id:"update-the-repository",depth:3},{value:"Stop The Containers",id:"stop-the-containers",depth:3},{value:"Pull New Images",id:"pull-new-images",depth:3},{value:"Start monter Again",id:"start-monter-again",depth:3},{value:"Troubleshooting",id:"troubleshooting",depth:2},{value:"Containers Won’t Start",id:"containers-wont-start",depth:3},{value:"Cannot Access monetr in the Browser",id:"cannot-access-monetr-in-the-browser",depth:3},{value:"Update Issues After Pulling New Images",id:"update-issues-after-pulling-new-images",depth:3},{value:"Need More Help?",id:"need-more-help",depth:3}]}let p=(0,r.e)(function(e){let{toc:s=c(e)}=e,i={a:"a",br:"br",code:"code",h1:"h1",h2:"h2",h3:"h3",li:"li",p:"p",pre:"pre",span:"span",strong:"strong",ul:"ul",...(0,a.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(i.h1,{children:"Docker Compose"}),"\n",(0,n.jsxs)(i.p,{children:["Self-hosting monetr via Docker Compose is the simplest and officially supported way to run monetr yourself. This guide\nassumes Docker is already installed on your system. If not, please refer to ",(0,n.jsx)(i.a,{href:"https://docs.docker.com/engine/install/",children:"Docker’s Installation\nGuide"})," to set it up."]}),"\n",(0,n.jsx)(i.p,{children:"monetr’s container images are built with every tagged release and are available on both:"}),"\n",(0,n.jsxs)(i.ul,{children:["\n",(0,n.jsx)(i.li,{children:(0,n.jsx)(i.a,{href:"https://hub.docker.com/r/monetr/monetr",children:"DockerHub"})}),"\n",(0,n.jsx)(i.li,{children:(0,n.jsx)(i.a,{href:"https://github.com/monetr/monetr/pkgs/container/monetr",children:"GitHub Container Registry (GHCR)"})}),"\n"]}),"\n",(0,n.jsx)(i.p,{children:"Images from both registries are identical for the same version tag, so feel free to use your preferred registry."}),"\n",(0,n.jsx)(i.h2,{id:s[0].id,children:s[0].value}),"\n",(0,n.jsx)(i.p,{children:"Each monetr release provides two types of container image tags:"}),"\n",(0,n.jsx)(i.h3,{id:s[1].id,children:s[1].value}),"\n",(0,n.jsxs)(i.p,{children:["The ",(0,n.jsx)(i.code,{children:"latest"})," tag always points to the most recent version of monetr. For example:",(0,n.jsx)(i.br,{}),"\n",(0,n.jsx)(i.code,{children:"ghcr.io/monetr/monetr:latest"})]}),"\n",(0,n.jsx)(o.P,{type:"warning",children:(0,n.jsxs)(i.p,{children:[(0,n.jsx)(i.strong,{children:"Note"}),(0,n.jsx)(i.br,{}),"\n","Using the ",(0,n.jsx)(i.code,{children:"latest"})," tag is convenient but can lead to unexpected behavior if updates introduce breaking changes."]})}),"\n",(0,n.jsx)(i.h3,{id:s[2].id,children:s[2].value}),"\n",(0,n.jsxs)(i.p,{children:["Versioned tags, such as ",(0,n.jsx)(i.code,{children:"0.18.31"}),", refer to specific releases. For example:",(0,n.jsx)(i.br,{}),"\n",(0,n.jsx)(i.code,{children:"ghcr.io/monetr/monetr:0.18.31"})]}),"\n",(0,n.jsxs)(i.p,{children:["Version tags are recommended for stability. By pinning a version, you can control updates and easily roll back if needed. monetr’s version numbers use a ",(0,n.jsx)(i.code,{children:"v"})," prefix (e.g., ",(0,n.jsx)(i.code,{children:"v0.18.31"}),"), but container tags omit this prefix."]}),"\n",(0,n.jsx)(i.h2,{id:s[3].id,children:s[3].value}),"\n",(0,n.jsxs)(i.p,{children:["The easiest way to start monetr is to use the provided\n",(0,n.jsx)(i.a,{href:"https://github.com/monetr/monetr/blob/main/docker-compose.yaml",children:(0,n.jsx)(i.code,{children:"docker-compose.yaml"})})," located in the project’s root\ndirectory."]}),"\n",(0,n.jsxs)(h,{children:[(0,n.jsx)(i.h3,{id:s[4].id,children:s[4].value}),(0,n.jsx)(i.p,{children:"To get the compose file, first clone the monetr repository:"}),(0,n.jsx)(i.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Clone monetr",children:(0,n.jsx)(i.code,{children:(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"git"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" clone"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" https://github.com/monetr/monetr.git"})]})})}),(0,n.jsx)(i.h3,{id:s[5].id,children:s[5].value}),(0,n.jsx)(i.p,{children:"To run monetr, execute the following command in your terminal from the root directory of monetr’s repository."}),(0,n.jsx)(i.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Start monetr",children:(0,n.jsx)(i.code,{children:(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" up"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:" -d"})]})})}),(0,n.jsx)(i.p,{children:"This will start the monetr server, as well as the database it needs and a redis server."}),(0,n.jsx)(i.h3,{id:s[6].id,children:s[6].value}),(0,n.jsxs)(i.p,{children:["Once monetr is finished starting, you should be able to access it in your browser via ",(0,n.jsx)(i.code,{children:"http://localhost:4000"}),"."]})]}),"\n",(0,n.jsx)(o.P,{type:"warning",children:(0,n.jsxs)(i.p,{children:["Sign ups are enabled by default from the ",(0,n.jsx)(i.code,{children:"docker-compose.yaml"})," provided. If you are exposing your monetr instance to\nthe public internet; it is recommended you disable sign ups after you have created your own login."]})}),"\n",(0,n.jsx)(i.h2,{id:s[7].id,children:s[7].value}),"\n",(0,n.jsx)(i.p,{children:"If you are already running monetr and want to upgrade to a more recent version you can perform the following steps."}),"\n",(0,n.jsx)(i.p,{children:"Please make sure to review the release notes for monter before upgrading, as it will include any breaking changes you\nshould be aware of."}),"\n",(0,n.jsxs)(h,{children:[(0,n.jsx)(i.h3,{id:s[8].id,children:s[8].value}),(0,n.jsx)(i.p,{children:"In your cloned monetr directory, run the following command to retrieve the latest changes."}),(0,n.jsx)(i.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Retrieve changes",children:(0,n.jsx)(i.code,{children:(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"git"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" pull"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:" --rebase"})]})})}),(0,n.jsx)(i.p,{children:"If you encounter a conflict while performing the pull, this means that some changes you may have made locally might\ncause problems with the latest version of monetr. Make sure to resolve these conflicts before moving onto the next step."}),(0,n.jsx)(i.h3,{id:s[9].id,children:s[9].value}),(0,n.jsx)(i.p,{children:"You’ll need to stop the containers running before upgrading to make sure there are not conflicts."}),(0,n.jsx)(i.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Stop monetr",children:(0,n.jsx)(i.code,{children:(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" stop"})]})})}),(0,n.jsx)(i.h3,{id:s[10].id,children:s[10].value}),(0,n.jsx)(i.p,{children:"Once the containers have stopped you can run the following command to update the monetr image:"}),(0,n.jsx)(i.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Docker pull",children:(0,n.jsx)(i.code,{children:(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" pull"})]})})}),(0,n.jsx)(i.h3,{id:s[11].id,children:s[11].value}),(0,n.jsx)(i.p,{children:"Once the new images have been pulled onto your local machine you can restart the server via docker compose:"}),(0,n.jsx)(i.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Docker start",children:(0,n.jsx)(i.code,{children:(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" start"})]})})}),(0,n.jsx)(i.p,{children:"Things like database migrations are automatically run when using the provided compose file."})]}),"\n",(0,n.jsx)(i.h2,{id:s[12].id,children:s[12].value}),"\n",(0,n.jsx)(i.p,{children:"If you encounter issues while setting up or running monetr, here are some common problems and their solutions:"}),"\n",(0,n.jsx)(i.h3,{id:s[13].id,children:s[13].value}),"\n",(0,n.jsx)(i.p,{children:"If the containers fail to start or exit immediately:"}),"\n",(0,n.jsxs)(i.ul,{children:["\n",(0,n.jsxs)(i.li,{children:["Check the logs using:","\n",(0,n.jsx)(i.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"",children:(0,n.jsx)(i.code,{children:(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" logs"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:" -f"})]})})}),"\n"]}),"\n"]}),"\n",(0,n.jsx)(i.h3,{id:s[14].id,children:s[14].value}),"\n",(0,n.jsxs)(i.p,{children:["If ",(0,n.jsx)(i.code,{children:"http://localhost:4000"})," doesn’t load:"]}),"\n",(0,n.jsx)(i.p,{children:"Verify the containers are running using:"}),"\n",(0,n.jsx)(i.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"",children:(0,n.jsx)(i.code,{children:(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" ps"})]})})}),"\n",(0,n.jsxs)(i.p,{children:["Ensure no other services are using port 4000. Modify the ",(0,n.jsx)(i.code,{children:"docker-compose.yaml"})," file to use a different port if needed.\nCheck firewall or network settings on your machine."]}),"\n",(0,n.jsx)(i.h3,{id:s[15].id,children:s[15].value}),"\n",(0,n.jsx)(i.p,{children:"If monetr doesn’t work correctly after an update:"}),"\n",(0,n.jsxs)(i.ul,{children:["\n",(0,n.jsxs)(i.li,{children:["Check for breaking changes in the ",(0,n.jsx)(i.a,{href:"https://github.com/monetr/monetr/releases",children:"Release Notes"})]}),"\n",(0,n.jsxs)(i.li,{children:["Run","\n",(0,n.jsx)(i.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"",children:(0,n.jsxs)(i.code,{children:[(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" stop"})]}),"\n",(0,n.jsxs)(i.span,{children:[(0,n.jsx)(i.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" up"}),(0,n.jsx)(i.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:" -d"})]})]})}),"\n","This should recreate the containers for monetr without deleting any data on the volumes."]}),"\n"]}),"\n",(0,n.jsx)(i.h3,{id:s[16].id,children:s[16].value}),"\n",(0,n.jsx)(i.p,{children:"If these solutions don’t resolve your issue:"}),"\n",(0,n.jsxs)(i.ul,{children:["\n",(0,n.jsxs)(i.li,{children:["Check the ",(0,n.jsx)(i.a,{href:"https://github.com/monetr/monetr/issues",children:"monetr GitHub Issues"})," for similar problems."]}),"\n",(0,n.jsx)(i.li,{children:"Create a new issue with detailed logs and steps to reproduce the problem."}),"\n",(0,n.jsxs)(i.li,{children:["Reach out for assistance on ",(0,n.jsx)(i.a,{href:"https://discord.gg/68wTCXrhuq",children:"Discord"}),"."]}),"\n"]})]})},"/documentation/install/docker",{filePath:"src/pages/documentation/install/docker.mdx",timestamp:1733364629e3,pageMap:t.O,frontMatter:{title:"Self-Host with Docker Compose",description:"Learn how to self-host monetr using Docker Compose. Follow step-by-step instructions to set up monetr, manage updates, and troubleshoot common issues for a seamless self-hosting experience."},title:"Self-Host with Docker Compose"},"undefined"==typeof RemoteContent?c:RemoteContent.useTOC)}},e=>{var s=s=>e(e.s=s);e.O(0,[5684,636,6593,8792],()=>s(2162)),_N_E=e.O()}]);