"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[6211],{6211:function(t,e,r){r.d(e,{diagram:function(){return k}});var a=r(6388),i=r(8472),n=r(6357),s=r(9500);let l=[];for(let t=0;t<256;++t)l.push((t+256).toString(16).slice(1));function unsafeStringify(t,e=0){return l[t[e+0]]+l[t[e+1]]+l[t[e+2]]+l[t[e+3]]+"-"+l[t[e+4]]+l[t[e+5]]+"-"+l[t[e+6]]+l[t[e+7]]+"-"+l[t[e+8]]+l[t[e+9]]+"-"+l[t[e+10]]+l[t[e+11]]+l[t[e+12]]+l[t[e+13]]+l[t[e+14]]+l[t[e+15]]}var c=/^(?:[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}|00000000-0000-0000-0000-000000000000)$/i,esm_browser_parse=function(t){let e;if(!("string"==typeof t&&c.test(t)))throw TypeError("Invalid UUID");let r=new Uint8Array(16);return r[0]=(e=parseInt(t.slice(0,8),16))>>>24,r[1]=e>>>16&255,r[2]=e>>>8&255,r[3]=255&e,r[4]=(e=parseInt(t.slice(9,13),16))>>>8,r[5]=255&e,r[6]=(e=parseInt(t.slice(14,18),16))>>>8,r[7]=255&e,r[8]=(e=parseInt(t.slice(19,23),16))>>>8,r[9]=255&e,r[10]=(e=parseInt(t.slice(24,36),16))/1099511627776&255,r[11]=e/4294967296&255,r[12]=e>>>24&255,r[13]=e>>>16&255,r[14]=e>>>8&255,r[15]=255&e,r};function stringToBytes(t){t=unescape(encodeURIComponent(t));let e=[];for(let r=0;r<t.length;++r)e.push(t.charCodeAt(r));return e}function v35(t,e,r){function generateUUID(t,a,i,n){var s;if("string"==typeof t&&(t=stringToBytes(t)),"string"==typeof a&&(a=esm_browser_parse(a)),(null===(s=a)||void 0===s?void 0:s.length)!==16)throw TypeError("Namespace must be array-like (16 iterable integer values, 0-255)");let l=new Uint8Array(16+t.length);if(l.set(a),l.set(t,a.length),(l=r(l))[6]=15&l[6]|e,l[8]=63&l[8]|128,i){n=n||0;for(let t=0;t<16;++t)i[n+t]=l[t];return i}return unsafeStringify(l)}try{generateUUID.name=t}catch(t){}return generateUUID.DNS="6ba7b810-9dad-11d1-80b4-00c04fd430c8",generateUUID.URL="6ba7b811-9dad-11d1-80b4-00c04fd430c8",generateUUID}function f(t,e,r,a){switch(t){case 0:return e&r^~e&a;case 1:case 3:return e^r^a;case 2:return e&r^e&a^r&a}}function ROTL(t,e){return t<<e|t>>>32-e}let h=v35("v5",80,function(t){let e=[1518500249,1859775393,2400959708,3395469782],r=[1732584193,4023233417,2562383102,271733878,3285377520];if("string"==typeof t){let e=unescape(encodeURIComponent(t));t=[];for(let r=0;r<e.length;++r)t.push(e.charCodeAt(r))}else Array.isArray(t)||(t=Array.prototype.slice.call(t));t.push(128);let a=t.length/4+2,i=Math.ceil(a/16),n=Array(i);for(let e=0;e<i;++e){let r=new Uint32Array(16);for(let a=0;a<16;++a)r[a]=t[64*e+4*a]<<24|t[64*e+4*a+1]<<16|t[64*e+4*a+2]<<8|t[64*e+4*a+3];n[e]=r}n[i-1][14]=(t.length-1)*8/4294967296,n[i-1][14]=Math.floor(n[i-1][14]),n[i-1][15]=(t.length-1)*8&4294967295;for(let t=0;t<i;++t){let a=new Uint32Array(80);for(let e=0;e<16;++e)a[e]=n[t][e];for(let t=16;t<80;++t)a[t]=ROTL(a[t-3]^a[t-8]^a[t-14]^a[t-16],1);let i=r[0],s=r[1],l=r[2],c=r[3],h=r[4];for(let t=0;t<80;++t){let r=Math.floor(t/20),n=ROTL(i,5)+f(r,s,l,c)+h+e[r]+a[t]>>>0;h=c,c=l,l=ROTL(s,30)>>>0,s=i,i=n}r[0]=r[0]+i>>>0,r[1]=r[1]+s>>>0,r[2]=r[2]+l>>>0,r[3]=r[3]+c>>>0,r[4]=r[4]+h>>>0}return[r[0]>>24&255,r[0]>>16&255,r[0]>>8&255,255&r[0],r[1]>>24&255,r[1]>>16&255,r[1]>>8&255,255&r[1],r[2]>>24&255,r[2]>>16&255,r[2]>>8&255,255&r[2],r[3]>>24&255,r[3]>>16&255,r[3]>>8&255,255&r[3],r[4]>>24&255,r[4]>>16&255,r[4]>>8&255,255&r[4]]});r(7693),r(7608),r(1699);var d=function(){var o=function(t,e,r,a){for(r=r||{},a=t.length;a--;r[t[a]]=e);return r},t=[6,8,10,20,22,24,26,27,28],e=[1,10],r=[1,11],a=[1,12],i=[1,13],n=[1,14],s=[1,15],l=[1,21],c=[1,22],h=[1,23],d=[1,24],y=[1,25],u=[6,8,10,13,15,18,19,20,22,24,26,27,28,41,42,43,44,45],p=[1,34],_=[27,28,46,47],E=[41,42,43,44,45],g=[17,34],m=[1,54],O=[1,53],b=[17,34,36,38],k={trace:function(){},yy:{},symbols_:{error:2,start:3,ER_DIAGRAM:4,document:5,EOF:6,line:7,SPACE:8,statement:9,NEWLINE:10,entityName:11,relSpec:12,":":13,role:14,BLOCK_START:15,attributes:16,BLOCK_STOP:17,SQS:18,SQE:19,title:20,title_value:21,acc_title:22,acc_title_value:23,acc_descr:24,acc_descr_value:25,acc_descr_multiline_value:26,ALPHANUM:27,ENTITY_NAME:28,attribute:29,attributeType:30,attributeName:31,attributeKeyTypeList:32,attributeComment:33,ATTRIBUTE_WORD:34,attributeKeyType:35,COMMA:36,ATTRIBUTE_KEY:37,COMMENT:38,cardinality:39,relType:40,ZERO_OR_ONE:41,ZERO_OR_MORE:42,ONE_OR_MORE:43,ONLY_ONE:44,MD_PARENT:45,NON_IDENTIFYING:46,IDENTIFYING:47,WORD:48,$accept:0,$end:1},terminals_:{2:"error",4:"ER_DIAGRAM",6:"EOF",8:"SPACE",10:"NEWLINE",13:":",15:"BLOCK_START",17:"BLOCK_STOP",18:"SQS",19:"SQE",20:"title",21:"title_value",22:"acc_title",23:"acc_title_value",24:"acc_descr",25:"acc_descr_value",26:"acc_descr_multiline_value",27:"ALPHANUM",28:"ENTITY_NAME",34:"ATTRIBUTE_WORD",36:"COMMA",37:"ATTRIBUTE_KEY",38:"COMMENT",41:"ZERO_OR_ONE",42:"ZERO_OR_MORE",43:"ONE_OR_MORE",44:"ONLY_ONE",45:"MD_PARENT",46:"NON_IDENTIFYING",47:"IDENTIFYING",48:"WORD"},productions_:[0,[3,3],[5,0],[5,2],[7,2],[7,1],[7,1],[7,1],[9,5],[9,4],[9,3],[9,1],[9,7],[9,6],[9,4],[9,2],[9,2],[9,2],[9,1],[11,1],[11,1],[16,1],[16,2],[29,2],[29,3],[29,3],[29,4],[30,1],[31,1],[32,1],[32,3],[35,1],[33,1],[12,3],[39,1],[39,1],[39,1],[39,1],[39,1],[40,1],[40,1],[14,1],[14,1],[14,1]],performAction:function(t,e,r,a,i,n,s){var l=n.length-1;switch(i){case 1:break;case 2:case 6:case 7:this.$=[];break;case 3:n[l-1].push(n[l]),this.$=n[l-1];break;case 4:case 5:case 19:case 43:case 27:case 28:case 31:this.$=n[l];break;case 8:a.addEntity(n[l-4]),a.addEntity(n[l-2]),a.addRelationship(n[l-4],n[l],n[l-2],n[l-3]);break;case 9:a.addEntity(n[l-3]),a.addAttributes(n[l-3],n[l-1]);break;case 10:a.addEntity(n[l-2]);break;case 11:a.addEntity(n[l]);break;case 12:a.addEntity(n[l-6],n[l-4]),a.addAttributes(n[l-6],n[l-1]);break;case 13:a.addEntity(n[l-5],n[l-3]);break;case 14:a.addEntity(n[l-3],n[l-1]);break;case 15:case 16:this.$=n[l].trim(),a.setAccTitle(this.$);break;case 17:case 18:this.$=n[l].trim(),a.setAccDescription(this.$);break;case 20:case 41:case 42:case 32:this.$=n[l].replace(/"/g,"");break;case 21:case 29:this.$=[n[l]];break;case 22:n[l].push(n[l-1]),this.$=n[l];break;case 23:this.$={attributeType:n[l-1],attributeName:n[l]};break;case 24:this.$={attributeType:n[l-2],attributeName:n[l-1],attributeKeyTypeList:n[l]};break;case 25:this.$={attributeType:n[l-2],attributeName:n[l-1],attributeComment:n[l]};break;case 26:this.$={attributeType:n[l-3],attributeName:n[l-2],attributeKeyTypeList:n[l-1],attributeComment:n[l]};break;case 30:n[l-2].push(n[l]),this.$=n[l-2];break;case 33:this.$={cardA:n[l],relType:n[l-1],cardB:n[l-2]};break;case 34:this.$=a.Cardinality.ZERO_OR_ONE;break;case 35:this.$=a.Cardinality.ZERO_OR_MORE;break;case 36:this.$=a.Cardinality.ONE_OR_MORE;break;case 37:this.$=a.Cardinality.ONLY_ONE;break;case 38:this.$=a.Cardinality.MD_PARENT;break;case 39:this.$=a.Identification.NON_IDENTIFYING;break;case 40:this.$=a.Identification.IDENTIFYING}},table:[{3:1,4:[1,2]},{1:[3]},o(t,[2,2],{5:3}),{6:[1,4],7:5,8:[1,6],9:7,10:[1,8],11:9,20:e,22:r,24:a,26:i,27:n,28:s},o(t,[2,7],{1:[2,1]}),o(t,[2,3]),{9:16,11:9,20:e,22:r,24:a,26:i,27:n,28:s},o(t,[2,5]),o(t,[2,6]),o(t,[2,11],{12:17,39:20,15:[1,18],18:[1,19],41:l,42:c,43:h,44:d,45:y}),{21:[1,26]},{23:[1,27]},{25:[1,28]},o(t,[2,18]),o(u,[2,19]),o(u,[2,20]),o(t,[2,4]),{11:29,27:n,28:s},{16:30,17:[1,31],29:32,30:33,34:p},{11:35,27:n,28:s},{40:36,46:[1,37],47:[1,38]},o(_,[2,34]),o(_,[2,35]),o(_,[2,36]),o(_,[2,37]),o(_,[2,38]),o(t,[2,15]),o(t,[2,16]),o(t,[2,17]),{13:[1,39]},{17:[1,40]},o(t,[2,10]),{16:41,17:[2,21],29:32,30:33,34:p},{31:42,34:[1,43]},{34:[2,27]},{19:[1,44]},{39:45,41:l,42:c,43:h,44:d,45:y},o(E,[2,39]),o(E,[2,40]),{14:46,27:[1,49],28:[1,48],48:[1,47]},o(t,[2,9]),{17:[2,22]},o(g,[2,23],{32:50,33:51,35:52,37:m,38:O}),o([17,34,37,38],[2,28]),o(t,[2,14],{15:[1,55]}),o([27,28],[2,33]),o(t,[2,8]),o(t,[2,41]),o(t,[2,42]),o(t,[2,43]),o(g,[2,24],{33:56,36:[1,57],38:O}),o(g,[2,25]),o(b,[2,29]),o(g,[2,32]),o(b,[2,31]),{16:58,17:[1,59],29:32,30:33,34:p},o(g,[2,26]),{35:60,37:m},{17:[1,61]},o(t,[2,13]),o(b,[2,30]),o(t,[2,12])],defaultActions:{34:[2,27],41:[2,22]},parseError:function(t,e){if(e.recoverable)this.trace(t);else{var r=Error(t);throw r.hash=e,r}},parse:function(t){var e=this,r=[0],a=[],i=[null],n=[],s=this.table,l="",c=0,h=0,d=n.slice.call(arguments,1),y=Object.create(this.lexer),u={yy:{}};for(var p in this.yy)Object.prototype.hasOwnProperty.call(this.yy,p)&&(u.yy[p]=this.yy[p]);y.setInput(t,u.yy),u.yy.lexer=y,u.yy.parser=this,void 0===y.yylloc&&(y.yylloc={});var _=y.yylloc;n.push(_);var E=y.options&&y.options.ranges;function lex(){var t;return"number"!=typeof(t=a.pop()||y.lex()||1)&&(t instanceof Array&&(t=(a=t).pop()),t=e.symbols_[t]||t),t}"function"==typeof u.yy.parseError?this.parseError=u.yy.parseError:this.parseError=Object.getPrototypeOf(this).parseError;for(var g,m,O,b,k,R,N,T,x={};;){if(m=r[r.length-1],this.defaultActions[m]?O=this.defaultActions[m]:(null==g&&(g=lex()),O=s[m]&&s[m][g]),void 0===O||!O.length||!O[0]){var A="";for(k in T=[],s[m])this.terminals_[k]&&k>2&&T.push("'"+this.terminals_[k]+"'");A=y.showPosition?"Parse error on line "+(c+1)+":\n"+y.showPosition()+"\nExpecting "+T.join(", ")+", got '"+(this.terminals_[g]||g)+"'":"Parse error on line "+(c+1)+": Unexpected "+(1==g?"end of input":"'"+(this.terminals_[g]||g)+"'"),this.parseError(A,{text:y.match,token:this.terminals_[g]||g,line:y.yylineno,loc:_,expected:T})}if(O[0]instanceof Array&&O.length>1)throw Error("Parse Error: multiple actions possible at state: "+m+", token: "+g);switch(O[0]){case 1:r.push(g),i.push(y.yytext),n.push(y.yylloc),r.push(O[1]),g=null,h=y.yyleng,l=y.yytext,c=y.yylineno,_=y.yylloc;break;case 2:if(R=this.productions_[O[1]][1],x.$=i[i.length-R],x._$={first_line:n[n.length-(R||1)].first_line,last_line:n[n.length-1].last_line,first_column:n[n.length-(R||1)].first_column,last_column:n[n.length-1].last_column},E&&(x._$.range=[n[n.length-(R||1)].range[0],n[n.length-1].range[1]]),void 0!==(b=this.performAction.apply(x,[l,h,c,u.yy,O[1],i,n].concat(d))))return b;R&&(r=r.slice(0,-1*R*2),i=i.slice(0,-1*R),n=n.slice(0,-1*R)),r.push(this.productions_[O[1]][0]),i.push(x.$),n.push(x._$),N=s[r[r.length-2]][r[r.length-1]],r.push(N);break;case 3:return!0}}return!0}};function Parser(){this.yy={}}return k.lexer={EOF:1,parseError:function(t,e){if(this.yy.parser)this.yy.parser.parseError(t,e);else throw Error(t)},setInput:function(t,e){return this.yy=e||this.yy||{},this._input=t,this._more=this._backtrack=this.done=!1,this.yylineno=this.yyleng=0,this.yytext=this.matched=this.match="",this.conditionStack=["INITIAL"],this.yylloc={first_line:1,first_column:0,last_line:1,last_column:0},this.options.ranges&&(this.yylloc.range=[0,0]),this.offset=0,this},input:function(){var t=this._input[0];return this.yytext+=t,this.yyleng++,this.offset++,this.match+=t,this.matched+=t,t.match(/(?:\r\n?|\n).*/g)?(this.yylineno++,this.yylloc.last_line++):this.yylloc.last_column++,this.options.ranges&&this.yylloc.range[1]++,this._input=this._input.slice(1),t},unput:function(t){var e=t.length,r=t.split(/(?:\r\n?|\n)/g);this._input=t+this._input,this.yytext=this.yytext.substr(0,this.yytext.length-e),this.offset-=e;var a=this.match.split(/(?:\r\n?|\n)/g);this.match=this.match.substr(0,this.match.length-1),this.matched=this.matched.substr(0,this.matched.length-1),r.length-1&&(this.yylineno-=r.length-1);var i=this.yylloc.range;return this.yylloc={first_line:this.yylloc.first_line,last_line:this.yylineno+1,first_column:this.yylloc.first_column,last_column:r?(r.length===a.length?this.yylloc.first_column:0)+a[a.length-r.length].length-r[0].length:this.yylloc.first_column-e},this.options.ranges&&(this.yylloc.range=[i[0],i[0]+this.yyleng-e]),this.yyleng=this.yytext.length,this},more:function(){return this._more=!0,this},reject:function(){return this.options.backtrack_lexer?(this._backtrack=!0,this):this.parseError("Lexical error on line "+(this.yylineno+1)+". You can only invoke reject() in the lexer when the lexer is of the backtracking persuasion (options.backtrack_lexer = true).\n"+this.showPosition(),{text:"",token:null,line:this.yylineno})},less:function(t){this.unput(this.match.slice(t))},pastInput:function(){var t=this.matched.substr(0,this.matched.length-this.match.length);return(t.length>20?"...":"")+t.substr(-20).replace(/\n/g,"")},upcomingInput:function(){var t=this.match;return t.length<20&&(t+=this._input.substr(0,20-t.length)),(t.substr(0,20)+(t.length>20?"...":"")).replace(/\n/g,"")},showPosition:function(){var t=this.pastInput(),e=Array(t.length+1).join("-");return t+this.upcomingInput()+"\n"+e+"^"},test_match:function(t,e){var r,a,i;if(this.options.backtrack_lexer&&(i={yylineno:this.yylineno,yylloc:{first_line:this.yylloc.first_line,last_line:this.last_line,first_column:this.yylloc.first_column,last_column:this.yylloc.last_column},yytext:this.yytext,match:this.match,matches:this.matches,matched:this.matched,yyleng:this.yyleng,offset:this.offset,_more:this._more,_input:this._input,yy:this.yy,conditionStack:this.conditionStack.slice(0),done:this.done},this.options.ranges&&(i.yylloc.range=this.yylloc.range.slice(0))),(a=t[0].match(/(?:\r\n?|\n).*/g))&&(this.yylineno+=a.length),this.yylloc={first_line:this.yylloc.last_line,last_line:this.yylineno+1,first_column:this.yylloc.last_column,last_column:a?a[a.length-1].length-a[a.length-1].match(/\r?\n?/)[0].length:this.yylloc.last_column+t[0].length},this.yytext+=t[0],this.match+=t[0],this.matches=t,this.yyleng=this.yytext.length,this.options.ranges&&(this.yylloc.range=[this.offset,this.offset+=this.yyleng]),this._more=!1,this._backtrack=!1,this._input=this._input.slice(t[0].length),this.matched+=t[0],r=this.performAction.call(this,this.yy,this,e,this.conditionStack[this.conditionStack.length-1]),this.done&&this._input&&(this.done=!1),r)return r;if(this._backtrack)for(var n in i)this[n]=i[n];return!1},next:function(){if(this.done)return this.EOF;this._input||(this.done=!0),this._more||(this.yytext="",this.match="");for(var t,e,r,a,i=this._currentRules(),n=0;n<i.length;n++)if((r=this._input.match(this.rules[i[n]]))&&(!e||r[0].length>e[0].length)){if(e=r,a=n,this.options.backtrack_lexer){if(!1!==(t=this.test_match(r,i[n])))return t;if(!this._backtrack)return!1;e=!1;continue}if(!this.options.flex)break}return e?!1!==(t=this.test_match(e,i[a]))&&t:""===this._input?this.EOF:this.parseError("Lexical error on line "+(this.yylineno+1)+". Unrecognized text.\n"+this.showPosition(),{text:"",token:null,line:this.yylineno})},lex:function(){return this.next()||this.lex()},begin:function(t){this.conditionStack.push(t)},popState:function(){return this.conditionStack.length-1>0?this.conditionStack.pop():this.conditionStack[0]},_currentRules:function(){return this.conditionStack.length&&this.conditionStack[this.conditionStack.length-1]?this.conditions[this.conditionStack[this.conditionStack.length-1]].rules:this.conditions.INITIAL.rules},topState:function(t){return(t=this.conditionStack.length-1-Math.abs(t||0))>=0?this.conditionStack[t]:"INITIAL"},pushState:function(t){this.begin(t)},stateStackSize:function(){return this.conditionStack.length},options:{"case-insensitive":!0},performAction:function(t,e,r,a){switch(r){case 0:return this.begin("acc_title"),22;case 1:return this.popState(),"acc_title_value";case 2:return this.begin("acc_descr"),24;case 3:return this.popState(),"acc_descr_value";case 4:this.begin("acc_descr_multiline");break;case 5:this.popState();break;case 6:return"acc_descr_multiline_value";case 7:return 10;case 8:case 15:case 20:break;case 9:return 8;case 10:return 28;case 11:return 48;case 12:return 4;case 13:return this.begin("block"),15;case 14:return 36;case 16:return 37;case 17:case 18:return 34;case 19:return 38;case 21:return this.popState(),17;case 22:case 54:return e.yytext[0];case 23:return 18;case 24:return 19;case 25:case 29:case 30:case 43:return 41;case 26:case 27:case 28:case 36:case 38:case 45:return 43;case 31:case 32:case 33:case 34:case 35:case 37:case 44:return 42;case 39:case 40:case 41:case 42:return 44;case 46:return 45;case 47:case 50:case 51:case 52:return 46;case 48:case 49:return 47;case 53:return 27;case 55:return 6}},rules:[/^(?:accTitle\s*:\s*)/i,/^(?:(?!\n||)*[^\n]*)/i,/^(?:accDescr\s*:\s*)/i,/^(?:(?!\n||)*[^\n]*)/i,/^(?:accDescr\s*\{\s*)/i,/^(?:[\}])/i,/^(?:[^\}]*)/i,/^(?:[\n]+)/i,/^(?:\s+)/i,/^(?:[\s]+)/i,/^(?:"[^"%\r\n\v\b\\]+")/i,/^(?:"[^"]*")/i,/^(?:erDiagram\b)/i,/^(?:\{)/i,/^(?:,)/i,/^(?:\s+)/i,/^(?:\b((?:PK)|(?:FK)|(?:UK))\b)/i,/^(?:(.*?)[~](.*?)*[~])/i,/^(?:[\*A-Za-z_][A-Za-z0-9\-_\[\]\(\)]*)/i,/^(?:"[^"]*")/i,/^(?:[\n]+)/i,/^(?:\})/i,/^(?:.)/i,/^(?:\[)/i,/^(?:\])/i,/^(?:one or zero\b)/i,/^(?:one or more\b)/i,/^(?:one or many\b)/i,/^(?:1\+)/i,/^(?:\|o\b)/i,/^(?:zero or one\b)/i,/^(?:zero or more\b)/i,/^(?:zero or many\b)/i,/^(?:0\+)/i,/^(?:\}o\b)/i,/^(?:many\(0\))/i,/^(?:many\(1\))/i,/^(?:many\b)/i,/^(?:\}\|)/i,/^(?:one\b)/i,/^(?:only one\b)/i,/^(?:1\b)/i,/^(?:\|\|)/i,/^(?:o\|)/i,/^(?:o\{)/i,/^(?:\|\{)/i,/^(?:\s*u\b)/i,/^(?:\.\.)/i,/^(?:--)/i,/^(?:to\b)/i,/^(?:optionally to\b)/i,/^(?:\.-)/i,/^(?:-\.)/i,/^(?:[A-Za-z_][A-Za-z0-9\-_]*)/i,/^(?:.)/i,/^(?:$)/i],conditions:{acc_descr_multiline:{rules:[5,6],inclusive:!1},acc_descr:{rules:[3],inclusive:!1},acc_title:{rules:[1],inclusive:!1},block:{rules:[14,15,16,17,18,19,20,21,22],inclusive:!1},INITIAL:{rules:[0,2,4,7,8,9,10,11,12,13,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55],inclusive:!0}}},Parser.prototype=k,k.Parser=Parser,new Parser}();d.parser=d;let y={},u=[],addEntity=function(t,e){return void 0===y[t]?(y[t]={attributes:[],alias:e},a.l.info("Added new entity :",t)):y[t]&&!y[t].alias&&e&&(y[t].alias=e,a.l.info(`Add alias '${e}' to entity '${t}'`)),y[t]},p={Cardinality:{ZERO_OR_ONE:"ZERO_OR_ONE",ZERO_OR_MORE:"ZERO_OR_MORE",ONE_OR_MORE:"ONE_OR_MORE",ONLY_ONE:"ONLY_ONE",MD_PARENT:"MD_PARENT"},Identification:{NON_IDENTIFYING:"NON_IDENTIFYING",IDENTIFYING:"IDENTIFYING"},getConfig:()=>(0,a.c)().er,addEntity,addAttributes:function(t,e){let r,i=addEntity(t);for(r=e.length-1;r>=0;r--)i.attributes.push(e[r]),a.l.debug("Added attribute ",e[r].attributeName)},getEntities:()=>y,addRelationship:function(t,e,r,i){let n={entityA:t,roleA:e,entityB:r,relSpec:i};u.push(n),a.l.debug("Added new relationship :",n)},getRelationships:()=>u,clear:function(){y={},u=[],(0,a.t)()},setAccTitle:a.s,getAccTitle:a.g,setAccDescription:a.b,getAccDescription:a.a,setDiagramTitle:a.q,getDiagramTitle:a.r},_={ONLY_ONE_START:"ONLY_ONE_START",ONLY_ONE_END:"ONLY_ONE_END",ZERO_OR_ONE_START:"ZERO_OR_ONE_START",ZERO_OR_ONE_END:"ZERO_OR_ONE_END",ONE_OR_MORE_START:"ONE_OR_MORE_START",ONE_OR_MORE_END:"ONE_OR_MORE_END",ZERO_OR_MORE_START:"ZERO_OR_MORE_START",ZERO_OR_MORE_END:"ZERO_OR_MORE_END",MD_PARENT_END:"MD_PARENT_END",MD_PARENT_START:"MD_PARENT_START"},E={ERMarkers:_,insertMarkers:function(t,e){let r;t.append("defs").append("marker").attr("id",_.MD_PARENT_START).attr("refX",0).attr("refY",7).attr("markerWidth",190).attr("markerHeight",240).attr("orient","auto").append("path").attr("d","M 18,7 L9,13 L1,7 L9,1 Z"),t.append("defs").append("marker").attr("id",_.MD_PARENT_END).attr("refX",19).attr("refY",7).attr("markerWidth",20).attr("markerHeight",28).attr("orient","auto").append("path").attr("d","M 18,7 L9,13 L1,7 L9,1 Z"),t.append("defs").append("marker").attr("id",_.ONLY_ONE_START).attr("refX",0).attr("refY",9).attr("markerWidth",18).attr("markerHeight",18).attr("orient","auto").append("path").attr("stroke",e.stroke).attr("fill","none").attr("d","M9,0 L9,18 M15,0 L15,18"),t.append("defs").append("marker").attr("id",_.ONLY_ONE_END).attr("refX",18).attr("refY",9).attr("markerWidth",18).attr("markerHeight",18).attr("orient","auto").append("path").attr("stroke",e.stroke).attr("fill","none").attr("d","M3,0 L3,18 M9,0 L9,18"),(r=t.append("defs").append("marker").attr("id",_.ZERO_OR_ONE_START).attr("refX",0).attr("refY",9).attr("markerWidth",30).attr("markerHeight",18).attr("orient","auto")).append("circle").attr("stroke",e.stroke).attr("fill","white").attr("cx",21).attr("cy",9).attr("r",6),r.append("path").attr("stroke",e.stroke).attr("fill","none").attr("d","M9,0 L9,18"),(r=t.append("defs").append("marker").attr("id",_.ZERO_OR_ONE_END).attr("refX",30).attr("refY",9).attr("markerWidth",30).attr("markerHeight",18).attr("orient","auto")).append("circle").attr("stroke",e.stroke).attr("fill","white").attr("cx",9).attr("cy",9).attr("r",6),r.append("path").attr("stroke",e.stroke).attr("fill","none").attr("d","M21,0 L21,18"),t.append("defs").append("marker").attr("id",_.ONE_OR_MORE_START).attr("refX",18).attr("refY",18).attr("markerWidth",45).attr("markerHeight",36).attr("orient","auto").append("path").attr("stroke",e.stroke).attr("fill","none").attr("d","M0,18 Q 18,0 36,18 Q 18,36 0,18 M42,9 L42,27"),t.append("defs").append("marker").attr("id",_.ONE_OR_MORE_END).attr("refX",27).attr("refY",18).attr("markerWidth",45).attr("markerHeight",36).attr("orient","auto").append("path").attr("stroke",e.stroke).attr("fill","none").attr("d","M3,9 L3,27 M9,18 Q27,0 45,18 Q27,36 9,18"),(r=t.append("defs").append("marker").attr("id",_.ZERO_OR_MORE_START).attr("refX",18).attr("refY",18).attr("markerWidth",57).attr("markerHeight",36).attr("orient","auto")).append("circle").attr("stroke",e.stroke).attr("fill","white").attr("cx",48).attr("cy",18).attr("r",6),r.append("path").attr("stroke",e.stroke).attr("fill","none").attr("d","M0,18 Q18,0 36,18 Q18,36 0,18"),(r=t.append("defs").append("marker").attr("id",_.ZERO_OR_MORE_END).attr("refX",39).attr("refY",18).attr("markerWidth",57).attr("markerHeight",36).attr("orient","auto")).append("circle").attr("stroke",e.stroke).attr("fill","white").attr("cx",9).attr("cy",18).attr("r",6),r.append("path").attr("stroke",e.stroke).attr("fill","none").attr("d","M21,18 Q39,0 57,18 Q39,36 21,18")}},g=/[^\dA-Za-z](\W)*/g,m={},O=new Map,drawAttributes=(t,e,r)=>{let i=m.entityPadding/3,n=m.entityPadding/3,s=.85*m.fontSize,l=e.node().getBBox(),c=[],h=!1,d=!1,y=0,u=0,p=0,_=0,E=l.height+2*i,g=1;r.forEach(t=>{void 0!==t.attributeKeyTypeList&&t.attributeKeyTypeList.length>0&&(h=!0),void 0!==t.attributeComment&&(d=!0)}),r.forEach(r=>{let n=`${e.node().id}-attr-${g}`,l=0,m=(0,a.v)(r.attributeType),O=t.append("text").classed("er entityLabel",!0).attr("id",`${n}-type`).attr("x",0).attr("y",0).style("dominant-baseline","middle").style("text-anchor","left").style("font-family",(0,a.c)().fontFamily).style("font-size",s+"px").text(m),b=t.append("text").classed("er entityLabel",!0).attr("id",`${n}-name`).attr("x",0).attr("y",0).style("dominant-baseline","middle").style("text-anchor","left").style("font-family",(0,a.c)().fontFamily).style("font-size",s+"px").text(r.attributeName),k={};k.tn=O,k.nn=b;let R=O.node().getBBox(),N=b.node().getBBox();if(y=Math.max(y,R.width),u=Math.max(u,N.width),l=Math.max(R.height,N.height),h){let e=void 0!==r.attributeKeyTypeList?r.attributeKeyTypeList.join(","):"",i=t.append("text").classed("er entityLabel",!0).attr("id",`${n}-key`).attr("x",0).attr("y",0).style("dominant-baseline","middle").style("text-anchor","left").style("font-family",(0,a.c)().fontFamily).style("font-size",s+"px").text(e);k.kn=i;let c=i.node().getBBox();p=Math.max(p,c.width),l=Math.max(l,c.height)}if(d){let e=t.append("text").classed("er entityLabel",!0).attr("id",`${n}-comment`).attr("x",0).attr("y",0).style("dominant-baseline","middle").style("text-anchor","left").style("font-family",(0,a.c)().fontFamily).style("font-size",s+"px").text(r.attributeComment||"");k.cn=e;let i=e.node().getBBox();_=Math.max(_,i.width),l=Math.max(l,i.height)}k.height=l,c.push(k),E+=l+2*i,g+=1});let O=4;h&&(O+=2),d&&(O+=2);let b=y+u+p+_,k={width:Math.max(m.minEntityWidth,Math.max(l.width+2*m.entityPadding,b+n*O)),height:r.length>0?E:Math.max(m.minEntityHeight,l.height+2*m.entityPadding)};if(r.length>0){let r=Math.max(0,(k.width-b-n*O)/(O/2));e.attr("transform","translate("+k.width/2+","+(i+l.height/2)+")");let a=l.height+2*i,s="attributeBoxOdd";c.forEach(e=>{let l=a+i+e.height/2;e.tn.attr("transform","translate("+n+","+l+")");let c=t.insert("rect","#"+e.tn.node().id).classed(`er ${s}`,!0).attr("x",0).attr("y",a).attr("width",y+2*n+r).attr("height",e.height+2*i),E=parseFloat(c.attr("x"))+parseFloat(c.attr("width"));e.nn.attr("transform","translate("+(E+n)+","+l+")");let g=t.insert("rect","#"+e.nn.node().id).classed(`er ${s}`,!0).attr("x",E).attr("y",a).attr("width",u+2*n+r).attr("height",e.height+2*i),m=parseFloat(g.attr("x"))+parseFloat(g.attr("width"));if(h){e.kn.attr("transform","translate("+(m+n)+","+l+")");let c=t.insert("rect","#"+e.kn.node().id).classed(`er ${s}`,!0).attr("x",m).attr("y",a).attr("width",p+2*n+r).attr("height",e.height+2*i);m=parseFloat(c.attr("x"))+parseFloat(c.attr("width"))}d&&(e.cn.attr("transform","translate("+(m+n)+","+l+")"),t.insert("rect","#"+e.cn.node().id).classed(`er ${s}`,"true").attr("x",m).attr("y",a).attr("width",_+2*n+r).attr("height",e.height+2*i)),a+=e.height+2*i,s="attributeBoxOdd"===s?"attributeBoxEven":"attributeBoxOdd"})}else k.height=Math.max(m.minEntityHeight,E),e.attr("transform","translate("+k.width/2+","+k.height/2+")");return k},drawEntities=function(t,e,r){let i;let n=Object.keys(e);return n.forEach(function(n){let s=generateId(n,"entity");O.set(n,s);let l=t.append("g").attr("id",s);i=void 0===i?s:i;let c="text-"+s,h=l.append("text").classed("er entityLabel",!0).attr("id",c).attr("x",0).attr("y",0).style("dominant-baseline","middle").style("text-anchor","middle").style("font-family",(0,a.c)().fontFamily).style("font-size",m.fontSize+"px").text(e[n].alias??n),{width:d,height:y}=drawAttributes(l,h,e[n].attributes),u=l.insert("rect","#"+c).classed("er entityBox",!0).attr("x",0).attr("y",0).attr("width",d).attr("height",y),p=u.node().getBBox();r.setNode(s,{width:p.width,height:p.height,shape:"rect",id:s})}),i},adjustEntities=function(t,e){e.nodes().forEach(function(r){void 0!==r&&void 0!==e.node(r)&&t.select("#"+r).attr("transform","translate("+(e.node(r).x-e.node(r).width/2)+","+(e.node(r).y-e.node(r).height/2)+" )")})},getEdgeName=function(t){return(t.entityA+t.roleA+t.entityB).replace(/\s/g,"")},b=0,drawRelationshipFromLayout=function(t,e,r,i,s){b++;let l=r.edge(O.get(e.entityA),O.get(e.entityB),getEdgeName(e)),c=(0,n.jvg)().x(function(t){return t.x}).y(function(t){return t.y}).curve(n.$0Z),h=t.insert("path","#"+i).classed("er relationshipLine",!0).attr("d",c(l.points)).style("stroke",m.stroke).style("fill","none");e.relSpec.relType===s.db.Identification.NON_IDENTIFYING&&h.attr("stroke-dasharray","8,8");let d="";switch(m.arrowMarkerAbsolute&&(d=(d=(d=window.location.protocol+"//"+window.location.host+window.location.pathname+window.location.search).replace(/\(/g,"\\(")).replace(/\)/g,"\\)")),e.relSpec.cardA){case s.db.Cardinality.ZERO_OR_ONE:h.attr("marker-end","url("+d+"#"+E.ERMarkers.ZERO_OR_ONE_END+")");break;case s.db.Cardinality.ZERO_OR_MORE:h.attr("marker-end","url("+d+"#"+E.ERMarkers.ZERO_OR_MORE_END+")");break;case s.db.Cardinality.ONE_OR_MORE:h.attr("marker-end","url("+d+"#"+E.ERMarkers.ONE_OR_MORE_END+")");break;case s.db.Cardinality.ONLY_ONE:h.attr("marker-end","url("+d+"#"+E.ERMarkers.ONLY_ONE_END+")");break;case s.db.Cardinality.MD_PARENT:h.attr("marker-end","url("+d+"#"+E.ERMarkers.MD_PARENT_END+")")}switch(e.relSpec.cardB){case s.db.Cardinality.ZERO_OR_ONE:h.attr("marker-start","url("+d+"#"+E.ERMarkers.ZERO_OR_ONE_START+")");break;case s.db.Cardinality.ZERO_OR_MORE:h.attr("marker-start","url("+d+"#"+E.ERMarkers.ZERO_OR_MORE_START+")");break;case s.db.Cardinality.ONE_OR_MORE:h.attr("marker-start","url("+d+"#"+E.ERMarkers.ONE_OR_MORE_START+")");break;case s.db.Cardinality.ONLY_ONE:h.attr("marker-start","url("+d+"#"+E.ERMarkers.ONLY_ONE_START+")");break;case s.db.Cardinality.MD_PARENT:h.attr("marker-start","url("+d+"#"+E.ERMarkers.MD_PARENT_START+")")}let y=h.node().getTotalLength(),u=h.node().getPointAtLength(.5*y),p="rel"+b,_=t.append("text").classed("er relationshipLabel",!0).attr("id",p).attr("x",u.x).attr("y",u.y).style("text-anchor","middle").style("dominant-baseline","middle").style("font-family",(0,a.c)().fontFamily).style("font-size",m.fontSize+"px").text(e.roleA),g=_.node().getBBox();t.insert("rect","#"+p).classed("er relationshipLabelBox",!0).attr("x",u.x-g.width/2).attr("y",u.y-g.height/2).attr("width",g.width).attr("height",g.height)};function generateId(t="",e=""){let r=t.replace(g,"");return`${strWithHyphen(e)}${strWithHyphen(r)}${h(t,"28e9f9db-3c8d-5aa5-9faf-44286ae5937c")}`}function strWithHyphen(t=""){return t.length>0?`${t}-`:""}let k={parser:d,db:p,renderer:{setConf:function(t){let e=Object.keys(t);for(let r of e)m[r]=t[r]},draw:function(t,e,r,l){var c;let h,d;m=(0,a.c)().er,a.l.info("Drawing ER diagram");let y=(0,a.c)().securityLevel;"sandbox"===y&&(h=(0,n.Ys)("#i"+e));let u="sandbox"===y?(0,n.Ys)(h.nodes()[0].contentDocument.body):(0,n.Ys)("body"),p=u.select(`[id='${e}']`);E.insertMarkers(p,m),d=new i.k({multigraph:!0,directed:!0,compound:!1}).setGraph({rankdir:m.layoutDirection,marginx:20,marginy:20,nodesep:100,edgesep:100,ranksep:100}).setDefaultEdgeLabel(function(){return{}});let _=drawEntities(p,l.db.getEntities(),d),g=((c=l.db.getRelationships()).forEach(function(t){d.setEdge(O.get(t.entityA),O.get(t.entityB),{relationship:t},getEdgeName(t))}),c);(0,s.bK)(d),adjustEntities(p,d),g.forEach(function(t){drawRelationshipFromLayout(p,t,d,_,l)});let b=m.diagramPadding;a.u.insertTitle(p,"entityTitleText",m.titleTopMargin,l.db.getDiagramTitle());let k=p.node().getBBox(),R=k.width+2*b,N=k.height+2*b;(0,a.i)(p,N,R,m.useMaxWidth),p.attr("viewBox",`${k.x-b} ${k.y-b} ${R} ${N}`)}},styles:t=>`
  .entityBox {
    fill: ${t.mainBkg};
    stroke: ${t.nodeBorder};
  }

  .attributeBoxOdd {
    fill: ${t.attributeBackgroundColorOdd};
    stroke: ${t.nodeBorder};
  }

  .attributeBoxEven {
    fill:  ${t.attributeBackgroundColorEven};
    stroke: ${t.nodeBorder};
  }

  .relationshipLabelBox {
    fill: ${t.tertiaryColor};
    opacity: 0.7;
    background-color: ${t.tertiaryColor};
      rect {
        opacity: 0.5;
      }
  }

    .relationshipLine {
      stroke: ${t.lineColor};
    }

  .entityTitleText {
    text-anchor: middle;
    font-size: 18px;
    fill: ${t.textColor};
  }    
  #MD_PARENT_START {
    fill: #f5f5f5 !important;
    stroke: ${t.lineColor} !important;
    stroke-width: 1;
  }
  #MD_PARENT_END {
    fill: #f5f5f5 !important;
    stroke: ${t.lineColor} !important;
    stroke-width: 1;
  }
  
`}}}]);