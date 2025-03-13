(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[2857],{2162:(e,i,s)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/install/docker",function(){return s(6698)}])},6698:(e,i,s)=>{"use strict";s.r(i),s.d(i,{default:()=>h,useTOC:()=>d});var n=s(2540),r=s(7933),t=s(7170),a=s(8795),l=s(1785),o=s(8126);function d(e){let i={code:"code",...(0,a.R)()};return[{value:"Image Tags",id:"image-tags",depth:2},{value:(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(i.code,{children:"latest"})," Tag"]}),id:"latest-tag",depth:3},{value:"Versioned Tags",id:"versioned-tags",depth:3},{value:"Docker Compose",id:"docker-compose",depth:2},{value:"Clone The Repository",id:"clone-the-repository",depth:3},{value:"Configure The Server",id:"configure-the-server",depth:3},{value:"Start The Server",id:"start-the-server",depth:3},{value:"Open monetr",id:"open-monetr",depth:3},{value:"Updating Via Docker Compose",id:"updating-via-docker-compose",depth:2},{value:"Update The Repository",id:"update-the-repository",depth:3},{value:"Stop The Containers",id:"stop-the-containers",depth:3},{value:"Pull New Images",id:"pull-new-images",depth:3},{value:"Start monetr Again",id:"start-monetr-again",depth:3},{value:"Troubleshooting",id:"troubleshooting",depth:2},{value:"Containers Won’t Start",id:"containers-wont-start",depth:3},{value:"Permission Error",id:"permission-error",depth:4},{value:"Cannot Access monetr in the Browser",id:"cannot-access-monetr-in-the-browser",depth:3},{value:"Update Issues After Pulling New Images",id:"update-issues-after-pulling-new-images",depth:3},{value:"Need More Help?",id:"need-more-help",depth:3},{value:"Uninstalling",id:"uninstalling",depth:2}]}let h=(0,r.e)(function(e){let{toc:i=d(e)}=e,s={a:"a",br:"br",code:"code",em:"em",h1:"h1",h2:"h2",h3:"h3",h4:"h4",li:"li",p:"p",pre:"pre",span:"span",strong:"strong",ul:"ul",...(0,a.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(s.h1,{children:"Docker Compose"}),"\n",(0,n.jsxs)(s.p,{children:["Self-hosting monetr via Docker Compose is the simplest and officially supported way to run monetr yourself. This guide\nassumes Docker is already installed on your system. If not, please refer to ",(0,n.jsx)(s.a,{href:"https://docs.docker.com/engine/install/",children:"Docker’s Installation\nGuide"})," to set it up."]}),"\n",(0,n.jsx)(s.p,{children:"monetr’s container images are built with every tagged release and are available on both:"}),"\n",(0,n.jsxs)(s.ul,{children:["\n",(0,n.jsx)(s.li,{children:(0,n.jsx)(s.a,{href:"https://hub.docker.com/r/monetr/monetr",children:"DockerHub"})}),"\n",(0,n.jsx)(s.li,{children:(0,n.jsx)(s.a,{href:"https://github.com/monetr/monetr/pkgs/container/monetr",children:"GitHub Container Registry (GHCR)"})}),"\n"]}),"\n",(0,n.jsx)(s.p,{children:"Images from both registries are identical for the same version tag, so feel free to use your preferred registry."}),"\n",(0,n.jsx)(s.h2,{id:i[0].id,children:i[0].value}),"\n",(0,n.jsx)(s.p,{children:"Each monetr release provides two types of container image tags:"}),"\n",(0,n.jsx)(s.h3,{id:i[1].id,children:i[1].value}),"\n",(0,n.jsxs)(s.p,{children:["The ",(0,n.jsx)(s.code,{children:"latest"})," tag always points to the most recent version of monetr. For example:",(0,n.jsx)(s.br,{}),"\n",(0,n.jsx)(s.code,{children:"ghcr.io/monetr/monetr:latest"})]}),"\n",(0,n.jsx)(l.P,{type:"warning",children:(0,n.jsxs)(s.p,{children:[(0,n.jsx)(s.strong,{children:"Note"}),(0,n.jsx)(s.br,{}),"\n","Using the ",(0,n.jsx)(s.code,{children:"latest"})," tag is convenient but can lead to unexpected behavior if updates introduce breaking changes."]})}),"\n",(0,n.jsx)(s.h3,{id:i[2].id,children:i[2].value}),"\n",(0,n.jsxs)(s.p,{children:["Versioned tags, such as ",(0,n.jsx)(s.code,{children:"0.18.31"}),", refer to specific releases. For example:",(0,n.jsx)(s.br,{}),"\n",(0,n.jsx)(s.code,{children:"ghcr.io/monetr/monetr:0.18.31"})]}),"\n",(0,n.jsxs)(s.p,{children:["Version tags are recommended for stability. By pinning a version, you can control updates and easily roll back if\nneeded. monetr’s version numbers use a ",(0,n.jsx)(s.code,{children:"v"})," prefix (e.g., ",(0,n.jsx)(s.code,{children:"v0.18.31"}),"), but container tags omit this prefix."]}),"\n",(0,n.jsx)(s.h2,{id:i[3].id,children:i[3].value}),"\n",(0,n.jsxs)(s.p,{children:["The easiest way to start monetr is to use the provided\n",(0,n.jsx)(s.a,{href:"https://github.com/monetr/monetr/blob/main/docker-compose.yaml",children:(0,n.jsx)(s.code,{children:"docker-compose.yaml"})})," located in the project’s root\ndirectory."]}),"\n",(0,n.jsxs)(o.g,{children:[(0,n.jsx)(s.h3,{id:i[4].id,children:i[4].value}),(0,n.jsx)(s.p,{children:"To get the compose file, first clone the monetr repository:"}),(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Clone monetr",children:(0,n.jsxs)(s.code,{children:[(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"git"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" clone"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" https://github.com/monetr/monetr.git"})]}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"cd"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" monetr"})]})]})}),(0,n.jsx)(s.h3,{id:i[5].id,children:i[5].value}),(0,n.jsxs)(s.p,{children:["The provided compose file includes some defaults that should be sufficient to get monetr started and to try out manual\nbudgeting. But if you want to change anything; like adding Plaid credentials or setting up a proper domain name, you’ll\nneed to configure monetr. The recommended way to do this is to pass environment variables for the parameters you want to\nchange. The easiest way to do this is to create a ",(0,n.jsx)(s.code,{children:".env"})," file somewhere outside the monetr repository folder and when\nrunning the Docker commands below, include the flag ",(0,n.jsx)(s.code,{children:"--env-file=${YOUR FILE PATH}"}),". This will apply your customizations\nto the compose file without needing to modify the provided file."]}),(0,n.jsxs)(s.p,{children:["If you want to use a config file though you will need to modify the compose file to use one, or you will need to create\na config file within the default volume mount that gets created. To use a config file adjust the ",(0,n.jsx)(s.code,{children:"command"})," for the\nmonetr service in the compose file to look like this:"]}),(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"docker-compose.yaml",children:(0,n.jsxs)(s.code,{children:[(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    command"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(s.span,{"data-highlighted-line":"",children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:"      - "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"-c"})]}),"\n",(0,n.jsxs)(s.span,{"data-highlighted-line":"",children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:"      - "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"/etc/monetr/config.yaml"})]}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:"      - "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"serve"})]}),"\n",(0,n.jsx)(s.span,{children:(0,n.jsx)(s.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"      # Setup the database and perform migrations."})}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:"      - "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"--migrate"})]}),"\n",(0,n.jsx)(s.span,{children:(0,n.jsx)(s.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"      # Since certificates will not have been created, make some."})}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:"      - "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"--generate-certificates"})]})]})}),(0,n.jsxs)(s.p,{children:["You can replace the path to the config file with any path you want as long as it is readable by monetr. You can specify\nmultiple configuration file if you need to by passing ",(0,n.jsx)(s.code,{children:"-c ${file}"})," multiple times ",(0,n.jsx)(s.em,{children:"before"})," the ",(0,n.jsx)(s.code,{children:"serve"})," command."]}),(0,n.jsx)(l.P,{type:"warning",children:(0,n.jsxs)(s.p,{children:["Environment variables can take priority over values in the configuration file. If you are not seeing the behavior\nyou’re expecting with your configuration changes, make sure that the environment variable for that configuration\nparameter is not specified with an incorrect or ",(0,n.jsx)(s.strong,{children:"blank"})," value. A blank value in the environment variable may cause\nunusual behaviors."]})}),(0,n.jsx)(s.h3,{id:i[6].id,children:i[6].value}),(0,n.jsx)(s.p,{children:"To run monetr, execute the following command in your terminal from the root directory of monetr’s repository."}),(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Start monetr",children:(0,n.jsx)(s.code,{children:(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" up"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:" --wait"})]})})}),(0,n.jsx)(s.p,{children:"This will start the monetr server, as well as the database it needs and a Valkey server. It will wait for everything to\nbe healthy before letting you continue."}),(0,n.jsx)(s.h3,{id:i[7].id,children:i[7].value}),(0,n.jsxs)(s.p,{children:["Once monetr is finished starting, you should be able to access it in your browser via ",(0,n.jsx)(s.code,{children:"http://localhost:4000"}),"."]}),(0,n.jsxs)(l.P,{type:"info",children:[(0,n.jsx)(s.p,{children:"monetr may be accessible from other URLs like the host’s IP address, but it will only set the authentication cookie\n(as well as other things like email links) based on the external URL configuration. If you are having trouble logging\nin, make sure you are accessing monetr from the same URL that it logs as “externalUrl” when it starts."}),(0,n.jsxs)(s.p,{children:["You can configure the external URL here: ",(0,n.jsx)(s.a,{href:"/documentation/configure/server",children:"Server Configuration"})]})]})]}),"\n",(0,n.jsx)(l.P,{type:"warning",children:(0,n.jsxs)(s.p,{children:["Sign ups are enabled by default from the ",(0,n.jsx)(s.code,{children:"docker-compose.yaml"})," provided. If you are exposing your monetr instance to\nthe public internet; it is recommended you disable sign ups after you have created your own login."]})}),"\n",(0,n.jsx)(s.h2,{id:i[8].id,children:i[8].value}),"\n",(0,n.jsx)(s.p,{children:"If you are already running monetr and want to upgrade to a more recent version you can perform the following steps."}),"\n",(0,n.jsx)(s.p,{children:"Please make sure to review the release notes for monetr before upgrading, as it will include any breaking changes you\nshould be aware of."}),"\n",(0,n.jsxs)(o.g,{children:[(0,n.jsx)(s.h3,{id:i[9].id,children:i[9].value}),(0,n.jsx)(s.p,{children:"In your cloned monetr directory, run the following command to retrieve the latest changes."}),(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Retrieve changes",children:(0,n.jsx)(s.code,{children:(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"git"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" pull"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:" --rebase"})]})})}),(0,n.jsx)(s.p,{children:"If you encounter a conflict while performing the pull, this means that some changes you may have made locally might\ncause problems with the latest version of monetr. Make sure to resolve these conflicts before moving onto the next step."}),(0,n.jsx)(s.h3,{id:i[10].id,children:i[10].value}),(0,n.jsx)(s.p,{children:"You’ll need to stop the containers running before upgrading to make sure there are not conflicts."}),(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Stop monetr",children:(0,n.jsx)(s.code,{children:(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" stop"})]})})}),(0,n.jsx)(s.h3,{id:i[11].id,children:i[11].value}),(0,n.jsx)(s.p,{children:"Once the containers have stopped you can run the following command to update the monetr image:"}),(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Docker pull",children:(0,n.jsx)(s.code,{children:(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" pull"})]})})}),(0,n.jsx)(s.h3,{id:i[12].id,children:i[12].value}),(0,n.jsx)(s.p,{children:"Once the new images have been pulled onto your local machine you can restart the server via docker compose:"}),(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Docker start",children:(0,n.jsx)(s.code,{children:(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" up"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:" --wait"})]})})}),(0,n.jsx)(s.p,{children:"Things like database migrations are automatically run when using the provided compose file."})]}),"\n",(0,n.jsx)(s.h2,{id:i[13].id,children:i[13].value}),"\n",(0,n.jsx)(s.p,{children:"If you encounter issues while setting up or running monetr, here are some common problems and their solutions:"}),"\n",(0,n.jsx)(s.h3,{id:i[14].id,children:i[14].value}),"\n",(0,n.jsx)(s.p,{children:"If the containers fail to start or exit immediately:"}),"\n",(0,n.jsxs)(s.ul,{children:["\n",(0,n.jsxs)(s.li,{children:["Check the logs using:","\n",(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"",children:(0,n.jsx)(s.code,{children:(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" logs"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:" -f"})]})})}),"\n"]}),"\n"]}),"\n",(0,n.jsx)(s.h4,{id:i[15].id,children:i[15].value}),"\n",(0,n.jsx)(s.p,{children:"If you are getting a permission denied error in your logs similar to:"}),"\n",(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"plaintext","data-word-wrap":"",children:(0,n.jsx)(s.code,{children:(0,n.jsx)(s.span,{children:(0,n.jsx)(s.span,{children:"failed to write private key: open /etc/monetr/ed25519.key: permission denied"})})})}),"\n",(0,n.jsx)(s.p,{children:"Then it is possible the permissions for your volume is not setup properly for the docker compose. This can happen if you\nare using host path volume mounts."}),"\n",(0,n.jsxs)(s.p,{children:["On Linux or macOS run the ",(0,n.jsx)(s.code,{children:"id"})," command in your terminal, you should get something like this:"]}),"\n",(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"",children:(0,n.jsxs)(s.code,{children:[(0,n.jsx)(s.span,{children:(0,n.jsx)(s.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"id"})}),"\n",(0,n.jsx)(s.span,{children:(0,n.jsx)(s.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"# uid=1000(elliotcourant) gid=1000(elliotcourant) ..."})})]})}),"\n",(0,n.jsxs)(s.p,{children:["Those two numbers could be anything on your system, take those two numbers and add the following line to the ",(0,n.jsx)(s.code,{children:"monetr"}),"\nservice in your docker compose file:"]}),"\n",(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"docker-compose.yaml",children:(0,n.jsxs)(s.code,{children:[(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"services"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsx)(s.span,{children:(0,n.jsx)(s.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"  # ..."})}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  monetr"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsx)(s.span,{children:(0,n.jsx)(s.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"    # ..."})}),"\n",(0,n.jsxs)(s.span,{"data-highlighted-line":"",children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"    user"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"1000:1000"})]})]})}),"\n",(0,n.jsxs)(s.p,{children:["Replacing the pairs of ",(0,n.jsx)(s.code,{children:"1000"})," with the values of ",(0,n.jsx)(s.code,{children:"uid"})," and ",(0,n.jsx)(s.code,{children:"gid"})," respectively."]}),"\n",(0,n.jsx)(s.p,{children:"Then try to start the compose file again. This should alleviate any permission issues with host path mounts as it will\nmake the container match your own user’s permissions."}),"\n",(0,n.jsx)(s.h3,{id:i[16].id,children:i[16].value}),"\n",(0,n.jsxs)(s.p,{children:["If ",(0,n.jsx)(s.code,{children:"http://localhost:4000"})," doesn’t load:"]}),"\n",(0,n.jsx)(s.p,{children:"Verify the containers are running using:"}),"\n",(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"",children:(0,n.jsx)(s.code,{children:(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" ps"})]})})}),"\n",(0,n.jsxs)(s.p,{children:["Ensure no other services are using port 4000. Modify the ",(0,n.jsx)(s.code,{children:"docker-compose.yaml"})," file to use a different port if needed.\nCheck firewall or network settings on your machine."]}),"\n",(0,n.jsx)(s.h3,{id:i[17].id,children:i[17].value}),"\n",(0,n.jsx)(s.p,{children:"If monetr doesn’t work correctly after an update:"}),"\n",(0,n.jsxs)(s.ul,{children:["\n",(0,n.jsxs)(s.li,{children:["Check for breaking changes in the ",(0,n.jsx)(s.a,{href:"https://github.com/monetr/monetr/releases",children:"Release Notes"})]}),"\n",(0,n.jsxs)(s.li,{children:["Run","\n",(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"",children:(0,n.jsxs)(s.code,{children:[(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" stop"})]}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" up"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:" -d"})]})]})}),"\n","This should recreate the containers for monetr without deleting any data on the volumes."]}),"\n"]}),"\n",(0,n.jsx)(s.h3,{id:i[18].id,children:i[18].value}),"\n",(0,n.jsx)(s.p,{children:"If these solutions don’t resolve your issue:"}),"\n",(0,n.jsxs)(s.ul,{children:["\n",(0,n.jsxs)(s.li,{children:["Check the ",(0,n.jsx)(s.a,{href:"https://github.com/monetr/monetr/issues",children:"monetr GitHub Issues"})," for similar problems."]}),"\n",(0,n.jsx)(s.li,{children:"Create a new issue with detailed logs and steps to reproduce the problem."}),"\n",(0,n.jsxs)(s.li,{children:["Reach out for assistance on ",(0,n.jsx)(s.a,{href:"https://discord.gg/68wTCXrhuq",children:"Discord"}),"."]}),"\n"]}),"\n",(0,n.jsx)(s.h2,{id:i[19].id,children:i[19].value}),"\n",(0,n.jsx)(l.P,{type:"warning",children:(0,n.jsx)(s.p,{children:"This will remove all of your data stored for monetr, please be careful as this data cannot be recovered unless you\nhave created a backup yourself somewhere."})}),"\n",(0,n.jsx)(s.p,{children:"To uninstall monetr via Docker Compose you can run the following command:"}),"\n",(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"",children:(0,n.jsx)(s.code,{children:(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"docker"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" compose"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" down"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:" --remove-orphans"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:" -v"})]})})})]})},"/documentation/install/docker",{filePath:"src/pages/documentation/install/docker.mdx",timestamp:1741908512e3,pageMap:t.O,frontMatter:{title:"Self-Host with Docker Compose",description:"Learn how to self-host monetr using Docker Compose. Follow step-by-step instructions to set up monetr, manage updates, and troubleshoot common issues for a seamless self-hosting experience."},title:"Self-Host with Docker Compose"},"undefined"==typeof RemoteContent?d:RemoteContent.useTOC)},1785:(e,i,s)=>{"use strict";s.d(i,{P:()=>o});var n=s(2540),r=s(1750),t=s(6877);let a={default:"\uD83D\uDCA1",error:"\uD83D\uDEAB",info:(0,n.jsx)(t.KS,{className:"_mt-1"}),warning:"⚠️"},l={default:(0,r.A)("_border-orange-100 _bg-orange-50 _text-orange-800 dark:_border-orange-400/30 dark:_bg-orange-400/20 dark:_text-orange-300"),error:(0,r.A)("_border-red-200 _bg-red-100 _text-red-900 dark:_border-red-200/30 dark:_bg-red-900/30 dark:_text-red-200"),info:(0,r.A)("_border-blue-200 _bg-blue-100 _text-blue-900 dark:_border-blue-200/30 dark:_bg-blue-900/30 dark:_text-blue-200"),warning:(0,r.A)("_border-yellow-100 _bg-yellow-50 _text-yellow-900 dark:_border-yellow-200/30 dark:_bg-yellow-700/30 dark:_text-yellow-200")};function o({children:e,type:i="default",emoji:s=a[i]}){return(0,n.jsxs)("div",{className:(0,r.A)("nextra-callout _overflow-x-auto _mt-6 _flex _rounded-lg _border _py-2 ltr:_pr-4 rtl:_pl-4","contrast-more:_border-current contrast-more:dark:_border-current",l[i]),children:[(0,n.jsx)("div",{className:"_select-none _text-xl ltr:_pl-3 ltr:_pr-2 rtl:_pr-3 rtl:_pl-2",style:{fontFamily:'"Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"'},children:s}),(0,n.jsx)("div",{className:"_w-full _min-w-0 _leading-7",children:e})]})}},8126:(e,i,s)=>{"use strict";s.d(i,{g:()=>a});var n=s(2540),r=s(1750),t=s(3696);function a({children:e,className:i,style:s,...a}){let l=(0,t.useId)().replaceAll(":","");return(0,n.jsx)("div",{className:(0,r.A)("nextra-steps _ms-4 _mb-12 _border-s _border-gray-200 _ps-6","dark:_border-neutral-800",i),style:{...s,"--counter-id":l},...a,children:e})}}},e=>{var i=i=>e(e.s=i);e.O(0,[7933,7170,636,6593,8792],()=>i(2162)),_N_E=e.O()}]);