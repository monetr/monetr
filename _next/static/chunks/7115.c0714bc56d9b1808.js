"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[7115],{7115:(e,t,o)=>{o.d(t,{diagram:()=>E});var l=o(5224),r=o(6357),a=o(1724),i=o(9523),n=o(6746);o(7693),o(7608),o(1699);let s=(e,t,o)=>{let{parentById:l}=o,r=new Set,a=e;for(;a;){if(r.add(a),a===t)return a;a=l[a]}for(a=t;a;){if(r.has(a))return a;a=l[a]}return"root"},d=new n,c={},h={},p={},u=async function(e,t,o,l,r,n,s){let d=o.select(`[id="${t}"]`).insert("g").attr("class","nodes"),c=Object.keys(e);return await Promise.all(c.map(async function(t){let o,s;let c=e[t],h="default";c.classes.length>0&&(h=c.classes.join(" ")),h+=" flowchart-label";let u=(0,i.k)(c.styles),b=void 0!==c.text?c.text:c.id,y={width:0,height:0},g=[{id:c.id+"-west",layoutOptions:{"port.side":"WEST"}},{id:c.id+"-east",layoutOptions:{"port.side":"EAST"}},{id:c.id+"-south",layoutOptions:{"port.side":"SOUTH"}},{id:c.id+"-north",layoutOptions:{"port.side":"NORTH"}}],f=0,w="",k={};switch(c.type){case"round":f=5,w="rect";break;case"square":case"group":default:w="rect";break;case"diamond":w="question",k={portConstraints:"FIXED_SIDE"};break;case"hexagon":w="hexagon";break;case"odd":case"odd_right":w="rect_left_inv_arrow";break;case"lean_right":w="lean_right";break;case"lean_left":w="lean_left";break;case"trapezoid":w="trapezoid";break;case"inv_trapezoid":w="inv_trapezoid";break;case"circle":w="circle";break;case"ellipse":w="ellipse";break;case"stadium":w="stadium";break;case"subroutine":w="subroutine";break;case"cylinder":w="cylinder";break;case"doublecircle":w="doublecircle"}let x={labelStyle:u.labelStyle,shape:w,labelText:b,labelType:c.labelType,rx:f,ry:f,class:h,style:u.style,id:c.id,link:c.link,linkTarget:c.linkTarget,tooltip:r.db.getTooltip(c.id)||"",domId:r.db.lookUpDomId(c.id),haveCallback:c.haveCallback,width:"group"===c.type?500:void 0,dir:c.dir,type:c.type,props:c.props,padding:(0,i.F)().flowchart.padding};if("group"!==x.type)o=(s=await (0,a.e)(d,x,c.dir)).node().getBBox();else{l.createElementNS("http://www.w3.org/2000/svg","text");let{shapeSvg:e,bbox:t}=await (0,a.l)(d,x,void 0,!0);y.width=t.width,y.wrappingWidth=(0,i.F)().flowchart.wrappingWidth,y.height=t.height,y.labelNode=e.node(),x.labelData=y}let m={id:c.id,ports:"diamond"===c.type?g:[],layoutOptions:k,labelText:b,labelData:y,domId:r.db.lookUpDomId(c.id),width:null==o?void 0:o.width,height:null==o?void 0:o.height,type:c.type,el:s,parent:n.parentById[c.id]};p[x.id]=m})),s},b=(e,t,o)=>{let l={TB:{in:{north:"north"},out:{south:"west",west:"east",east:"south"}},LR:{in:{west:"west"},out:{east:"south",south:"north",north:"east"}},RL:{in:{east:"east"},out:{west:"north",north:"south",south:"west"}},BT:{in:{south:"south"},out:{north:"east",east:"west",west:"north"}}};return l.TD=l.TB,l[o][t][e]},y=(e,t,o)=>{if(i.l.info("getNextPort",{node:e,edgeDirection:t,graphDirection:o}),!c[e])switch(o){case"TB":case"TD":c[e]={inPosition:"north",outPosition:"south"};break;case"BT":c[e]={inPosition:"south",outPosition:"north"};break;case"RL":c[e]={inPosition:"east",outPosition:"west"};break;case"LR":c[e]={inPosition:"west",outPosition:"east"}}let l="in"===t?c[e].inPosition:c[e].outPosition;return"in"===t?c[e].inPosition=b(c[e].inPosition,t,o):c[e].outPosition=b(c[e].outPosition,t,o),l},g=(e,t)=>{let o=e.start,l=e.end,r=o,a=l,i=p[o],n=p[l];return i&&n?("diamond"===i.type&&(o=`${o}-${y(o,"out",t)}`),"diamond"===n.type&&(l=`${l}-${y(l,"in",t)}`),{source:o,target:l,sourceId:r,targetId:a}):{source:o,target:l}},f=function(e,t,o,l){let n,s;i.l.info("abc78 edges = ",e);let d=l.insert("g").attr("class","edgeLabels"),c={},p=t.db.getDirection();if(void 0!==e.defaultStyle){let t=(0,i.k)(e.defaultStyle);n=t.style,s=t.labelStyle}return e.forEach(function(t){let l="L-"+t.start+"-"+t.end;void 0===c[l]?c[l]=0:c[l]++,i.l.info("abc78 new entry",l,c[l]);let u=l+"-"+c[l];i.l.info("abc78 new link id to be used is",l,u,c[l]);let b="LS-"+t.start,y="LE-"+t.end,f={style:"",labelStyle:""};switch(f.minlen=t.length||1,"arrow_open"===t.type?f.arrowhead="none":f.arrowhead="normal",f.arrowTypeStart="arrow_open",f.arrowTypeEnd="arrow_open",t.type){case"double_arrow_cross":f.arrowTypeStart="arrow_cross";case"arrow_cross":f.arrowTypeEnd="arrow_cross";break;case"double_arrow_point":f.arrowTypeStart="arrow_point";case"arrow_point":f.arrowTypeEnd="arrow_point";break;case"double_arrow_circle":f.arrowTypeStart="arrow_circle";case"arrow_circle":f.arrowTypeEnd="arrow_circle"}let w="",k="";switch(t.stroke){case"normal":w="fill:none;",void 0!==n&&(w=n),void 0!==s&&(k=s),f.thickness="normal",f.pattern="solid";break;case"dotted":f.thickness="normal",f.pattern="dotted",f.style="fill:none;stroke-width:2px;stroke-dasharray:3;";break;case"thick":f.thickness="thick",f.pattern="solid",f.style="stroke-width: 3.5px;fill:none;"}if(void 0!==t.style){let e=(0,i.k)(t.style);w=e.style,k=e.labelStyle}f.style=f.style+=w,f.labelStyle=f.labelStyle+=k,void 0!==t.interpolate?f.curve=(0,i.n)(t.interpolate,r.c_6):void 0!==e.defaultInterpolate?f.curve=(0,i.n)(e.defaultInterpolate,r.c_6):f.curve=(0,i.n)(h.curve,r.c_6),void 0===t.text?void 0!==t.style&&(f.arrowheadStyle="fill: #333"):(f.arrowheadStyle="fill: #333",f.labelpos="c"),f.labelType=t.labelType,f.label=t.text.replace(i.e.lineBreakRegex,"\n"),void 0===t.style&&(f.style=f.style||"stroke: #333; stroke-width: 1.5px;fill:none;"),f.labelStyle=f.labelStyle.replace("color:","fill:"),f.id=u,f.classes="flowchart-link "+b+" "+y;let x=(0,a.f)(d,f),{source:m,target:v,sourceId:T,targetId:$}=g(t,p);i.l.debug("abc78 source and target",m,v),o.edges.push({id:"e"+t.start+t.end,sources:[m],targets:[v],sourceId:T,targetId:$,labelEl:x,labels:[{width:f.width,height:f.height,orgWidth:f.width,orgHeight:f.height,text:f.label,layoutOptions:{"edgeLabels.inline":"true","edgeLabels.placement":"CENTER"}}],edgeData:f})}),o},w=function(e,t,o,l,r){let i="";l&&(i=(i=(i=window.location.protocol+"//"+window.location.host+window.location.pathname+window.location.search).replace(/\(/g,"\\(")).replace(/\)/g,"\\)")),(0,a.m)(e,t,i,r,o)},k=function(e){let t={parentById:{},childrenById:{}},o=e.getSubGraphs();return i.l.info("Subgraphs - ",o),o.forEach(function(e){e.nodes.forEach(function(o){t.parentById[o]=e.id,void 0===t.childrenById[e.id]&&(t.childrenById[e.id]=[]),t.childrenById[e.id].push(o)})}),o.forEach(function(e){e.id,void 0!==t.parentById[e.id]&&t.parentById[e.id]}),t},x=function(e,t,o){let l=s(e,t,o);if(void 0===l||"root"===l)return{x:0,y:0};let r=p[l].offset;return{x:r.posX,y:r.posY}},m=function(e,t,o,l,i,n){let s=x(t.sourceId,t.targetId,i),d=t.sections[0].startPoint,c=t.sections[0].endPoint,h=(t.sections[0].bendPoints?t.sections[0].bendPoints:[]).map(e=>[e.x+s.x,e.y+s.y]),p=[[d.x+s.x,d.y+s.y],...h,[c.x+s.x,c.y+s.y]],{x:u,y:b}=(0,a.k)(t.edgeData),y=(0,r.jvg)().x(u).y(b).curve(r.c_6),g=e.insert("path").attr("d",y(p)).attr("class","path "+o.classes).attr("fill","none"),f=e.insert("g").attr("class","edgeLabel"),k=(0,r.Ys)(f.node().appendChild(t.labelEl)),m=k.node().firstChild.getBoundingClientRect();k.attr("width",m.width),k.attr("height",m.height),f.attr("transform",`translate(${t.labels[0].x+s.x}, ${t.labels[0].y+s.y})`),w(g,o,l.type,l.arrowMarkerAbsolute,n)},v=(e,t)=>{e.forEach(e=>{e.children||(e.children=[]);let o=t.childrenById[e.id];o&&o.forEach(t=>{e.children.push(p[t])}),v(e.children,t)})},T=async function(e,t,o,l){var n;let s,h;l.db.clear(),p={},c={},l.db.setGen("gen-2"),l.parser.parse(e);let b=(0,r.Ys)("body").append("div").attr("style","height:400px").attr("id","cy"),y={id:"root",layoutOptions:{"elk.hierarchyHandling":"INCLUDE_CHILDREN","org.eclipse.elk.padding":"[top=100, left=100, bottom=110, right=110]","elk.layered.spacing.edgeNodeBetweenLayers":"30","elk.direction":"DOWN"},children:[],edges:[]};switch(i.l.info("Drawing flowchart using v3 renderer",d),l.db.getDirection()){case"BT":y.layoutOptions["elk.direction"]="UP";break;case"TB":y.layoutOptions["elk.direction"]="DOWN";break;case"LR":y.layoutOptions["elk.direction"]="RIGHT";break;case"RL":y.layoutOptions["elk.direction"]="LEFT"}let{securityLevel:g,flowchart:w}=(0,i.F)();"sandbox"===g&&(s=(0,r.Ys)("#i"+t));let x="sandbox"===g?(0,r.Ys)(s.nodes()[0].contentDocument.body):(0,r.Ys)("body"),T="sandbox"===g?s.nodes()[0].contentDocument:document,_=x.select(`[id="${t}"]`);(0,a.a)(_,["point","circle","cross"],l.type,t);let E=l.db.getVertices(),C=l.db.getSubGraphs();i.l.info("Subgraphs - ",C);for(let e=C.length-1;e>=0;e--)h=C[e],l.db.addVertex(h.id,{text:h.title,type:h.labelType},"group",void 0,h.classes,h.dir);let B=_.insert("g").attr("class","subgraphs"),S=k(l.db);y=await u(E,t,x,T,l,S,y);let I=_.insert("g").attr("class","edges edgePath");y=f(l.db.getEdges(),l,y,_),Object.keys(p).forEach(e=>{let t=p[e];t.parent||y.children.push(t),void 0!==S.childrenById[e]&&(t.labels=[{text:t.labelText,layoutOptions:{"nodeLabels.placement":"[H_CENTER, V_TOP, INSIDE]"},width:t.labelData.width,height:t.labelData.height}],delete t.x,delete t.y,delete t.width,delete t.height)}),v(y.children,S),i.l.info("after layout",JSON.stringify(y,null,2));let P=await d.layout(y);$(0,0,P.children,_,B,l,0),i.l.info("after layout",P),null==(n=P.edges)||n.map(e=>{m(I,e,e.edgeData,l,S,t)}),(0,i.o)({},_,w.diagramPadding,w.useMaxWidth),b.remove()},$=(e,t,o,l,r,a,n)=>{o.forEach(function(o){if(o){if(p[o.id].offset={posX:o.x+e,posY:o.y+t,x:e,y:t,depth:n,width:o.width,height:o.height},"group"===o.type){let l=r.insert("g").attr("class","subgraph");l.insert("rect").attr("class","subgraph subgraph-lvl-"+n%5+" node").attr("x",o.x+e).attr("y",o.y+t).attr("width",o.width).attr("height",o.height);let a=l.insert("g").attr("class","label"),s=(0,i.F)().flowchart.htmlLabels?o.labelData.width/2:0;a.attr("transform",`translate(${o.labels[0].x+e+o.x+s}, ${o.labels[0].y+t+o.y+3})`),a.node().appendChild(o.labelData.labelNode),i.l.info("Id (UGH)= ",o.type,o.labels)}else i.l.info("Id (UGH)= ",o.id),o.el.attr("transform",`translate(${o.x+e+o.width/2}, ${o.y+t+o.height/2})`)}}),o.forEach(function(o){o&&"group"===o.type&&$(e+o.x,t+o.y,o.children,l,r,a,n+1)})},_=e=>{let t="";for(let o=0;o<5;o++)t+=`
      .subgraph-lvl-${o} {
        fill: ${e[`surface${o}`]};
        stroke: ${e[`surfacePeer${o}`]};
      }
    `;return t},E={db:l.d,renderer:{getClasses:function(e,t){return i.l.info("Extracting classes"),t.db.getClasses()},draw:T},parser:l.p,styles:e=>`.label {
    font-family: ${e.fontFamily};
    color: ${e.nodeTextColor||e.textColor};
  }
  .cluster-label text {
    fill: ${e.titleColor};
  }
  .cluster-label span {
    color: ${e.titleColor};
  }

  .label text,span {
    fill: ${e.nodeTextColor||e.textColor};
    color: ${e.nodeTextColor||e.textColor};
  }

  .node rect,
  .node circle,
  .node ellipse,
  .node polygon,
  .node path {
    fill: ${e.mainBkg};
    stroke: ${e.nodeBorder};
    stroke-width: 1px;
  }

  .node .label {
    text-align: center;
  }
  .node.clickable {
    cursor: pointer;
  }

  .arrowheadPath {
    fill: ${e.arrowheadColor};
  }

  .edgePath .path {
    stroke: ${e.lineColor};
    stroke-width: 2.0px;
  }

  .flowchart-link {
    stroke: ${e.lineColor};
    fill: none;
  }

  .edgeLabel {
    background-color: ${e.edgeLabelBackground};
    rect {
      opacity: 0.85;
      background-color: ${e.edgeLabelBackground};
      fill: ${e.edgeLabelBackground};
    }
    text-align: center;
  }

  .cluster rect {
    fill: ${e.clusterBkg};
    stroke: ${e.clusterBorder};
    stroke-width: 1px;
  }

  .cluster text {
    fill: ${e.titleColor};
  }

  .cluster span {
    color: ${e.titleColor};
  }
  /* .cluster div {
    color: ${e.titleColor};
  } */

  div.mermaidTooltip {
    position: absolute;
    text-align: center;
    max-width: 200px;
    padding: 2px;
    font-family: ${e.fontFamily};
    font-size: 12px;
    background: ${e.tertiaryColor};
    border: 1px solid ${e.border2};
    border-radius: 2px;
    pointer-events: none;
    z-index: 100;
  }

  .flowchartTitleText {
    text-anchor: middle;
    font-size: 18px;
    fill: ${e.textColor};
  }
  .subgraph {
    stroke-width:2;
    rx:3;
  }
  // .subgraph-lvl-1 {
  //   fill:#ccc;
  //   // stroke:black;
  // }

  .flowchart-label text {
    text-anchor: middle;
  }

  ${_(e)}
`}}}]);