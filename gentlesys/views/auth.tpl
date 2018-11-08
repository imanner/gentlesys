<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8"> 
   <meta name="viewport" content="width=device-width,initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no"/>
   <title>Gentlesys {{.Title}}</title>
   <link rel="stylesheet" href="http://apps.bdimg.com/libs/bootstrap/3.2.0/css/bootstrap.min.css">
   <script src="http://apps.bdimg.com/libs/jquery/2.1.4/jquery.min.js"></script>
   <script src="/static/bootsp/js/jquery.md5.js"></script>
   <script src="http://apps.bdimg.com/libs/bootstrap/3.2.0/js/bootstrap.min.js"></script>
    <style>
    .text {
   	color:#009900;
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
        		<label class="col-md-2 control-label text">账号</label>
        		<div class="col-md-4">
        			<input type="text" class="form-control" id="name" 
        				   placeholder="请输入ID或名称">
        		</div>
        	</div>
        	<div class="row form-group">
        		<label class="col-md-2 control-label text">密码</label>
        		<div class="col-md-4">
        			<input type="password" class="form-control" id="passwd" 
        				   placeholder="密码">
        		    
        		</div>
        	</div>
        </form>
        <div class="col-md-offset-2">
        <button id="log" type="submit" class="btn navbar-btn btn-default" onclick="login()">登录</button> 
        <p id="info" class="alert">网站提示：您还没有登录，输入账号密码登录！</p>
        <p class="alert"><a href="register">没有账号？点击注册 </a><a href="findpd">忘记密码？点击找回</a></p>
        </div>
</div>
        
         <script>
    		function login() {
    	    	var name = document.getElementById("name").value;
    			var passwd = document.getElementById("passwd").value;

    			if (name.length < 1) {
    				document.getElementById("info").innerHTML=("用户名为空！");
    				return 
    			} else if(name.length > 32) {
    				document.getElementById("info").innerHTML=("用户名长度不能超过32个字符！");
    				return 
    			} 

    			if (passwd.length < 6) {
    				document.getElementById("info").innerHTML=("密码不对！");
    				return 
    			} else if(passwd.length > 32) {
    				document.getElementById("info").innerHTML=("密码长度不能超过32个字符！");
    				return 
    			}

    			var md5Pwd=$.md5(passwd);
    			
    			document.getElementById("info").innerHTML=("正在登录中...");
    			var btn = $("#log"); 
    			btn.attr("disabled", true);

    			$.ajax({
    		          async:true,
    		          cache:false,
    		          timeout:10000,
    		          type:"POST",
    		          url:"/auth",
    		          data:{
    			       name_:name,
    				   passwd_:md5Pwd,
    			      },
    		          error:function(jqXHR, textStatus, errorThrown){
    		            if(textStatus=="timeout"){
    		              document.getElementById("info").innerHTML=("登录超时，请重试...");
    		            }else{
    		              document.getElementById("info").innerHTML=("登录失败!");
    		            }
    					btn.attr("disabled", false);
    		          },
    		          success:function(msg){
    		          	if ("[0]" != msg.substr(0,3)) {
							if ("[4]" != msg.substr(0,3)) {
								btn.attr("disabled", false);
							}
    						document.getElementById("info").innerHTML=(msg);
    		          	} else {
    						document.getElementById("info").innerHTML=(msg);
    						window.location.href = "/";
    		          	}
    		            
    		          }
    		        });
    		}
    </script>    
</div>

</div>
</body>
</html>
