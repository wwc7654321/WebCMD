function $ID(id)
{
	return document.getElementById(id);
}
var skey,sKeyReq;
function init(sesskey)
{
	skey=sesskey;
	document.body.onload=function(){
		try{
			sKeyReq=hex_md5(sKey);
		}catch(e){
			return;
		}
		$ID("sKey").value=sKeyReq;
		$ID("form1").onsubmit=function(){
			setTimeout(function(){$ID("cmd").value="";$ID("cmdtxt").value="";},300);
		};
		$ID("ml").onchange=function(){
			$ID("cmdtxt").style.display="none";
			$ID("cmd").style.display="inline";
		};
		$ID("sl").onchange=function(){
			$ID("cmdtxt").style.display="inline";
			$ID("cmd").style.display="none";
			$ID("cmdtxt").value=$ID("cmd").value;
			$ID("cmdtxt").style.width=$ID("cmd").style.width;
		};
		$ID("cmdtxt").onchange=function(){
			$ID("cmd").value=$ID("cmdtxt").value;
		};
		$ID("sl").checked="checked";
		$ID("sl").onchange();
		Build();
	};
}


function Build() {
	pref = "";
	/*if(document.location.port)
	{
		pref=":"+document.location.port;
	}*/
    ws = new WebSocket('ws://'+document.location.host+pref+'/ws');

    ws.onopen = OnOpen;
    ws.onmessage = OnMessage;
}

function OnOpen(event) {
    ws.send(sKeyReq);
}

function OnMessage(event) {
    $ID("stdout").value += event.data;
};