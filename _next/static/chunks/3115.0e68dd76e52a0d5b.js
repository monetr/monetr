"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[3115],{3115:function(e,t,r){r.d(t,{diagram:function(){return $}});var a=r(5116),o=r(6357),l=r(7656),i=r(6388),n=r(6746);r(7693),r(7608),r(1699);let s=(e,t,r)=>{let{parentById:a}=r,o=new Set,l=e;for(;l;){if(o.add(l),l===t)return l;l=a[l]}for(l=t;l;){if(o.has(l))return l;l=a[l]}return"root"},d=new n,c={},h={},p={},u=async function(e,t,r,a,o,n,s){let d=r.select(`[id="${t}"]`),c=d.insert("g").attr("class","nodes"),h=Object.keys(e);return await Promise.all(h.map(async function(t){let r,s;let d=e[t],h="default";d.classes.length>0&&(h=d.classes.join(" ")),h+=" flowchart-label";let u=(0,i.k)(d.styles),b=void 0!==d.text?d.text:d.id,g={width:0,height:0},y=[{id:d.id+"-west",layoutOptions:{"port.side":"WEST"}},{id:d.id+"-east",layoutOptions:{"port.side":"EAST"}},{id:d.id+"-south",layoutOptions:{"port.side":"SOUTH"}},{id:d.id+"-north",layoutOptions:{"port.side":"NORTH"}}],w=0,k="",f={};switch(d.type){case"round":w=5,k="rect";break;case"square":case"group":default:k="rect";break;case"diamond":k="question",f={portConstraints:"FIXED_SIDE"};break;case"hexagon":k="hexagon";break;case"odd":case"odd_right":k="rect_left_inv_arrow";break;case"lean_right":k="lean_right";break;case"lean_left":k="lean_left";break;case"trapezoid":k="trapezoid";break;case"inv_trapezoid":k="inv_trapezoid";break;case"circle":k="circle";break;case"ellipse":k="ellipse";break;case"stadium":k="stadium";break;case"subroutine":k="subroutine";break;case"cylinder":k="cylinder";break;case"doublecircle":k="doublecircle"}let x={labelStyle:u.labelStyle,shape:k,labelText:b,labelType:d.labelType,rx:w,ry:w,class:h,style:u.style,id:d.id,link:d.link,linkTarget:d.linkTarget,tooltip:o.db.getTooltip(d.id)||"",domId:o.db.lookUpDomId(d.id),haveCallback:d.haveCallback,width:"group"===d.type?500:void 0,dir:d.dir,type:d.type,props:d.props,padding:(0,i.E)().flowchart.padding};if("group"!==x.type)r=(s=await (0,l.e)(c,x,d.dir)).node().getBBox();else{a.createElementNS("http://www.w3.org/2000/svg","text");let{shapeSvg:e,bbox:t}=await (0,l.l)(c,x,void 0,!0);g.width=t.width,g.wrappingWidth=(0,i.E)().flowchart.wrappingWidth,g.height=t.height,g.labelNode=e.node(),x.labelData=g}let m={id:d.id,ports:"diamond"===d.type?y:[],layoutOptions:f,labelText:b,labelData:g,domId:o.db.lookUpDomId(d.id),width:null==r?void 0:r.width,height:null==r?void 0:r.height,type:d.type,el:s,parent:n.parentById[d.id]};p[x.id]=m})),s},b=(e,t,r)=>{let a={TB:{in:{north:"north"},out:{south:"west",west:"east",east:"south"}},LR:{in:{west:"west"},out:{east:"south",south:"north",north:"east"}},RL:{in:{east:"east"},out:{west:"north",north:"south",south:"west"}},BT:{in:{south:"south"},out:{north:"east",east:"west",west:"north"}}};return a.TD=a.TB,i.l.info("abc88",r,t,e),a[r][t][e]},g=(e,t,r)=>{if(i.l.info("getNextPort abc88",{node:e,edgeDirection:t,graphDirection:r}),!c[e])switch(r){case"TB":case"TD":c[e]={inPosition:"north",outPosition:"south"};break;case"BT":c[e]={inPosition:"south",outPosition:"north"};break;case"RL":c[e]={inPosition:"east",outPosition:"west"};break;case"LR":c[e]={inPosition:"west",outPosition:"east"}}let a="in"===t?c[e].inPosition:c[e].outPosition;return"in"===t?c[e].inPosition=b(c[e].inPosition,t,r):c[e].outPosition=b(c[e].outPosition,t,r),a},y=(e,t)=>{let r=e.start,a=e.end,o=r,l=a,i=p[r],n=p[a];return i&&n?("diamond"===i.type&&(r=`${r}-${g(r,"out",t)}`),"diamond"===n.type&&(a=`${a}-${g(a,"in",t)}`),{source:r,target:a,sourceId:o,targetId:l}):{source:r,target:a}},w=function(e,t,r,a){let n,s;i.l.info("abc78 edges = ",e);let d=a.insert("g").attr("class","edgeLabels"),c={},p=t.db.getDirection();if(void 0!==e.defaultStyle){let t=(0,i.k)(e.defaultStyle);n=t.style,s=t.labelStyle}return e.forEach(function(t){let a="L-"+t.start+"-"+t.end;void 0===c[a]?c[a]=0:c[a]++,i.l.info("abc78 new entry",a,c[a]);let u=a+"-"+c[a];i.l.info("abc78 new link id to be used is",a,u,c[a]);let b="LS-"+t.start,g="LE-"+t.end,w={style:"",labelStyle:""};switch(w.minlen=t.length||1,"arrow_open"===t.type?w.arrowhead="none":w.arrowhead="normal",w.arrowTypeStart="arrow_open",w.arrowTypeEnd="arrow_open",t.type){case"double_arrow_cross":w.arrowTypeStart="arrow_cross";case"arrow_cross":w.arrowTypeEnd="arrow_cross";break;case"double_arrow_point":w.arrowTypeStart="arrow_point";case"arrow_point":w.arrowTypeEnd="arrow_point";break;case"double_arrow_circle":w.arrowTypeStart="arrow_circle";case"arrow_circle":w.arrowTypeEnd="arrow_circle"}let k="",f="";switch(t.stroke){case"normal":k="fill:none;",void 0!==n&&(k=n),void 0!==s&&(f=s),w.thickness="normal",w.pattern="solid";break;case"dotted":w.thickness="normal",w.pattern="dotted",w.style="fill:none;stroke-width:2px;stroke-dasharray:3;";break;case"thick":w.thickness="thick",w.pattern="solid",w.style="stroke-width: 3.5px;fill:none;"}if(void 0!==t.style){let e=(0,i.k)(t.style);k=e.style,f=e.labelStyle}w.style=w.style+=k,w.labelStyle=w.labelStyle+=f,void 0!==t.interpolate?w.curve=(0,i.n)(t.interpolate,o.c_6):void 0!==e.defaultInterpolate?w.curve=(0,i.n)(e.defaultInterpolate,o.c_6):w.curve=(0,i.n)(h.curve,o.c_6),void 0===t.text?void 0!==t.style&&(w.arrowheadStyle="fill: #333"):(w.arrowheadStyle="fill: #333",w.labelpos="c"),w.labelType=t.labelType,w.label=t.text.replace(i.e.lineBreakRegex,"\n"),void 0===t.style&&(w.style=w.style||"stroke: #333; stroke-width: 1.5px;fill:none;"),w.labelStyle=w.labelStyle.replace("color:","fill:"),w.id=u,w.classes="flowchart-link "+b+" "+g;let x=(0,l.f)(d,w),{source:m,target:_,sourceId:E,targetId:T}=y(t,p);i.l.debug("abc78 source and target",m,_),r.edges.push({id:"e"+t.start+t.end,sources:[m],targets:[_],sourceId:E,targetId:T,labelEl:x,labels:[{width:w.width,height:w.height,orgWidth:w.width,orgHeight:w.height,text:w.label,layoutOptions:{"edgeLabels.inline":"true","edgeLabels.placement":"CENTER"}}],edgeData:w})}),r},k=function(e,t,r,a,o){let l="";switch(a&&(l=(l=(l=window.location.protocol+"//"+window.location.host+window.location.pathname+window.location.search).replace(/\(/g,"\\(")).replace(/\)/g,"\\)")),t.arrowTypeStart){case"arrow_cross":e.attr("marker-start","url("+l+"#"+o+"_"+r+"-crossStart)");break;case"arrow_point":e.attr("marker-start","url("+l+"#"+o+"_"+r+"-pointStart)");break;case"arrow_barb":e.attr("marker-start","url("+l+"#"+o+"_"+r+"-barbStart)");break;case"arrow_circle":e.attr("marker-start","url("+l+"#"+o+"_"+r+"-circleStart)");break;case"aggregation":e.attr("marker-start","url("+l+"#"+o+"_"+r+"-aggregationStart)");break;case"extension":e.attr("marker-start","url("+l+"#"+o+"_"+r+"-extensionStart)");break;case"composition":e.attr("marker-start","url("+l+"#"+o+"_"+r+"-compositionStart)");break;case"dependency":e.attr("marker-start","url("+l+"#"+o+"_"+r+"-dependencyStart)");break;case"lollipop":e.attr("marker-start","url("+l+"#"+o+"_"+r+"-lollipopStart)")}switch(t.arrowTypeEnd){case"arrow_cross":e.attr("marker-end","url("+l+"#"+o+"_"+r+"-crossEnd)");break;case"arrow_point":e.attr("marker-end","url("+l+"#"+o+"_"+r+"-pointEnd)");break;case"arrow_barb":e.attr("marker-end","url("+l+"#"+o+"_"+r+"-barbEnd)");break;case"arrow_circle":e.attr("marker-end","url("+l+"#"+o+"_"+r+"-circleEnd)");break;case"aggregation":e.attr("marker-end","url("+l+"#"+o+"_"+r+"-aggregationEnd)");break;case"extension":e.attr("marker-end","url("+l+"#"+o+"_"+r+"-extensionEnd)");break;case"composition":e.attr("marker-end","url("+l+"#"+o+"_"+r+"-compositionEnd)");break;case"dependency":e.attr("marker-end","url("+l+"#"+o+"_"+r+"-dependencyEnd)");break;case"lollipop":e.attr("marker-end","url("+l+"#"+o+"_"+r+"-lollipopEnd)")}},f=function(e){let t={parentById:{},childrenById:{}},r=e.getSubGraphs();return i.l.info("Subgraphs - ",r),r.forEach(function(e){e.nodes.forEach(function(r){t.parentById[r]=e.id,void 0===t.childrenById[e.id]&&(t.childrenById[e.id]=[]),t.childrenById[e.id].push(r)})}),r.forEach(function(e){e.id,void 0!==t.parentById[e.id]&&t.parentById[e.id]}),t},x=function(e,t,r){let a=s(e,t,r);if(void 0===a||"root"===a)return{x:0,y:0};let o=p[a].offset;return{x:o.posX,y:o.posY}},m=function(e,t,r,a,i,n){let s=x(t.sourceId,t.targetId,i),d=t.sections[0].startPoint,c=t.sections[0].endPoint,h=t.sections[0].bendPoints?t.sections[0].bendPoints:[],p=h.map(e=>[e.x+s.x,e.y+s.y]),u=[[d.x+s.x,d.y+s.y],...p,[c.x+s.x,c.y+s.y]],{x:b,y:g}=(0,l.j)(t.edgeData),y=(0,o.jvg)().x(b).y(g).curve(o.c_6),w=e.insert("path").attr("d",y(u)).attr("class","path "+r.classes).attr("fill","none"),f=e.insert("g").attr("class","edgeLabel"),m=(0,o.Ys)(f.node().appendChild(t.labelEl)),_=m.node().firstChild.getBoundingClientRect();m.attr("width",_.width),m.attr("height",_.height),f.attr("transform",`translate(${t.labels[0].x+s.x}, ${t.labels[0].y+s.y})`),k(w,r,a.type,a.arrowMarkerAbsolute,n)},_=(e,t)=>{e.forEach(e=>{e.children||(e.children=[]);let r=t.childrenById[e.id];r&&r.forEach(t=>{e.children.push(p[t])}),_(e.children,t)})},E=async function(e,t,r,a){var n;let s,h;a.db.clear(),p={},c={},a.db.setGen("gen-2"),a.parser.parse(e);let b=(0,o.Ys)("body").append("div").attr("style","height:400px").attr("id","cy"),g={id:"root",layoutOptions:{"elk.hierarchyHandling":"INCLUDE_CHILDREN","org.eclipse.elk.padding":"[top=100, left=100, bottom=110, right=110]","elk.layered.spacing.edgeNodeBetweenLayers":"30","elk.direction":"DOWN"},children:[],edges:[]};switch(i.l.info("Drawing flowchart using v3 renderer",d),a.db.getDirection()){case"BT":g.layoutOptions["elk.direction"]="UP";break;case"TB":g.layoutOptions["elk.direction"]="DOWN";break;case"LR":g.layoutOptions["elk.direction"]="RIGHT";break;case"RL":g.layoutOptions["elk.direction"]="LEFT"}let{securityLevel:y,flowchart:k}=(0,i.E)();"sandbox"===y&&(s=(0,o.Ys)("#i"+t));let x="sandbox"===y?(0,o.Ys)(s.nodes()[0].contentDocument.body):(0,o.Ys)("body"),E="sandbox"===y?s.nodes()[0].contentDocument:document,v=x.select(`[id="${t}"]`);(0,l.a)(v,["point","circle","cross"],a.type,t);let $=a.db.getVertices(),S=a.db.getSubGraphs();i.l.info("Subgraphs - ",S);for(let e=S.length-1;e>=0;e--)h=S[e],a.db.addVertex(h.id,{text:h.title,type:h.labelType},"group",void 0,h.classes,h.dir);let C=v.insert("g").attr("class","subgraphs"),B=f(a.db);g=await u($,t,x,E,a,B,g);let I=v.insert("g").attr("class","edges edgePath"),P=a.db.getEdges();g=w(P,a,g,v);let D=Object.keys(p);D.forEach(e=>{let t=p[e];t.parent||g.children.push(t),void 0!==B.childrenById[e]&&(t.labels=[{text:t.labelText,layoutOptions:{"nodeLabels.placement":"[H_CENTER, V_TOP, INSIDE]"},width:t.labelData.width,height:t.labelData.height}],delete t.x,delete t.y,delete t.width,delete t.height)}),_(g.children,B),i.l.info("after layout",JSON.stringify(g,null,2));let L=await d.layout(g);T(0,0,L.children,v,C,a,0),i.l.info("after layout",L),null==(n=L.edges)||n.map(e=>{m(I,e,e.edgeData,a,B,t)}),(0,i.o)({},v,k.diagramPadding,k.useMaxWidth),b.remove()},T=(e,t,r,a,o,l,n)=>{r.forEach(function(r){if(r){if(p[r.id].offset={posX:r.x+e,posY:r.y+t,x:e,y:t,depth:n,width:r.width,height:r.height},"group"===r.type){let a=o.insert("g").attr("class","subgraph");a.insert("rect").attr("class","subgraph subgraph-lvl-"+n%5+" node").attr("x",r.x+e).attr("y",r.y+t).attr("width",r.width).attr("height",r.height);let l=a.insert("g").attr("class","label"),s=(0,i.E)().flowchart.htmlLabels?r.labelData.width/2:0;l.attr("transform",`translate(${r.labels[0].x+e+r.x+s}, ${r.labels[0].y+t+r.y+3})`),l.node().appendChild(r.labelData.labelNode),i.l.info("Id (UGH)= ",r.type,r.labels)}else i.l.info("Id (UGH)= ",r.id),r.el.attr("transform",`translate(${r.x+e+r.width/2}, ${r.y+t+r.height/2})`)}}),r.forEach(function(r){r&&"group"===r.type&&T(e+r.x,t+r.y,r.children,a,o,l,n+1)})},v=e=>{let t="";for(let r=0;r<5;r++)t+=`
      .subgraph-lvl-${r} {
        fill: ${e[`surface${r}`]};
        stroke: ${e[`surfacePeer${r}`]};
      }
    `;return t},$={db:a.d,renderer:{getClasses:function(e,t){return i.l.info("Extracting classes"),t.db.getClasses()},draw:E},parser:a.p,styles:e=>`.label {
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

  ${v(e)}
`}}}]);