(self.webpackChunk_monetr_stories=self.webpackChunk_monetr_stories||[]).push([["main"],{"../tailwind.config.js":function(e,r,t){e.exports={important:!0,darkMode:"class",future:{hoverOnlyWhenSupported:!0},plugins:[t("../node_modules/tailwindcss-animate/index.js")],theme:{extend:{backgroundImage:{"gradient-radial":"radial-gradient(var(--tw-gradient-stops))","gradient-conic":"conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))"},keyframes:{"accordion-down":{from:{height:"0"},to:{height:"var(--radix-accordion-content-height)"}},"accordion-up":{from:{height:"var(--radix-accordion-content-height)"},to:{height:"0"}}},animation:{"accordion-down":"accordion-down 0.2s ease-out","accordion-up":"accordion-up 0.2s ease-out"},colors:{border:"hsl(var(--border))",input:"hsl(var(--input))",ring:"hsl(var(--ring))",background:"hsl(var(--background))",foreground:"hsl(var(--foreground))",primary:{DEFAULT:"hsl(var(--primary))",foreground:"hsl(var(--primary-foreground))"},secondary:{DEFAULT:"hsl(var(--secondary))",foreground:"hsl(var(--secondary-foreground))"},destructive:{DEFAULT:"hsl(var(--destructive))",foreground:"hsl(var(--destructive-foreground))"},muted:{DEFAULT:"hsl(var(--muted))",foreground:"hsl(var(--muted-foreground))"},accent:{DEFAULT:"hsl(var(--accent))",foreground:"hsl(var(--accent-foreground))"},popover:{DEFAULT:"hsl(var(--popover))",foreground:"hsl(var(--popover-foreground))"},card:{DEFAULT:"hsl(var(--card))",foreground:"hsl(var(--card-foreground))"},monetr:{brand:{DEFAULT:"#4E1AA0"},background:{subtle:"",DEFAULT:"#F8F8F8",emphasis:""},border:{DEFAULT:""},content:{subtle:"#6b7280",DEFAULT:"#111827",emphasis:""}},"dark-monetr":{red:{DEFAULT:"#ef4444"},green:{DEFAULT:"#22c55e"},blue:{DEFAULT:"#3b82f6"},brand:{bright:"#CFB9F4",faint:"#AC84EB",muted:"#9461E5",subtle:"#5D1FC1",DEFAULT:"#4E1AA0"},background:{subtle:"#27272a",DEFAULT:"#19161f",emphasis:"#3f3f46",focused:"#131118",bright:"#fafafa"},border:{subtle:"#27272a",DEFAULT:"#3f3f46",string:"#71717a"},content:{muted:"#52525b",subtle:"#a1a1aa",DEFAULT:"#d4d4d8",emphasis:"#fafafa"},popover:{DEFAULT:"0 0% 100%",foreground:"222.2 47.4% 11.2%"}}}},aspectRatio:{"video-vertical":"9/16"},animation:{"ping-slow":"ping 2s cubic-bezier(0, 0, 0.2, 1) infinite"},colors:{inherit:"inherit",current:"currentColor",transparent:"transparent",black:"#000",white:"#fff",slate:{50:"#f8fafc",100:"#f1f5f9",200:"#e2e8f0",300:"#cbd5e1",400:"#94a3b8",500:"#64748b",600:"#475569",700:"#334155",800:"#1e293b",900:"#0f172a"},gray:{50:"#f9fafb",100:"#f3f4f6",200:"#e5e7eb",300:"#d1d5db",400:"#9ca3af",500:"#6b7280",600:"#4b5563",700:"#374151",800:"#1f2937",900:"#111827"},zinc:{50:"#fafafa",100:"#f4f4f5",200:"#e4e4e7",300:"#d4d4d8",400:"#a1a1aa",500:"#71717a",600:"#52525b",700:"#3f3f46",800:"#27272a",900:"#19161f"},neutral:{50:"#fafafa",100:"#f5f5f5",200:"#e5e5e5",300:"#d4d4d4",400:"#a3a3a3",500:"#737373",600:"#525252",700:"#404040",800:"#262626",900:"#171717"},stone:{50:"#fafaf9",100:"#f5f5f4",200:"#e7e5e4",300:"#d6d3d1",400:"#a8a29e",500:"#78716c",600:"#57534e",700:"#44403c",800:"#292524",900:"#1c1917"},red:{50:"#fef2f2",100:"#fee2e2",200:"#fecaca",300:"#fca5a5",400:"#f87171",500:"#ef4444",600:"#dc2626",700:"#b91c1c",800:"#991b1b",900:"#7f1d1d"},orange:{50:"#fff7ed",100:"#ffedd5",200:"#fed7aa",300:"#fdba74",400:"#fb923c",500:"#f97316",600:"#ea580c",700:"#c2410c",800:"#9a3412",900:"#7c2d12"},amber:{50:"#fffbeb",100:"#fef3c7",200:"#fde68a",300:"#fcd34d",400:"#fbbf24",500:"#f59e0b",600:"#d97706",700:"#b45309",800:"#92400e",900:"#78350f"},yellow:{50:"#fefce8",100:"#fef9c3",200:"#fef08a",300:"#fde047",400:"#facc15",500:"#eab308",600:"#ca8a04",700:"#a16207",800:"#854d0e",900:"#713f12"},lime:{50:"#f7fee7",100:"#ecfccb",200:"#d9f99d",300:"#bef264",400:"#a3e635",500:"#84cc16",600:"#65a30d",700:"#4d7c0f",800:"#3f6212",900:"#365314"},green:{50:"#f0fdf4",100:"#dcfce7",200:"#bbf7d0",300:"#86efac",400:"#4ade80",500:"#22c55e",600:"#16a34a",700:"#15803d",800:"#166534",900:"#14532d"},emerald:{50:"#ecfdf5",100:"#d1fae5",200:"#a7f3d0",300:"#6ee7b7",400:"#34d399",500:"#10b981",600:"#059669",700:"#047857",800:"#065f46",900:"#064e3b"},teal:{50:"#f0fdfa",100:"#ccfbf1",200:"#99f6e4",300:"#5eead4",400:"#2dd4bf",500:"#14b8a6",600:"#0d9488",700:"#0f766e",800:"#115e59",900:"#134e4a"},cyan:{50:"#ecfeff",100:"#cffafe",200:"#a5f3fc",300:"#67e8f9",400:"#22d3ee",500:"#06b6d4",600:"#0891b2",700:"#0e7490",800:"#155e75",900:"#164e63"},sky:{50:"#f0f9ff",100:"#e0f2fe",200:"#bae6fd",300:"#7dd3fc",400:"#38bdf8",500:"#0ea5e9",600:"#0284c7",700:"#0369a1",800:"#075985",900:"#0c4a6e"},blue:{50:"#eff6ff",100:"#dbeafe",200:"#bfdbfe",300:"#93c5fd",400:"#60a5fa",500:"#3b82f6",600:"#2563eb",700:"#1d4ed8",800:"#1e40af",900:"#1e3a8a"},indigo:{50:"#eef2ff",100:"#e0e7ff",200:"#c7d2fe",300:"#a5b4fc",400:"#818cf8",500:"#6366f1",600:"#4f46e5",700:"#4338ca",800:"#3730a3",900:"#312e81"},violet:{50:"#f5f3ff",100:"#ede9fe",200:"#ddd6fe",300:"#c4b5fd",400:"#a78bfa",500:"#8b5cf6",600:"#7c3aed",700:"#6d28d9",800:"#5b21b6",900:"#4c1d95"},purple:{50:"#EDE5FB",100:"#D8C6F6",200:"#B591ED",300:"#8E58E4",400:"#6823D7",500:"#4E1AA0",600:"#3E157F",700:"#2F1060",800:"#200B42",900:"#0F051F"},fuchsia:{50:"#fdf4ff",100:"#fae8ff",200:"#f5d0fe",300:"#f0abfc",400:"#e879f9",500:"#d946ef",600:"#c026d3",700:"#a21caf",800:"#86198f",900:"#701a75"},pink:{50:"#fdf2f8",100:"#fce7f3",200:"#fbcfe8",300:"#f9a8d4",400:"#f472b6",500:"#ec4899",600:"#db2777",700:"#be185d",800:"#9d174d",900:"#831843"},rose:{50:"#fff1f2",100:"#ffe4e6",200:"#fecdd3",300:"#fda4af",400:"#fb7185",500:"#f43f5e",600:"#e11d48",700:"#be123c",800:"#9f1239",900:"#881337"}}}}},"../interface/src/styles/styles.css":function(e,r,t){e.hot.accept()},"./preview.css":function(e,r,t){e.hot.accept()},"@storybook/addons":function(e,r,t){e.exports=__STORYBOOK_MODULE_ADDONS__},"@storybook/channel-postmessage":function(e,r,t){e.exports=__STORYBOOK_MODULE_CHANNEL_POSTMESSAGE__},"@storybook/channel-websocket":function(e,r,t){e.exports=__STORYBOOK_MODULE_CHANNEL_WEBSOCKET__},"@storybook/client-logger":function(e,r,t){e.exports=__STORYBOOK_MODULE_CLIENT_LOGGER__},"@storybook/core-events":function(e,r,t){e.exports=__STORYBOOK_MODULE_CORE_EVENTS__},"@storybook/preview-api":function(e,r,t){e.exports=__STORYBOOK_MODULE_PREVIEW_API__},"../interface/tailwind.config.cjs":function(e,r,t){let a=t("../tailwind.config.js");e.exports={...a,content:["src/**/*.@(js|jsx|ts|tsx)"]}},"./preview.tsx":function(e,r,t){"use strict";Object.defineProperty(r,"__esModule",{value:!0});!function(e,r){for(var t in r)Object.defineProperty(e,t,{enumerable:!0,get:r[t]})}(r,{useTheme:function(){return _},default:function(){return y},globalTypes:function(){return E}});var a=t("../node_modules/react/jsx-runtime.js");t("../node_modules/@fontsource-variable/inter/index.css"),t("../node_modules/react/index.js");var o=t.ir(t("../node_modules/@ebay/nice-modal-react/lib/esm/index.js")),n=t.ir(t("../node_modules/@mui/icons-material/Done.js")),s=t.ir(t("../node_modules/@mui/icons-material/Error.js")),f=t.ir(t("../node_modules/@mui/icons-material/Info.js")),d=t.ir(t("../node_modules/@mui/icons-material/Warning.js")),i=t("../node_modules/@mui/material/index.js"),c=t("../node_modules/@storybook/addon-viewport/dist/index.mjs"),u=t("@storybook/addons"),l=t("../node_modules/@tanstack/react-query/build/lib/index.mjs"),b=t("../node_modules/notistack/notistack.esm.js"),m=t("../interface/src/theme.ts"),p=t.ir(t("../interface/src/util/query.ts")),h=t("../node_modules/storybook-addon-react-router-v6/dist/index.mjs"),g=t("../node_modules/storycap/lib-esm/index.js");t("../interface/src/styles/styles.css"),t("./preview.css");let _=e=>{let[{theme:r}]=(0,u.useGlobals)();return(0,u.useEffect)(()=>{document.querySelector("html")?.setAttribute("class",r||"dark")},[r]),e()},v={decorators:[g.withScreenshot,_,(e,r)=>{let t={error:(0,a.jsx)(s.default,{className:"mr-2.5"}),success:(0,a.jsx)(n.default,{className:"mr-2.5"}),warning:(0,a.jsx)(d.default,{className:"mr-2.5"}),info:(0,a.jsx)(f.default,{className:"mr-2.5"})},c=new l.QueryClient({defaultOptions:{queries:{staleTime:6e5,queryFn:p.default}}});return(0,a.jsx)(l.QueryClientProvider,{client:c,children:(0,a.jsx)(i.ThemeProvider,{theme:m.newTheme,children:(0,a.jsx)(b.SnackbarProvider,{maxSnack:5,iconVariant:t,children:(0,a.jsxs)(o.default.Provider,{children:[(0,a.jsx)(i.CssBaseline,{}),(0,a.jsx)(e,{})]})})})})},h.withRouter],args:{},parameters:{screenshot:{viewport:{width:1280,height:720,isMobile:!1,hasTouch:!1},delay:3e3},viewport:{viewports:{desktop:{name:"Desktop",styles:{width:"1280px",height:"720px"}},...c.INITIAL_VIEWPORTS}},actions:{argTypesRegex:"^on[A-Z].*"},controls:{matchers:{color:/(background|color)$/i,date:/Date$/}}}};var y=v;let E={theme:{name:"Toggle theme",description:"Global theme for components",defaultValue:"dark",toolbar:{icon:"circlehollow",items:["dark"],showName:!0,dynamicTitle:!0}}}},"../interface/src/api/api.ts":function(e,r,t){"use strict";Object.defineProperty(r,"__esModule",{value:!0});!function(e,r){for(var t in r)Object.defineProperty(e,t,{enumerable:!0,get:r[t]})}(r,{NewClient:function(){return n},default:function(){return f}});var a=t.ir(t("../node_modules/@sentry/react/build/esm/index.js")),o=t.ir(t("../node_modules/axios/index.js"));function n(e){let r=o.default.create(e);return r.interceptors.request.use(e=>e,e=>Promise.reject(e)),r.interceptors.response.use(e=>e,e=>(500===e.response.status&&a.captureException(e),Promise.reject(e))),r}let s=n({baseURL:"/api"});var f=s},"../interface/src/theme.ts":function(e,r,t){"use strict";Object.defineProperty(r,"__esModule",{value:!0});!function(e,r){for(var t in r)Object.defineProperty(e,t,{enumerable:!0,get:r[t]})}(r,{newTheme:function(){return i},default:function(){return u}});var a=t("../node_modules/@mui/material/index.js"),o=t.ir(t("../interface/tailwind.config.cjs")),n=t.ir(t("../node_modules/tailwindcss/resolveConfig.js"));let s=(0,n.default)(o.default),f=s.theme.colors.purple["500"],d="#FF5798",i=(0,a.createTheme)({typography:{fontFamily:"Inter Variable,Helvetica,Arial,sans-serif"},shape:{borderRadius:10},palette:{mode:"dark",text:{secondary:s.theme.colors.zinc["50"]},background:{default:s.theme.colors.zinc["900"],paper:s.theme.colors.zinc["900"]}}}),c=(0,a.createTheme)({typography:{fontFamily:"Helvetica,Arial,sans-serif"},shape:{borderRadius:10},components:{MuiAppBar:{styleOverrides:{root:{backgroundColor:f,backgroundImage:"none"}}},MuiInputBase:{styleOverrides:{root:{height:56}}},MuiTextField:{styleOverrides:{root:{height:56}}},MuiInputLabel:{styleOverrides:{root:{}}}},palette:{mode:"light",primary:{main:f,light:f,dark:"#712ddd",contrastText:"#FFFFFF"},secondary:{main:d,contrastText:"#FFFFFF"},background:{default:"#F8F8F8"}}});var u=c},"../interface/src/util/query.ts":function(e,r,t){"use strict";Object.defineProperty(r,"__esModule",{value:!0}),Object.defineProperty(r,"default",{enumerable:!0,get:function(){return o}});var a=t.ir(t("../interface/src/util/request.ts"));async function o(e){let r=1===e.queryKey.length?"GET":"POST",{data:t}=await (0,a.default)().request({url:`${e.queryKey[0]}`,method:r,params:e.pageParam&&{offset:e.pageParam},data:2===e.queryKey.length&&e.queryKey[1]}).catch(e=>{switch(e.response.status){case 404:case 500:case 502:throw e;default:return e.response}});return t}},"../interface/src/util/request.ts":function(e,r,t){"use strict";Object.defineProperty(r,"__esModule",{value:!0}),Object.defineProperty(r,"default",{enumerable:!0,get:function(){return o}});var a=t.ir(t("../interface/src/api/api.ts"));function o(){return a.default}}},function(e){e.O(0,["4"],function(){return e(e.s="./node_modules/.rspack-virtual-module/storybook-config-entry.js")}),e.O()}]);