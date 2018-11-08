<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	 <meta name="viewport" content="width=device-width,initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no"/>
   <title>Gentlesys {{.Title}}</title>
   <link rel="stylesheet" href="http://apps.bdimg.com/libs/bootstrap/3.2.0/css/bootstrap.min.css">
   <script src="http://apps.bdimg.com/libs/jquery/2.1.4/jquery.min.js"></script>
   <script src="http://apps.bdimg.com/libs/bootstrap/3.2.0/js/bootstrap.min.js"></script>
    <style>
    .text {
   	color:#009900;
   	}
	.des-center {
	    text-align: center;
	    font-size:24px;
	    color:#4876FF;
	    padding-bottom:5px;
	}
	</style>
</head>
<body style="padding-top:55px;">
{{str2html .Navigation}}
<div class="container-fluid">

<div class="row">

    <div class="col-md-6 col-md-offset-2">
        <form class="form-horizontal" role="form">
        	<div class="row form-group">
        		<label  class="col-md-2 control-label text">账号</label>
        		<div class="col-md-4">
        			<input type="text" class="form-control" id="name" 
        				   placeholder="请输入ID或名称">
        		    
        		</div>
        	</div>
        </form>

        <div class="col-md-offset-2">
            <button id="log" type="submit" class="btn btn-default navbar-btn" onclick="reset()">找回密码</button>
            <p id="info" class="alert alert-info">网站提示：请输入账号名称。</p>
        </div>
    </div>
     <script>
		function reset() {
	    	var name = document.getElementById("name").value;
			
			if (name.length < 1) {
				document.getElementById("info").innerHTML=("用户名为空！");
				return 
			} else if(name.length > 32) {
				document.getElementById("info").innerHTML=("用户名长度不能超过32个字符！");
				return 
			} 

			document.getElementById("info").innerHTML=("正在找回密码中...");
			var btn = $("#log"); 
			btn.attr("disabled", true);

			$.ajax({
		          async:true,
		          cache:false,
		          timeout:10000,
		          type:"POST",
		          url:"/findpd",
		          data:{
			       name_:name,
			      },
		          error:function(jqXHR, textStatus, errorThrown){
		            if(textStatus=="timeout"){
		              document.getElementById("info").innerHTML=("找回超时，请重试...");
		            }else{
		              document.getElementById("info").innerHTML=("找回失败!");
		            }
					btn.attr("disabled", false);
		          },
		          success:function(msg){
		          	if ("[0]" != msg.substr(0,3)) {
						btn.attr("disabled", false);
						document.getElementById("info").innerHTML=(msg);
		          	} else {
						document.getElementById("info").innerHTML=(msg);
		          	}
		            
		          }
		        });
		}
        </script>    
</div>

</div>
</body>
</html>