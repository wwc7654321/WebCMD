function $ID(id)
{
	return document.getElementById(id);
}
var skey;
function init(sesskey)
{
	skey=sesskey;
	document.body.onload=function(){
		$ID("form1").onsubmit=OnSubmit;
	}
}

function OnSubmit()
{
	$ID("cmd").value="";
}